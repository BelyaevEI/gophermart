package userepository

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"time"

	"github.com/BelyaevEI/gophermart/internal/database"
	"github.com/BelyaevEI/gophermart/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	db *database.Database
}

func New(db *database.Database) *User {
	return &User{
		db: db,
	}
}

// Get user from db for authorization
func (u *User) GetUser(login string) models.User {

	return models.User{}
}

// Check unique login
func (u *User) CheckUserLogin(ctx context.Context, login string) bool {
	err := u.db.CheckUserLogin(ctx, login)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false
	}
	return true
}

// Save user in system
func (u *User) SaveUser(ctx context.Context, login, password, key, token string) error {

	// Generate hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = u.db.SaveUser(ctx, login, string(hashedPassword), key)
	if err != nil {
		return err
	}

	err = u.db.SaveToken(ctx, token, key)
	if err != nil {
		return err
	}
	return nil
}

// Check user password
func (u *User) Authorization(ctx context.Context, login, password string) error {

	var user models.User

	row := u.db.GetUser(ctx, login)
	if err := row.Scan(&user.Login, &user.Password, &user.SecretKey); err != nil {
		return err
	}

	return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
}

func (u *User) GetKey(ctx context.Context, token string) (string, error) {
	var key string

	row := u.db.GetKey(ctx, token)
	if err := row.Scan(&key); err != nil {
		return "", err
	}
	return key, nil
}

func (u *User) GetAccrualUser(ctx context.Context, token string) (float64, error) {
	accrual, err := u.db.GetAccrualUser(ctx, token)
	if err != nil {
		return 0, err
	}

	return accrual, nil
}

func (u *User) GetWithdrawnUser(ctx context.Context, token string) (float64, error) {
	withdrawn, err := u.db.GetWithdrawnUser(ctx, token)
	if err != nil {
		return 0, err
	}
	return withdrawn, nil
}

func (u *User) UpdateWithdrawn(ctx context.Context, token string, sum float64, order int) error {
	return u.db.UpdateWithdrawn(ctx, token, sum, order)
}

func (u *User) GetUSerScore(ctx context.Context, token string) (float64, error) {
	score, err := u.db.GetUSerScore(ctx, token)
	if err != nil {
		return 0, err
	}
	return score, nil
}

func (u *User) UpdateScores(ctx context.Context, token string, score float64) error {
	return u.db.UpdateScores(ctx, token, score)
}

func (u *User) GetAllWithdrawn(ctx context.Context, token string) ([]models.AllWithdrawn, error) {
	withdrawns, err := u.db.GetAllWithdrawn(ctx, token)
	if err != nil {
		return withdrawns, err
	}

	// Sort list descending
	sort.SliceStable(withdrawns, func(i, j int) bool {
		timeI, _ := time.Parse("2006-01-02T15:04:05Z07:00", withdrawns[i].Processed)
		timeJ, _ := time.Parse("2006-01-02T15:04:05Z07:00", withdrawns[j].Processed)
		return timeI.Before(timeJ)
	})

	return withdrawns, nil

}
