package models

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/spanner"
	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	jwt "github.com/dgrijalva/jwt-go"
	"google.golang.org/api/iterator"
)

const (
	PaymentStatePending = "pending"
	PaymentStatePaid    = "paid"
)

const users_DDL = `
CREATE TABLE users (
	user_id	          STRING(36) NOT NULL,
	identity_number   INT64 NOT NULL,
	full_name         STRING(512) NOT NULL,
	access_token      STRING(512) NOT NULL,
	avatar_url        STRING(1024) NOT NULL,
	trace_id          STRING(36) NOT NULL,
	state             STRING(128) NOT NULL,
	subscribed_at     TIMESTAMP NOT NULL,
) PRIMARY KEY(user_id);

CREATE INDEX users_by_subscribed ON users(subscribed_at) STORING(full_name);
`

type User struct {
	UserId         string
	IdentityNumber int64
	FullName       string
	AccessToken    string
	AvatarURL      string
	TraceId        string
	State          string
	SubscribedAt   time.Time

	AuthenticationToken string
}

var usersCols = []string{"user_id", "identity_number", "full_name", "access_token", "avatar_url", "trace_id", "state", "subscribed_at"}

func (u *User) values() []interface{} {
	return []interface{}{u.UserId, u.IdentityNumber, u.FullName, u.AccessToken, u.AvatarURL, u.TraceId, u.State, u.SubscribedAt}
}

func AuthenticateUserByOAuth(ctx context.Context, authorizationCode string) (*User, error) {
	accessToken, scope, err := bot.OAuthGetAccessToken(ctx, config.ClientId, config.ClientSecret, authorizationCode, "")
	if err != nil {
		return nil, session.ServerError(ctx, err)
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
		user, queryErr = findUserById(ctx, fmt.Sprint(claims["jti"]))
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
	return user, nil
}

func createUser(ctx context.Context, accessToken, userId, identityNumber, fullName, avatarURL string) (*User, error) {
	id, err := bot.UuidFromString(userId)
	if err != nil {
		return nil, session.ForbiddenError(ctx)
	}
	if avatarURL == "" {
		avatarURL = "https://images.mixin.one/E2y0BnTopFK9qey0YI-8xV3M82kudNnTaGw0U5SU065864SsewNUo6fe9kDF1HIzVYhXqzws4lBZnLj1lPsjk-0=s128"
	}
	num, _ := strconv.ParseInt(identityNumber, 10, 64)
	authenticationToken, err := generateAuthenticationToken(ctx, id.String(), accessToken)
	if err != nil {
		return nil, session.ServerError(ctx, err)
	}
	user, err := findUserById(ctx, userId)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	if user == nil {
		user = &User{
			UserId:         userId,
			IdentityNumber: num,
			TraceId:        bot.UuidNewV4().String(),
			FullName:       fullName,
			State:          PaymentStatePending,
		}
	}
	user.AuthenticationToken = authenticationToken
	user.AccessToken = accessToken
	user.AvatarURL = avatarURL
	if err := session.Database(ctx).Apply(ctx, []*spanner.Mutation{
		spanner.InsertOrUpdate("users", usersCols, user.values()),
	}, "users", "INSERT", "createUser"); err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return user, nil
}

func (user *User) UpdateProfile(ctx context.Context, fullName string) error {
	fullName = strings.TrimSpace(fullName)
	if fullName == "" {
		return nil
	}
	user.FullName = fullName
	if err := session.Database(ctx).Apply(ctx, []*spanner.Mutation{
		spanner.Update("users", []string{"user_id", "full_name"}, []interface{}{user.UserId, user.FullName}),
	}, "users", "UPDATE", "UpdateProfile"); err != nil {
		return session.TransactionError(ctx, err)
	}
	return nil
}

func (user *User) Subscribe(ctx context.Context) error {
	if !user.SubscribedAt.IsZero() {
		return nil
	}
	user.SubscribedAt = time.Now()
	if err := session.Database(ctx).Apply(ctx, []*spanner.Mutation{
		spanner.Update("users", []string{"user_id", "subscribed_at"}, []interface{}{user.UserId, user.SubscribedAt}),
	}, "users", "UPDATE", "Subscribe"); err != nil {
		return session.TransactionError(ctx, err)
	}
	return nil
}

func (user *User) Unsubscribe(ctx context.Context) error {
	if user.SubscribedAt.IsZero() {
		return nil
	}
	user.SubscribedAt = time.Time{}
	if err := session.Database(ctx).Apply(ctx, []*spanner.Mutation{
		spanner.Update("users", []string{"user_id", "subscribed_at"}, []interface{}{user.UserId, user.SubscribedAt}),
	}, "users", "UPDATE", "Subscribe"); err != nil {
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
	t := time.Now()
	message := &Message{
		MessageId: bot.UuidNewV4().String(),
		UserId:    config.ClientId,
		Category:  "PLAIN_TEXT",
		Data:      []byte(fmt.Sprintf(config.MessageTipsJoin, user.FullName)),
		CreatedAt: t,
		UpdatedAt: t,
		State:     MessageStatePending,
	}
	user.State, user.SubscribedAt = PaymentStatePaid, time.Now()
	if err := session.Database(ctx).Apply(ctx, []*spanner.Mutation{
		spanner.Insert("messages", messagesCols, message.values()),
		spanner.Update("users", []string{"user_id", "state", "subscribed_at"}, []interface{}{user.UserId, user.State, user.SubscribedAt}),
	}, "users", "UPDATE", "Payment"); err != nil {
		return session.TransactionError(ctx, err)
	}
	return nil
}

func Subscribers(ctx context.Context, offset time.Time, num int64) ([]*User, error) {
	if num > 20000 {
		user, err := findUserByIdentityNumber(ctx, num)
		if err != nil {
			return nil, session.TransactionError(ctx, err)
		} else if user == nil {
			return nil, nil
		}
		return []*User{user}, nil
	}
	ids, _, err := subscribedUserIds(ctx, offset, 200)
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return nil, nil
	}
	stmt := spanner.Statement{
		SQL:    fmt.Sprintf("SELECT %s FROM users WHERE user_id IN UNNEST(@user_ids)", strings.Join(usersCols, ",")),
		Params: map[string]interface{}{"user_ids": ids},
	}
	it := session.Database(ctx).Query(ctx, stmt, "users", "Subscribers")
	defer it.Stop()

	var users []*User
	for {
		row, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return users, session.TransactionError(ctx, err)
		}
		user, err := userFromRow(row)
		if err != nil {
			return users, session.TransactionError(ctx, err)
		}
		users = append(users, user)
	}
	sort.Slice(users, func(i, j int) bool { return users[i].SubscribedAt.Before(users[j].SubscribedAt) })
	return users, nil
}

func SubscribersCount(ctx context.Context) (int64, error) {
	stmt := spanner.Statement{
		SQL:    "SELECT COUNT(*) AS count FROM users@{FORCE_INDEX=users_by_subscribed} WHERE subscribed_at>@subscribed_at",
		Params: map[string]interface{}{"subscribed_at": time.Time{}},
	}
	it := session.Database(ctx).Query(ctx, stmt, "users", "SubscribersCount")
	defer it.Stop()

	row, err := it.Next()
	if err == iterator.Done {
		return 0, nil
	} else if err != nil {
		return 0, session.TransactionError(ctx, err)
	}

	var count int64
	if err := row.Columns(&count); err != nil {
		return 0, session.TransactionError(ctx, err)
	}
	return count, nil
}

func (user *User) DeleteUser(ctx context.Context, id string) error {
	if !config.Operators[user.UserId] {
		return nil
	}

	if err := session.Database(ctx).Apply(ctx, []*spanner.Mutation{
		spanner.Delete("users", spanner.Key{id}),
	}, "users", "DELETE", "DeleteUser"); err != nil {
		return session.TransactionError(ctx, err)
	}
	return nil
}

func (user *User) GetRole() string {
	if config.Operators[user.UserId] {
		return "admin"
	}
	return "user"
}

func subscribedUserIds(ctx context.Context, subscribedAt time.Time, limit int) ([]string, time.Time, error) {
	var ids []string
	stmt := spanner.Statement{
		SQL:    fmt.Sprintf("SELECT user_id,subscribed_at FROM users@{FORCE_INDEX=users_by_subscribed} WHERE subscribed_at>@subscribed_at ORDER BY subscribed_at LIMIT %d", limit),
		Params: map[string]interface{}{"subscribed_at": subscribedAt},
	}
	it := session.Database(ctx).Query(ctx, stmt, "users", "SubscribedUsers")
	defer it.Stop()

	for {
		row, err := it.Next()
		if err == iterator.Done {
			return ids, subscribedAt, nil
		} else if err != nil {
			return ids, subscribedAt, session.TransactionError(ctx, err)
		}
		var id string
		if err := row.Columns(&id, &subscribedAt); err != nil {
			return ids, subscribedAt, session.TransactionError(ctx, err)
		}
		ids = append(ids, id)
	}
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
	return findUserById(ctx, userId)
}

func findUserById(ctx context.Context, userId string) (*User, error) {
	txn := session.Database(ctx).ReadOnlyTransaction()
	defer txn.Close()

	user, err := readUser(ctx, txn, userId)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return user, nil
}

func findUserByIdentityNumber(ctx context.Context, num int64) (*User, error) {
	txn := session.Database(ctx).ReadOnlyTransaction()
	defer txn.Close()

	stmt := spanner.Statement{
		SQL:    fmt.Sprintf("SELECT %s FROM users WHERE identity_number=@identity_number LIMIT 1", strings.Join(usersCols, ",")),
		Params: map[string]interface{}{"identity_number": num},
	}
	it := session.Database(ctx).Query(ctx, stmt, "users", "findUserByIdentityNumber")
	defer it.Stop()

	row, err := it.Next()
	if err == iterator.Done {
		return nil, nil
	} else if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	user, err := userFromRow(row)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return user, nil
}

func readUser(ctx context.Context, txn durable.Transaction, userId string) (*User, error) {
	it := txn.Read(ctx, "users", spanner.Key{userId}, usersCols)
	defer it.Stop()

	row, err := it.Next()
	if err == iterator.Done {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return userFromRow(row)
}

func userFromRow(row *spanner.Row) (*User, error) {
	var u User
	err := row.Columns(&u.UserId, &u.IdentityNumber, &u.FullName, &u.AccessToken, &u.AvatarURL, &u.TraceId, &u.State, &u.SubscribedAt)
	return &u, err
}

var nameColorSet = []string{"#AA4848", "#B0665E", "#EF8A44", "#A09555", "#727234", "#9CAD23", "#AA9100", "#C49B4B", "#A47758", "#DF694C", "#D65859", "#C2405A", "#A75C96", "#BD637C", "#8F7AC5", "#7983C2", "#728DB8", "#5977C2", "#5E6DA2", "#3D98D0", "#5E97A1", "#4EABAA", "#63A082", "#877C9B", "#AA66C3", "#BB5334", "#667355", "#668899", "#83BE44", "#BBA600", "#429AB6", "#75856F", "#88A299", "#B3798E", "#447899", "#D79200", "#728DB8", "#DD637C", "#887C66", "#BE6C2C", "#9B6D77", "#B69370", "#976236", "#9D77A5", "#8A660E", "#5E935E", "#9B8484", "#92B288"}
