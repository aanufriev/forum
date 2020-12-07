package delivery

import (
	"encoding/json"
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
