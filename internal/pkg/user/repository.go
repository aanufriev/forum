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
	Update(user models.User) (models.User, error)
	CheckIfUserExists(nickname string) (string, error)
	GetUserNicknameWithEmail(email string) (string, error)
	GetUserIDByNickname(nickname string) (int, error)
}
