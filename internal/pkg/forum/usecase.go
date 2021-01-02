package forum

import "github.com/aanufriev/forum/internal/pkg/models"

type Usecase interface {
	Create(forum models.Forum) error
	Get(slug string) (models.Forum, error)
	CreateThread(model *models.Thread) error
	CheckForum(slug string) (string, error)
	GetThreads(slug string, limit string, since string, desc string) ([]models.Thread, error)
	CreatePosts(slugOrID string, posts []models.Post) ([]models.Post, error)
	GetThread(slugOrID string) (models.Thread, error)
	Vote(vote models.Vote) (models.Thread, error)
	GetPosts(slugOrID string, limit int, sort string, order string, since string) ([]models.Post, error)
	UpdateThread(slugOrID string, thread models.Thread) (models.Thread, error)
	GetUsersFromForum(slug string, limit int, since string, desc string) ([]models.User, error)
	GetPostDetaild(id string) (models.Post, error)
	UpdatePost(post models.Post) (models.Post, error)
	ClearService() error
	GetServiceInfo() (models.ServiceInfo, error)
	CheckThread(slugOrID string) error
}
