package jwt

import (
	"github.com/digkill/sso-auth-go/internal/domain/models"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// NewToken creates new JWT token for given user and app.
// test need
func NewToken(user models.User, app models.App, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	// Information adding in token
	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(duration).Unix()
	claims["app_id"] = app.ID

	// Token subscription, use secret key app
	tokenString, err := token.SignedString([]byte(app.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
