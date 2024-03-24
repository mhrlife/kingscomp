package integrationtest

import (
	"context"
	"fmt"
	"github.com/redis/rueidis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"kingscomp/internal/repository/redis"
	"kingscomp/internal/scoreboard"
	"testing"
)

type ScoreboardSuite struct {
	suite.Suite
	rc rueidis.Client
	sb *scoreboard.Scoreboard

	ctx context.Context
}

func TestScoreboard(t *testing.T) {
	rc, err := redis.NewRedisClient(fmt.Sprintf("localhost:%s", redisPort))
	assert.NoError(t, err)
	s := &ScoreboardSuite{
		sb:  scoreboard.NewScoreboard(rc),
		rc:  rc,
		ctx: context.Background(),
	}
	suite.Run(t, s)
}

func (s *ScoreboardSuite) BeforeTest(suiteName, testName string) {
	s.rc.Do(context.Background(), s.rc.B().Flushall().Build())
}

func (s *ScoreboardSuite) TestScoreboardSimple() {
	for i := 1; i <= 100; i++ {
		s.NoError(s.sb.Register(s.ctx, int64(i), int64(10*i)))
	}
	info, err := s.sb.Get(s.ctx, scoreboard.GetScoreboardArgs{
		Type:       scoreboard.ScoreboardHourly,
		FirstCount: 10,
		AccountID:  430,
	})
	s.NoError(err)
	s.Len(info.Tops, 10)
	s.Equal(int64(1000), info.Tops[0].Score)
	s.Equal(int64(100), info.Tops[0].AccountID)
	s.False(info.Found)

	info, err = s.sb.Get(s.ctx, scoreboard.GetScoreboardArgs{
		Type:       scoreboard.ScoreboardHourly,
		FirstCount: 10,
		AccountID:  5,
	})
	s.NoError(err)
	s.True(info.Found)
	s.Equal(int64(50), info.UserScore)
	s.Equal(int64(95), info.UserRank)

}
