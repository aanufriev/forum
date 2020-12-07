package user

import (
	"fmt"

	"github.com/aanufriev/forum/internal/pkg/models"
)

var (
	ErrUserDoesntExists = fmt.Errorf("user exists")
)

type Repository interface {
	Create(user models.User) error
	Get(nickname string) (models.User, error)
	GetUsersWithNicknameAndEmail(nickname, email string) ([]models.User, error)
}
