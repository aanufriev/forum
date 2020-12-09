package forum

import "github.com/aanufriev/forum/internal/pkg/models"

type Usecase interface {
	Create(forum models.Forum) error
	Get(slug string) (models.Forum, error)
	CreateThread(model *models.Thread) error
	CheckForum(slug string) (string, error)
	GetThreads(slug string, limit string, since string, desc string) ([]models.Thread, error)
}
