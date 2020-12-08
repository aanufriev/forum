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

func (u UserUsecase) Get(nickname string) (models.User, error) {
	return u.userRepository.Get(nickname)
}

func (u UserUsecase) GetUsersWithNicknameAndEmail(nickname, email string) ([]models.User, error) {
	return u.userRepository.GetUsersWithNicknameAndEmail(nickname, email)
}

func (u UserUsecase) Update(model models.User) error {
	return u.userRepository.Update(model)
}

func (u UserUsecase) CheckIfUserExists(nickname string) error {
	return u.userRepository.CheckIfUserExists(nickname)
}
