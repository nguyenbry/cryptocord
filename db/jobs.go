package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrUnique = errors.New("unique constraint")

type Job struct {
	Id  uuid.UUID `json:"id"`
	Url string    `json:"url"`
}

type JobsService struct {
	db *Db
}

func NewJobsService(db *Db) *JobsService {
	return &JobsService{db}
}

func (j *JobsService) Create(ctx context.Context, url string) (*uuid.UUID, error) {
	sql := `INSERT INTO jobs (id, url) VALUES ($1, $2) RETURNING id`

	var id uuid.UUID
	err := j.db.Pool.QueryRow(ctx, sql, uuid.New(), url).Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, fmt.Errorf("failed to insert: %w: %w", ErrUnique, pgErr)
			}
		}
		return nil, fmt.Errorf("failed to insert: %v", err)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return &id, nil
	}

}

func (j *JobsService) Get(id uuid.UUID) (*Job, error) {
	sql := `SELECT id, url FROM jobs WHERE id = $1`

	job := Job{}
	err := j.db.Pool.QueryRow(context.TODO(), sql, id).Scan(&job.Id, &job.Url)

	if err != nil {
		return nil, err
	}

	return &job, nil
}

func (j *JobsService) All(ctx context.Context) ([]Job, error) {
	rows, err := j.db.Pool.Query(ctx, `SELECT id, url FROM jobs`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var jobs []Job

	for rows.Next() {
		var id uuid.UUID
		var url string

		err := rows.Scan(&id, &url)

		if err != nil {
			return nil, err
		}

		jobs = append(jobs, Job{
			Id:  id,
			Url: url,
		})
	}

	err = rows.Err()

	if err != nil {
		return nil, err
	}

	return jobs, nil
}
