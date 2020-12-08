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
		`SELECT nickname, fullname, email, about FROM users
		WHERE nickname = $1`,
		nickname,
	).Scan(&model.Nickname, &model.Fullname, &model.Email, &model.About)

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
		WHERE lower(nickname) = lower($1) OR lower(email) = lower($2)`,
		nickname, email,
	)
	if err != nil {
		return nil, fmt.Errorf(`couldn't get users with nickname '%v' and email '%v'. Error: %w`, nickname, email, err)
	}
	defer rows.Close()

	users := make([]models.User, 0, 5)
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

func (u UserRepository) Update(model models.User) error {
	result, err := u.db.Exec(
		`UPDATE users SET fullname = $1, email = $2, about = $3
		WHERE nickname = $4`,
		model.Fullname, model.Email, model.About, model.Nickname,
	)

	if err != nil {
		if err != sql.ErrConnDone {
			return user.ErrDataConflict
		}
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return user.ErrUserDoesntExists
	}

	return nil
}
