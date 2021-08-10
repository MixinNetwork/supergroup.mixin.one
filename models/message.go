package models

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"
	"unicode/utf8"

	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/gofrs/uuid"
	"golang.org/x/crypto/curve25519"
)

const (
	MessageStatePending = "pending"
	MessageStateSuccess = "success"

	MessageCategoryPlainText           = "PLAIN_TEXT"
	MessageCategoryPlainImage          = "PLAIN_IMAGE"
	MessageCategoryPlainVideo          = "PLAIN_VIDEO"
	MessageCategoryPlainLive           = "PLAIN_LIVE"
	MessageCategoryPlainData           = "PLAIN_DATA"
	MessageCategoryPlainSticker        = "PLAIN_STICKER"
	MessageCategoryPlainContact        = "PLAIN_CONTACT"
	MessageCategoryPlainAudio          = "PLAIN_AUDIO"
	MessageCategoryPlainPost           = "PLAIN_POST"
	MessageCategoryPlainTranscript     = "PLAIN_TRANSCRIPT"
	MessageCategoryEncryptedPost       = "ENCRYPTED_POST"
	MessageCategoryEncryptedText       = "ENCRYPTED_TEXT"
	MessageCategoryEncryptedImage      = "ENCRYPTED_IMAGE"
	MessageCategoryEncryptedVideo      = "ENCRYPTED_VIDEO"
	MessageCategoryEncryptedLive       = "ENCRYPTED_LIVE"
	MessageCategoryEncryptedAudio      = "ENCRYPTED_AUDIO"
	MessageCategoryEncryptedData       = "ENCRYPTED_DATA"
	MessageCategoryEncryptedSticker    = "ENCRYPTED_STICKER"
	MessageCategoryEncryptedContact    = "ENCRYPTED_CONTACT"
	MessageCategoryEncryptedLocation   = "ENCRYPTED_LOCATION"
	MessageCategoryEncryptedTranscript = "ENCRYPTED_TRANSCRIPT"
	MessageCategoryAppCard             = "APP_CARD"
	MessageCategoryAppButtonGroup      = "APP_BUTTON_GROUP"
	MessageCategoryMessageRecall       = "MESSAGE_RECALL"
)

type Message struct {
	MessageId        string
	UserId           string
	Category         string
	QuoteMessageId   string
	Data             string
	Silent           bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
	State            string
	LastDistributeAt time.Time

	FullName sql.NullString
}

var messagesCols = []string{"message_id", "user_id", "category", "quote_message_id", "data", "silent", "created_at", "updated_at", "state", "last_distribute_at"}

func (m *Message) values() []interface{} {
	return []interface{}{m.MessageId, m.UserId, m.Category, m.QuoteMessageId, m.Data, m.Silent, m.CreatedAt, m.UpdatedAt, m.State, m.LastDistributeAt}
}

func messageFromRow(row durable.Row) (*Message, error) {
	var m Message
	err := row.Scan(&m.MessageId, &m.UserId, &m.Category, &m.QuoteMessageId, &m.Data, &m.Silent, &m.CreatedAt, &m.UpdatedAt, &m.State, &m.LastDistributeAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &m, err
}

func CreateMessage(ctx context.Context, user *User, messageId, category, quoteMessageId, data string, silent bool, createdAt, updatedAt time.Time) (*Message, error) {
	if len(data) > 5*1024 {
		return nil, notifyTooLarge(ctx, messageId, category, user.UserId, user.FullName)
	}
	switch category {
	case MessageCategoryPlainText,
		MessageCategoryPlainImage,
		MessageCategoryPlainVideo,
		MessageCategoryPlainLive,
		MessageCategoryPlainData,
		MessageCategoryPlainSticker,
		MessageCategoryPlainContact,
		MessageCategoryPlainAudio,
		MessageCategoryPlainPost,
		MessageCategoryPlainTranscript,
		MessageCategoryEncryptedPost,
		MessageCategoryEncryptedText,
		MessageCategoryEncryptedImage,
		MessageCategoryEncryptedVideo,
		MessageCategoryEncryptedLive,
		MessageCategoryEncryptedAudio,
		MessageCategoryEncryptedData,
		MessageCategoryEncryptedSticker,
		MessageCategoryEncryptedContact,
		MessageCategoryEncryptedLocation,
		MessageCategoryEncryptedTranscript,
		MessageCategoryAppCard,
		MessageCategoryAppButtonGroup,
		MessageCategoryMessageRecall:
	default:
		return nil, nil
	}
	if !user.isAdmin() && user.UserId != config.AppConfig.Mixin.ClientId {
		b, err := ReadProhibitedProperty(ctx)
		if err != nil {
			return nil, err
		} else if b {
			return nil, nil
		}
		system := config.AppConfig.System
		switch category {
		case MessageCategoryPlainImage, MessageCategoryEncryptedImage:
			if !system.ImageMessageEnable {
				return nil, nil
			}
		case MessageCategoryPlainVideo, MessageCategoryEncryptedVideo:
			if !system.VideoMessageEnable {
				return nil, nil
			}
		case MessageCategoryPlainLive, MessageCategoryEncryptedLive:
			if !system.LiveMessageEnable {
				return nil, nil
			}
		case MessageCategoryPlainContact, MessageCategoryEncryptedContact:
			if !system.ContactMessageEnable {
				return nil, nil
			}
		case MessageCategoryPlainAudio, MessageCategoryEncryptedAudio:
			if !system.AudioMessageEnable {
				return nil, nil
			}
		}

		switch category {
		case MessageCategoryMessageRecall:
		default:
			if !durable.Allow(user.UserId) {
				text := base64.RawURLEncoding.EncodeToString([]byte(config.AppConfig.MessageTemplate.MessageTipsTooMany))
				err = CreateSystemDistributedMessage(ctx, user, MessageCategoryPlainText, text)
				return nil, err
			}
		}
	}

	switch category {
	case MessageCategoryEncryptedPost,
		MessageCategoryEncryptedText,
		MessageCategoryEncryptedImage,
		MessageCategoryEncryptedVideo,
		MessageCategoryEncryptedLive,
		MessageCategoryEncryptedAudio,
		MessageCategoryEncryptedData,
		MessageCategoryEncryptedSticker,
		MessageCategoryEncryptedContact,
		MessageCategoryPlainTranscript,
		MessageCategoryEncryptedLocation:
		var err error
		data, err = decryptMessageData(data)
		if err != nil || data == "" {
			return nil, err
		}
	}

	if user.isAdmin() && quoteMessageId != "" {
		switch category {
		case MessageCategoryPlainText, MessageCategoryEncryptedText:
			if id, _ := bot.UuidFromString(quoteMessageId); id.String() == quoteMessageId {
				bytes, err := base64.RawURLEncoding.DecodeString(data)
				if err != nil {
					return nil, err
				}
				upper := strings.ToUpper(strings.TrimSpace(string(bytes)))
				switch upper {
				case "BAN", "KICK", "DELETE", "REMOVE":
					dm, err := FindDistributedMessage(ctx, quoteMessageId)
					if err != nil || dm == nil {
						return nil, err
					}
					if upper == "BAN" {
						_, err = user.CreateBlacklist(ctx, dm.UserId)
						if err != nil {
							return nil, err
						}
					}
					if upper == "KICK" {
						err = user.DeleteUser(ctx, dm.UserId)
						if err != nil {
							return nil, err
						}
					}
					quoteMessageId = ""
					category = MessageCategoryMessageRecall
					data = base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"message_id":"%s"}`, dm.ParentId)))
				}
			}
		}
	}

	message := &Message{
		MessageId:        messageId,
		UserId:           user.UserId,
		Category:         category,
		Data:             data,
		Silent:           silent,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
		State:            MessageStatePending,
		LastDistributeAt: genesisStartedAt(),
	}

	if quoteMessageId != "" {
		if id, _ := uuid.FromString(quoteMessageId); id.String() == quoteMessageId {
			message.QuoteMessageId = quoteMessageId
			dm, err := FindDistributedMessage(ctx, quoteMessageId)
			if err != nil {
				return nil, err
			}
			if dm != nil {
				message.QuoteMessageId = dm.ParentId
			}
		}
	}
	if category == MessageCategoryMessageRecall {
		bytes, err := base64.RawURLEncoding.DecodeString(data)
		if err != nil {
			return nil, session.BadDataError(ctx)
		}
		var recallMessage RecallMessage
		err = json.Unmarshal(bytes, &recallMessage)
		if err != nil {
			return nil, session.BadDataError(ctx)
		}
		m, err := FindMessage(ctx, recallMessage.MessageId)
		if err != nil || m == nil {
			return nil, err
		}
		if m.UserId != user.UserId && !user.isAdmin() {
			return nil, nil
		}
		if user.isAdmin() {
			message.UserId = m.UserId
		}
	}
	query := durable.PrepareQuery("INSERT INTO messages (%s) VALUES (%s) ON CONFLICT (message_id) DO NOTHING", messagesCols)
	_, err := session.Database(ctx).ExecContext(ctx, query, message.values()...)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return message, nil
}

func createSystemRewardMessage(ctx context.Context, tx *sql.Tx, r *Reward, user, receipt *User, asset *Asset) error {
	label := fmt.Sprintf(config.AppConfig.MessageTemplate.MessageRewardLabel, user.FullName, receipt.FullName, r.Amount, asset.Symbol)
	if utf8.RuneCountInString(label) > 36 {
		label = fmt.Sprintf(config.AppConfig.MessageTemplate.MessageRewardLabel, FirstNStringInRune(user.FullName, 5), FirstNStringInRune(receipt.FullName, 5), r.Amount, asset.Symbol)
	}
	if utf8.RuneCountInString(label) > 36 {
		label = fmt.Sprintf(FirstNStringInRune(label, 30))
	}
	action := config.AppConfig.Service.HTTPResourceHost + "/broadcasters"
	colors := []string{"#AA4848", "#B0665E", "#EF8A44", "#A09555", "#727234", "#9CAD23", "#AA9100", "#C49B4B", "#A47758", "#DF694C", "#D65859", "#C2405A", "#A75C96", "#BD637C", "#8F7AC5", "#7983C2", "#728DB8", "#5977C2", "#5E6DA2", "#3D98D0", "#5E97A1"}
	btns, err := json.Marshal([]interface{}{map[string]string{
		"label":  label,
		"action": action,
		"color":  colors[rand.Intn(len(colors))],
	}})
	if err != nil {
		return session.ServerError(ctx, err)
	}
	data := base64.RawURLEncoding.EncodeToString(btns)
	return createSystemMessage(ctx, tx, MessageCategoryAppButtonGroup, data)
}

func createSystemJoinMessage(ctx context.Context, tx *sql.Tx, user *User) error {
	data := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(config.AppConfig.MessageTemplate.MessageTipsJoin, user.FullName)))
	return createSystemMessage(ctx, tx, MessageCategoryPlainText, data)
}

func createSystemMessage(ctx context.Context, tx *sql.Tx, category, data string) error {
	mixin := config.AppConfig.Mixin
	t := time.Now()
	message := &Message{
		MessageId:        bot.UuidNewV4().String(),
		UserId:           mixin.ClientId,
		Category:         category,
		Data:             data,
		CreatedAt:        t,
		UpdatedAt:        t,
		State:            MessageStatePending,
		LastDistributeAt: genesisStartedAt(),
	}
	query := durable.PrepareQuery("INSERT INTO messages (%s) VALUES (%s) ON CONFLICT (message_id) DO NOTHING", messagesCols)
	_, err := tx.ExecContext(ctx, query, message.values()...)
	return err
}

func PendingMessages(ctx context.Context, limit int64) ([]*Message, error) {
	var messages []*Message
	query := fmt.Sprintf("SELECT %s FROM messages WHERE state=$1 ORDER BY state,updated_at LIMIT $2", strings.Join(messagesCols, ","))
	rows, err := session.Database(ctx).QueryContext(ctx, query, MessageStatePending, limit)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	for rows.Next() {
		m, err := messageFromRow(rows)
		if err != nil {
			return nil, session.TransactionError(ctx, err)
		}
		messages = append(messages, m)
	}
	return messages, nil
}

func LoopClearUpSuccessMessages(ctx context.Context) (int64, error) {
	query := fmt.Sprintf("DELETE FROM messages WHERE message_id IN (SELECT message_id FROM messages WHERE state='success' AND updated_at<$1 LIMIT 100)")
	r, err := session.Database(ctx).ExecContext(ctx, query, time.Now().Add(-365*24*time.Hour))
	if err != nil {
		return 0, session.ServerError(ctx, err)
	}
	return r.RowsAffected()
}

func FindMessage(ctx context.Context, id string) (*Message, error) {
	query := fmt.Sprintf("SELECT %s FROM messages WHERE message_id=$1", strings.Join(messagesCols, ","))
	row := session.Database(ctx).QueryRowContext(ctx, query, id)
	message, err := messageFromRow(row)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return message, nil
}

func LatestMessageWithUser(ctx context.Context, limit int64) ([]*Message, error) {
	query := "SELECT messages.message_id,messages.category,messages.data,messages.created_at,users.full_name FROM messages LEFT JOIN users ON messages.user_id=users.user_id ORDER BY updated_at DESC LIMIT $1"
	rows, err := session.Database(ctx).QueryContext(ctx, query, limit)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var m Message
		err := rows.Scan(&m.MessageId, &m.Category, &m.Data, &m.CreatedAt, &m.FullName)
		if err != nil {
			return nil, session.TransactionError(ctx, err)
		}
		if m.Category == MessageCategoryPlainText {
			data, _ := base64.RawURLEncoding.DecodeString(m.Data)
			m.Data = string(data)
		} else {
			m.Data = ""
		}
		messages = append(messages, &m)
	}
	return messages, nil
}

func readLatestMessages(ctx context.Context, limit int64) ([]*Message, error) {
	var messages []*Message
	query := fmt.Sprintf("SELECT %s FROM messages WHERE state=$1 ORDER BY updated_at DESC LIMIT $2", strings.Join(messagesCols, ","))
	rows, err := session.Database(ctx).QueryContext(ctx, query, MessageStateSuccess, limit)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	defer rows.Close()

	for rows.Next() {
		m, err := messageFromRow(rows)
		if err != nil {
			return nil, session.TransactionError(ctx, err)
		}
		messages = append(messages, m)
	}
	return messages, nil
}

func readLatestMessagesInTx(ctx context.Context, tx *sql.Tx, userId string, limit int64) ([]*Message, error) {
	var messages []*Message
	query := fmt.Sprintf("SELECT %s FROM messages WHERE state=$1 ORDER BY updated_at DESC LIMIT $2", strings.Join(messagesCols, ","))
	rows, err := tx.QueryContext(ctx, query, MessageStateSuccess, limit)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	defer rows.Close()

	for rows.Next() {
		m, err := messageFromRow(rows)
		if err != nil {
			return nil, session.TransactionError(ctx, err)
		}
		if m.UserId == userId {
			continue
		}
		messages = append(messages, m)
	}
	return messages, nil
}

type RecallMessage struct {
	MessageId string `json:"message_id"`
}

func FirstNStringInRune(s string, n int) string {
	if utf8.RuneCountInString(s) <= n+3 {
		return s
	}
	return string([]rune(s)[:n]) + "..."
}

func decryptMessageData(data string) (string, error) {
	bytes, err := base64.RawURLEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	size := 16 + 48 // session id bytes + encypted key bytes size
	total := len(bytes)
	if total < 1+2+32+size+12 {
		return "", nil
	}
	sessionLen := int(binary.LittleEndian.Uint16(bytes[1:3]))
	mixin := config.AppConfig.Mixin
	prefixSize := 35 + sessionLen*size
	var key []byte
	for i := 35; i < prefixSize; i += size {
		if uid, _ := bot.UuidFromBytes(bytes[i : i+16]); uid.String() == mixin.SessionId {
			private, err := base64.RawURLEncoding.DecodeString(mixin.SessionKey)
			if err != nil {
				return "", err
			}
			var dst, priv, pub [32]byte
			copy(pub[:], bytes[3:35])
			bot.PrivateKeyToCurve25519(&priv, ed25519.PrivateKey(private))
			curve25519.ScalarMult(&dst, &priv, &pub)

			block, err := aes.NewCipher(dst[:])
			if err != nil {
				return "", err
			}
			iv := bytes[i+16 : i+16+aes.BlockSize]
			key = bytes[i+16+aes.BlockSize : i+size]
			mode := cipher.NewCBCDecrypter(block, iv)
			mode.CryptBlocks(key, key)
			key = key[:16]
			break
		}
	}
	if len(key) != 16 {
		return "", nil
	}
	nonce := bytes[prefixSize : prefixSize+12]
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", nil // TODO
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", nil // TODO
	}
	plaintext, err := aesgcm.Open(nil, nonce, bytes[prefixSize+12:], nil)
	if err != nil {
		return "", nil // TODO
	}
	return base64.RawURLEncoding.EncodeToString(plaintext), nil
}

func EncryptMessageData(data string, sessions []*Session) (string, error) {
	dataBytes, err := base64.RawURLEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	key := make([]byte, 16)
	_, err = rand.Read(key)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, 12)
	_, err = rand.Read(nonce)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	ciphertext := aesgcm.Seal(nil, nonce, dataBytes, nil)

	var sessionLen [2]byte
	binary.LittleEndian.PutUint16(sessionLen[:], uint16(len(sessions)))

	mixin := config.AppConfig.Mixin
	privateBytes, err := base64.RawURLEncoding.DecodeString(mixin.SessionKey)
	if err != nil {
		return "", err
	}

	var pub [32]byte
	private := ed25519.PrivateKey(privateBytes)
	bot.PublicKeyToCurve25519(&pub, ed25519.PublicKey(private[32:]))

	var sessionsBytes []byte
	for _, s := range sessions {
		clientPublic, err := base64.RawURLEncoding.DecodeString(s.PublicKey)
		if err != nil {
			return "", err
		}
		var dst, priv, clientPub [32]byte
		copy(clientPub[:], clientPublic[:])
		bot.PrivateKeyToCurve25519(&priv, private)
		curve25519.ScalarMult(&dst, &priv, &clientPub)

		block, err := aes.NewCipher(dst[:])
		if err != nil {
			return "", err
		}
		padding := aes.BlockSize - len(key)%aes.BlockSize
		padtext := bytes.Repeat([]byte{byte(padding)}, padding)
		shared := make([]byte, len(key))
		copy(shared[:], key[:])
		shared = append(shared, padtext...)
		ciphertext := make([]byte, aes.BlockSize+len(shared))
		iv := ciphertext[:aes.BlockSize]
		_, err = rand.Read(iv)
		if err != nil {
			return "", err
		}
		mode := cipher.NewCBCEncrypter(block, iv)
		mode.CryptBlocks(ciphertext[aes.BlockSize:], shared)
		id, err := bot.UuidFromString(s.SessionID)
		if err != nil {
			return "", err
		}
		sessionsBytes = append(sessionsBytes, id.Bytes()...)
		sessionsBytes = append(sessionsBytes, ciphertext...)
	}

	result := []byte{1}
	result = append(result, sessionLen[:]...)
	result = append(result, pub[:]...)
	result = append(result, sessionsBytes...)
	result = append(result, nonce[:]...)
	result = append(result, ciphertext...)
	return base64.RawURLEncoding.EncodeToString(result), nil
}
