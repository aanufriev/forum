package delivery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/aanufriev/forum/configs"
	"github.com/aanufriev/forum/internal/pkg/forum"
	"github.com/aanufriev/forum/internal/pkg/models"
	"github.com/aanufriev/forum/internal/pkg/user"
	"github.com/gorilla/mux"
)

type ForumDelivery struct {
	forumUsecase forum.Usecase
	userUsecase  user.Usecase
}

func New(forumUsecase forum.Usecase, userUsecase user.Usecase) ForumDelivery {
	return ForumDelivery{
		forumUsecase: forumUsecase,
		userUsecase:  userUsecase,
	}
}

func (f ForumDelivery) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	forum := models.Forum{}
	err := json.NewDecoder(r.Body).Decode(&forum)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nickname, err := f.userUsecase.CheckIfUserExists(forum.User)
	if err != nil {
		msg := models.Message{
			Text: fmt.Sprintf("Can't find user with id #%v\n", forum.User),
		}

		w.WriteHeader(http.StatusNotFound)
		err = json.NewEncoder(w).Encode(msg)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}
	forum.User = nickname

	err = f.forumUsecase.Create(forum)
	if err != nil {
		existingForum, err := f.forumUsecase.Get(forum.Slug)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusConflict)
		err = json.NewEncoder(w).Encode(existingForum)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(forum)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (f ForumDelivery) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	slug := mux.Vars(r)["slug"]
	forum, err := f.forumUsecase.Get(slug)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		msg := models.Message{
			Text: fmt.Sprintf("Can't find forum with slug: %v", slug),
		}

		_ = json.NewEncoder(w).Encode(msg)
		return
	}

	err = json.NewEncoder(w).Encode(forum)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (f ForumDelivery) CreateThread(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	slug := mux.Vars(r)["slug"]

	thread := &models.Thread{}
	err := json.NewDecoder(r.Body).Decode(thread)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	thread.Forum = slug

	nickname, err := f.userUsecase.CheckIfUserExists(thread.Author)
	if err != nil {
		msg := models.Message{
			Text: fmt.Sprintf("Can't find user with id #%v\n", thread.Author),
		}

		w.WriteHeader(http.StatusNotFound)
		err = json.NewEncoder(w).Encode(msg)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}
	thread.Author = nickname

	err = f.forumUsecase.CreateThread(thread)
	if err != nil {
		if err == forum.ErrForumDoesntExists {
			w.WriteHeader(http.StatusNotFound)
			msg := models.Message{
				Text: fmt.Sprintf("Can't find thread forum by slug: %v", thread.Forum),
			}

			_ = json.NewEncoder(w).Encode(msg)
			return
		}

		existedThread, err := f.forumUsecase.GetThread(*thread.Slug)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(existedThread)
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(thread)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (f ForumDelivery) GetThreads(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	slug := mux.Vars(r)["slug"]

	_, err := f.forumUsecase.CheckForum(slug)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		msg := models.Message{
			Text: fmt.Sprintf("Can't find forum by slug: %v", slug),
		}

		_ = json.NewEncoder(w).Encode(msg)
		return
	}

	limit := r.URL.Query().Get(configs.Limit)
	desc := r.URL.Query().Get(configs.Desc)
	since := r.URL.Query().Get(configs.Since)

	threads, err := f.forumUsecase.GetThreads(slug, limit, since, desc)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(threads)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (f ForumDelivery) CreatePosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	slugOrID := mux.Vars(r)["slug_or_id"]

	posts := []models.Post{}
	err := json.NewDecoder(r.Body).Decode(&posts)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	created := time.Now()

	for idx := range posts {
		posts[idx].Created = created
	}

	posts, err = f.forumUsecase.CreatePosts(slugOrID, posts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(posts)
}

func (f ForumDelivery) GetThread(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	slugOrID := mux.Vars(r)["slug_or_id"]

	threads, err := f.forumUsecase.GetThread(slugOrID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(threads)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (f ForumDelivery) Vote(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	slugOrID := mux.Vars(r)["slug_or_id"]

	vote := models.Vote{}
	err := json.NewDecoder(r.Body).Decode(&vote)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	vote.Slug = slugOrID
	id, err := strconv.Atoi(slugOrID)
	if err != nil {
		id = 0
	}
	vote.ID = id

	thread, err := f.forumUsecase.Vote(vote)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.NewEncoder(w).Encode(thread)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (f ForumDelivery) GetPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	slugOrID := mux.Vars(r)["slug_or_id"]

	limitParam := r.URL.Query().Get(configs.Limit)
	limit, err := strconv.Atoi(limitParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sortParam := r.URL.Query().Get(configs.Sort)
	descParam := r.URL.Query().Get(configs.Desc)
	if descParam == "" {
		descParam = "false"
	}

	sinceParam := r.URL.Query().Get(configs.Since)

	posts, err := f.forumUsecase.GetPosts(slugOrID, limit, sortParam, descParam, sinceParam)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(posts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (f ForumDelivery) UpdateThread(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	slugOrID := mux.Vars(r)["slug_or_id"]

	thread := models.Thread{}
	err := json.NewDecoder(r.Body).Decode(&thread)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	thread, err = f.forumUsecase.UpdateThread(slugOrID, thread)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(thread)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
