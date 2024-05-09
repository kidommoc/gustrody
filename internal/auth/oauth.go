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

func VerifyToken(token string, session string) (username string, err error) {
	parsed, err := jwt.Parse(token, func(tok *jwt.Token) (interface{}, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("jwt method error")
		}
		return []byte("penguin"), nil // should loaded from .env
	})

	// error
	if !parsed.Valid {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return "", errors.New("token expired")
		default:
			return "", errors.New("invalid token")
		}
	}

	username, err = parsed.Claims.GetIssuer()
	if err != nil {
		return "", errors.New("invalid token")
	}
	sess, err := parsed.Claims.GetSubject()
	if err != nil {
		return "", errors.New("invalid token")
	}

	if sess != session {
		return "", errors.New("unmatched session")
	}
	return username, nil
}

func Login(username string, password string) (session string, oauth *OauthToken, err error) {
	p, err := db.GetPasswordOfUser(username)
	if err != nil {
		// handle error: user not found
		return "", nil, err
	}
	if p != password {
		// handle error: incorrect password
		return "", nil, errors.New("incorrect password")
	}
	session = generateSession()
	db.SetSession(session, username)
	oauth = &OauthToken{
		Token:   generateToken(username, session, TOKEN_EXPIRE),
		Refresh: generateToken(username, session, REFRESH_EXPIRE),
	}
	return session, oauth, nil
}

func RefreshToken(session string, refresh string) (oauth *OauthToken, err error) {
	/* validate */
	username, err := VerifyToken(refresh, session)
	if err != nil {
		return nil, err
	}
	oauth = &OauthToken{
		Token:   generateToken(username, session, TOKEN_EXPIRE),
		Refresh: generateToken(username, session, REFRESH_EXPIRE),
	}
	return oauth, nil
}
