package entity

import "time"

type Account struct {
	ID        int64     `json:"id"`
	FirstName string    `json:"first_name"`
	Username  string    `json:"username"`
	JoinedAt  time.Time `json:"joined_at"`

	DisplayName string `json:"display_name"`
	State       string `json:"state"`

	CurrentLobby string `json:"current_lobby"`
	InQueue      bool   `json:"in_queue"`
}

func (a Account) EntityID() ID {
	return NewID("account", a.ID)
}
