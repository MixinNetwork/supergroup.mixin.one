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
		b, err := ReadProhibitedProperty(ctx)
		if err != nil {
			return nil, err
		}
		s["prohibited"] = b
	}
	return s, nil
}
