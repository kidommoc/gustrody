package db

import (
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/kidommoc/gustrody/internal/utils"
)

type Post struct {
	ID      string
	User    string
	Date    int64
	Content string
}

var postDb = make(map[string]*Post)

// should load from .env
var site = "localhost:8000"

func initPostDb() {
	SetPost("u1", "1:u1u1u1u1")
	time.Sleep(time.Second)
	SetPost("u2", "1:u2u2u2u2")
	time.Sleep(time.Second)
	SetPost("u1", "2:u1u1u1u1")
	time.Sleep(time.Second)
	SetPost("u3", "1:u3u3u3u3")
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
	if postDb[id] == nil {
		return nil, utils.NewErr(ErrNotFound, "post")
	}
	p := *postDb[id]
	return &p, nil
}

func QueryPostsByUser(user string, asec bool) (l []*Post, err utils.Err) {
	for _, v := range postDb {
		if v.User == user {
			p := *v
			l = append(l, &p)
		}
	}
	sort.Slice(l, func(i, j int) bool {
		if asec {
			return l[i].Date < l[j].Date
		} else {
			return l[i].Date > l[j].Date
		}
	})
	return l, err
}

func SetPost(user string, content string) utils.Err {
	id := site + "/posts/" + uuid.New().String()
	date := time.Now()
	p := &Post{
		ID:      id,
		User:    user,
		Date:    date.Unix(),
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
	p.Date = time.Now().Unix()
	return nil
}

func RemovePost(id string) utils.Err {
	p := postDb[id]
	if p == nil {
		return utils.NewErr(ErrNotFound, "post")
	}
	delete(postDb, id)
	return nil
}
