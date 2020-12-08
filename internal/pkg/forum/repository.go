package forum

import "github.com/aanufriev/forum/internal/pkg/models"

type Repository interface {
	Create(forum models.Forum) error
	Get(slug string) (models.Forum, error)
}
