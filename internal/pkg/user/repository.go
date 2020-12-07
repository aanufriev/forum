package user

import (
	"fmt"

	"github.com/aanufriev/forum/internal/pkg/models"
)

var (
	ErrUserDoesntExists = fmt.Errorf("user exists")
	ErrDataConflict     = fmt.Errorf("data conflict")
)

type Repository interface {
	Create(user models.User) error
	Get(nickname string) (models.User, error)
	GetUsersWithNicknameAndEmail(nickname, email string) ([]models.User, error)
	Update(user models.User) error
}
