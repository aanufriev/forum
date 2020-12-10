package repository

import (
	"database/sql"
	"fmt"
	"strconv"

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
	var forumSlug string

	err := f.db.QueryRow(
		"SELECT slug FROM forums WHERE lower(slug) = lower($1)",
		thread.Forum,
	).Scan(&forumSlug)

	if err == sql.ErrNoRows {
		return forum.ErrForumDoesntExists
	}

	if err != nil {
		return err
	}

	thread.Forum = forumSlug

	err = f.db.QueryRow(
		`INSERT INTO threads (author, created, forum, msg, title, slug)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		thread.Author, thread.Created, thread.Forum, thread.Message, thread.Title, thread.Slug,
	).Scan(&thread.ID)

	if err != nil {
		return fmt.Errorf("couldn't create thread. Error: %w", err)
	}

	return nil
}

func (f ForumRepository) CheckForum(slug string) (string, error) {
	err := f.db.QueryRow(
		"SELECT slug FROM forums WHERE lower(slug) = lower($1)",
		slug,
	).Scan(&slug)

	if err != nil {
		return "", fmt.Errorf("couldn't get forum with slug '%v'. Error: %w", slug, err)
	}

	return slug, nil
}

func (f ForumRepository) GetThreads(slug string, limit string, since string, desc string) ([]models.Thread, error) {
	query := `SELECT author, created, forum, id, msg, slug, title FROM threads
	WHERE lower(forum) = lower($1)`

	args := make([]interface{}, 0, 4)
	args = append(args, slug)

	var operator string
	if desc == "" || desc == "false" {
		operator = ">"
	} else {
		operator = "<"
	}

	if since != "" {
		query += fmt.Sprintf(" AND created %v= $2", operator)
		args = append(args, since)
	}

	if desc == "" || desc == "false" {
		desc = "ASC"
	} else {
		desc = "DESC"
	}
	query += fmt.Sprintf(" ORDER BY created %v", desc)

	rows, err := f.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return nil, err
	}

	threads := make([]models.Thread, 0, limitInt)
	var thread models.Thread
	var idx int
	for rows.Next() {
		idx++
		if idx == limitInt+1 {
			break
		}
		err = rows.Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.ID, &thread.Message, &thread.Slug, &thread.Title)
		if err != nil {
			return nil, err
		}

		threads = append(threads, thread)
	}

	return threads, nil
}

func (f ForumRepository) CreatePosts(slug string, id int, posts []models.Post) ([]models.Post, error) {
	var forum string
	err := f.db.QueryRow(
		`SELECT f.slug, t.id, t.slug FROM forums AS f
		JOIN threads AS t
		ON lower(f.slug) = lower(t.forum)
		WHERE lower(t.slug) = lower($1) or t.id = $2`,
		slug, id,
	).Scan(&forum, &id, &slug)

	if err != nil {
		return nil, err
	}

	for idx, post := range posts {
		posts[idx].Forum = forum
		if id != 0 {
			posts[idx].Thread = id
		} else {
			posts[idx].Slug = slug
		}
		err := f.db.QueryRow(
			"INSERT INTO posts (author, msg, parent, thread) VALUES ($1, $2, $3, $4) RETURNING id",
			post.Author, post.Message, post.Parent, post.Thread,
		).Scan(&posts[idx].ID)

		if err != nil {
			return nil, err
		}
	}

	return posts, nil
}

func (f ForumRepository) GetThread(slug string) (models.Thread, error) {
	var thread models.Thread
	err := f.db.QueryRow(
		`SELECT author, created, forum, id, msg, slug, title FROM threads
		WHERE lower(slug) = lower($1)`,
		slug,
	).Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.ID, &thread.Message, &thread.Slug, &thread.Title)

	if err != nil {
		return models.Thread{}, err
	}

	return thread, nil
}
