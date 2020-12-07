package user

import "github.com/aanufriev/forum/internal/pkg/models"

type Usecase interface {
	Create(user models.User) error
	Get(nickname string) (models.User, error)
	GetUsersWithNicknameAndEmail(nickname, email string) ([]models.User, error)
}
