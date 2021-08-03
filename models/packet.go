package models

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	bot "github.com/MixinNetwork/bot-api-go-client"
	number "github.com/MixinNetwork/go-number"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/gofrs/uuid"
)

const (
	PacketStateInitial  = "INITIAL"
	PacketStatePaid     = "PAID"
	PacketStateExpired  = "EXPIRED"
	PacketStateRefunded = "REFUNDED"

	PacketSizeLimit = 500
	shareShardId    = "c94ac88f-4671-3976-b60a-09064f1811e8"
)

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

var packetsCols = []string{"packet_id", "user_id", "asset_id", "amount", "greeting", "total_count", "remaining_count", "remaining_amount", "state", "created_at"}

func (p *Packet) values() []interface{} {
	return []interface{}{p.PacketId, p.UserId, p.AssetId, p.Amount, p.Greeting, p.TotalCount, p.RemainingCount, p.RemainingAmount, p.State, p.CreatedAt}
}

func packetFromRow(row durable.Row) (*Packet, error) {
	var p Packet
	err := row.Scan(&p.PacketId, &p.UserId, &p.AssetId, &p.Amount, &p.Greeting, &p.TotalCount, &p.RemainingCount, &p.RemainingAmount, &p.State, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func (current *User) Prepare(ctx context.Context) (int64, error) {
	sum, err := SubscribersCount(ctx)
	if sum > PacketSizeLimit {
		sum = PacketSizeLimit
	}
	return sum, err
}

func (current *User) CreatePacket(ctx context.Context, assetId string, amount number.Decimal, totalCount int64, greeting string) (*Packet, error) {
	if !current.isAdmin() {
		b, err := ReadProhibitedProperty(ctx)
		if err != nil {
			return nil, err
		} else if b {
			return nil, session.ForbiddenError(ctx)
		}
	}
	asset, err := current.ShowAsset(ctx, assetId)
	if err != nil {
		return nil, err
	}
	if config.AppConfig.System.PriceAssetsEnable {
		if number.FromString(asset.PriceUSD).Cmp(number.Zero()) <= 0 {
			return nil, session.BadDataError(ctx)
		}
	}
	u, _ := bot.UserMe(ctx, current.AccessToken)
	if u != nil {
		name := strings.TrimSpace(u.FullName)
		if name != current.FullName || u.AvatarURL != current.AvatarURL {
			if name != "" {
				current.FullName = name
			}
			current.AvatarURL = u.AvatarURL
			if _, err = session.Database(ctx).ExecContext(ctx, "UPDATE users SET (full_name, avatar_url)=($1,$2) WHERE user_id=$3", current.FullName, current.AvatarURL, current.UserId); err != nil {
				session.TransactionError(ctx, err)
			}
		}
	}
	return current.createPacket(ctx, asset, amount, totalCount, greeting)
}

func (current *User) createPacket(ctx context.Context, asset *Asset, amount number.Decimal, totalCount int64, greeting string) (*Packet, error) {
	if amount.Cmp(number.FromString("0.0001")) < 0 {
		return nil, session.BadDataError(ctx)
	}
	if utf8.RuneCountInString(greeting) > 36 {
		greeting = string([]rune(greeting)[:36])
	}
	amount = amount.RoundFloor(8)
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

	query := durable.PrepareQuery("INSERT INTO packets (%s) VALUES (%s)", packetsCols)
	_, err = session.Database(ctx).ExecContext(ctx, query, packet.values()...)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return packet, nil
}

func PayPacket(ctx context.Context, packetId string, assetId, amount string) (*Packet, error) {
	var packet *Packet
	err := session.Database(ctx).RunInTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		packet, err = readPacketWithAssetAndUser(ctx, tx, packetId)
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
		_, err = tx.ExecContext(ctx, "UPDATE packets SET state=$1 WHERE packet_id=$2", packet.State, packet.PacketId)
		if err != nil {
			return err
		}
		return handlePacketExpiration(ctx, tx, packet)
	})
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return packet, nil
}

func ShowPacket(ctx context.Context, packetId string) (*Packet, error) {
	var packet *Packet
	err := session.Database(ctx).RunInTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		packet, err = readPacketWithAssetAndUser(ctx, tx, packetId)
		if err != nil || packet == nil {
			return err
		}
		return handlePacketExpiration(ctx, tx, packet)
	})
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	if packet != nil {
		err = packet.GetParticipants(ctx)
		if err != nil {
			return nil, session.TransactionError(ctx, err)
		}
	}
	return packet, nil
}

var mutexeSet map[string]*sync.Mutex

func (current *User) ClaimPacket(ctx context.Context, packetId string) (*Packet, error) {
	if current.State != PaymentStatePaid {
		return nil, session.ForbiddenError(ctx)
	}
	packet, err := ShowPacket(ctx, packetId)
	if err != nil || packet == nil {
		return nil, err
	}
	if packet.State != PacketStatePaid {
		return packet, nil
	}
	if packet.RemainingCount > packet.TotalCount {
		return nil, session.InsufficientAccountBalanceError(ctx)
	}
	if number.FromString(packet.RemainingAmount).Cmp(number.FromString(packet.Amount)) > 0 {
		return nil, session.InsufficientAccountBalanceError(ctx)
	}

	err = session.Database(ctx).RunInTransaction(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable}, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		packet, err = readPacketWithAssetAndUser(ctx, tx, packetId)
		if err != nil || packet == nil {
			return err
		}
		err = handlePacketExpiration(ctx, tx, packet)
		if err != nil {
			return err
		}
		var userId string
		err = tx.QueryRowContext(ctx, "SELECT user_id FROM participants WHERE packet_id=$1 AND user_id=$2", packet.PacketId, current.UserId).Scan(&userId)
		if err == sql.ErrNoRows {
			err = handlePacketClaim(ctx, tx, packet, current.UserId)
			if err != nil {
				return err
			}
			b, err := readProhibitedStatus(ctx, tx)
			if err == nil && !b {
				dm, err := buildDistributeMessage(ctx, bot.UuidNewV4().String(), bot.UuidNewV4().String(), "", config.AppConfig.Mixin.ClientId, packet.UserId, MessageCategoryPlainText, base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(config.AppConfig.MessageTemplate.GroupOpenedRedPacket, current.FullName))), false)
				if err != nil {
					return err
				}
				query := durable.PrepareQuery("INSERT INTO distributed_messages (%s) VALUES (%s)", distributedMessagesCols)
				_, err = tx.ExecContext(ctx, query, dm.values()...)
				return err
			}
		}
		return err
	})
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	err = packet.GetParticipants(ctx)
	if err != nil {
		return nil, err
	}
	return packet, nil
}

func RefundPacket(ctx context.Context, packetId string) (*Packet, error) {
	var packet *Packet
	err := session.Database(ctx).RunInTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		packet, err = readPacketWithAssetAndUser(ctx, tx, packetId)
		if err != nil || packet == nil {
			return err
		}
		err = handlePacketExpiration(ctx, tx, packet)
		if err != nil {
			return err
		}
		if packet.State != PacketStateExpired {
			return nil
		}
		packet.State = PacketStateRefunded
		_, err = tx.ExecContext(ctx, "UPDATE packets SET state=$1 WHERE packet_id=$2", packet.State, packet.PacketId)
		return err
	})
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
	mixin := config.AppConfig.Mixin
	_, err = bot.CreateTransfer(ctx, in, mixin.ClientId, mixin.SessionId, mixin.SessionKey, mixin.SessionAssetPIN, mixin.PinToken)
	if err != nil {
		return nil, session.ServerError(ctx, err)
	}

	return RefundPacket(ctx, packetId)
}

func ListExpiredPackets(ctx context.Context, limit int) ([]string, error) {
	var packetIds []string
	query := "SELECT packet_id FROM packets WHERE state IN ($1, $2) AND created_at<$3 LIMIT $4"
	rows, err := session.Database(ctx).QueryContext(ctx, query, PacketStatePaid, PacketStateExpired, time.Now().Add(-25*time.Hour), limit)
	if err != nil {
		return packetIds, session.TransactionError(ctx, err)
	}
	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return packetIds, session.TransactionError(ctx, err)
		}
		packetIds = append(packetIds, id)
	}
	return packetIds, nil
}

func handlePacketClaim(ctx context.Context, tx *sql.Tx, packet *Packet, userId string) error {
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
	_, err := tx.ExecContext(ctx, "UPDATE packets SET (remaining_count, remaining_amount)=($1,$2) WHERE packet_id=$3", packet.RemainingCount, packet.RemainingAmount, packet.PacketId)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "INSERT INTO participants (packet_id,user_id,amount) VALUES ($1, $2, $3)", packet.PacketId, userId, amount.Persist())
	return err
}

func handlePacketExpiration(ctx context.Context, tx *sql.Tx, packet *Packet) error {
	if packet.State != PacketStatePaid {
		return nil
	}
	if packet.RemainingCount == 0 || number.FromString(packet.RemainingAmount).Exhausted() {
		packet.State = PacketStateRefunded
	} else if packet.CreatedAt.Before(time.Now().Add(-24 * time.Hour)) {
		packet.State = PacketStateExpired
	}
	if packet.State == PacketStatePaid {
		return nil
	}
	_, err := tx.ExecContext(ctx, "UPDATE packets SET state=$1 WHERE packet_id=$2", packet.State, packet.PacketId)
	return err
}

func readPacketWithAssetAndUser(ctx context.Context, tx *sql.Tx, packetId string) (*Packet, error) {
	packet, err := readPacket(ctx, tx, packetId)
	if err != nil || packet == nil {
		return nil, err
	}
	packet.Asset, err = findAssetById(ctx, tx, packet.AssetId)
	if err != nil {
		return nil, err
	}
	if packet.Asset == nil {
		return nil, nil
	}
	packet.User, err = findUserById(ctx, tx, packet.UserId)
	if err != nil {
		return nil, err
	}
	if packet.User == nil {
		return nil, nil
	}
	return packet, nil
}

func readPacket(ctx context.Context, tx *sql.Tx, packetId string) (*Packet, error) {
	query := fmt.Sprintf("SELECT %s FROM packets WHERE packet_id=$1", strings.Join(packetsCols, ","))
	row := tx.QueryRowContext(ctx, query, packetId)
	p, err := packetFromRow(row)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return p, nil
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

func init() {
	mutexeSet = make(map[string]*sync.Mutex, 0)
}
