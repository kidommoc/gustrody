package posts

import (
	"github.com/google/uuid"
	"github.com/kidommoc/gustrody/internal/database"
	"github.com/kidommoc/gustrody/internal/users"
)

// should load from .env
var maxContentLength = 1000
var site = "127.0.0.1:8000"

type Post struct {
	ID          string          `json:"id"`
	User        *users.UserInfo `json:"user"`
	SharedBy    *users.UserInfo `json:"sharedBy,omitempty"`
	PublishedAt string          `json:"publishedAt"`
	Content     string          `json:"content"`
	Likes       int             `json:"likes"`
	Shares      int             `json:"shares"`
	Replyings   []*Post         `json:"replyings,omitempty"`
	Replies     []*Post         `json:"replies,omitempty"`
}

func newID() string {
	return site + "/posts/" + uuid.New().String()
}

func fullID(id string) string {
	return site + "/posts/" + id
}

// services

type PostService struct {
	db   database.IPostDb
	user *users.UserService
}

func NewService(db database.IPostDb, us *users.UserService) *PostService {
	return &PostService{
		db:   db,
		user: us,
	}
}
