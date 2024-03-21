package gameserver

import "time"

var (
	DefaultReminderToReadyAfter = time.Minute * 30
	DefaultReadyDeadline        = time.Minute * 45
	DefaultQuestionTimeout      = time.Second * 30
	DefaultLobbyAge             = time.Minute * 15
	DefaultGetReadyDuration     = time.Second * 5
)
