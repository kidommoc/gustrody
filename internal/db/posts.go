package db

import (
	"time"

	"github.com/google/uuid"
	"github.com/kidommoc/gustrody/internal/utils"
)

type Post struct {
	ID      string
	User    string
	Date    string // should use timestamp
	Content string
}

var postDb = make(map[string]*Post)

func initPostDb() {
}

func checkPost(id string) bool {
	if postDb[id] != nil {
		return true
	} else {
		return false
	}
}

func checkPostOwner(id string, user string) bool {
	if p := postDb[id]; p != nil && p.User == user {
		return true
	} else {
		return false
	}
}

func now() string {
	return time.Now().Format(time.RFC822)
}

func QueryPostOwner(id string) (username string, err utils.Err) {
	post := postDb[id]
	if post == nil {
		return "", utils.NewErr(ErrNotFound, "post")
	}
	return post.User, nil
}

func QueryPostByID(id string) (post *Post, err utils.Err) {
	post = postDb[id]
	if post == nil {
		return nil, utils.NewErr(ErrNotFound, "post")
	}
	return post, nil
}

func QueryPostsByUser(user string) (list []*Post, err utils.Err) {
	for _, v := range postDb {
		if v.User == user {
			list = append(list, v)
		}
	}
	return list, err
}

func SetPost(user string, content string) utils.Err {
	id := uuid.New().String()
	date := now()
	p := &Post{
		ID:      id,
		User:    user,
		Date:    date,
		Content: content,
	}
	postDb[id] = p
	return nil
}

func UpdatePost(id string, content string) utils.Err {
	p := postDb[id]
	if p == nil {
		return utils.NewErr(ErrNotFound, "post")
	}
	p.Content = content
	p.Date = now()
	return nil
}

func RemovePost(id string) utils.Err {
	p := postDb[id]
	if p == nil {
		return utils.NewErr(ErrNotFound, "post")
	}
	postDb[id] = nil
	return nil
}
