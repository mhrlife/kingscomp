package scoreboard

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/rueidis"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"kingscomp/internal/repository"
	"strconv"
	"time"
)

type Scoreboard struct {
	rdb rueidis.Client
}

func NewScoreboard(rdb rueidis.Client) *Scoreboard {
	return &Scoreboard{rdb: rdb}
}

func (s *Scoreboard) Register(ctx context.Context, accountId int64, score int64) error {
	keys := []string{
		fmt.Sprintf("scoreboard:%s", time.Now().Format("2006-01-02")),
		fmt.Sprintf("scoreboard:%s:%02d", time.Now().Format("2006-01-02"), time.Now().Hour()),
	}
	cmds := lo.Map(keys, func(key string, _ int) rueidis.Completed {
		return s.rdb.B().Zincrby().Key(key).Increment(float64(score)).
			Member(strconv.FormatInt(accountId, 10)).Build()
	})

	cmds = append(cmds, lo.Map(keys, func(key string, _ int) rueidis.Completed {
		return s.rdb.B().Expire().Key(key).Seconds(int64((time.Hour * 24).Seconds())).Build()
	})...)

	err := repository.ReduceRedisResponseError(s.rdb.DoMulti(ctx,
		cmds...,
	))
	if err != nil {
		logrus.WithError(err).Errorln("couldn't update the scoreboard")
	}
	return err
}

type ScoreboardType int

const (
	ScoreboardDaily ScoreboardType = iota
	ScoreboardHourly
)

type GetScoreboardArgs struct {
	Type       ScoreboardType
	FirstCount int
	AccountID  int64
}
type Score struct {
	AccountID int64 `json:"account_id"`
	Score     int64 `json:"score"`
}
type Info struct {
	Type      ScoreboardType
	Tops      []Score
	UserScore int64
	UserRank  int64
	Found     bool
}

func (s *Scoreboard) Get(ctx context.Context, args GetScoreboardArgs) (Info, error) {
	keys := map[ScoreboardType]string{
		ScoreboardDaily:  fmt.Sprintf("scoreboard:%s", time.Now().Format("2006-01-02")),
		ScoreboardHourly: fmt.Sprintf("scoreboard:%s:%02d", time.Now().Format("2006-01-02"), time.Now().Hour()),
	}
	key := keys[args.Type]

	cmdTopScore := s.rdb.B().Zrevrange().Key(key).Start(0).Stop(int64(args.FirstCount - 1)).Withscores().Build()
	cmdRank := s.rdb.B().Zrevrank().Key(key).Member(strconv.FormatInt(args.AccountID, 10)).Withscore().Build()
	result := s.rdb.DoMulti(ctx, cmdTopScore, cmdRank)
	err := repository.ReduceRedisResponseError(result)
	tops, err := result[0].AsZScores()
	if err != nil {
		logrus.WithError(err).Errorln("couldn't fetch top scoreboard")
		return Info{}, err
	}

	info := Info{Type: args.Type}
	info.Tops = lo.Map(tops, func(item rueidis.ZScore, _ int) Score {
		accountId, _ := strconv.ParseInt(item.Member, 10, 64)
		return Score{
			AccountID: accountId,
			Score:     int64(item.Score),
		}
	})

	userInfo, err := result[1].AsIntSlice()
	if err != nil {
		if errors.Is(err, rueidis.Nil) {
			return info, nil
		}
		logrus.WithError(err).Errorln("couldn't fetch user's scoreboard")
		return Info{}, err
	}
	info.UserScore = userInfo[1]
	info.UserRank = userInfo[0] + 1
	info.Found = true
	return info, nil
}
