package gameserver

import "time"

var (
	DefaultReminderToReadyAfter = time.Second * 30
	DefaultReadyDeadline        = time.Second * 45
	DefaultQuestionTimeout      = time.Second * 10
	DefaultLobbyAge             = time.Minute * 15
	DefaultGetReadyDuration     = time.Second * 5
)
