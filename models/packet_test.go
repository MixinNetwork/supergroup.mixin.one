package models

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"testing"
	"time"

	bot "github.com/MixinNetwork/bot-api-go-client/v2"
	number "github.com/MixinNetwork/go-number"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/stretchr/testify/assert"
)

func TestPacketCRUD(t *testing.T) {
	assert := assert.New(t)
	ctx := setupTestContext()
	defer teardownTestContext(ctx)

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	assert.Nil(err)
	public := base64.RawURLEncoding.EncodeToString(pub)
	private := base64.RawURLEncoding.EncodeToString(priv)
	authorizationID := bot.UuidNewV4().String()

	user, err := createUser(ctx, public, private, authorizationID, "", bot.UuidNewV4().String(), "1000", "name", "http://localhost")
	assert.Nil(err)
	assert.NotNil(user)
	err = user.Payment(ctx)
	assert.Nil(err)
	err = user.Subscribe(ctx)
	assert.Nil(err)
	sum, err := user.Prepare(ctx)
	assert.Nil(err)
	assert.Equal(int64(1), sum)
	err = user.Payment(ctx)
	assert.Nil(err)

	li, err := createUser(ctx, public, private, authorizationID, "", bot.UuidNewV4().String(), "1001", "Li", "http://localhost")
	assert.Nil(err)
	assert.NotNil(li)
	err = li.Payment(ctx)
	assert.Nil(err)
	err = li.Subscribe(ctx)
	assert.Nil(err)
	sum, err = user.Prepare(ctx)
	assert.Nil(err)
	assert.Equal(int64(2), sum)
	err = li.Payment(ctx)
	assert.Nil(err)

	asset := &Asset{
		AssetId:  bot.UuidNewV4().String(),
		Symbol:   "XIN",
		Name:     "Mixin",
		IconURL:  "http://mixin.one",
		PriceBTC: "0",
		PriceUSD: "0",
		Balance:  "100",
	}
	err = upsertAssets(ctx, []*Asset{asset})
	assert.Nil(err)
	packet, err := li.createPacket(ctx, asset, number.FromString("1"), 2, "Hello Packet")
	assert.Nil(err)
	assert.NotNil(packet)
	assert.Equal(PacketStateInitial, packet.State)
	packet, err = PayPacket(ctx, packet.PacketId, asset.AssetId, "1")
	assert.Nil(err)
	assert.NotNil(packet)
	assert.Equal(PacketStatePaid, packet.State)
	packet, err = ShowPacket(ctx, packet.PacketId)
	assert.Nil(err)
	assert.NotNil(packet)
	assert.Equal("1", packet.Amount)
	assert.Equal(int64(2), packet.TotalCount)
	packet, err = li.ClaimPacket(ctx, packet.PacketId)
	assert.Nil(err)
	assert.NotNil(packet)
	assert.Equal(PacketStatePaid, packet.State)
	packet, err = RefundPacket(ctx, packet.PacketId)
	assert.Nil(err)
	assert.NotNil(packet)
	assert.Equal(PacketStatePaid, packet.State)
	packet, err = user.ClaimPacket(ctx, packet.PacketId)
	assert.Nil(err)
	assert.NotNil(packet)
	assert.Equal(PacketStatePaid, packet.State)
	packet, err = ShowPacket(ctx, packet.PacketId)
	assert.Nil(err)
	assert.NotNil(packet)
	assert.Equal(int64(0), packet.RemainingCount)
	assert.Equal("0", packet.RemainingAmount)
	assert.Len(packet.Participants, 2)
	packet, err = li.createPacket(ctx, asset, number.FromString("1"), 2, "Hello Packet Hello Packet Hello Packet Hello Packet Hello")
	assert.Nil(err)
	assert.NotNil(packet)
	packet, err = PayPacket(ctx, packet.PacketId, asset.AssetId, "1")
	assert.Nil(err)
	assert.NotNil(packet)
	assert.Equal(PacketStatePaid, packet.State)
	_, err = session.Database(ctx).ExecContext(ctx, "UPDATE packets SET created_at=$1 WHERE packet_id=$2", time.Now().Add(-25*time.Hour), packet.PacketId)
	assert.Nil(err)
	packet, err = ShowPacket(ctx, packet.PacketId)
	assert.Nil(err)
	assert.NotNil(packet)
	assert.Equal(PacketStateExpired, packet.State)
	ids, err := ListExpiredPackets(ctx, 100)
	assert.Nil(err)
	assert.Len(ids, 1)
	packet, err = RefundPacket(ctx, packet.PacketId)
	assert.Nil(err)
	assert.NotNil(packet)
	assert.Equal(PacketStateRefunded, packet.State)
	packet, err = testReadPacketWithRelation(ctx, packet.PacketId)
	assert.Nil(err)
	assert.NotNil(packet)
	packet, err = testReadPacketWithRelation(ctx, bot.UuidNewV4().String())
	assert.Nil(err)
	assert.Nil(packet)
}

func testReadPacketWithRelation(ctx context.Context, packetId string) (*Packet, error) {
	var packet *Packet
	err := session.Database(ctx).RunInTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		packet, err = readPacketWithAssetAndUser(ctx, tx, packetId)
		return err
	})
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return packet, err
}
