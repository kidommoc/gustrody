package posts

import (
	"github.com/google/uuid"
	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/users"
)

type Post struct {
	ID          string          `json:"id"`
	User        *users.UserInfo `json:"user"`
	ReplyTo     *users.UserInfo `json:"replyTo,omitempty"`
	SharedBy    *users.UserInfo `json:"sharedBy,omitempty"`
	PublishedAt string          `json:"publishedAt"`
	Content     string          `json:"content"`
	Likes       int64           `json:"likes"`
	Shares      int64           `json:"shares"`
	Attachments []string        `json:"attachments"`
	Replyings   []*Post         `json:"replyings,omitempty"`
	Replies     []*Post         `json:"replies,omitempty"`
}

// services

type PostService struct {
	site             string
	maxContentLength int
	db               models.IPostDb
	user             *users.UserService
}

func NewService(db models.IPostDb, us *users.UserService, c ...config.Config) *PostService {
	var cfg config.Config
	if len(c) == 0 {
		cfg = config.Get()
	} else {
		cfg = c[0]
	}
	return &PostService{
		site:             cfg.Site,
		maxContentLength: cfg.MaxContentLength,
		db:               db,
		user:             us,
	}
}

func (service *PostService) newID() string {
	return service.site + "/posts/" + uuid.New().String()
}

func (service *PostService) fullID(short string) string {
	return service.site + "/posts/" + short
}
