package models

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"cloud.google.com/go/spanner"
	bot "github.com/MixinNetwork/bot-api-go-client"
	number "github.com/MixinNetwork/go-number"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/api/iterator"
)

const participants_DDL = `
CREATE TABLE participants (
	packet_id         STRING(36) NOT NULL,
	user_id	          STRING(36) NOT NULL,
	amount            STRING(128) NOT NULL,
	created_at        TIMESTAMP NOT NULL,
	paid_at           TIMESTAMP,
) PRIMARY KEY(packet_id, user_id),
INTERLEAVE IN PARENT packets ON DELETE CASCADE;

CREATE INDEX participants_by_created_paid ON participants(created_at, paid_at) STORING(amount);
`

type Participant struct {
	PacketId  string
	UserId    string
	Amount    string
	CreatedAt time.Time

	FullName  string
	AvatarURL string
}

func (packet *Packet) GetParticipants(ctx context.Context) error {
	txn := session.Database(ctx).ReadOnlyTransaction()
	defer txn.Close()

	it := txn.Query(ctx, spanner.Statement{
		SQL:    "SELECT user_id,amount,created_at FROM participants WHERE packet_id=@packet_id LIMIT 1000",
		Params: map[string]interface{}{"packet_id": packet.PacketId},
	})
	defer it.Stop()

	var participants []*Participant
	var userIds []string
	for {
		row, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return session.TransactionError(ctx, err)
		}
		var p Participant
		err = row.Columns(&p.UserId, &p.Amount, &p.CreatedAt)
		if err != nil {
			return session.TransactionError(ctx, err)
		}
		participants = append(participants, &p)
		userIds = append(userIds, p.UserId)
	}

	cit := txn.Query(ctx, spanner.Statement{
		SQL:    "SELECT user_id,full_name,avatar_url FROM users WHERE user_id IN UNNEST(@user_ids)",
		Params: map[string]interface{}{"user_ids": userIds},
	})
	defer cit.Stop()

	var userInfo = make(map[string]*User)
	for {
		row, err := cit.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return session.TransactionError(ctx, err)
		}
		var u User
		err = row.Columns(&u.UserId, &u.FullName, &u.AvatarURL)
		if err != nil {
			return session.TransactionError(ctx, err)
		}
		userInfo[u.UserId] = &u
	}

	for _, p := range participants {
		if userInfo[p.UserId] != nil {
			p.FullName = userInfo[p.UserId].FullName
			p.AvatarURL = userInfo[p.UserId].AvatarURL
		}
	}

	sort.Slice(participants, func(i, j int) bool { return participants[i].CreatedAt.Before(participants[j].CreatedAt) })
	packet.Participants = participants
	return nil
}

func ListPendingParticipants(ctx context.Context, limit int) ([]*Participant, error) {
	query := fmt.Sprintf("SELECT packet_id,user_id,amount FROM participants@{FORCE_INDEX=participants_by_created_paid} WHERE paid_at IS NULL ORDER BY created_at LIMIT %d", limit)
	it := session.Database(ctx).Query(ctx, spanner.Statement{SQL: query}, "participants", "ListPendingParticipants")
	defer it.Stop()

	var participants []*Participant
	for {
		row, err := it.Next()
		if err == iterator.Done {
			return participants, nil
		} else if err != nil {
			return participants, session.TransactionError(ctx, err)
		}
		var p Participant
		err = row.Columns(&p.PacketId, &p.UserId, &p.Amount)
		if err != nil {
			return participants, session.TransactionError(ctx, err)
		}
		participants = append(participants, &p)
	}
}

func SendParticipantTransfer(ctx context.Context, packetId, userId string, amount string) error {
	traceId, err := generateParticipantId(packetId, userId)
	if err != nil {
		return session.ServerError(ctx, err)
	}

	txn := session.Database(ctx).ReadOnlyTransaction()
	defer txn.Close()
	packet, err := readPacketWithAssetAndUser(ctx, txn, packetId)
	if err != nil {
		return session.TransactionError(ctx, err)
	}
	memo := fmt.Sprintf("来自 %s 的红包", packet.User.FullName)
	if strings.TrimSpace(packet.User.FullName) == "" {
		memo = "来自无名氏的红包"
	}
	if count := utf8.RuneCountInString(memo); count > 100 {
		name := string([]rune(packet.User.FullName)[:16])
		memo = fmt.Sprintf("来自 %s 的红包", name)
	}

	in := &bot.TransferInput{
		AssetId:     packet.AssetId,
		RecipientId: userId,
		Amount:      number.FromString(amount),
		TraceId:     traceId,
		Memo:        memo,
	}
	if !number.FromString(amount).Exhausted() {
		err = bot.CreateTransfer(ctx, in, config.ClientId, config.SessionId, config.SessionKey, config.SessionAssetPIN, config.PinToken)
		if err != nil {
			return session.ServerError(ctx, err)
		}
	}

	_, err = session.Database(ctx).ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		return txn.BufferWrite([]*spanner.Mutation{
			spanner.Update("participants", []string{"packet_id", "user_id", "paid_at"}, []interface{}{packetId, userId, time.Now()}),
		})
	}, "packets", "UPDATE", "SendParticipantTransfer")
	if err != nil {
		return session.TransactionError(ctx, err)
	}
	return nil
}

func generateParticipantId(packetId, userId string) (string, error) {
	minId, maxId := packetId, userId
	if strings.Compare(packetId, userId) > 0 {
		maxId, minId = packetId, userId
	}
	h := md5.New()
	io.WriteString(h, minId)
	io.WriteString(h, maxId)
	sum := h.Sum(nil)
	sum[6] = (sum[6] & 0x0f) | 0x30
	sum[8] = (sum[8] & 0x3f) | 0x80
	id, err := uuid.FromBytes(sum)
	return id.String(), err
}
