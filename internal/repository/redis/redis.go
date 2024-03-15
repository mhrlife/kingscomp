package redis

import (
	"github.com/redis/rueidis"
)

func NewRedisClient(address string) (rueidis.Client, error) {
	return rueidis.NewClient(rueidis.ClientOption{InitAddress: []string{address}})
}
