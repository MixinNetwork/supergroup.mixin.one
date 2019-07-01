package models

import (
	"context"
)

func ReadStatistic(ctx context.Context, user *User) (map[string]interface{}, error) {
	s := make(map[string]interface{}, 0)
	count, err := PaidMemberCount(ctx)
	if err != nil {
		return nil, err
	}
	s["users_count"] = count
	s["prohibited"] = false
	if user != nil && user.isAdmin() {
		p, err := ReadProperty(ctx, ProhibitedMessage)
		if err != nil {
			return nil, err
		}
		if p != nil && p.Value == "true" {
			s["prohibited"] = true
		}
	}
	return s, nil
}
