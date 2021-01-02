package delivery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
		_, err = f.userUsecase.CheckIfUserExists(posts[idx].Author)
		if err != nil {
			msg := models.Message{
				Text: fmt.Sprintf("Can't find post author by nickname: %v", posts[idx].Author),
			}

			w.WriteHeader(http.StatusNotFound)
			err = json.NewEncoder(w).Encode(msg)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			return
		}
	}

	_, err = f.forumUsecase.GetThread(slugOrID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		msg := models.Message{
			Text: fmt.Sprintf("Can't find post thread by id: %v", slugOrID),
		}

		_ = json.NewEncoder(w).Encode(msg)
		return
	}

	posts, err = f.forumUsecase.CreatePosts(slugOrID, posts)
	if err != nil {
		w.WriteHeader(http.StatusConflict)
		msg := models.Message{
			Text: "Parent post was created in another thread",
		}

		_ = json.NewEncoder(w).Encode(msg)
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
		w.WriteHeader(http.StatusNotFound)
		msg := models.Message{
			Text: fmt.Sprintf("Can't find thread by slug: %v", slugOrID),
		}

		_ = json.NewEncoder(w).Encode(msg)
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
		w.WriteHeader(http.StatusNotFound)
		msg := models.Message{
			Text: fmt.Sprintf("Can't find thread by slug: %v", slugOrID),
		}

		_ = json.NewEncoder(w).Encode(msg)
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

	err := f.forumUsecase.CheckThread(slugOrID)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusNotFound)
		msg := models.Message{
			Text: fmt.Sprintf("Can't find thread by slug: %v", slugOrID),
		}

		_ = json.NewEncoder(w).Encode(msg)
		return
	}

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

	err := f.forumUsecase.CheckThread(slugOrID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		msg := models.Message{
			Text: fmt.Sprintf("Can't find thread by slug: %v", slugOrID),
		}

		_ = json.NewEncoder(w).Encode(msg)
		return
	}

	thread := models.Thread{}
	err = json.NewDecoder(r.Body).Decode(&thread)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if thread.Title == "" && thread.Message == "" {
		thread, err = f.forumUsecase.GetThread(slugOrID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		thread, err = f.forumUsecase.UpdateThread(slugOrID, thread)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	err = json.NewEncoder(w).Encode(thread)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (f ForumDelivery) GetUsersFromForum(w http.ResponseWriter, r *http.Request) {
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

	limitParam := r.URL.Query().Get(configs.Limit)
	limit, err := strconv.Atoi(limitParam)
	if err != nil && limitParam != "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	descParam := r.URL.Query().Get(configs.Desc)
	if descParam == "" {
		descParam = "false"
	}

	sinceParam := r.URL.Query().Get(configs.Since)

	users, err := f.forumUsecase.GetUsersFromForum(slug, limit, sinceParam, descParam)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (f ForumDelivery) GetPostDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	id := mux.Vars(r)["id"]

	post, err := f.forumUsecase.GetPostDetails(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		msg := models.Message{
			Text: fmt.Sprintf("Can't find post with id: %v", id),
		}

		_ = json.NewEncoder(w).Encode(msg)
		return
	}

	postInfo := models.PostInfo{
		Post: post,
	}

	related := r.URL.Query().Get("related")
	if strings.Contains(related, "user") {
		author, err := f.userUsecase.Get(post.Author)
		if err != nil {
			msg := models.Message{
				Text: fmt.Sprintf("Can't find user with id #%v\n", post.Author),
			}

			w.WriteHeader(http.StatusNotFound)
			err = json.NewEncoder(w).Encode(msg)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			return
		}
		postInfo.Author = &author
	}

	if strings.Contains(related, "thread") {
		thread, err := f.forumUsecase.GetThread(strconv.Itoa(post.Thread))
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			msg := models.Message{
				Text: fmt.Sprintf("Can't find thread forum by slug: %v", post.Thread),
			}

			_ = json.NewEncoder(w).Encode(msg)
			return
		}
		postInfo.Thread = &thread
	}

	if strings.Contains(related, "forum") {
		forum, err := f.forumUsecase.Get(post.Forum)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			msg := models.Message{
				Text: fmt.Sprintf("Can't find forum by slug: %v", post.Forum),
			}

			_ = json.NewEncoder(w).Encode(msg)
			return
		}
		postInfo.Forum = &forum
	}

	err = json.NewEncoder(w).Encode(postInfo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (f ForumDelivery) UpdatePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	id := mux.Vars(r)["id"]

	idInt, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	post := models.Post{}
	err = json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if post.Message == "" {
		post, err := f.forumUsecase.GetPostDetails(id)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			msg := models.Message{
				Text: fmt.Sprintf("Can't find post with id: %v", id),
			}

			_ = json.NewEncoder(w).Encode(msg)
			return
		}

		err = json.NewEncoder(w).Encode(post)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}
	post.ID = idInt

	post, err = f.forumUsecase.UpdatePost(post)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		msg := models.Message{
			Text: fmt.Sprintf("Can't find post with id: %v", id),
		}

		_ = json.NewEncoder(w).Encode(msg)
		return
	}

	err = json.NewEncoder(w).Encode(post)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (f ForumDelivery) ClearService(w http.ResponseWriter, r *http.Request) {
	err := f.forumUsecase.ClearService()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (f ForumDelivery) GetServiceInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	info, err := f.forumUsecase.GetServiceInfo()
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(info)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
