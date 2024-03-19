package repository

import (
	"errors"
	"github.com/redis/rueidis"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

func ReduceRedisResponseError(rr []rueidis.RedisResult, skipErrors ...error) error {
	return lo.Reduce(rr, func(agg error, item rueidis.RedisResult, index int) error {
		if agg != nil {
			return agg
		}

		err := item.Error()
		for _, skipError := range skipErrors {
			if errors.Is(err, skipError) {
				return nil
			}
		}

		if err != nil {
			logrus.WithError(err).WithField("index", index).Errorln("redis response reduce error")
		}
		return err
	}, nil)
}
