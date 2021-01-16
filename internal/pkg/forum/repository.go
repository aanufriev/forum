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
	GetThreadByID(id int) (models.Thread, error)
	GetThreadBySlug(slug string) (models.Thread, error)
	Vote(vote models.Vote) (models.Thread, error)
	GetPosts(slug string, id int, limit int, order string, since string) ([]models.Post, error)
	GetPostsTree(slug string, id int, limit int, order string, since string) ([]models.Post, error)
	GetPostsParentTree(slug string, id int, limit int, order string, since string) ([]models.Post, error)
	UpdateThread(thread models.Thread) (models.Thread, error)
	GetUsersFromForum(slug string, limit int, since string, desc string) ([]models.User, error)
	GetPostDetails(id string) (models.Post, error)
	UpdatePost(post models.Post) (models.Post, error)
	ClearService() error
	GetServiceInfo() (models.ServiceInfo, error)
	CheckThreadByID(id int) error
	CheckThreadBySlug(slug string) error
}
