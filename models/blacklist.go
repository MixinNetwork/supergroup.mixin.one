package models

import (
	"context"

	"cloud.google.com/go/spanner"
	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"google.golang.org/api/iterator"
)

const blacklist_DDL = `
CREATE TABLE blacklists (
	user_id	          STRING(36) NOT NULL,
) PRIMARY KEY(user_id);
`

type Blacklist struct {
	UserId string
}

func (user *User) CreateBlacklist(ctx context.Context, userId string) (*Blacklist, error) {
	_, err := bot.UuidFromString(userId)
	if err != nil {
		return nil, session.ForbiddenError(ctx)
	}
	if !operators[user.UserId] {
		return nil, nil
	}
	user, err = findUserById(ctx, userId)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	} else if user == nil {
		return nil, nil
	}

	session.Database(ctx).Apply(ctx, []*spanner.Mutation{
		spanner.Delete("users", spanner.Key{userId}),
		spanner.Insert("blacklists", []string{"user_id"}, []interface{}{userId}),
	}, "blacklists", "INSERT", "CreateBlacklist")

	return &Blacklist{UserId: userId}, nil
}

func readBlacklist(ctx context.Context, userId string) (*Blacklist, error) {
	it := session.Database(ctx).Read(ctx, "blacklists", spanner.Key{userId}, []string{"user_id"}, "readBlacklist")
	defer it.Stop()

	row, err := it.Next()
	if err == iterator.Done {
		return nil, nil
	} else if err != nil {
		return nil, session.TransactionError(ctx, err)
	}

	var b Blacklist
	if err := row.Columns(&b.UserId); err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return &b, nil
}
