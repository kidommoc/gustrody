package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/utils"
)

// token and session

const TOKEN_EXPIRE = 5             // 5 hours
const REFRESH_EXPIRE = 5 * 24 * 14 // 14 days

type OauthToken struct {
	Token   string `json:"token"`
	Refresh string `json:"refresh"`
}

func NewOauth(u string, s string) *OauthToken {
	return &OauthToken{
		Token:   generateToken(u, s, TOKEN_EXPIRE),
		Refresh: generateToken(u, s, REFRESH_EXPIRE),
	}
}

func generateSession() string {
	return utils.GenerateRamdonHexString(32)
}

func generateToken(u string, s string, exp uint) string {
	now := time.Now()
	key := []byte("penguin") // should loaded from .env

	tokenExpired := now.Add(time.Duration(exp) * time.Hour)
	claims := &jwt.RegisteredClaims{
		ID:        utils.GenerateRamdonHexString(16),
		Issuer:    u,
		Subject:   s,
		ExpiresAt: jwt.NewNumericDate(tokenExpired),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(key)
	if err != nil {
		// handle err
	}

	return signed
}

// oauth service

type OauthService struct {
	lg logging.Logger
	db models.IAuthDb
}

func NewService(db models.IAuthDb, lg logging.Logger) *OauthService {
	return &OauthService{
		lg: lg,
		db: db,
	}
}

func (service *OauthService) VerifyToken(token, session string) (username string, err error) {
	parsed, e := jwt.Parse(token, func(tok *jwt.Token) (interface{}, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("jwt method error")
		}
		return []byte("penguin"), nil // should loaded from .env
	})

	if !parsed.Valid {
		switch {
		case errors.Is(e, jwt.ErrTokenExpired):
			return "", ErrExpired
		default:
			return "", ErrInvalid
		}
	}

	username, e = parsed.Claims.GetIssuer()
	if e != nil {
		return "", ErrInvalid
	}
	sess, e := parsed.Claims.GetSubject()
	if e != nil {
		return "", ErrInvalid
	}

	if sess != session {
		return "", ErrWrongSession
	}
	return username, nil
}

func (service *OauthService) Login(username, password string) (session string, oauth OauthToken, err error) {
	p, err := service.db.QueryPasswordOfUser(username)
	if err != nil {
		return "", oauth, ErrUserNotFound
	}
	if p != password {
		return "", oauth, ErrWrongPassword
	}
	session = generateSession()
	// service.db.SetSession(session, username)
	return session, *NewOauth(username, session), nil
}

func (service *OauthService) RefreshToken(session, refresh string) (oauth OauthToken, err error) {
	/* validate */
	username, err := service.VerifyToken(refresh, session)
	if err != nil {
		return oauth, err
	}
	return *NewOauth(username, session), nil
}
