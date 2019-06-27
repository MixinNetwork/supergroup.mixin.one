package models

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	jwt "github.com/dgrijalva/jwt-go"
)

const (
	PaymentStatePending = "pending"
	PaymentStatePaid    = "paid"

	UserActivePeriod = 5 * time.Minute
)

const users_DDL = `
CREATE TABLE IF NOT EXISTS users (
	user_id	          VARCHAR(36) PRIMARY KEY CHECK (user_id ~* '^[0-9a-f-]{36,36}$'),
	identity_number   BIGINT NOT NULL,
	full_name         VARCHAR(512) NOT NULL DEFAULT '',
	access_token      VARCHAR(512) NOT NULL DEFAULT '',
	avatar_url        VARCHAR(1024) NOT NULL DEFAULT '',
	trace_id          VARCHAR(36) NOT NULL CHECK (trace_id ~* '^[0-9a-f-]{36,36}$'),
	state             VARCHAR(128) NOT NULL,
	active_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	subscribed_at     TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS users_identityx ON users(identity_number);
CREATE INDEX IF NOT EXISTS users_subscribedx ON users(subscribed_at);
CREATE INDEX IF NOT EXISTS users_activex ON users(active_at);
`

type User struct {
	UserId         string
	IdentityNumber int64
	FullName       string
	AccessToken    string
	AvatarURL      string
	TraceId        string
	State          string
	ActiveAt       time.Time
	SubscribedAt   time.Time

	isNew               bool
	AuthenticationToken string
}

var usersCols = []string{"user_id", "identity_number", "full_name", "access_token", "avatar_url", "trace_id", "state", "active_at", "subscribed_at"}

func (u *User) values() []interface{} {
	return []interface{}{u.UserId, u.IdentityNumber, u.FullName, u.AccessToken, u.AvatarURL, u.TraceId, u.State, u.ActiveAt, u.SubscribedAt}
}

func AuthenticateUserByOAuth(ctx context.Context, authorizationCode string) (*User, error) {
	accessToken, scope, err := bot.OAuthGetAccessToken(ctx, config.Get().Mixin.ClientId, config.Get().Mixin.ClientSecret, authorizationCode, "")
	if err != nil {
		return nil, err
	}
	if !strings.Contains(scope, "PROFILE:READ") {
		return nil, session.ForbiddenError(ctx)
	}

	me, err := bot.UserMe(ctx, accessToken)
	if err != nil {
		return nil, session.ServerError(ctx, err)
	}
	return createUser(ctx, accessToken, me.UserId, me.IdentityNumber, me.FullName, me.AvatarURL)
}

func createUser(ctx context.Context, accessToken, userId, identityNumber, fullName, avatarURL string) (*User, error) {
	id, err := bot.UuidFromString(userId)
	if err != nil {
		return nil, session.ForbiddenError(ctx)
	}
	if avatarURL == "" {
		avatarURL = "https://images.mixin.one/E2y0BnTopFK9qey0YI-8xV3M82kudNnTaGw0U5SU065864SsewNUo6fe9kDF1HIzVYhXqzws4lBZnLj1lPsjk-0=s128"
	}
	identity, _ := strconv.ParseInt(identityNumber, 10, 64)
	authenticationToken, err := generateAuthenticationToken(ctx, id.String(), accessToken)
	if err != nil {
		return nil, session.ServerError(ctx, err)
	}
	user, err := FindUser(ctx, userId)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	if user == nil {
		user = &User{
			UserId:         userId,
			IdentityNumber: identity,
			TraceId:        bot.UuidNewV4().String(),
			State:          PaymentStatePending,
			ActiveAt:       time.Now(),
			isNew:          true,
		}
		if !config.Get().System.PayToJoin {
			item, err := readBlacklist(ctx, user.UserId)
			if err != nil {
				return nil, session.TransactionError(ctx, err)
			} else if item != nil {
				return nil, session.ForbiddenError(ctx)
			}
			user.State = PaymentStatePaid
			user.SubscribedAt = time.Now()
		}
		err = createConversation(ctx, "CONTACT", userId)
		if err != nil {
			return nil, session.ServerError(ctx, err)
		}
	}
	if strings.TrimSpace(fullName) != "" {
		user.FullName = fullName
	}
	user.AccessToken = accessToken
	user.AvatarURL = avatarURL
	user.AuthenticationToken = authenticationToken

	if user.isNew {
		err = session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
			if user.State == PaymentStatePaid {
				if err := createSystemJoinMessage(ctx, tx, user); err != nil {
					return err
				}
			}
			params, positions := compileTableQuery(usersCols)
			_, err := tx.ExecContext(ctx, fmt.Sprintf("INSERT INTO users (%s) VALUES (%s)", params, positions), user.values()...)
			return err
		})
		if err != nil {
			return nil, session.TransactionError(ctx, err)
		}
		return user, nil
	}

	params, positions := compileTableQuery([]string{"full_name", "access_token", "avatar_url"})
	_, err = session.Database(ctx).Exec(fmt.Sprintf("UPDATE users SET (%s)=(%s) WHERE user_id='%s'", params, positions, user.UserId), user.FullName, user.AccessToken, user.AvatarURL)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return user, nil
}

func createConversation(ctx context.Context, category, participantId string) error {
	if config.Get().Service.Environment == "test" {
		return nil
	}
	conversationId := bot.UniqueConversationId(config.Get().Mixin.ClientId, participantId)
	participant := bot.Participant{
		UserId: participantId,
		Role:   "",
	}
	participants := []bot.Participant{
		participant,
	}
	_, err := bot.CreateConversation(ctx, category, conversationId, participants, config.Get().Mixin.ClientId, config.Get().Mixin.SessionId, config.Get().Mixin.SessionKey)
	return err
}

func AuthenticateUserByToken(ctx context.Context, authenticationToken string) (*User, error) {
	var user *User = nil
	var queryErr error = nil
	token, err := jwt.Parse(authenticationToken, func(token *jwt.Token) (interface{}, error) {
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, session.BadDataError(ctx)
		}
		_, ok = token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, session.BadDataError(ctx)
		}
		user, queryErr = FindUser(ctx, fmt.Sprint(claims["jti"]))
		if queryErr != nil {
			return nil, queryErr
		}
		if user == nil {
			return nil, nil
		}
		sum := sha256.Sum256([]byte(user.AccessToken))
		return sum[:], nil
	})

	if queryErr != nil {
		return nil, queryErr
	}
	if err != nil || !token.Valid {
		return nil, nil
	}
	if user.ActiveAt.Before(time.Now().Add(-1 * UserActivePeriod)) {
		err = PingUserActiveAt(ctx, user.UserId)
		if err != nil {
			session.Logger(ctx).Error("handleMessage PingUserActiveAt", err)
		}
	}
	return user, nil
}

func (user *User) UpdateProfile(ctx context.Context, fullName string) error {
	fullName = strings.TrimSpace(fullName)
	if fullName == "" {
		return nil
	}
	user.FullName = fullName
	query := "UPDATE users SET full_name=$1 WHERE user_id=$2"
	if _, err := session.Database(ctx).ExecContext(ctx, query, user.FullName, user.UserId); err != nil {
		return session.TransactionError(ctx, err)
	}
	return nil
}

func (user *User) Subscribe(ctx context.Context) error {
	if user.SubscribedAt.After(genesisStartedAt()) {
		return nil
	}
	user.SubscribedAt = time.Now()
	query := "UPDATE users SET subscribed_at=$1 WHERE user_id=$2"
	if _, err := session.Database(ctx).ExecContext(ctx, query, user.SubscribedAt, user.UserId); err != nil {
		return session.TransactionError(ctx, err)
	}
	return nil
}

func (user *User) Unsubscribe(ctx context.Context) error {
	if user.SubscribedAt.Before(genesisStartedAt()) || user.SubscribedAt.Equal(genesisStartedAt()) {
		return nil
	}
	user.SubscribedAt = time.Time{}
	query := "UPDATE users SET subscribed_at=$1 WHERE user_id=$2"
	if _, err := session.Database(ctx).ExecContext(ctx, query, user.SubscribedAt, user.UserId); err != nil {
		return session.TransactionError(ctx, err)
	}
	return nil
}

func (user *User) Payment(ctx context.Context) error {
	if user.State != PaymentStatePending {
		return nil
	}
	item, err := readBlacklist(ctx, user.UserId)
	if err != nil {
		return err
	} else if item != nil {
		return nil
	}

	user.State, user.SubscribedAt = PaymentStatePaid, time.Now()
	messages, err := readLastestMessages(ctx, 10)
	if err != nil {
		return err
	}

	sort.Slice(messages, func(i, j int) bool { return messages[i].CreatedAt.Before(messages[j].CreatedAt) })
	var values bytes.Buffer
	for i, msg := range messages {
		if msg.Category == MessageCategoryMessageRecall {
			var recallMessage RecallMessage
			data, err := base64.StdEncoding.DecodeString(msg.Data)
			if err != nil {
				return session.BadDataError(ctx)
			}
			err = json.Unmarshal(data, &recallMessage)
			if err != nil {
				return session.BadDataError(ctx)
			}

			r := RecallMessage{
				MessageId: UniqueConversationId(user.UserId, recallMessage.MessageId),
			}
			data, err = json.Marshal(r)
			if err != nil {
				return session.BadDataError(ctx)
			}
			msg.Data = base64.StdEncoding.EncodeToString(data)
		}

		messageId := UniqueConversationId(user.UserId, msg.MessageId)
		dm, err := createDistributeMessage(ctx, messageId, msg.MessageId, "", msg.UserId, user.UserId, msg.Category, msg.Data)
		if err != nil {
			session.TransactionError(ctx, err)
		}
		if i > 0 {
			values.WriteString(",")
		}
		values.WriteString(distributedMessageValuesString(dm.MessageId, dm.ConversationId, dm.RecipientId, dm.UserId, dm.ParentId, dm.QuoteMessageId, dm.Shard, dm.Category, dm.Data, dm.Status))
	}
	err = session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		v := values.String()
		if v != "" {
			dquery := fmt.Sprintf("INSERT INTO distributed_messages (%s) VALUES %s ON CONFLICT (message_id) DO NOTHING", strings.Join(distributedMessagesCols, ","), values.String())
			_, err := tx.ExecContext(ctx, dquery)
			if err != nil {
				return err
			}
		}
		if err := createSystemJoinMessage(ctx, tx, user); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, "UPDATE users SET (state,subscribed_at)=($1,$2) WHERE user_id=$3", user.State, user.SubscribedAt, user.UserId)
		return err
	})
	if err != nil {
		return session.TransactionError(ctx, err)
	}
	return nil
}

func Subscribers(ctx context.Context, offset time.Time, identity int64, keywords string) ([]*User, error) {
	if identity > 20000 {
		return findUsersByIdentityNumber(ctx, identity)
	}
	if len(keywords) != 0 {
		return findUsersByKeywords(ctx, keywords)
	}
	users, err := subscribedUsers(ctx, offset, 200)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	sort.Slice(users, func(i, j int) bool { return users[i].SubscribedAt.Before(users[j].SubscribedAt) })
	return users, nil
}

func SubscribersCount(ctx context.Context) (int64, error) {
	query := "SELECT COUNT(*) FROM users WHERE subscribed_at>$1"
	var count int64
	err := session.Database(ctx).QueryRowContext(ctx, query, genesisStartedAt()).Scan(&count)
	if err != nil {
		return 0, session.TransactionError(ctx, err)
	}
	return count, nil
}

func (user *User) DeleteUser(ctx context.Context, id string) error {
	if !config.Get().System.Operators[user.UserId] {
		return nil
	}
	_, err := session.Database(ctx).ExecContext(ctx, fmt.Sprintf("DELETE FROM users WHERE user_id=$1"), id)
	if err != nil {
		return session.TransactionError(ctx, err)
	}
	return nil
}

func (user *User) GetRole() string {
	if config.Get().System.Operators[user.UserId] {
		return "admin"
	}
	return "user"
}

func (user *User) isAdmin() bool {
	if config.Get().System.Operators[user.UserId] {
		return true
	}
	return false
}

func subscribedUsers(ctx context.Context, subscribedAt time.Time, limit int) ([]*User, error) {
	var users []*User
	query := fmt.Sprintf("SELECT %s FROM users WHERE subscribed_at>$1 ORDER BY subscribed_at LIMIT %d", strings.Join(usersCols, ","), limit)
	rows, err := session.Database(ctx).QueryContext(ctx, query, subscribedAt)
	if err != nil {
		return users, session.TransactionError(ctx, err)
	}
	defer rows.Close()

	for rows.Next() {
		u, err := userFromRow(rows)
		if err != nil {
			return users, session.TransactionError(ctx, err)
		}
		users = append(users, u)
	}
	return users, nil
}

func generateAuthenticationToken(ctx context.Context, userId, accessToken string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Id:        userId,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 365).Unix(),
	})
	sum := sha256.Sum256([]byte(accessToken))
	return token.SignedString(sum[:])
}

func FindUser(ctx context.Context, userId string) (*User, error) {
	var user *User
	err := session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		user, err = findUserById(ctx, tx, userId)
		return err
	})
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return user, nil
}

func PingUserActiveAt(ctx context.Context, userId string) error {
	query := "UPDATE users SET active_at=$1 WHERE user_id=$2"
	_, err := session.Database(ctx).ExecContext(ctx, query, time.Now(), userId)
	if err != nil {
		return session.TransactionError(ctx, err)
	}
	return nil
}

func findUsersByIdentityNumber(ctx context.Context, identity int64) ([]*User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE identity_number=$1", strings.Join(usersCols, ","))
	row := session.Database(ctx).QueryRowContext(ctx, query, identity)
	user, err := userFromRow(row)
	if err == sql.ErrNoRows {
		return []*User{}, nil
	} else if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return []*User{user}, nil
}

func findUsersByKeywords(ctx context.Context, keywords string) ([]*User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE LOWER(full_name) LIKE LOWER($1)", strings.Join(usersCols, ","))
	rows, err := session.Database(ctx).QueryContext(ctx, query, fmt.Sprintf("%%%s%%", keywords))
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		p, err := userFromRow(rows)
		if err != nil {
			return nil, session.TransactionError(ctx, err)
		}
		users = append(users, p)
	}
	return users, nil
}

func findUserById(ctx context.Context, tx *sql.Tx, userId string) (*User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE user_id=$1", strings.Join(usersCols, ","))
	row := tx.QueryRowContext(ctx, query, userId)
	user, err := userFromRow(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func userFromRow(row durable.Row) (*User, error) {
	var u User
	err := row.Scan(&u.UserId, &u.IdentityNumber, &u.FullName, &u.AccessToken, &u.AvatarURL, &u.TraceId, &u.State, &u.ActiveAt, &u.SubscribedAt)
	return &u, err
}

func genesisStartedAt() time.Time {
	startedAt, _ := time.Parse(time.RFC3339, "2017-01-01T00:00:00Z")
	return startedAt
}

func compileTableQuery(fields []string) (string, string) {
	var params, positions bytes.Buffer
	for i, f := range fields {
		if i != 0 {
			params.WriteString(",")
			positions.WriteString(",")
		}
		params.WriteString(f)
		positions.WriteString(fmt.Sprintf("$%d", i+1))
	}
	return params.String(), positions.String()
}

func (u *User) GetFullName() string {
	if u.FullName != "" {
		return u.FullName
	}
	return "Null"
}
