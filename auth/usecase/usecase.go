package usecase

import (
	"context"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/rs/zerolog/log"
	"platform-backend/auth"
	"platform-backend/models"
)

type AuthUseCase struct {
	userRepo auth.UserRepository
	jwtSecret []byte
}

func NewAuthUseCase(userRepo auth.UserRepository, jwtSecret []byte) *AuthUseCase {
	return &AuthUseCase{
		userRepo: userRepo,
		jwtSecret: jwtSecret,
	}
}

func (a *AuthUseCase) SignUp(ctx context.Context, user *models.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = jwt.MapClaims{"account_name": user.AccountName}

	signed, err := token.SignedString(a.jwtSecret)
	if err != nil {
		log.Debug().Msgf("JWT token sign error, %s", err.Error())
		return "", err
	}

	hasUser, err := a.userRepo.HasUser(context.Background(), user.AccountName)
	if err != nil {
		log.Debug().Msgf("User existing check error, %s", err.Error())
		return "", err
	}
	if !hasUser {
		if err := a.userRepo.AddUser(ctx, user); err != nil {
			return "", err
		}
	}

	return signed, nil
}

func (a *AuthUseCase) ParseToken(ctx context.Context, accessToken string) (*models.User, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return a.jwtSecret, nil
	})

	if err != nil {
		return nil, auth.ErrInvalidAccessToken
	}

	claimsMap := token.Claims.(jwt.MapClaims)
	if _, ok := claimsMap["account_name"]; !ok {
		return nil, auth.ErrInvalidAccessToken
	}

	user, err := a.userRepo.GetUser(ctx, claimsMap["account_name"].(string))
	if err != nil {
		return nil, auth.ErrUserNotFound
	}

	return user, nil
}
