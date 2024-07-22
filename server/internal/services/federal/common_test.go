package federal

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"slices"

	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/models"
)

const localServerAddr = "127.0.0.1:7999"

type mockingUserAccountDb struct {
	data map[string]struct {
		pub string
		pri string
	}
}

func newMockingUserAccountDb() *mockingUserAccountDb {
	return &mockingUserAccountDb{
		data: make(map[string]struct {
			pub string
			pri string
		}),
	}
}

// never used
func (db *mockingUserAccountDb) SetUser(user *models.User) error {
	return nil
}

func (db *mockingUserAccountDb) QueryUserKeys(username string) (pub string, pri string, err error) {
	return db.data[username].pub, db.data[username].pri, nil
}

// never used
func (db *mockingUserAccountDb) QueryUserPreferences(username string) (pf *models.Preferences, err error) {
	return nil, nil
}

// never used
func (db *mockingUserAccountDb) UpdateUserPreferences(username string, pf *models.Preferences) error {
	return nil
}

type mockingUserInfoDb struct {
}

func newMockingUserInfoDb() *mockingUserInfoDb {
	return &mockingUserInfoDb{}
}

func (db *mockingUserInfoDb) IsUserExist(username string) bool {
	return true
}

// never used
func (db *mockingUserInfoDb) QueryUser(username string) (user models.User, err error) {
	return
}

// never used
func (db *mockingUserInfoDb) UpdateUser(user *models.User) error {
	return nil
}

type mockingUserFollowDb struct {
	sig chan bool
}

func newMockingUserFollowDb(sig chan bool) *mockingUserFollowDb {
	return &mockingUserFollowDb{sig: sig}
}

// never used
func (db *mockingUserFollowDb) IsFollowing(username, target models.UD) bool {
	return true
}

// never used
func (db *mockingUserFollowDb) QueryUserFollowInfo(username string) (follows int64, followed int64, err error) {
	return
}

// never used
func (db *mockingUserFollowDb) QueryUserFollowings(username string) (list []models.UD, err error) {
	return
}

// never used
func (db *mockingUserFollowDb) QueryUserFollowers(username string) (list []models.UD, err error) {
	return
}

func (db *mockingUserFollowDb) SetFollow(from, to models.UD) error {
	db.sig <- true
	return nil
}

func (db *mockingUserFollowDb) RemoveFollow(from, to models.UD) error {
	db.sig <- true
	return nil
}

type mockingUserForeignDb struct {
	data map[models.UD]models.ForeignUser
}

func newMockingUserForeignDb() *mockingUserForeignDb {
	return &mockingUserForeignDb{
		data: make(map[models.UD]models.ForeignUser, 5),
	}
}

func (db *mockingUserForeignDb) IsForeignExist(username models.UD) bool {
	return db.data[username].ID != ""
}

func (db *mockingUserForeignDb) GetForeignUserByUD(username models.UD) (user models.ForeignUser, err error) {
	if db.data[username].ID == "" {
		return user, models.ErrNotFound
	}
	return db.data[username], nil
}

func (db *mockingUserForeignDb) GetForeignUserByID(id string) (user models.ForeignUser, err error) {
	for _, v := range db.data {
		if v.ID == id {
			return v, nil
		}
	}
	return user, models.ErrNotFound
}

func (db *mockingUserForeignDb) SetForeignUser(user *models.ForeignUser) error {
	if user.Username == models.NewUD("") {
		return models.ErrFormat
	}
	db.data[user.Username] = *user
	return nil
}

func (db *mockingUserForeignDb) GetInboxes(usernames []models.UD) (inboxes []string, err error) {
	inboxes = make([]string, 0, 5)
	for k, v := range db.data {
		if slices.Contains(usernames, k) && !slices.Contains(inboxes, v.Inbox) {
			inboxes = append(inboxes, v.Inbox)
		}
	}
	return
}

type mockingFileService struct {
	scheme string
	domain string
}

func newMockingFileService(cfg config.Config) *mockingFileService {
	return &mockingFileService{scheme: cfg.Scheme, domain: cfg.Domain}
}

func (service *mockingFileService) StoreImage(buf []byte) (url string, mediaType string, err error) {
	h := sha256.New()
	h.Write(buf)
	filename := hex.EncodeToString(h.Sum(nil))

	var ext string
	buffer := bytes.NewBuffer(buf)
	_, t, e := image.Decode(buffer)
	if e != nil {
		return "", "", errors.New("cannot decode file to image")
	}
	switch t {
	case "jpg":
		ext = "jpeg"
	case "jpeg":
		ext = "jpeg"
	case "png":
		ext = "png"
	default:
		return "", "", errors.New("wrong file type: " + t)
	}
	url = fmt.Sprintf("%s://%s/img/%s.%s", service.scheme, service.domain, filename, ext)
	return url, "image/" + ext, nil
}
