package forum

import (
	"fmt"

	"github.com/aanufriev/forum/internal/pkg/models"
)

var (
	ErrForumDoesntExists = fmt.Errorf("forum not exists")
	ErrDataConflict      = fmt.Errorf("data conflict")
)

type Repository interface {
	Create(forum models.Forum) error
	Get(slug string) (models.Forum, error)
	CreateThread(model *models.Thread) error
	CheckForum(slug string) (string, error)
	GetThreads(slug string, limit string, since string, desc string) ([]models.Thread, error)
	CreatePosts(slug string, id int, posts []models.Post) ([]models.Post, error)
	GetThread(slug string) (models.Thread, error)
}
