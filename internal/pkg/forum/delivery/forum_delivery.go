package delivery

import (
	"encoding/json"
	"fmt"
	"net/http"

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

	thread := &models.Thread{}
	err := json.NewDecoder(r.Body).Decode(thread)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

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
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(threads)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
