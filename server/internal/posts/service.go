package posts

import (
	"github.com/google/uuid"
	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/users"
	"github.com/kidommoc/gustrody/internal/utils"
)

type AttachImg struct {
	Type string `json:"type"`
	Url  string `json:"url"`
	Alt  string `json:"alt"`
}

func ToModelImg(i AttachImg) models.Img {
	return models.Img{
		Url: i.Url,
		Alt: i.Alt,
	}
}

type Post struct {
	ID          string          `json:"id"`
	Url         string          `json:"url"`
	User        *users.UserInfo `json:"user"`
	Date        string          `json:"date"`
	ReplyTo     *users.UserInfo `json:"replyTo,omitempty"`
	SharedBy    *users.UserInfo `json:"sharedBy,omitempty"`
	Visibility  string          `json:"visibility"`
	Content     string          `json:"content"`
	Likes       int64           `json:"likes"`
	Shares      int64           `json:"shares"`
	Attachments []AttachImg     `json:"attachments,omitempty"`
	Replyings   []*Post         `json:"replyings,omitempty"`
	Replies     []*Post         `json:"replies,omitempty"`
}

// services

type PostService struct {
	site             string
	maxContentLength int
	maxImgInPost     int
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
		maxImgInPost:     cfg.MaxImgInPost,
		db:               db,
		user:             us,
	}
}

func (service *PostService) newID() string {
	return uuid.New().String()
}

func (service *PostService) getUrl(id string) string {
	return service.site + "/posts/" + id
}

func (service *PostService) checkPermission(user string, target string, postID string, vsb utils.Vsb) bool {
	switch vsb {
	case utils.Vsb_FOLLOWER:
		if user == "" || (user != target && !service.user.IsFollowing(user, target)) {
			return false
		}
	case utils.Vsb_DIRECT:
		//
	}
	return true
}
