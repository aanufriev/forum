package user

import "github.com/aanufriev/forum/internal/pkg/models"

type Repository interface {
	Create(user models.User) error
	GetUsersWithNicknameAndEmail(nickname, email string) ([]models.User, error)
}
