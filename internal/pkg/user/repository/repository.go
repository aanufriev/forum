package repository

import (
	"database/sql"
	"fmt"

	"github.com/aanufriev/forum/internal/pkg/models"
	"github.com/aanufriev/forum/internal/pkg/user"
)

type UserRepository struct {
	db *sql.DB
}

func New(db *sql.DB) user.Repository {
	return UserRepository{
		db: db,
	}
}

func (u UserRepository) Create(model models.User) error {
	_, err := u.db.Exec(
		"INSERT INTO users (nickname, fullname, email, about) VALUES ($1, $2, $3, $4)",
		model.Nickname, model.Fullname, model.Email, model.About,
	)

	if err != nil {
		return fmt.Errorf("couldn't insert user: %v. Error: %w", model, err)
	}

	return nil
}

func (u UserRepository) Get(nickname string) (models.User, error) {
	var model models.User
	err := u.db.QueryRow(
		`SELECT id, nickname, fullname, email, about FROM users
		WHERE nickname = $1`,
		nickname,
	).Scan(&model.ID, &model.Nickname, &model.Fullname, &model.Email, &model.About)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, user.ErrUserDoesntExists
		}
		return models.User{}, fmt.Errorf("couldn't get user with nickname '%v'. Error: %w", nickname, err)
	}

	return model, nil
}

func (u UserRepository) GetUsersWithNicknameAndEmail(nickname, email string) ([]models.User, error) {
	rows, err := u.db.Query(
		`SELECT nickname, fullname, email, about FROM users
		WHERE nickname = $1 OR email = $2`,
		nickname, email,
	)
	if err != nil {
		return nil, fmt.Errorf(`couldn't get users with nickname '%v' and email '%v'. Error: %w`, nickname, email, err)
	}
	defer rows.Close()

	users := make([]models.User, 0, 2)
	user := models.User{}
	for rows.Next() {
		err = rows.Scan(&user.Nickname, &user.Fullname, &user.Email, &user.About)
		if err != nil {
			return nil, fmt.Errorf(`couldn't get users with nickname '%v' and email '%v'. Error: %w`, nickname, email, err)
		}

		users = append(users, user)
	}

	return users, nil
}

func (u UserRepository) Update(model models.User) (models.User, error) {
	userFromDB, err := u.Get(model.Nickname)
	if err != nil {
		return models.User{}, err
	}

	if model.Fullname == nil {
		model.Fullname = userFromDB.Fullname
	}

	if model.Email == nil {
		model.Email = userFromDB.Email
	}

	if model.About == nil {
		model.About = userFromDB.About
	}

	_, err = u.db.Exec(
		`UPDATE users SET fullname = $1, email = $2, about = $3
		WHERE id = $4`,
		model.Fullname, model.Email, model.About, userFromDB.ID,
	)

	if err != nil {
		if err != sql.ErrConnDone {
			return models.User{}, user.ErrDataConflict
		}
		return models.User{}, err
	}

	return model, nil
}

func (u UserRepository) CheckIfUserExists(nickname string) (string, error) {
	err := u.db.QueryRow(
		"SELECT nickname FROM users WHERE nickname = $1",
		nickname,
	).Scan(&nickname)

	if err != nil {
		return "", fmt.Errorf("user doesnt exist: %w", err)
	}

	return nickname, nil
}

func (u UserRepository) GetUserNicknameWithEmail(email string) (string, error) {
	var nickname string
	err := u.db.QueryRow(
		"SELECT nickname FROM users WHERE email = $1",
		email,
	).Scan(&nickname)

	if err != nil {
		return "", fmt.Errorf(`couldn't get user nickname with email '%v'. Error: %w`, email, err)
	}

	return nickname, nil
}

func (u UserRepository) GetUserIDByNickname(nickname string) (int, error) {
	var id int
	err := u.db.QueryRow(
		"SELECT id FROM users WHERE nickname = $1",
		nickname,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}
