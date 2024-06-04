package posts

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/kidommoc/gustrody/internal/db"
	"github.com/kidommoc/gustrody/internal/users"
	"github.com/kidommoc/gustrody/internal/utils"
)

// should load from .env
var maxContentLength = 1000

type Post struct {
	ID          string         `json:"id"`
	User        users.UserInfo `json:"user"`
	PublishedAt string         `json:"publishedAt"`
	Content     string         `json:"content"`
}

func Get(postId string) (post Post, err utils.Err) {
	result, e := db.QueryPostByID(postId)
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
	post.User = user
	post.PublishedAt = time.Unix(result.Date, 0).Format(time.RFC3339)
	post.Content = result.Content
	return post, nil
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
	result, e := db.QueryPostsByUser(username, false)
	if e != nil {
		switch {
		// default:
		}
	}

	fmt.Println(len(result))
	for _, p := range result {
		fmt.Println(*p)
		post := new(Post)
		post.ID = p.ID
		post.User = user
		post.PublishedAt = time.Unix(p.Date, 0).Format(time.RFC3339)
		post.Content = p.Content
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
	if e := db.SetPost(username, content); e != nil {
		switch {
		// default:
		}
	}
	return nil
}

func Edit(username string, postId string, content string) utils.Err {
	if content == "" {
		return utils.NewErr(ErrContent, "empty")
	}
	if utf8.RuneCountInString(content) > maxContentLength {
		return utils.NewErr(ErrContent, "long")
	}
	owner, e := db.QueryPostOwner(postId)
	if e != nil {
		switch {
		case e.Code() == db.ErrNotFound && e.Error() == "post":
			return utils.NewErr(ErrPostNotFound)
			// default:
		}
	}
	if owner != username {
		return utils.NewErr(ErrOwner)
	}
	if e := db.UpdatePost(postId, content); e != nil {
		switch {
		// default:
		}
	}
	return nil
}

func Remove(username string, postId string) utils.Err {
	owner, e := db.QueryPostOwner(postId)
	if e != nil {
		switch {
		case e.Code() == db.ErrNotFound && e.Error() == "post":
			return utils.NewErr(ErrPostNotFound)
			// default:
		}
	}
	if owner != username {
		return utils.NewErr(ErrOwner)
	}
	if e := db.RemovePost(postId); e != nil {
		switch {
		// default:
		}
	}
	return nil
}
