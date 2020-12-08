package delivery

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/aanufriev/forum/internal/pkg/models"
	"github.com/aanufriev/forum/internal/pkg/user"
	"github.com/gorilla/mux"
)

type UserDelivery struct {
	userUsecase user.Usecase
}

func New(userUsecase user.Usecase) UserDelivery {
	return UserDelivery{
		userUsecase: userUsecase,
	}
}

func (u UserDelivery) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	nickname := mux.Vars(r)["nickname"]

	user := models.User{}
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user.Nickname = nickname

	err = u.userUsecase.Create(user)
	if err != nil {
		users, err := u.userUsecase.GetUsersWithNicknameAndEmail(nickname, *user.Email)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusConflict)
		err = json.NewEncoder(w).Encode(users)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (u UserDelivery) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	nickname := mux.Vars(r)["nickname"]

	profile, err := u.userUsecase.Get(nickname)
	if err != nil {

		if errors.Is(err, user.ErrUserDoesntExists) {
			w.WriteHeader(http.StatusNotFound)
			msg := models.ErrorMessage{
				Message: fmt.Sprintf("Can't find user with id #%v\n", nickname),
			}

			err = json.NewEncoder(w).Encode(msg)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(profile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (u UserDelivery) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	nickname := mux.Vars(r)["nickname"]

	profile := models.User{}
	err := json.NewDecoder(r.Body).Decode(&profile)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	profile.Nickname = nickname

	err = u.userUsecase.Update(profile)
	if err != nil {

		if errors.Is(err, user.ErrUserDoesntExists) {
			w.WriteHeader(http.StatusNotFound)
		} else if errors.Is(err, user.ErrDataConflict) {
			w.WriteHeader(http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		msg := models.ErrorMessage{
			Message: fmt.Sprintf("Can't find user with id #%v\n", nickname),
		}

		err = json.NewEncoder(w).Encode(msg)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	err = json.NewEncoder(w).Encode(profile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
