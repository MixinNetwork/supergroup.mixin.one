package models

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"math/rand"
	"time"
	"unicode/utf8"

	"cloud.google.com/go/spanner"
	bot "github.com/MixinNetwork/bot-api-go-client"
	number "github.com/MixinNetwork/go-number"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/api/iterator"
)

const (
	PacketStateInitial  = "INITIAL"
	PacketStatePaid     = "PAID"
	PacketStateExpired  = "EXPIRED"
	PacketStateRefunded = "REFUNDED"
)

const packets_DDL = `
CREATE TABLE packets (
	packet_id         STRING(36) NOT NULL,
	user_id	          STRING(36) NOT NULL,
	asset_id          STRING(36) NOT NULL,
	amount            STRING(128) NOT NULL,
	greeting          STRING(36) NOT NULL,
	total_count       INT64 NOT NULL,
	remaining_count   INT64 NOT NULL,
	remaining_amount  STRING(128) NOT NULL,
	state             STRING(36) NOT NULL,
	created_at        TIMESTAMP NOT NULL,
) PRIMARY KEY(packet_id);

CREATE INDEX packets_by_state_created ON packets(state, created_at);
`

var packetsCols = []string{"packet_id", "user_id", "asset_id", "amount", "greeting", "total_count", "remaining_count", "remaining_amount", "state", "created_at"}

func (p *Packet) values() []interface{} {
	return []interface{}{p.PacketId, p.UserId, p.AssetId, p.Amount, p.Greeting, p.TotalCount, p.RemainingCount, p.RemainingAmount, p.State, p.CreatedAt}
}

type Packet struct {
	PacketId        string
	UserId          string
	AssetId         string
	Amount          string
	Greeting        string
	TotalCount      int64
	RemainingCount  int64
	RemainingAmount string
	State           string
	CreatedAt       time.Time

	User         *User
	Asset        *Asset
	Participants []*Participant
}

func (current *User) Prepare(ctx context.Context) (int, error) {
	sum, err := SubscribersCount(ctx)
	return int(sum), err
}

func (current *User) CreatePacket(ctx context.Context, assetId string, amount number.Decimal, totalCount int64, greeting string) (*Packet, error) {
	if amount.Cmp(number.FromString("0.0001")) < 0 {
		return nil, session.BadDataError(ctx)
	}
	if utf8.RuneCountInString(greeting) > 36 {
		return nil, session.BadDataError(ctx)
	}
	asset, err := current.ShowAsset(ctx, assetId)
	if err != nil {
		return nil, session.ServerError(ctx, err)
	}
	if number.FromString(asset.Balance).Cmp(amount) < 0 {
		return nil, session.InsufficientAccountBalanceError(ctx)
	}
	participantsCount, err := current.Prepare(ctx)
	if err != nil {
		return nil, err
	}
	if totalCount <= 0 || totalCount > int64(participantsCount) {
		return nil, session.BadDataError(ctx)
	}
	packet := &Packet{
		PacketId:        bot.UuidNewV4().String(),
		UserId:          current.UserId,
		AssetId:         asset.AssetId,
		Amount:          amount.Persist(),
		Greeting:        greeting,
		TotalCount:      totalCount,
		RemainingCount:  totalCount,
		RemainingAmount: amount.Persist(),
		State:           PacketStateInitial,
		CreatedAt:       time.Now(),
		User:            current,
		Asset:           asset,
	}
	err = session.Database(ctx).Apply(ctx, []*spanner.Mutation{
		spanner.Insert("packets", packetsCols, packet.values()),
	}, "packets", "INSERT", "CreatePacket")
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return packet, nil
}

func PayPacket(ctx context.Context, packetId string, assetId, amount string) (*Packet, error) {
	var packet *Packet
	_, err := session.Database(ctx).ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		var err error
		packet, err = readPacketWithAssetAndUser(ctx, txn, packetId)
		if err != nil || packet == nil {
			return err
		}
		if packet.State != PacketStateInitial {
			return nil
		}
		if assetId != packet.AssetId || number.FromString(amount).Cmp(number.FromString(packet.Amount)) < 0 {
			return nil
		}
		packet.State = PacketStatePaid
		txn.BufferWrite([]*spanner.Mutation{
			spanner.Update("packets", []string{"packet_id", "state"}, []interface{}{packetId, packet.State}),
		})
		return handlePacketExpiration(ctx, txn, packet)
	}, "packets", "UPDATE", "PayPacket")
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return packet, nil
}

func ShowPacket(ctx context.Context, packetId string) (*Packet, error) {
	var packet *Packet
	_, err := session.Database(ctx).ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		var err error
		packet, err = readPacketWithAssetAndUser(ctx, txn, packetId)
		if err != nil || packet == nil {
			return err
		}
		return handlePacketExpiration(ctx, txn, packet)
	}, "packets", "SELECT", "ShowPacket")
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	err = packet.GetParticipants(ctx)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return packet, nil
}

func (current *User) ClaimPacket(ctx context.Context, packetId string) (*Packet, error) {
	packet, err := ShowPacket(ctx, packetId)
	if err != nil || packet == nil {
		return nil, err
	}
	if packet.State != PacketStatePaid {
		return packet, nil
	}
	_, err = session.Database(ctx).ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		packet, err = readPacketWithAssetAndUser(ctx, txn, packetId)
		if err != nil || packet == nil {
			return err
		}
		err = handlePacketExpiration(ctx, txn, packet)
		if err != nil {
			return err
		}
		it := txn.Read(ctx, "participants", spanner.Key{packet.PacketId, current.UserId}, []string{"packet_id", "user_id"})
		defer it.Stop()

		_, err := it.Next()
		if err == iterator.Done {
			return handlePacketClaim(ctx, txn, packet, current.UserId)
		} else if err != nil {
			return err
		} else {
			return nil
		}
	}, "participants", "INSERT", "ClaimPacket")
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	err = packet.GetParticipants(ctx)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	if packet != nil {
		mutation := createDistributeMessage(ctx, config.ClientId, packet.UserId, "PLAIN_TEXT", []byte(fmt.Sprintf("%s 打开了你的红包", current.FullName)))
		session.Database(ctx).Apply(ctx, []*spanner.Mutation{mutation}, "distributed_messages", "INSERT", "ClaimPacket")
	}
	return packet, nil
}

func RefundPacket(ctx context.Context, packetId string) (*Packet, error) {
	var packet *Packet
	_, err := session.Database(ctx).ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		var err error
		packet, err = readPacketWithAssetAndUser(ctx, txn, packetId)
		if err != nil || packet == nil {
			return err
		}
		err = handlePacketExpiration(ctx, txn, packet)
		if err != nil {
			return err
		}
		if packet.State != PacketStateExpired {
			return nil
		}
		packet.State = PacketStateRefunded
		return txn.BufferWrite([]*spanner.Mutation{
			spanner.Update("packets", []string{"packet_id", "state"}, []interface{}{packet.PacketId, packet.State}),
		})
	}, "packets", "UPDATE", "RefundPacket")
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	err = packet.GetParticipants(ctx)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return packet, nil
}

func SendPacketRefundTransfer(ctx context.Context, packetId string) (*Packet, error) {
	traceId, err := generatePacketRefundId(packetId)
	if err != nil {
		return nil, session.ServerError(ctx, err)
	}

	packet, err := ShowPacket(ctx, packetId)
	if err != nil || packet == nil {
		return nil, err
	}
	if packet.State != PacketStateExpired {
		return packet, nil
	}

	in := &bot.TransferInput{
		AssetId:     packet.AssetId,
		RecipientId: packet.UserId,
		Amount:      number.FromString(packet.RemainingAmount),
		TraceId:     traceId,
		Memo:        "",
	}
	err = bot.CreateTransfer(ctx, in, config.ClientId, config.SessionId, config.SessionKey, config.SessionAssetPIN, config.PinToken)
	if err != nil {
		return nil, session.ServerError(ctx, err)
	}

	return RefundPacket(ctx, packetId)
}

func ListExpiredPackets(ctx context.Context, limit int) ([]string, error) {
	it := session.Database(ctx).Query(ctx, spanner.Statement{
		SQL:    fmt.Sprintf("SELECT packet_id FROM packets@{FORCE_INDEX=packets_by_state_created} WHERE state IN UNNEST(@states) AND created_at<@threshold LIMIT %d", limit),
		Params: map[string]interface{}{"states": []string{PacketStatePaid, PacketStateExpired}, "threshold": time.Now().Add(-25 * time.Hour)},
	}, "packets", "ListExpiredPackets")
	defer it.Stop()

	var packetIds []string
	for {
		row, err := it.Next()
		if err == iterator.Done {
			return packetIds, nil
		} else if err != nil {
			return packetIds, session.TransactionError(ctx, err)
		}
		var id string
		err = row.Columns(&id)
		if err != nil {
			return packetIds, session.TransactionError(ctx, err)
		}
		packetIds = append(packetIds, id)
	}
}

func handlePacketClaim(ctx context.Context, txn *spanner.ReadWriteTransaction, packet *Packet, userId string) error {
	if packet.State != PacketStatePaid {
		return nil
	}
	amount := number.FromString(packet.RemainingAmount)
	if packet.RemainingCount > 1 && amount.Cmp(number.FromString("0.000001")) > 0 {
		amount = amount.Mul(number.FromString("2")).Div(number.FromString(fmt.Sprint(packet.RemainingCount)))
		if amount.Cmp(number.FromString("0.000001")) > 0 {
			rand.Seed(time.Now().UnixNano())
			for {
				amount = amount.Mul(number.FromString(fmt.Sprint(rand.Float64())))
				for d := int32(1); d < 8; d++ {
					round := amount.RoundFloor(d)
					if !round.Exhausted() {
						amount = round
						break
					}
				}
				if !amount.Exhausted() {
					break
				}
			}
		}
	}
	amount = number.FromString(amount.PresentFloor())
	packet.RemainingCount = packet.RemainingCount - 1
	packet.RemainingAmount = number.FromString(packet.RemainingAmount).Sub(amount).Persist()
	return txn.BufferWrite([]*spanner.Mutation{
		spanner.Update("packets", []string{"packet_id", "remaining_count", "remaining_amount"}, []interface{}{packet.PacketId, packet.RemainingCount, packet.RemainingAmount}),
		spanner.Insert("participants", []string{"packet_id", "user_id", "amount", "created_at"}, []interface{}{packet.PacketId, userId, amount.Persist(), time.Now()}),
	})
}

func handlePacketExpiration(ctx context.Context, txn *spanner.ReadWriteTransaction, packet *Packet) error {
	if packet.State != PacketStatePaid {
		return nil
	}
	if packet.RemainingCount == 0 || number.FromString(packet.RemainingAmount).Exhausted() {
		packet.State = PacketStateRefunded
	} else if packet.CreatedAt.Before(time.Now().Add(-24 * time.Hour)) {
		packet.State = PacketStateExpired
	}
	return txn.BufferWrite([]*spanner.Mutation{
		spanner.Update("packets", []string{"packet_id", "state"}, []interface{}{packet.PacketId, packet.State}),
	})
}

func readPacketWithAssetAndUser(ctx context.Context, txn durable.Transaction, packetId string) (*Packet, error) {
	packet, err := readPacket(ctx, txn, packetId)
	if err != nil || packet == nil {
		return nil, err
	}
	packet.Asset, err = readAsset(ctx, txn, packet.AssetId)
	if err != nil {
		return nil, err
	}
	packet.User, err = readUser(ctx, txn, packet.UserId)
	if err != nil {
		return nil, err
	}
	return packet, nil
}

func readPacket(ctx context.Context, txn durable.Transaction, packetId string) (*Packet, error) {
	it := txn.Read(ctx, "packets", spanner.Key{packetId}, packetsCols)
	defer it.Stop()

	row, err := it.Next()
	if err == iterator.Done {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return packetFromRow(row)
}

func ReadPackageWithRelation(ctx context.Context, packetId string) (*Packet, error) {
	txn := session.Database(ctx).ReadOnlyTransaction()
	defer txn.Close()

	return readPacketWithAssetAndUser(ctx, txn, packetId)
}

func packetFromRow(row *spanner.Row) (*Packet, error) {
	var p Packet
	err := row.Columns(&p.PacketId, &p.UserId, &p.AssetId, &p.Amount, &p.Greeting, &p.TotalCount, &p.RemainingCount, &p.RemainingAmount, &p.State, &p.CreatedAt)
	return &p, err
}

func generatePacketRefundId(packetId string) (string, error) {
	h := md5.New()
	io.WriteString(h, packetId)
	io.WriteString(h, "REFUND")
	sum := h.Sum(nil)
	sum[6] = (sum[6] & 0x0f) | 0x30
	sum[8] = (sum[8] & 0x3f) | 0x80
	id, err := uuid.FromBytes(sum)
	return id.String(), err
}
