package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/wise-hub/gateway/internal/model"
	"golang.org/x/crypto/bcrypt"
)


func GetUserDataFromDB(pool *pgxpool.Pool, userName, password string) (*model.UserData, error) {
	
	query := `SELECT user_id, username, password FROM app1.users WHERE username = $1;`

	var userID int
	var username, hashedPass string

	err := pool.QueryRow(context.Background(), query, userName).Scan(&userID, &username, &hashedPass)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(password))
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	return &model.UserData{
		UserID:    userID,
		Username:  username,
		Roles:     []string{"role1", "role2"},
		Accounts:  []string{"acc1", "acc2"},  
	}, nil
	
}
