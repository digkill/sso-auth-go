package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/digkill/sso-auth-go/internal/domain/models"
	"github.com/digkill/sso-auth-go/internal/lib/jwt"
	"github.com/digkill/sso-auth-go/internal/lib/logger/sl"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type UserStorage interface {
	SaveUser(ctx context.Context, email string, password []byte) (uid int64, err error)
	User(ctx context.Context, email string) (models.User, error)
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		password []byte,
	) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}

func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration) *Auth {
	return &Auth{
		userSaver:    userSaver,
		userProvider: userProvider,
		log:          log,
		appProvider:  appProvider,
		tokenTTL:     tokenTTL, // Lifetime return token
	}
}

// RegisterNewUser registers new user in the system and returns user ID.
// If user with given username already exists, returns error.
func (a *Auth) RegisterNewUser(ctx context.Context, email string, pass string) (int64, error) {
	// method - имя текущей функции и пакета. Такую метку удобно
	// добавлять в логи и в текст ошибок, чтобы легче было искать хвосты
	// в случае поломок.
	const method = "Auth.RegisterNewUser"

	// Создаём локальный объект логгера с доп. полями, содержащими полезную инфу
	// о текущем вызове функции
	log := a.log.With(
		slog.String("method", method),
		slog.String("email", email),
	)

	log.Info("Registering new user")

	// Generated hash and salt for password.
	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Failed to generate password hash", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", method, err)
	}

	// User save DB
	id, err := a.userSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		log.Error("Failed to save new user", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", method, err)
	}

	return id, nil
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Login checks if user with given credentials exists in the system and returns access token.
//
// If user exists, but password is incorrect, returns error.
// If user doesn't exist, returns error.
func (a *Auth) Login(ctx context.Context,
	email string,
	password string,
	appID int,
) (string, error) {
	const method = "Auth.Login"

	log := a.log.With(
		slog.String("method", method),
		slog.String("username", email),
		// password либо не логируем, либо логируем в замаскированном виде
	)

	log.Info("attempting to login user")

	user, err := a.userProvider.User(ctx, email)
	if err != nil {

		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("User not found", sl.Err(err))
		}

		a.log.Error("Failed to get user", sl.Err(err))

		return "", fmt.Errorf("%s: %w", method, err)
	}

	// Information get about app
	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", method, err)
	}

	log.Info("user logged in successfully")

	// Create token authorization
	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("Failed to create token", sl.Err(err))

		return "", fmt.Errorf("%s: %w", method, err)
	}

	return token, nil
}
