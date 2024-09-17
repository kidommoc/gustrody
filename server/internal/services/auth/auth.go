package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/utils"
)

var ErrInvalid = errors.New("Invalid")
var ErrExpired = errors.New("Expired")
var ErrWrongSession = errors.New("WrongSession")
var ErrWrongPassword = errors.New("WrongPassword")
var ErrUserNotFound = errors.New("UserNotFound")

// token and session

const TOKEN_EXPIRE = 5             // 5 hours
const REFRESH_EXPIRE = 5 * 24 * 14 // 14 days

func generateSession() string {
	return utils.GenerateRamdonHexString(32)
}

func generateToken(u string, s string, exp uint, k string) string {
	now := time.Now()

	tokenExpired := now.Add(time.Duration(exp) * time.Hour)
	claims := &jwt.RegisteredClaims{
		ID:        utils.GenerateRamdonHexString(16),
		Issuer:    u,
		Subject:   s,
		ExpiresAt: jwt.NewNumericDate(tokenExpired),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(k)
	if err != nil {
		// handle err
	}

	return signed
}

// oauth service

type OauthService struct {
	lg  logging.Logger
	db  models.IUserAccount
	key string
}

func NewService(db models.IUserAccount, lg logging.Logger, key string) *OauthService {
	return &OauthService{
		lg:  lg,
		db:  db,
		key: key,
	}
}

func (service *OauthService) VerifyToken(token, session string) (username string, left time.Duration, err error) {
	parsed, e := jwt.Parse(token, func(tok *jwt.Token) (interface{}, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("jwt method error")
		}
		return []byte(service.key), nil // should loaded from .env
	})

	if !parsed.Valid {
		switch {
		case errors.Is(e, jwt.ErrTokenExpired):
			return "", 0, ErrExpired
		default:
			return "", 0, ErrInvalid
		}
	}

	username, e = parsed.Claims.GetIssuer()
	if e != nil {
		return "", 0, ErrInvalid
	}
	sess, e := parsed.Claims.GetSubject()
	if e != nil {
		return "", 0, ErrInvalid
	}

	if sess != session {
		return "", 0, ErrWrongSession
	}
	exp, e := parsed.Claims.GetExpirationTime()
	if e != nil {
		return "", 0, ErrInvalid
	}
	left = exp.Time.Sub(time.Now())
	return username, left, nil
}

func (service *OauthService) Login(username, password string) (
	session string,
	token string, refresh string,
	err error,
) {
	p, err := service.db.QueryPassword(username)
	if err != nil {
		return "", "", "", ErrUserNotFound
	}
	password = string(utils.SHA256Hash(password))
	if p != password {
		return "", "", "", ErrWrongPassword
	}
	session = generateSession()
	// service.db.SetSession(session, username)
	token = generateToken(username, session, TOKEN_EXPIRE, service.key)
	refresh = generateToken(username, session, TOKEN_EXPIRE, service.key)
	return session, token, refresh, nil
}

func (service *OauthService) RefreshToken(refresh, session string) (token, nre string, err error) {
	username, left, err := service.VerifyToken(refresh, session)
	if err != nil {
		return "", "", err
	}
	token = generateToken(username, session, TOKEN_EXPIRE, service.key)
	if left < REFRESH_EXPIRE/2*time.Hour {
		nre = generateToken(username, session, TOKEN_EXPIRE, service.key)
	}
	return token, refresh, nil
}
