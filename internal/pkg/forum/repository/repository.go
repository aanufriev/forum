package repository

import (
	"database/sql"
	"fmt"

	"github.com/aanufriev/forum/internal/pkg/forum"
	"github.com/aanufriev/forum/internal/pkg/models"
)

type ForumRepository struct {
	db *sql.DB
}

func New(db *sql.DB) forum.Repository {
	return ForumRepository{
		db: db,
	}
}

func (f ForumRepository) Create(model models.Forum) error {
	_, err := f.db.Exec(
		"INSERT INTO forums (slug, title, user_nickname) VALUES($1, $2, $3)",
		model.Slug, model.Title, model.User,
	)

	if err != nil {
		return fmt.Errorf("couldn't create new forum. Error: %w", err)
	}

	return nil
}

func (f ForumRepository) Get(slug string) (models.Forum, error) {
	var model models.Forum
	err := f.db.QueryRow(
		`SELECT slug, title, user_nickname FROM forums
		WHERE lower(slug) = lower($1)`,
		slug,
	).Scan(&model.Slug, &model.Title, &model.User)

	if err != nil {
		return models.Forum{}, fmt.Errorf("couldn't get forum with slug '%v'. Error: %w", slug, err)
	}

	return model, nil
}

func (f ForumRepository) CreateThread(thread *models.Thread) error {
	err := f.db.QueryRow(
		`INSERT INTO threads (author, created, forum, msg, title)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		thread.Author, thread.Created, thread.Forum, thread.Message, thread.Title,
	).Scan(&thread.ID)

	if err != nil {
		return fmt.Errorf("couldn't create thread. Error: %w", err)
	}

	return nil
}
