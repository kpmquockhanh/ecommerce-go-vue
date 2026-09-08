package repositories

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DeadLetterRepository interface {
	List(ctx context.Context, limit, offset int) ([]DeadLetter, error)
	Count(ctx context.Context) (int, error)
	FindByID(ctx context.Context, id int) (*DeadLetter, error)
	MarkRetried(ctx context.Context, id int) error
	Delete(ctx context.Context, id int) error
	Persist(ctx context.Context, queueName string, jobType string, payload []byte, errMsg string, retryCount int) error
}

type DeadLetter struct {
	ID           int             `json:"id"`
	JobType      string          `json:"job_type"`
	QueueName    string          `json:"queue_name"`
	Payload      json.RawMessage `json:"payload"`
	ErrorMessage string          `json:"error_message"`
	RetryCount   int             `json:"retry_count"`
	Status       string          `json:"status"`
	CreatedAt    string          `json:"created_at"`
	RetriedAt    *string         `json:"retried_at"`
}

type deadLetterRepository struct {
	db *pgxpool.Pool
}

func NewDeadLetterRepository(db *pgxpool.Pool) DeadLetterRepository {
	return &deadLetterRepository{db: db}
}

func (r *deadLetterRepository) List(ctx context.Context, limit, offset int) ([]DeadLetter, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, job_type, queue_name, payload, error_message, retry_count, status, created_at, retried_at
		 FROM dead_letters
		 WHERE status = 'failed'
		 ORDER BY created_at DESC
		 LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deadLetters []DeadLetter
	for rows.Next() {
		var dl DeadLetter
		if err := rows.Scan(&dl.ID, &dl.JobType, &dl.QueueName, &dl.Payload,
			&dl.ErrorMessage, &dl.RetryCount, &dl.Status, &dl.CreatedAt, &dl.RetriedAt); err != nil {
			return nil, err
		}
		deadLetters = append(deadLetters, dl)
	}
	return deadLetters, nil
}

func (r *deadLetterRepository) Count(ctx context.Context) (int, error) {
	var total int
	err := r.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM dead_letters WHERE status = 'failed'").Scan(&total)
	return total, err
}

func (r *deadLetterRepository) FindByID(ctx context.Context, id int) (*DeadLetter, error) {
	var dl DeadLetter
	err := r.db.QueryRow(ctx,
		`SELECT id, job_type, queue_name, payload, error_message, retry_count
		 FROM dead_letters WHERE id = $1 AND status = 'failed'`, id,
	).Scan(&dl.ID, &dl.JobType, &dl.QueueName, &dl.Payload, &dl.ErrorMessage, &dl.RetryCount)
	if err != nil {
		return nil, err
	}
	return &dl, nil
}

func (r *deadLetterRepository) MarkRetried(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx,
		"UPDATE dead_letters SET status = 'retried', retried_at = NOW() WHERE id = $1", id)
	return err
}

func (r *deadLetterRepository) Delete(ctx context.Context, id int) error {
	result, err := r.db.Exec(ctx,
		"DELETE FROM dead_letters WHERE id = $1", id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *deadLetterRepository) Persist(ctx context.Context, queueName string, jobType string, payload []byte, errMsg string, retryCount int) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO dead_letters (job_type, queue_name, payload, error_message, retry_count)
		 VALUES ($1, $2, $3, $4, $5)`,
		jobType, queueName, payload, errMsg, retryCount)
	return err
}
