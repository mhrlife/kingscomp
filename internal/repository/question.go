package repository

import (
	"context"
	"github.com/redis/rueidis"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"kingscomp/internal/entity"
)

var _ Question = &QuestionRedisRepository{}

type QuestionRedisRepository struct {
	*RedisCommonBehaviour[entity.Question]
}

func NewQuestionRedisRepository(client rueidis.Client) *QuestionRedisRepository {
	return &QuestionRedisRepository{
		NewRedisCommonBehaviour[entity.Question](client),
	}
}

func (q *QuestionRedisRepository) GetActiveQuestionsCount(ctx context.Context) (int64, error) {
	cmd := q.client.B().Llen().Key("active_questions").Build()
	return q.client.Do(ctx, cmd).ToInt64()
}

func (q *QuestionRedisRepository) GetActiveQuestions(ctx context.Context, index ...int64) ([]entity.Question, error) {
	cmds := make([]rueidis.Completed, len(index))
	for i, id := range index {
		cmds[i] = q.client.B().Lindex().Key("active_questions").Index(id).Build()
	}
	resp := q.client.DoMulti(ctx, cmds...)
	err := ReduceRedisResponseError(resp)
	if err != nil {
		logrus.WithError(err).Errorln("couldn't fetch active questions from redis")
		return nil, err
	}

	questionIds := lo.Map(resp, func(item rueidis.RedisResult, index int) entity.ID {
		s, _ := item.ToString()
		return entity.NewID("question", s)
	})

	return q.MGet(ctx, questionIds...)
}

func (q *QuestionRedisRepository) PushActiveQuestion(ctx context.Context, questions ...entity.Question) error {
	if err := q.MSet(ctx, questions...); err != nil {
		return err
	}

	ids := lo.Map(questions, func(item entity.Question, _ int) string {
		return item.ID
	})

	cmd := q.client.B().Rpush().Key("active_questions").Element(ids...).Build()
	if err := q.client.Do(ctx, cmd).Error(); err != nil {
		logrus.WithError(err).Errorln("couldn't push active questions")
		return err
	}
	return nil
}
