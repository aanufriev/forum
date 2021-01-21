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

func (f ForumUsecase) CreatePosts(thread models.Thread, posts []models.Post) ([]models.Post, error) {
	return f.forumRepository.CreatePosts(thread, posts)
}

func (f ForumUsecase) GetThread(slugOrID string) (models.Thread, error) {
	id, err := strconv.Atoi(slugOrID)
	if err != nil {
		return f.forumRepository.GetThreadBySlug(slugOrID)
	}

	return f.forumRepository.GetThreadByID(id)
}

func (f ForumUsecase) Vote(vote models.Vote) (models.Thread, error) {
	return f.forumRepository.Vote(vote)
}

func (f ForumUsecase) GetPosts(slugOrID string, limit int, sort string, order string, since string) ([]models.Post, error) {
	switch order {
	case "true":
		order = "DESC"
	case "false":
		order = "ASC"
	}

	switch sort {
	case "flat":
		return f.forumRepository.GetPosts(slugOrID, limit, order, since)
	case "tree":
		return f.forumRepository.GetPostsTree(slugOrID, limit, order, since)
	case "parent_tree":
		return f.forumRepository.GetPostsParentTree(slugOrID, limit, order, since)
	default:
		return f.forumRepository.GetPosts(slugOrID, limit, order, since)
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

func (f ForumUsecase) GetPostDetails(id string) (models.Post, error) {
	return f.forumRepository.GetPostDetails(id)
}

func (f ForumUsecase) UpdatePost(post models.Post) (models.Post, error) {
	return f.forumRepository.UpdatePost(post)
}

func (f ForumUsecase) ClearService() error {
	return f.forumRepository.ClearService()
}

func (f ForumUsecase) GetServiceInfo() (models.ServiceInfo, error) {
	return f.forumRepository.GetServiceInfo()
}

func (f ForumUsecase) CheckThread(slugOrID string) error {
	id, err := strconv.Atoi(slugOrID)
	if err != nil {
		_, err = f.forumRepository.CheckThreadBySlug(slugOrID)
		return err
	}

	_, err = f.forumRepository.CheckThreadByID(id)
	return err
}

func (f ForumUsecase) GetThreadIDAndForum(slugOrID string) (models.Thread, error) {
	return f.forumRepository.GetThreadIDAndForum(slugOrID)
}
