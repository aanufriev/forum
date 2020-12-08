package usecase

import (
	"github.com/aanufriev/forum/internal/pkg/forum"
	"github.com/aanufriev/forum/internal/pkg/models"
)

type ForumUsecase struct {
	forumRepository forum.Repository
}

func New(forumRepository forum.Repository) forum.Usecase {
	return ForumUsecase{
		forumRepository: forumRepository,
	}
}

func (f ForumUsecase) Create(forum models.Forum) error {
	return f.forumRepository.Create(forum)
}

func (f ForumUsecase) Get(slug string) (models.Forum, error) {
	return f.forumRepository.Get(slug)
}
