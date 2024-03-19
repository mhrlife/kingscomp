package gameserver

import "time"

var (
	DefaultReminderToReadyAfter = time.Second * 10
	DefaultReadyDeadline        = time.Second * 20
)
