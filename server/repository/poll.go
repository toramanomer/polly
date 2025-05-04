package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/toramanomer/polly/primitives"
)

var (
	ErrPollNotFound        = errors.New("poll not found")
	ErrOptionBelongsToPoll = errors.New("option does not belong to the poll")
	ErrPollExpired         = errors.New("poll expired")
)

type PollOption struct {
	ID       uuid.UUID `json:"id"`
	PollID   uuid.UUID `json:"pollID"`
	Text     string    `json:"text"`
	Position int       `json:"position"`
	Count    int       `json:"count"`
}

type Poll struct {
	ID        uuid.UUID           `json:"id"`
	UserID    uuid.UUID           `json:"userID"`
	Question  primitives.Question `json:"question"`
	CreatedAt time.Time           `json:"createdAt"`
	ExpiresAt time.Time           `json:"expiresAt"`
	Options   []PollOption        `json:"options"`
}

type NewPollParams struct {
	UserID    uuid.UUID
	Question  primitives.Question
	ExpiresAt time.Time
	Options   []string
}

func NewPoll(params NewPollParams) *Poll {

	pollID := uuid.New()
	pollOptions := make([]PollOption, len(params.Options))

	for i, option := range params.Options {
		pollOptions[i] = PollOption{
			ID:       uuid.New(),
			PollID:   pollID,
			Text:     option,
			Position: i,
		}
	}

	return &Poll{
		ID:        pollID,
		UserID:    params.UserID,
		Question:  params.Question,
		CreatedAt: time.Now(),
		ExpiresAt: params.ExpiresAt,
		Options:   pollOptions,
	}
}

type Vote struct {
	ID       uuid.UUID `json:"id"`
	PollID   uuid.UUID `json:"pollID"`
	OptionID uuid.UUID `json:"optionID"`
	VotedAt  time.Time `json:"votedAt"`
}

func NewVote(pollID uuid.UUID, optionID uuid.UUID) *Vote {
	return &Vote{
		ID:       uuid.New(),
		PollID:   pollID,
		OptionID: optionID,
		VotedAt:  time.Now(),
	}
}
