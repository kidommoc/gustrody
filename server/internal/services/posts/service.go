package posts

import (
	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/services/users"
	"github.com/kidommoc/gustrody/internal/utils"
)

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

type AttachImg struct {
	Type string `json:"type"`
	Url  string `json:"url"`
	Alt  string `json:"alt"`
}

func ToModelImg(i AttachImg) models.Img {
	return models.Img{
		Type: i.Type,
		Url:  i.Url,
		Alt:  i.Alt,
	}
}

// services

type PostDbs struct {
	Query models.IPostQuery
	Set   models.IPostSet
	Like  models.IPostLike
	Share models.IPostShare
}

type PostService struct {
	lg               logging.Logger
	site             string
	maxContentLength int
	maxImgInPost     int
	db               PostDbs
	user             *users.UserService
}

func NewService(us *users.UserService, dbs PostDbs, cfg config.Config, lg logging.Logger) *PostService {
	return &PostService{
		lg:               lg,
		site:             cfg.Site,
		maxContentLength: cfg.MaxContentLength,
		maxImgInPost:     cfg.MaxImgInPost,
		db:               dbs,
		user:             us,
	}
}

func (service *PostService) checkPermission(user string, post *models.Post) bool {
	switch post.Vsb {
	case utils.Vsb_FOLLOWER:
		target := post.User.String()
		if user == "" || (user != target && !service.user.IsFollowing(user, target)) {
			return false
		}
	case utils.Vsb_DIRECT:
		// not implement yet
	}
	return true
}
