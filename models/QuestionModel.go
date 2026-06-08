package models

import "errors"

type Question struct {
	Question      string   `bson:"question" json:"question" binding:"required"`
	Options       []string `bson:"options" json:"options" binding:"required,min=2"`
	CorrectAnswer string   `bson:"correct_answer" json:"correct_answer" binding:"required"`
	TimeLimit     int      `bson:"time_limit" json:"time_limit"`
}

func (q *Question) Validate() error {
	if q.Question == "" {
		return errors.New("question text is required")
	}
	if len(q.Options) < 2 {
		return errors.New("each question must have at least 2 options")
	}
	if q.CorrectAnswer == "" {
		return errors.New("correct_answer is required")
	}
	for _, opt := range q.Options {
		if opt == q.CorrectAnswer {
			return nil
		}
	}
	return errors.New("correct_answer must match one of the provided options")
}
