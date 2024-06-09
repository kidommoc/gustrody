package posts

import (
	"github.com/google/uuid"
	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/database"
	"github.com/kidommoc/gustrody/internal/users"
)

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

// services

type PostService struct {
	site             string
	maxContentLength int
	db               database.IPostDb
	user             *users.UserService
}

func NewService(db database.IPostDb, us *users.UserService, c ...config.EnvConfig) *PostService {
	var cfg config.EnvConfig
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

func (service *PostService) fullID(id string) string {
	return service.site + "/posts/" + id
}
