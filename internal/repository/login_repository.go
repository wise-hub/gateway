package repository

import (
	"context"
	"fmt"

	"fibank.bg/fis-gateway-ws/internal/model"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)


func GetUserDataFromDB(pool *pgxpool.Pool, userName, password string) (*model.UserData, error) {
	
	query := `
		SELECT user_id, username, password
		FROM app1.users 
		WHERE username = $1;
	`

	var userID int
	var username, hashedPass string

	// Execute the query
	err := pool.QueryRow(context.Background(), query, userName).Scan(&userID, &username, &hashedPass)
	if err != nil {
		return nil, fmt.Errorf("error fetching user data: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("error fetching user data: %w", err)
	}

	// Construct and return UserData without tokens
	return &model.UserData{
		UserID:    userID,
		Username:  username,
		Roles:     []string{"role1", "role2"}, // Replace with actual roles if needed
		Accounts:  []string{"acc1", "acc2"},  // Replace with actual accounts if needed
	}, nil
	
}
