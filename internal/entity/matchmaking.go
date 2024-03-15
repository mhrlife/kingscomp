package entity

type Lobby struct {
	ID            string  `json:"id"`
	Participants  []int64 `json:"participants"`
	CreatedAtUnix int64   `json:"created_at"`

	State string `json:"state"`
}

func (l Lobby) EntityID() ID {
	return NewID("lobby", l.ID)
}
