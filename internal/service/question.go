package service

import (
	"kingscomp/internal/repository"
)

type QuestionService struct {
	repository.Question
}

func NewQuestionService(rep repository.Question) *QuestionService {
	return &QuestionService{Question: rep}
}
