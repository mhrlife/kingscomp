package entity

type UserState struct {
	IsReady    bool `json:"isReady"`
	IsResigned bool `json:"isResigned"`

	CurrentQuestionIndex    int  `json:"currentQuestionIndex"`
	CurrentQuestionAnswered bool `json:"currentQuestionAnswered"`
}

type Lobby struct {
	ID            string  `json:"id"`
	Participants  []int64 `json:"participants"`
	CreatedAtUnix int64   `json:"created_at"`

	Questions []Question `json:"questions"`

	UserState map[int64]UserState `json:"userState"`

	State string `json:"state"`
}

func (l Lobby) EntityID() ID {
	return NewID("lobby", l.ID)
}

type Question struct {
	ID            string   `json:"id"`
	Question      string   `json:"question"`
	Answers       []string `json:"answers"`
	CorrectAnswer int      `json:"correctAnswer"`
}

func (q Question) EntityID() ID {
	return NewID("question", q.ID)
}
