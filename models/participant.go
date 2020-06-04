package models

import (
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	bot "github.com/MixinNetwork/bot-api-go-client"
	number "github.com/MixinNetwork/go-number"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/gofrs/uuid"
	"github.com/lib/pq"
)

type Participant struct {
	PacketId  string
	UserId    string
	Amount    string
	CreatedAt time.Time
	PaidAt    pq.NullTime

	FullName  string
	AvatarURL string
}

func participantFromRow(row durable.Row) (*Participant, error) {
	var p Participant
	err := row.Scan(&p.PacketId, &p.UserId, &p.Amount, &p.CreatedAt, &p.PaidAt, &p.FullName, &p.AvatarURL)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func (packet *Packet) GetParticipants(ctx context.Context) error {
	query := fmt.Sprintf("SELECT p.packet_id,p.user_id,p.amount,p.created_at,p.paid_at,u.full_name,u.avatar_url FROM participants p INNER JOIN users u ON p.user_id=u.user_id WHERE p.packet_id=$1 ORDER BY p.created_at")
	rows, err := session.Database(ctx).QueryContext(ctx, query, packet.PacketId)
	if err != nil {
		return session.TransactionError(ctx, err)
	}

	var participants []*Participant
	for rows.Next() {
		p, err := participantFromRow(rows)
		if err != nil {
			return session.TransactionError(ctx, err)
		}
		participants = append(participants, p)
	}
	packet.Participants = participants
	return nil
}

func ListPendingParticipants(ctx context.Context, limit int) ([]*Participant, error) {
	var participants []*Participant
	query := "SELECT packet_id,user_id,amount FROM participants WHERE paid_at IS NULL ORDER BY created_at LIMIT $1"
	rows, err := session.Database(ctx).QueryContext(ctx, query, limit)
	if err != nil {
		return participants, session.TransactionError(ctx, err)
	}

	for rows.Next() {
		var p Participant
		err = rows.Scan(&p.PacketId, &p.UserId, &p.Amount)
		if err != nil {
			return participants, session.TransactionError(ctx, err)
		}
		participants = append(participants, &p)
	}
	return participants, nil
}

func SendParticipantTransfer(ctx context.Context, packetId, userId string, amount string) error {
	traceId, err := generateParticipantId(packetId, userId)
	if err != nil {
		return session.ServerError(ctx, err)
	}
	err = session.Database(ctx).RunInTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
		packet, err := readPacketWithAssetAndUser(ctx, tx, packetId)
		if err != nil {
			return err
		}
		if packet == nil {
			_, err = tx.ExecContext(ctx, "DELETE FROM packets WHERE packet_id=$1", packetId)
			return err
		}
		memo := fmt.Sprintf(config.AppConfig.MessageTemplate.GroupRedPacketDesc, packet.User.FullName)
		if strings.TrimSpace(packet.User.FullName) == "" {
			memo = config.AppConfig.MessageTemplate.GroupRedPacketShortDesc
		}
		if count := utf8.RuneCountInString(memo); count > 100 {
			name := string([]rune(packet.User.FullName)[:16])
			memo = fmt.Sprintf(config.AppConfig.MessageTemplate.GroupRedPacketDesc, name)
		}
		in := &bot.TransferInput{
			AssetId:     packet.AssetId,
			RecipientId: userId,
			Amount:      number.FromString(amount),
			TraceId:     traceId,
			Memo:        memo,
		}
		if !number.FromString(amount).Exhausted() {
			err = bot.CreateTransfer(ctx, in, config.AppConfig.Mixin.ClientId, config.AppConfig.Mixin.SessionId, config.AppConfig.Mixin.SessionKey, config.AppConfig.Mixin.SessionAssetPIN, config.AppConfig.Mixin.PinToken)
			if err != nil {
				return err
			}
		}
		_, err = tx.ExecContext(ctx, "UPDATE participants SET paid_at=$1 WHERE packet_id=$2 AND user_id=$3", time.Now(), packetId, userId)
		return err
	})
	if err != nil {
		return session.ServerError(ctx, err)
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
