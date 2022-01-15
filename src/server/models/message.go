package models

import (
	"sort"
	"time"

	"github.com/jameycribbs/hare"
)

type Message struct {
	ID        int       `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Content   []byte    `json:"content"`
	Type      string    `json:"type"`
	Filename  string    `json:"filename"`
	Timestamp time.Time `json:"timestamp"`
}

func (m *Message) GetID() int {
	return m.ID
}

func (m *Message) SetID(id int) {
	m.ID = id
}

func (m *Message) AfterFind(db *hare.Database) error {
	*m = Message(*m)
	return nil
}

func QueryMessages(db *hare.Database, queryFn func(m Message) bool, limit int) ([]Message, error) {
	var results []Message
	var err error

	ids, err := db.IDs("messages")
	if err != nil {
		return nil, err
	}

	for _, id := range ids {
		m := Message{}
		if err := db.Find("messages", id, &m); err != nil {
			return nil, err
		}

		if queryFn(m) {
			results = append(results, m)
		}

		if limit != 0 && len(results) == limit {
			break
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.Before(results[j].Timestamp)
	})

	return results, err
}
