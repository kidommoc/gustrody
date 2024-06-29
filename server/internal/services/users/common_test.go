package users

import (
	"crypto/rsa"
	"testing"

	"github.com/kidommoc/gustrody/internal/models"
)

type uDb struct {
	t    *testing.T
	data map[string]*models.User
}

func newUdb(t *testing.T) *uDb {
	return &uDb{t, make(map[string]*models.User)}
}

// Account DB

type MockingAccountDb struct {
	data *uDb
}

func newMockingAccountDb(u *uDb) *MockingAccountDb {
	return &MockingAccountDb{u}
}

func (db *MockingAccountDb) SetUser(user *models.User) error {
	db.data.data[user.Username] = user
	return nil
}

// will never use
func (db *MockingAccountDb) QueryUserKeys(username string) (pub *rsa.PublicKey, pri *rsa.PrivateKey, err error) {
	return nil, nil, nil
}

func (db *MockingAccountDb) QueryUserPreferences(username string) (pf *models.Preferences, err error) {
	if db.data.data[username] == nil {
		db.data.t.Error("user is nil")
		return nil, nil
	}
	return &db.data.data[username].Preferences, nil
}

func (db *MockingAccountDb) UpdateUserPreferences(username string, pf *models.Preferences) error {
	if db.data.data[username] == nil {
		db.data.t.Error("user is nil")
		return nil
	}
	db.data.data[username].Preferences = *pf
	return nil
}

// Info DB

type MockingInfoDb struct {
	data *uDb
}

func newMockingInfoDb(u *uDb) *MockingInfoDb {
	return &MockingInfoDb{u}
}

// for simplicity, always true
func (db *MockingInfoDb) IsUserExist(username string) bool {
	return true
}

func (db *MockingInfoDb) QueryUser(username string) (user models.User, err error) {
	return models.User{}, nil
}

func (db *MockingInfoDb) UpdateUser(user *models.User) error {
	return nil
}

// Auth DB

type MockingAuthDb struct {
	data map[string]string
}

func newMockingAuthDb() *MockingAuthDb {
	return &MockingAuthDb{make(map[string]string)}
}

func (db *MockingAuthDb) QueryPasswordOfUser(username string) (password string, err error) {
	return db.data[username], nil
}

func (db *MockingAuthDb) SetUserPassword(username string, password string) error {
	db.data[username] = password
	return nil
}
