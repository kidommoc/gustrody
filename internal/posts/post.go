package posts

import (
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/kidommoc/gustrody/internal/db"
	"github.com/kidommoc/gustrody/internal/users"
	"github.com/kidommoc/gustrody/internal/utils"
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

func makePost(p *db.Post) (post Post, err utils.Err) {
	u, e := users.GetInfo(p.User)
	if e != nil {
		return post, utils.NewErr(ErrUserNotFound)
	}

	post = Post{
		ID:          p.ID,
		User:        &u,
		PublishedAt: time.Unix(p.Date, 0).Format(time.RFC3339),
		Content:     p.Content,
		Likes:       len(p.Likes),
		Shares:      len(p.Shares),
	}

	return post, nil
}

func Get(postID string) (post Post, err utils.Err) {
	result, e := db.QueryPostByID(fullID(postID))
	if e != nil {
		switch {
		case e.Code() == db.ErrNotFound && e.Error() == "post":
			return post, utils.NewErr(ErrPostNotFound)
			// default:
		}
	}

	post, e = makePost(&result)
	if e != nil {
		switch {
		case e.Code() == db.ErrNotFound && e.Error() == "user":
			return post, utils.NewErr(ErrOwner)
			// default:
		}
	}
	setReplies(&post)

	return post, nil
}

func GetLikes(postID string) (list []*users.UserInfo, err utils.Err) {
	result, e := db.QueryPostByID(fullID(postID))
	if e != nil {
		switch {
		case e.Code() == db.ErrNotFound && e.Error() == "post":
			return list, utils.NewErr(ErrPostNotFound)
			// default:
		}
	}

	list = make([]*users.UserInfo, 0, len(result.Likes))
	for _, u := range result.Likes {
		info, e := users.GetInfo(u)
		if e == nil {
			list = append(list, &info)
		}
	}

	return list, nil
}

func GetShares(postID string) (list []*users.UserInfo, err utils.Err) {
	result, e := db.QueryPostByID(fullID(postID))
	if e != nil {
		switch {
		case e.Code() == db.ErrNotFound && e.Error() == "post":
			return list, utils.NewErr(ErrPostNotFound)
			// default:
		}
	}

	list = make([]*users.UserInfo, 0, len(result.Shares))
	for _, u := range result.Shares {
		info, e := users.GetInfo(u)
		if e == nil {
			list = append(list, &info)
		}
	}

	return list, nil
}

func GetByUser(username string) (list []*Post, err utils.Err) {
	user, e := users.GetInfo(username)
	if e != nil {
		switch {
		case e.Code() == db.ErrNotFound && e.Error() == "user":
			return list, utils.NewErr(ErrUserNotFound)
			// default:
		}
	}

	posts, _ := db.QueryPostsByUser(username, true) // ascending by date
	shares, _ := db.QuerySharesByUser(username, true)
	iP := len(posts) - 1 // start from end (newest)
	iS := len(shares) - 1

	for iP >= 0 || iS >= 0 {
		var p *db.Post
		var flag bool // true: post, false: share
		var actor *users.UserInfo
		var sharedBy *users.UserInfo

		if iP < 0 {
			flag = false
		} else if iS < 0 {
			flag = true
		} else {
			if posts[iP].Date > shares[iS].Date {
				flag = true
			} else {
				flag = false
			}
		}

		if flag {
			p = posts[iP]
			actor = &user
			sharedBy = nil
			iP = iP - 1
		} else {
			shared, e := db.QueryPostByID(shares[iS].ID)
			if e != nil {
				continue
			}
			p = &shared
			u, e := users.GetInfo(p.User)
			if e != nil {
				continue
			}
			actor = &u
			sharedBy = &user
			iS = iS - 1
		}

		post, _ := makePost(p)
		post.User = actor
		post.SharedBy = sharedBy
		list = append(list, &post)
	}

	return list, nil
}

func New(username string, content string) utils.Err {
	if !db.IsUserExist(username) {
		return utils.NewErr(ErrUserNotFound)
	}
	if content == "" {
		return utils.NewErr(ErrContent, "empty")
	}
	if utf8.RuneCountInString(content) > maxContentLength {
		return utils.NewErr(ErrContent, "long")
	}

	id := newID()
	for db.IsPostExsit(id) {
		id = newID()
	}
	// note: should not return any error here
	db.SetPost(id, username, content)

	return nil
}

func Edit(username string, postID string, content string) utils.Err {
	if content == "" {
		return utils.NewErr(ErrContent, "empty")
	}
	if utf8.RuneCountInString(content) > maxContentLength {
		return utils.NewErr(ErrContent, "too long")
	}

	post, e := db.QueryPostByID(fullID(postID))
	if e != nil {
		switch {
		case e.Code() == db.ErrNotFound && e.Error() == "post":
			return utils.NewErr(ErrPostNotFound)
			// default:
		}
	}
	if post.User != username {
		return utils.NewErr(ErrOwner)
	}

	if e := db.UpdatePost(fullID(postID), content); e != nil {
		switch {
		// default:
		}
	}

	return nil
}

func Remove(username string, postID string) utils.Err {
	post, e := db.QueryPostByID(fullID(postID))
	if e != nil {
		switch {
		case e.Code() == db.ErrNotFound && e.Error() == "post":
			return utils.NewErr(ErrPostNotFound)
			// default:
		}
	}
	if post.User != username {
		return utils.NewErr(ErrOwner)
	}

	if e := db.RemovePost(fullID(postID)); e != nil {
		switch {
		// default:
		}
	}

	return nil
}
