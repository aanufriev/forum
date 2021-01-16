package delivery

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/aanufriev/forum/internal/pkg/models"
	"github.com/aanufriev/forum/internal/pkg/user"
	"github.com/valyala/fasthttp"
)

type UserDelivery struct {
	userUsecase user.Usecase
}

func New(userUsecase user.Usecase) UserDelivery {
	return UserDelivery{
		userUsecase: userUsecase,
	}
}

func (u UserDelivery) Create(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")
	nickname := ctx.UserValue("nickname").(string)

	var user models.User
	err := user.UnmarshalJSON(ctx.PostBody())
	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		return
	}
	user.Nickname = nickname

	err = u.userUsecase.Create(user)
	if err != nil {
		users, err := u.userUsecase.GetUsersWithNicknameAndEmail(nickname, *user.Email)
		if err != nil {
			ctx.SetStatusCode(http.StatusInternalServerError)
			return
		}

		ctx.SetStatusCode(http.StatusConflict)
		err = json.NewEncoder(ctx).Encode(users)
		if err != nil {
			ctx.SetStatusCode(http.StatusInternalServerError)
			return
		}
		return
	}

	ctx.SetStatusCode(http.StatusCreated)
	_ = json.NewEncoder(ctx).Encode(user)
}

func (u UserDelivery) Get(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")
	nickname := ctx.UserValue("nickname").(string)

	profile, err := u.userUsecase.Get(nickname)
	if err != nil {

		if errors.Is(err, user.ErrUserDoesntExists) {
			ctx.SetStatusCode(http.StatusNotFound)
			msg := models.Message{
				Text: fmt.Sprintf("Can't find user with id #%v\n", nickname),
			}

			err = json.NewEncoder(ctx).Encode(msg)
			if err != nil {
				ctx.SetStatusCode(http.StatusInternalServerError)
				return
			}
			return
		}

		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(ctx).Encode(profile)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}
}

func (u UserDelivery) Update(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")
	nickname := ctx.UserValue("nickname").(string)

	profile := models.User{}
	err := profile.UnmarshalJSON(ctx.PostBody())
	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		return
	}
	profile.Nickname = nickname

	fullProfile, err := u.userUsecase.Update(profile)
	if err != nil {
		var msg models.Message
		if errors.Is(err, user.ErrUserDoesntExists) {
			ctx.SetStatusCode(http.StatusNotFound)
			msg = models.Message{
				Text: fmt.Sprintf("Can't find user with id #%v\n", nickname),
			}
		} else if errors.Is(err, user.ErrDataConflict) {
			emailOwnerNickname, err := u.userUsecase.GetUserNicknameWithEmail(*profile.Email)
			if err != nil {
				ctx.SetStatusCode(http.StatusInternalServerError)
				return
			}

			ctx.SetStatusCode(http.StatusConflict)
			msg = models.Message{
				Text: fmt.Sprintf("This email is already registered by user: %v", emailOwnerNickname),
			}
		} else {
			ctx.SetStatusCode(http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(ctx).Encode(msg)
		if err != nil {
			ctx.SetStatusCode(http.StatusInternalServerError)
			return
		}
		return
	}

	err = json.NewEncoder(ctx).Encode(fullProfile)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}
}
