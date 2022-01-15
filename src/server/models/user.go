package models

import (
	"sort"

	"github.com/jameycribbs/hare"
)

type User struct {
	ID           int          `json:"id"`
	Username     string       `json:"username"`
	PasswordHash string       `json:"password_hash"`
	Friends      map[int]bool `json:"friends"`
}

func (u *User) GetID() int {
	return u.ID
}

func (u *User) SetID(id int) {
	u.ID = id
}

func (u *User) AfterFind(db *hare.Database) error {
	*u = User(*u)
	return nil
}

func QueryUsers(db *hare.Database, queryFn func(u User) bool, limit int) ([]User, error) {
	var results []User
	var err error

	ids, err := db.IDs("users")
	if err != nil {
		return nil, err
	}

	for _, id := range ids {
		u := User{}
		if err := db.Find("users", id, &u); err != nil {
			return nil, err
		}

		if queryFn(u) {
			results = append(results, u)
		}

		if limit != 0 && len(results) == limit {
			break
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].ID < results[j].ID
	})

	return results, err
}
