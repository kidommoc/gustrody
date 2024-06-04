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
}

func newID() string {
	return site + "/posts/" + uuid.New().String()
}

func fullID(id string) string {
	return site + "/posts/" + id
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
	user, e := users.GetInfo(result.User)
	if e != nil {
		switch {
		case e.Code() == db.ErrNotFound && e.Error() == "user":
			return post, utils.NewErr(ErrOwner)
			// default:
		}
	}

	post.ID = result.ID
	post.User = &user
	post.PublishedAt = time.Unix(result.Date, 0).Format(time.RFC3339)
	post.Content = result.Content
	post.Likes = len(result.Likes)
	post.Shares = len(result.Shares)
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
		post := new(Post)
		var p *db.Post
		var flag bool // true: post, false: share

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
			post.User = &user
			post.SharedBy = nil
			iP = iP - 1
		} else {
			p, e = db.QueryPostByID(shares[iS].ID)
			if e != nil {
				continue
			}
			u, e := users.GetInfo(p.User)
			if e != nil {
				continue
			}
			post.User = &u
			post.SharedBy = &user
			iS = iS - 1
		}

		post.ID = p.ID
		post.PublishedAt = time.Unix(p.Date, 0).Format(time.RFC3339)
		post.Content = p.Content
		post.Likes = len(p.Likes)
		post.Shares = len(p.Shares)
		list = append(list, post)
	}
	return list, nil
}

func New(username string, content string) utils.Err {
	if _, e := db.QueryUser(username); e != nil {
		switch {
		case e.Code() == db.ErrNotFound && e.Error() == "user":
			return utils.NewErr(ErrUserNotFound)
		}
	}
	if content == "" {
		return utils.NewErr(ErrContent, "empty")
	}
	if utf8.RuneCountInString(content) > maxContentLength {
		return utils.NewErr(ErrContent, "long")
	}
	if e := db.SetPost(newID(), username, content); e != nil {
		switch {
		// default:
		}
	}
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
