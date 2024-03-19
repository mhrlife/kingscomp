package integrationtest

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"kingscomp/internal/entity"
	"kingscomp/internal/repository"
	"kingscomp/internal/repository/redis"
	"testing"
)

func TestQuestions_ActiveQuestions(t *testing.T) {
	redisClient, err := redis.NewRedisClient(fmt.Sprintf("localhost:%s", redisPort))
	assert.NoError(t, err)
	ctx := context.Background()
	cb := repository.NewQuestionRedisRepository(redisClient)

	count, err := cb.GetActiveQuestionsCount(ctx)
	assert.Equal(t, int64(0), count)
	assert.NoError(t, err)

	err = cb.PushActiveQuestion(ctx, entity.Question{
		ID:            "id1",
		Question:      "q1",
		Answers:       []string{"a1", "a2"},
		CorrectAnswer: 1,
	}, entity.Question{
		ID:            "id2",
		Question:      "q2",
		Answers:       []string{"a1", "a2"},
		CorrectAnswer: 2,
	}, entity.Question{
		ID:            "id3",
		Question:      "q3",
		Answers:       []string{"a1", "a2"},
		CorrectAnswer: 2,
	})

	assert.NoError(t, err)
	count, err = cb.GetActiveQuestionsCount(ctx)
	assert.Equal(t, int64(3), count)
	assert.NoError(t, err)

	aqs, err := cb.GetActiveQuestions(ctx, 0, 2)
	assert.NoError(t, err)
	assert.Len(t, aqs, 2)

	assert.Equal(t, "id1", aqs[0].ID)
	assert.Equal(t, "id3", aqs[1].ID)

}
