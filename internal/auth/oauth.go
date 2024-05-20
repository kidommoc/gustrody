package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kidommoc/gustrody/internal/db"
	"github.com/kidommoc/gustrody/internal/utils"
)

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

func VerifyToken(token string, session string) (username string, err utils.Err) {
	parsed, e := jwt.Parse(token, func(tok *jwt.Token) (interface{}, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("jwt method error")
		}
		return []byte("penguin"), nil // should loaded from .env
	})

	if !parsed.Valid {
		switch {
		case errors.Is(e, jwt.ErrTokenExpired):
			return "", utils.NewErr(ErrExpired)
		default:
			return "", utils.NewErr(ErrInvalid)
		}
	}

	username, e = parsed.Claims.GetIssuer()
	if e != nil {
		return "", utils.NewErr(ErrInvalid)
	}
	sess, e := parsed.Claims.GetSubject()
	if e != nil {
		return "", utils.NewErr(ErrInvalid)
	}

	if sess != session {
		return "", utils.NewErr(ErrWrongSession)
	}
	return username, nil
}

func Login(username string, password string) (session string, oauth OauthToken, err utils.Err) {
	p, err := db.QueryPasswordOfUser(username)
	if err != nil {
		return "", oauth, utils.NewErr(ErrUserNotFound)
	}
	if p != password {
		return "", oauth, utils.NewErr(ErrWrongPassword)
	}
	session = generateSession()
	db.SetSession(session, username)
	return session, *NewOauth(username, session), nil
}

func RefreshToken(session string, refresh string) (oauth OauthToken, err utils.Err) {
	/* validate */
	username, err := VerifyToken(refresh, session)
	if err != nil {
		return oauth, err
	}
	return *NewOauth(username, session), nil
}
