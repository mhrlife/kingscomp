package gameserver

import "time"

var (
	DefaultReminderToReadyAfter = time.Second * 10
	DefaultReadyDeadline        = time.Second * 15
	DefaultQuestionTimeout      = time.Second * 10
	DefaultLobbyAge             = time.Minute * 15
	DefaultGetReadyDuration     = time.Second * 5
)
