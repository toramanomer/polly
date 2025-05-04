package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/toramanomer/polly/primitives"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

const createUser = `
	INSERT INTO users (id, username, email, password_hash)
	VALUES ($1, $2, $3, $4)`

func (r *Repository) CreateUser(ctx context.Context, user *User) error {
	_, err := r.db.Exec(ctx, createUser, user.ID, user.Username, user.Email, user.PasswordHash)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				if strings.Contains(pgErr.Message, "email") {
					return ErrEmailAlreadyExists
				}
				if strings.Contains(pgErr.Message, "username") {
					return ErrUsernameAlreadyExists
				}
			}
		}
	}

	return err
}

const getUserByEmail = `
	SELECT id, username, email, password_hash
	FROM users
	WHERE email = $1`

func (r *Repository) GetUserByEmail(ctx context.Context, email primitives.Email) (*User, error) {
	var user User

	err := r.db.
		QueryRow(ctx, getUserByEmail, email).
		Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

const insertPoll = `
	INSERT INTO polls (id, user_id, question, created_at, expires_at)
	VALUES ($1, $2, $3, $4, $5)`

const insertPollOptions = "INSERT INTO poll_options (id, poll_id, text, position) VALUES"

func (r *Repository) CreatePollWithOptions(ctx context.Context, poll *Poll) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	//-------------------- Insert poll
	_, err = tx.Exec(ctx, insertPoll,
		poll.ID, poll.UserID, poll.Question, poll.CreatedAt, poll.ExpiresAt)
	if err != nil {
		return fmt.Errorf("error inserting poll: %w", err)
	}
	//--------------------

	//-------------------- Insert poll options
	var (
		optionsCount = len(poll.Options)
		valueStrings = make([]string, 0, optionsCount)
		valueArgs    = make([]any, 0, optionsCount)
	)

	for i, option := range poll.Options {
		baseIndex := i * 4
		valueStrings = append(valueStrings,
			fmt.Sprintf(
				"($%d, $%d, $%d, $%d)",
				baseIndex+1, baseIndex+2, baseIndex+3, baseIndex+4),
		)
		valueArgs = append(valueArgs,
			option.ID, option.PollID, option.Text, option.Position)
	}

	optionsQuery := fmt.Sprintf(
		"%s %s",
		insertPollOptions, strings.Join(valueStrings, ","))

	_, err = tx.Exec(ctx, optionsQuery, valueArgs...)
	if err != nil {
		return fmt.Errorf("error inserting poll options: %w", err)
	}
	//--------------------

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

const deletePoll = `
	WITH
		to_delete AS (
			SELECT true AS exists, (user_id = $2) AS is_owner
			FROM polls
			WHERE id = $1
		),
		deleted	AS (
			DELETE FROM polls
			WHERE id = $1 AND user_id = $2
			RETURNING true AS deleted
		)
	SELECT
		COALESCE ((SELECT exists FROM to_delete), false) AS poll_exists,
		COALESCE ((SELECT is_owner FROM to_delete), false) AS is_owner,
		COALESCE ((SELECT deleted FROM deleted), false) AS deleted`

type DeletePollParams struct {
	PollID uuid.UUID
	UserID uuid.UUID
}

func (r *Repository) DeletePoll(ctx context.Context, arg DeletePollParams) error {
	row := r.db.QueryRow(ctx, deletePoll, arg.PollID, arg.UserID)

	var pollExists, isOwner, deleted bool
	err := row.Scan(&pollExists, &isOwner, &deleted)

	if err != nil {
		return fmt.Errorf("error deleting poll: %w", err)
	}

	switch {
	case !pollExists:
		return ErrPollNotFound
	case !isOwner:
		return ErrNotPollOwner
	case !deleted:
		return errors.New("unknown error occurred while deleting poll")
	}

	return nil
}

const getPollWithOptions = `
	SELECT
		p.id,
		p.user_id,
		p.question,
		p.created_at,
		p.expires_at,
		COALESCE(
		(
			SELECT jsonb_agg(json_build_object(
				'id', o.id,
				'poll_id', o.poll_id,
				'text', o.text,
				'position', o.position
			) ORDER BY o.position)
			FROM poll_options o
			WHERE o.poll_id = p.id
		),
		'[]'
		) AS options
	FROM polls p
	WHERE p.id = $1`

func (r *Repository) GetPollWithOptions(ctx context.Context, pollID uuid.UUID) (*Poll, error) {
	row := r.db.QueryRow(ctx, getPollWithOptions, pollID)

	var poll Poll
	err := row.Scan(
		&poll.ID,
		&poll.UserID,
		&poll.Question,
		&poll.CreatedAt,
		&poll.ExpiresAt,
		&poll.Options,
	)

	return &poll, err
}

const recordVote = `
	WITH
		poll_found AS (
			SELECT expires_at
			FROM polls
			WHERE id = $2
		),
		poll_active AS (
			SELECT 1 AS ok
			FROM poll_found
			WHERE expires_at > $5
		),
		option_valid AS (
			SELECT 1 AS ok
			FROM poll_options
			WHERE id = $3 AND poll_id = $2
		),
		insert_vote AS (
			INSERT INTO votes (id, poll_id, option_id, voted_at)
			SELECT $1, $2, $3, $4
			WHERE
				EXISTS (SELECT 1 FROM poll_active) AND
				EXISTS (SELECT 1 FROM option_valid)
			RETURNING 1
		)
	SELECT
		EXISTS (SELECT 1 FROM poll_found)	AS poll_exists,
		EXISTS (SELECT 1 FROM poll_active)	AS poll_active,
		EXISTS (SELECT 1 FROM option_valid)	AS option_valid,
		EXISTS (SELECT 1 FROM insert_vote)	AS inserted`

func (r *Repository) RecordVote(ctx context.Context, vote *Vote) error {
	row := r.db.QueryRow(ctx, recordVote,
		vote.ID, vote.PollID, vote.OptionID, vote.VotedAt, time.Now())

	var pollExists, pollActive, optionValid, inserted bool
	err := row.Scan(&pollExists, &pollActive, &optionValid, &inserted)

	switch {
	case err != nil:
		return fmt.Errorf("error inserting vote: %w", err)
	case !pollExists:
		return ErrPollNotFound
	case !pollActive:
		return ErrPollExpired
	case !optionValid:
		return ErrOptionBelongsToPoll
	case !inserted:
		return errors.New("unknown error occurred while inserting vote")
	}

	return nil
}

const getUserPollsWithStats = `
	SELECT
		polls.id,
		polls.user_id,
		polls.question,
		polls.created_at,
		polls.expires_at,
		jsonb_agg(json_build_object(
			'id', poll_options.id,
			'poll_id', poll_options.poll_id,
			'text', poll_options.text,
			'position', poll_options.position,
			'count', COALESCE(vs.vote_count, 0)
		) ORDER BY position ASC) AS options
	FROM polls
	JOIN poll_options ON poll_options.poll_id = polls.id
	LEFT JOIN (
		SELECT option_id, COUNT(*) AS vote_count
		FROM votes
		GROUP BY option_id
	) vs ON vs.option_id = poll_options.id
	WHERE user_id = $1
	GROUP BY polls.id
	ORDER BY polls.created_at DESC`

func (r *Repository) GetUserPollsWithStats(ctx context.Context, userID uuid.UUID) ([]Poll, error) {
	rows, err := r.db.Query(ctx, getUserPollsWithStats, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying user polls: %w", err)
	}
	defer rows.Close()

	var polls []Poll
	for rows.Next() {
		var poll Poll
		err := rows.Scan(
			&poll.ID,
			&poll.UserID,
			&poll.Question,
			&poll.CreatedAt,
			&poll.ExpiresAt,
			&poll.Options,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning poll: %w", err)
		}
		polls = append(polls, poll)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating polls: %w", err)
	}

	return polls, nil
}
