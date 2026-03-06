package repositories

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"myapp/models"
)

var ErrEmailExists = errors.New("email already exists")

func CreateUser(db *sql.DB, name, email, passwordHash string) (models.User, error) {
	var u models.User

	err := db.QueryRow(`
		INSERT INTO users (name, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, name, email
	`, name, email, passwordHash).Scan(&u.ID, &u.Name, &u.Email)

	if err != nil {
		// Postgres unique violation code = 23505
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return models.User{}, ErrEmailExists
		}
		return models.User{}, err
	}

	return u, nil
}

func GetUserByEmail(db *sql.DB, email string) (models.User, error) {
	var u models.User

	err := db.QueryRow(`
		SELECT id, name, email, password
		FROM users
		WHERE email = $1
	`, email).Scan(&u.ID, &u.Name, &u.Email, &u.Password)

	return u, err
}