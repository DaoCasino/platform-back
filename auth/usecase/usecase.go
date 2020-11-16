package usecase

import (
	"context"
	"encoding/hex"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/machinebox/graphql"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/sha3"
	"platform-backend/auth"
	"platform-backend/contracts"
	"platform-backend/models"
	"platform-backend/server/session_manager"
	"strconv"
	"time"
)

type AuthUseCase struct {
	userRepo        auth.UserRepository
	smRepo          session_manager.Repository
	contractUC      contracts.UseCase
	jwtSecret       []byte
	refreshTokenTTL int64
	accessTokenTTL  int64

	walletGqClient     *graphql.Client
	walletClientId     int64
	walletClientSecret string
}

func NewAuthUseCase(userRepo auth.UserRepository, smRepo session_manager.Repository, contractUC contracts.UseCase,
	jwtSecret []byte, accessTokenTTL int64, refreshTokenTTL int64,
	walletUrl string, walletClientId int64, walletClientSecret string) *AuthUseCase {
	return &AuthUseCase{
		userRepo:        userRepo,
		smRepo:          smRepo,
		contractUC:      contractUC,
		jwtSecret:       jwtSecret,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,

		walletGqClient:     graphql.NewClient(walletUrl),
		walletClientId:     walletClientId,
		walletClientSecret: walletClientSecret,
	}
}

func (a *AuthUseCase) ResolveUser(ctx context.Context, tmpToken string) (*models.User, error) {
	request := graphql.NewRequest(`
		mutation TokenValidate($token: String!, $client_id: Int!, $sign: String!) {
		  	tokenValidate(data: { key: $token client_id: $client_id }, sign: $sign) {
			  	result attachment user { email ref_token kyc_status account_name }
		  	}
	  	}
	`)

	request.Var("token", tmpToken)
	request.Var("client_id", a.walletClientId)

	// wallet require data hash salted with secret
	hash := sha3.NewLegacyKeccak256()
	strForHash := strconv.FormatInt(a.walletClientId, 10) + tmpToken + a.walletClientSecret
	_, err := hash.Write([]byte(strForHash))
	if err != nil {
		return nil, err
	}

	request.Var("sign", hex.EncodeToString(hash.Sum(nil)))

	response := &struct {
		TokenValidate struct {
			User struct {
				ID          int64  `json:"Id"`
				Email       string `json:"email"`
				AccountName string `json:"account_name"`
				RefToken    string `json:"ref_token"`
				KycStatus   string `json:"kyc_status"`
			} `json:"user"`
		} `json:"tokenValidate"`
	}{}

	err = a.walletGqClient.Run(ctx, request, response)
	if err != nil {
		log.Debug().Msgf("TokenValidate request error: %s", err.Error())
		return nil, err
	}

	if response.TokenValidate.User.AccountName == "" {
		return nil, errors.New("got empty account name from wallet")
	}

	if response.TokenValidate.User.Email == "" {
		return nil, errors.New("got empty email from wallet")
	}

	return &models.User{
		AccountName: response.TokenValidate.User.AccountName,
		Email:       response.TokenValidate.User.Email,
		AffiliateID: "",
	}, nil
}

func (a *AuthUseCase) SignUp(ctx context.Context, user *models.User, casinoName string) (string, string, error) {
	hasUser, err := a.userRepo.HasUser(ctx, user.AccountName)
	if err != nil {
		log.Debug().Msgf("User existing check error: %s", err.Error())
		return "", "", err
	}
	if !hasUser {
		if err := a.userRepo.AddUser(ctx, user); err != nil {
			log.Debug().Msgf("User add error: %s", err.Error())
			return "", "", err
		}

		if err := a.contractUC.SendNewPlayerToCasino(ctx, user.AccountName, casinoName); err != nil {
			log.Debug().Msgf("Send new player to casino error: %s", err.Error())
			return "", "", err
		}
	}

	hasEmail, err := a.userRepo.HasEmail(ctx, user.AccountName)
	if err != nil {
		log.Debug().Msgf("User email existing check error: %s", err.Error())
		return "", "", err
	}
	if !hasEmail {
		if err := a.userRepo.AddEmail(ctx, user); err != nil {
			log.Debug().Msgf("User email add error: %s", err.Error())
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

	claims := refreshToken.Claims.(jwt.MapClaims)
	err = a.userRepo.InvalidateSession(ctx, claims["account_name"].(string), int64(claims["nonce"].(float64)))
	if err != nil {
		return "", "", err
	}

	return a.generateTokens(ctx, refreshToken.Claims.(jwt.MapClaims)["account_name"].(string))
}

func (a *AuthUseCase) generateTokens(ctx context.Context, accountName string) (string, string, error) {
	newNonce, err := a.userRepo.AddNewSession(ctx, accountName)
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

func (a *AuthUseCase) Logout(ctx context.Context, accessToken string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	token, err := a.parseToken(accessToken)
	if err != nil {
		return err
	}
	if err := a.validateAccessToken(ctx, token); err != nil {
		return err
	}
	claims := token.Claims.(jwt.MapClaims)
	return a.userRepo.InvalidateSession(ctx, claims["account_name"].(string), int64(claims["nonce"].(float64)))
}

func (a *AuthUseCase) OptOut(ctx context.Context, accessToken string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	token, err := a.parseToken(accessToken)
	if err != nil {
		return err
	}
	if err := a.validateAccessToken(ctx, token); err != nil {
		return err
	}
	claims := token.Claims.(jwt.MapClaims)
	return a.userRepo.DeleteEmail(ctx, claims["account_name"].(string))
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

	active, err := a.userRepo.IsSessionActive(ctx, claims["account_name"].(string), int64(claims["nonce"].(float64)))
	if err != nil {
		return auth.ErrInvalidToken
	}

	if !active {
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
