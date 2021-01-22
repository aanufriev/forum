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
	"github.com/valyala/fasthttp"
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

func (f ForumDelivery) Create(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")

	var forum models.Forum
	err := json.Unmarshal(ctx.PostBody(), &forum)
	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		return
	}

	nickname, err := f.userUsecase.CheckIfUserExists(forum.User)
	if err != nil {
		msg := models.Message{
			Text: fmt.Sprintf("Can't find user with id #%v\n", forum.User),
		}

		ctx.SetStatusCode(http.StatusNotFound)
		err = json.NewEncoder(ctx).Encode(msg)
		if err != nil {
			ctx.SetStatusCode(http.StatusInternalServerError)
			return
		}
		return
	}
	forum.User = nickname

	err = f.forumUsecase.Create(forum)
	if err != nil {
		existingForum, err := f.forumUsecase.Get(forum.Slug)
		if err != nil {
			ctx.SetStatusCode(http.StatusInternalServerError)
			return
		}

		ctx.SetStatusCode(http.StatusConflict)
		err = json.NewEncoder(ctx).Encode(existingForum)
		if err != nil {
			ctx.SetStatusCode(http.StatusInternalServerError)
			return
		}
		return
	}

	ctx.SetStatusCode(http.StatusCreated)
	err = json.NewEncoder(ctx).Encode(forum)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}
}

func (f ForumDelivery) Get(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")

	slug := ctx.UserValue("slug").(string)
	forum, err := f.forumUsecase.Get(slug)
	if err != nil {
		ctx.SetStatusCode(http.StatusNotFound)
		msg := models.Message{
			Text: fmt.Sprintf("Can't find forum with slug: %v", slug),
		}

		_ = json.NewEncoder(ctx).Encode(msg)
		return
	}

	err = json.NewEncoder(ctx).Encode(forum)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}
}

func (f ForumDelivery) CreateThread(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")

	slug := ctx.UserValue("slug").(string)

	thread := &models.Thread{}
	err := json.Unmarshal(ctx.PostBody(), thread)
	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		return
	}
	thread.Forum = slug

	nickname, err := f.userUsecase.CheckIfUserExists(thread.Author)
	if err != nil {
		msg := models.Message{
			Text: fmt.Sprintf("Can't find user with id #%v\n", thread.Author),
		}

		ctx.SetStatusCode(http.StatusNotFound)
		err = json.NewEncoder(ctx).Encode(msg)
		if err != nil {
			ctx.SetStatusCode(http.StatusInternalServerError)
			return
		}
		return
	}
	thread.Author = nickname

	err = f.forumUsecase.CreateThread(thread)
	if err != nil {
		if err == forum.ErrForumDoesntExists {
			ctx.SetStatusCode(http.StatusNotFound)
			msg := models.Message{
				Text: fmt.Sprintf("Can't find thread forum by slug: %v", thread.Forum),
			}

			_ = json.NewEncoder(ctx).Encode(msg)
			return
		}

		existedThread, err := f.forumUsecase.GetThread(*thread.Slug)
		if err != nil {
			ctx.SetStatusCode(http.StatusInternalServerError)
			return
		}

		ctx.SetStatusCode(http.StatusConflict)
		_ = json.NewEncoder(ctx).Encode(existedThread)
		return
	}

	ctx.SetStatusCode(http.StatusCreated)
	err = json.NewEncoder(ctx).Encode(thread)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}
}

func (f ForumDelivery) GetThreads(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")

	slug := ctx.UserValue("slug").(string)

	_, err := f.forumUsecase.CheckForum(slug)
	if err != nil {
		ctx.SetStatusCode(http.StatusNotFound)
		msg := models.Message{
			Text: fmt.Sprintf("Can't find forum by slug: %v", slug),
		}

		_ = json.NewEncoder(ctx).Encode(msg)
		return
	}

	limit := string(ctx.URI().QueryArgs().Peek(configs.Limit))
	desc := string(ctx.URI().QueryArgs().Peek(configs.Desc))
	since := string(ctx.URI().QueryArgs().Peek(configs.Since))

	threads, err := f.forumUsecase.GetThreads(slug, limit, since, desc)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(ctx).Encode(threads)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}
}

func (f ForumDelivery) CreatePosts(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")

	slugOrID := ctx.UserValue("slug_or_id").(string)

	thread, err := f.forumUsecase.GetThreadIDAndForum(slugOrID)
	if err != nil {
		ctx.SetStatusCode(http.StatusNotFound)
		msg := models.Message{
			Text: fmt.Sprintf("Can't find post thread by id: %v", slugOrID),
		}

		_ = json.NewEncoder(ctx).Encode(msg)
		return
	}

	posts := make([]models.Post, 0)
	err = json.Unmarshal(ctx.PostBody(), &posts)
	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		return
	}

	if len(posts) == 0 {
		ctx.SetStatusCode(http.StatusCreated)
		_ = json.NewEncoder(ctx).Encode(posts)
		return
	}

	err = f.forumUsecase.CreatePosts(thread, posts)
	if err != nil {
		if err == forum.ErrWrongParent {
			fmt.Printf("%v CREATE POSTS ERR: %v\n", time.Now().Format("02.01.2006 15:04:05"), err)
			ctx.SetStatusCode(http.StatusConflict)
			msg := models.Message{
				Text: "Parent post was created in another thread",
			}

			_ = json.NewEncoder(ctx).Encode(msg)
			return
		}
		for _, post := range posts {
			_, err = f.userUsecase.CheckIfUserExists(post.Author)
			if err != nil {
				fmt.Printf("%v CHECK USER ERR: %v\n", time.Now().Format("02.01.2006 15:04:05"), err)
				msg := models.Message{
					Text: fmt.Sprintf("Can't find post author by nickname: %v", post.Author),
				}

				ctx.SetStatusCode(http.StatusNotFound)
				err = json.NewEncoder(ctx).Encode(msg)
				if err != nil {
					ctx.SetStatusCode(http.StatusInternalServerError)
					return
				}
				return
			}
		}
		ctx.SetStatusCode(http.StatusConflict)
		msg := models.Message{
			Text: "Parent post was created in another thread",
		}

		_ = json.NewEncoder(ctx).Encode(msg)
		return
	}

	ctx.SetStatusCode(http.StatusCreated)
	_ = json.NewEncoder(ctx).Encode(posts)
}

func (f ForumDelivery) GetThread(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")

	slugOrID := ctx.UserValue("slug_or_id").(string)

	threads, err := f.forumUsecase.GetThread(slugOrID)
	if err != nil {
		ctx.SetStatusCode(http.StatusNotFound)
		msg := models.Message{
			Text: fmt.Sprintf("Can't find thread by slug: %v", slugOrID),
		}

		_ = json.NewEncoder(ctx).Encode(msg)
		return
	}

	err = json.NewEncoder(ctx).Encode(threads)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}
}

func (f ForumDelivery) Vote(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")

	slugOrID := ctx.UserValue("slug_or_id").(string)

	vote := models.Vote{}
	err := json.Unmarshal(ctx.PostBody(), &vote)
	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
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
		ctx.SetStatusCode(http.StatusNotFound)
		msg := models.Message{
			Text: fmt.Sprintf("Can't find thread by slug: %v", slugOrID),
		}

		_ = json.NewEncoder(ctx).Encode(msg)
		return
	}

	err = json.NewEncoder(ctx).Encode(thread)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}
}

func (f ForumDelivery) GetPosts(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")

	slugOrID := ctx.UserValue("slug_or_id").(string)

	err := f.forumUsecase.CheckThread(slugOrID)
	if err != nil {
		ctx.SetStatusCode(http.StatusNotFound)
		msg := models.Message{
			Text: fmt.Sprintf("Can't find thread by slug: %v", slugOrID),
		}

		_ = json.NewEncoder(ctx).Encode(msg)
		return
	}

	limitParam := string(ctx.URI().QueryArgs().Peek(configs.Limit))
	limit, err := strconv.Atoi(limitParam)
	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		return
	}

	sortParam := string(ctx.URI().QueryArgs().Peek(configs.Sort))
	descParam := string(ctx.URI().QueryArgs().Peek(configs.Desc))
	if descParam == "" {
		descParam = "false"
	}

	sinceParam := string(ctx.URI().QueryArgs().Peek(configs.Since))

	posts, err := f.forumUsecase.GetPosts(slugOrID, limit, sortParam, descParam, sinceParam)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(ctx).Encode(posts)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}
}

func (f ForumDelivery) UpdateThread(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")

	slugOrID := ctx.UserValue("slug_or_id").(string)

	err := f.forumUsecase.CheckThread(slugOrID)
	if err != nil {
		ctx.SetStatusCode(http.StatusNotFound)
		msg := models.Message{
			Text: fmt.Sprintf("Can't find thread by slug: %v", slugOrID),
		}

		_ = json.NewEncoder(ctx).Encode(msg)
		return
	}

	thread := models.Thread{}
	err = json.Unmarshal(ctx.PostBody(), &thread)
	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		return
	}

	if thread.Title == "" && thread.Message == "" {
		thread, err = f.forumUsecase.GetThread(slugOrID)
		if err != nil {
			ctx.SetStatusCode(http.StatusBadRequest)
			return
		}
	} else {
		thread, err = f.forumUsecase.UpdateThread(slugOrID, thread)
		if err != nil {
			ctx.SetStatusCode(http.StatusInternalServerError)
			return
		}
	}

	err = json.NewEncoder(ctx).Encode(thread)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}
}

func (f ForumDelivery) GetUsersFromForum(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")

	slug := ctx.UserValue("slug").(string)

	_, err := f.forumUsecase.CheckForum(slug)
	if err != nil {
		ctx.SetStatusCode(http.StatusNotFound)
		msg := models.Message{
			Text: fmt.Sprintf("Can't find forum by slug: %v", slug),
		}

		_ = json.NewEncoder(ctx).Encode(msg)
		return
	}

	limitParam := string(ctx.URI().QueryArgs().Peek(configs.Limit))
	limit, err := strconv.Atoi(limitParam)
	if err != nil && limitParam != "" {
		ctx.SetStatusCode(http.StatusBadRequest)
		return
	}

	descParam := string(ctx.URI().QueryArgs().Peek(configs.Desc))
	if descParam == "" {
		descParam = "false"
	}

	sinceParam := string(ctx.URI().QueryArgs().Peek(configs.Since))

	users, err := f.forumUsecase.GetUsersFromForum(slug, limit, sinceParam, descParam)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(ctx).Encode(users)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}
}

func (f ForumDelivery) GetPostDetails(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")

	id := ctx.UserValue("id").(string)

	post, err := f.forumUsecase.GetPostDetails(id)
	if err != nil {
		ctx.SetStatusCode(http.StatusNotFound)
		msg := models.Message{
			Text: fmt.Sprintf("Can't find post with id: %v", id),
		}

		_ = json.NewEncoder(ctx).Encode(msg)
		return
	}

	postInfo := models.PostInfo{
		Post: post,
	}

	related := string(ctx.URI().QueryArgs().Peek("related"))
	if strings.Contains(related, "user") {
		author, err := f.userUsecase.Get(post.Author)
		if err != nil {
			msg := models.Message{
				Text: fmt.Sprintf("Can't find user with id #%v\n", post.Author),
			}

			ctx.SetStatusCode(http.StatusNotFound)
			err = json.NewEncoder(ctx).Encode(msg)
			if err != nil {
				ctx.SetStatusCode(http.StatusInternalServerError)
				return
			}
			return
		}
		postInfo.Author = &author
	}

	if strings.Contains(related, "thread") {
		thread, err := f.forumUsecase.GetThread(strconv.Itoa(post.Thread))
		if err != nil {
			ctx.SetStatusCode(http.StatusNotFound)
			msg := models.Message{
				Text: fmt.Sprintf("Can't find thread forum by slug: %v", post.Thread),
			}

			_ = json.NewEncoder(ctx).Encode(msg)
			return
		}
		postInfo.Thread = &thread
	}

	if strings.Contains(related, "forum") {
		forum, err := f.forumUsecase.Get(post.Forum)
		if err != nil {
			ctx.SetStatusCode(http.StatusNotFound)
			msg := models.Message{
				Text: fmt.Sprintf("Can't find forum by slug: %v", post.Forum),
			}

			_ = json.NewEncoder(ctx).Encode(msg)
			return
		}
		postInfo.Forum = &forum
	}

	err = json.NewEncoder(ctx).Encode(postInfo)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}
}

func (f ForumDelivery) UpdatePost(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")

	id := ctx.UserValue("id").(string)

	idInt, err := strconv.Atoi(id)
	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		return
	}

	var post models.Post
	err = json.Unmarshal(ctx.PostBody(), &post)
	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		return
	}

	if post.Message == "" {
		post, err := f.forumUsecase.GetPostDetails(id)
		if err != nil {
			ctx.SetStatusCode(http.StatusNotFound)
			msg := models.Message{
				Text: fmt.Sprintf("Can't find post with id: %v", id),
			}

			_ = json.NewEncoder(ctx).Encode(msg)
			return
		}

		err = json.NewEncoder(ctx).Encode(post)
		if err != nil {
			ctx.SetStatusCode(http.StatusInternalServerError)
			return
		}
		return
	}
	post.ID = idInt

	post, err = f.forumUsecase.UpdatePost(post)
	if err != nil {
		ctx.SetStatusCode(http.StatusNotFound)
		msg := models.Message{
			Text: fmt.Sprintf("Can't find post with id: %v", id),
		}

		_ = json.NewEncoder(ctx).Encode(msg)
		return
	}

	err = json.NewEncoder(ctx).Encode(post)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}
}

func (f ForumDelivery) ClearService(ctx *fasthttp.RequestCtx) {
	err := f.forumUsecase.ClearService()
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}
}

func (f ForumDelivery) GetServiceInfo(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")

	info, err := f.forumUsecase.GetServiceInfo()
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(ctx).Encode(info)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}
}
