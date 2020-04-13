package usecase

import (
	"context"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"platform-backend/auth"
	"platform-backend/models"
	"platform-backend/server/session_manager"
	"time"
)

type AuthUseCase struct {
	userRepo        auth.UserRepository
	smRepo          session_manager.Repository
	jwtSecret       []byte
	refreshTokenTTL int64
	accessTokenTTL  int64
}

func NewAuthUseCase(userRepo auth.UserRepository, smRepo session_manager.Repository,
	jwtSecret []byte, accessTokenTTL int64, refreshTokenTTL int64) *AuthUseCase {
	return &AuthUseCase{
		userRepo:        userRepo,
		smRepo:          smRepo,
		jwtSecret:       jwtSecret,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (a *AuthUseCase) SignUp(ctx context.Context, user *models.User) (string, string, error) {
	hasUser, err := a.userRepo.HasUser(context.Background(), user.AccountName)
	if err != nil {
		log.Debug().Msgf("User existing check error, %s", err.Error())
		return "", "", err
	}
	if !hasUser {
		if err := a.userRepo.AddUser(ctx, user); err != nil {
			return "", "", err
		}
	}

	return a.generateTokens(ctx, user.AccountName)
}

func (a *AuthUseCase) SignIn(ctx context.Context, accessToken string) (*models.User, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	token, err := a.parseToken(accessToken)
	if err != nil {
		return nil, err
	}
	if err := a.validateAccessToken(ctx, token); err != nil {
		return nil, err
	}

	user, err := a.userRepo.GetUser(ctx, token.Claims.(jwt.MapClaims)["account_name"].(string))
	if err != nil {
		return nil, auth.ErrUserNotFound
	}

	suid := ctx.Value("suid")
	if suid == nil {
		return nil, auth.ErrSessionNotFound
	}

	err = a.smRepo.SetUser(suid.(uuid.UUID), user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (a *AuthUseCase) RefreshToken(ctx context.Context, refreshTokenStr string) (string, string, error) {
	refreshToken, err := a.parseToken(refreshTokenStr)
	if err != nil {
		return "", "", err
	}

	err = a.validateRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", err
	}

	return a.generateTokens(ctx, refreshToken.Claims.(jwt.MapClaims)["account_name"].(string))
}

func (a *AuthUseCase) generateTokens(ctx context.Context, accountName string) (string, string, error) {
	err := a.userRepo.UpdateTokenNonce(ctx, accountName)
	if err != nil {
		return "", "", err
	}

	newNonce, err := a.userRepo.GetTokenNonce(ctx, accountName)
	if err != nil {
		return "", "", err
	}

	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshToken.Claims = jwt.MapClaims{
		"account_name": accountName,
		"iat":          time.Now().Unix(),
		"exp":          time.Now().Unix() + a.refreshTokenTTL,
		"nonce":        newNonce,
		"type":         "refresh",
	}

	accessToken := jwt.New(jwt.SigningMethodHS256)
	accessToken.Claims = jwt.MapClaims{
		"account_name": accountName,
		"iat":          time.Now().Unix(),
		"exp":          time.Now().Unix() + a.accessTokenTTL,
		"nonce":        newNonce,
		"type":         "access",
	}

	signedRefresh, err := refreshToken.SignedString(a.jwtSecret)
	if err != nil {
		return "", "", err
	}

	signedAccess, err := accessToken.SignedString(a.jwtSecret)
	if err != nil {
		return "", "", err
	}

	return signedRefresh, signedAccess, nil
}

func (a *AuthUseCase) validateRefreshToken(ctx context.Context, token *jwt.Token) error {
	if err := a.validateToken(ctx, token); err != nil {
		return err
	}

	if err := a.validateTokenType(token, "refresh"); err != nil {
		return err
	}

	return nil
}

func (a *AuthUseCase) validateAccessToken(ctx context.Context, token *jwt.Token) error {
	if err := a.validateToken(ctx, token); err != nil {
		return err
	}

	if err := a.validateTokenType(token, "access"); err != nil {
		return err
	}

	return nil
}

func (a *AuthUseCase) validateToken(ctx context.Context, token *jwt.Token) error {
	claims := token.Claims.(jwt.MapClaims)

	if _, ok := claims["account_name"]; !ok {
		return auth.ErrInvalidToken
	}
	if _, ok := claims["account_name"].(string); !ok {
		return auth.ErrInvalidToken
	}
	if _, ok := claims["type"]; !ok {
		return auth.ErrInvalidToken
	}
	if _, ok := claims["type"].(string); !ok {
		return auth.ErrInvalidToken
	}
	if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
		return auth.ErrExpiredToken
	}
	if !claims.VerifyIssuedAt(time.Now().Unix(), true) {
		return auth.ErrInvalidToken
	}
	if _, ok := claims["nonce"]; !ok {
		return auth.ErrInvalidToken
	}
	if _, ok := claims["nonce"].(float64); !ok {
		return auth.ErrInvalidToken
	}

	nonce, err := a.userRepo.GetTokenNonce(ctx, claims["account_name"].(string))
	if err != nil {
		return auth.ErrInvalidToken
	}

	tokenNonce := int64(claims["nonce"].(float64))
	if nonce > tokenNonce {
		return auth.ErrExpiredTokenNonce
	}

	return nil
}

func (a *AuthUseCase) validateTokenType(token *jwt.Token, reqType string) error {
	claims := token.Claims.(jwt.MapClaims)
	if claims["type"].(string) != reqType {
		return auth.ErrInvalidToken
	}
	return nil
}

func (a *AuthUseCase) parseToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid sign method")
		}
		return a.jwtSecret, nil
	})

	if err != nil {
		log.Debug().Msgf("Token parse error: %s", err.Error())
		return nil, auth.ErrCannotParseToken
	}

	return token, nil
}
