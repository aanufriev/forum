package usecase

import (
	"github.com/aanufriev/forum/internal/pkg/models"
	"github.com/aanufriev/forum/internal/pkg/user"
)

type UserUsecase struct {
	userRepository user.Repository
}

func New(userRepository user.Repository) user.Usecase {
	return UserUsecase{
		userRepository: userRepository,
	}
}

func (u UserUsecase) Create(model models.User) error {
	return u.userRepository.Create(model)
}

func (u UserUsecase) GetUsersWithNicknameAndEmail(nickname, email string) ([]models.User, error) {
	return u.userRepository.GetUsersWithNicknameAndEmail(nickname, email)
}
