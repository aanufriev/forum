package usecase

import (
	"strconv"

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

func (f ForumUsecase) CreateThread(thread *models.Thread) error {
	return f.forumRepository.CreateThread(thread)
}

func (f ForumUsecase) CheckForum(slug string) (string, error) {
	return f.forumRepository.CheckForum(slug)
}

func (f ForumUsecase) GetThreads(slug string, limit string, since string, desc string) ([]models.Thread, error) {
	return f.forumRepository.GetThreads(slug, limit, since, desc)
}

func (f ForumUsecase) CreatePosts(slugOrID string, posts []models.Post) ([]models.Post, error) {
	slug := slugOrID
	id, err := strconv.Atoi(slugOrID)
	if err != nil {
		id = 0
	}
	return f.forumRepository.CreatePosts(slug, id, posts)
}

func (f ForumUsecase) GetThread(slugOrID string) (models.Thread, error) {
	slug := slugOrID
	id, err := strconv.Atoi(slugOrID)
	if err != nil {
		id = 0
	}
	return f.forumRepository.GetThread(slug, id)
}

func (f ForumUsecase) Vote(vote models.Vote) (models.Thread, error) {
	return f.forumRepository.Vote(vote)
}

func (f ForumUsecase) GetPosts(slugOrID string, limit int, sort string, order string, since string) ([]models.Post, error) {
	slug := slugOrID
	id, err := strconv.Atoi(slugOrID)
	if err != nil {
		id = 0
	}

	switch order {
	case "true":
		order = "DESC"
	case "false":
		order = "ASC"
	}

	switch sort {
	case "flat":
		return f.forumRepository.GetPosts(slug, id, limit, order, since)
	case "tree":
		return f.forumRepository.GetPostsTree(slug, id, limit, order, since)
	case "parent_tree":
		return f.forumRepository.GetPostsParentTree(slug, id, limit, order, since)
	default:
		return f.forumRepository.GetPosts(slug, id, limit, order, since)
	}
}

func (f ForumUsecase) UpdateThread(slugOrID string, thread models.Thread) (models.Thread, error) {
	thread.Slug = &slugOrID
	id, err := strconv.Atoi(slugOrID)
	if err != nil {
		id = 0
	}

	thread.ID = id
	return f.forumRepository.UpdateThread(thread)
}

func (f ForumUsecase) GetUsersFromForum(slug string, limit int, since string, desc string) ([]models.User, error) {
	switch desc {
	case "true":
		desc = "DESC"
	case "false":
		desc = "ASC"
	}
	return f.forumRepository.GetUsersFromForum(slug, limit, since, desc)
}

func (f ForumUsecase) GetPostDetaild(id string) (models.Post, error) {
	return f.forumRepository.GetPostDetaild(id)
}
