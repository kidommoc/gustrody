package db

import (
	"slices"
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
	ReplyTo string
	Replies []string
	Likes   []string
	Shares  []string
}

type Share struct {
	ID   string // origin post id
	User string
	Date int64
}

var postDb = make(map[string]*Post)
var shareDb = make([]*Share, 0, 100)

func initPostDb() {
	site := "127.0.0.1:8000"
	id := func() string {
		return site + "/posts/" + uuid.New().String()
	}
	SetPost(id(), "u1", "1:u1u1u1u1")
	time.Sleep(time.Second)
	tmp1 := id()
	SetPost(tmp1, "u2", "1:u2u2u2u2")
	time.Sleep(time.Second)
	SetShare("u1", tmp1)
	time.Sleep(time.Second)
	SetPost(id(), "u1", "2:u1u1u1u1")
	time.Sleep(time.Second)
	tmp2 := id()
	SetPost(tmp2, "u3", "1:u3u3u3u3")
	time.Sleep(time.Second)
	SetShare("u1", tmp2)
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

func QueryPostByID(id string) (post *Post, err utils.Err) {
	if postDb[id] == nil {
		return nil, utils.NewErr(ErrNotFound, "post")
	}
	p := *postDb[id]
	return &p, nil
}

func QueryPostsByUser(user string, asec bool) (l []*Post, err utils.Err) {
	l = make([]*Post, 0)
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
	return l, nil
}

func QuerySharesByUser(user string, asec bool) (l []*Share, err utils.Err) {
	l = make([]*Share, 0)
	for _, v := range shareDb {
		if v.User == user {
			s := *v
			l = append(l, &s)
		}
	}
	sort.Slice(l, func(i, j int) bool {
		if asec {
			return l[i].Date < l[j].Date
		} else {
			return l[i].Date > l[j].Date
		}
	})
	return l, nil
}

func SetPost(id string, user string, content string) utils.Err {
	p := &Post{
		ID:      id,
		User:    user,
		Date:    time.Now().Unix(),
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

func SetLike(user string, id string) utils.Err {
	p := postDb[id]
	if p == nil {
		return utils.NewErr(ErrNotFound, "post")
	}
	for _, v := range p.Likes {
		if v == user {
			return nil
		}
	}
	p.Likes = append(p.Likes, user)
	return nil
}

func RemoveLike(user string, id string) utils.Err {
	p := postDb[id]
	if p == nil {
		return utils.NewErr(ErrNotFound, "post")
	}
	for i, v := range p.Likes {
		if v == user {
			p.Likes = slices.Delete(p.Likes, i, i+1)
			return nil
		}
	}
	return utils.NewErr(ErrNotFound, "like")
}

func SetShare(user string, id string) utils.Err {
	p := postDb[id]
	if p == nil {
		return utils.NewErr(ErrNotFound, "post")
	}
	for _, v := range p.Shares {
		if v == user {
			return nil
		}
	}
	p.Shares = append(p.Likes, user)
	shareDb = append(shareDb, &Share{
		ID:   id,
		User: user,
		Date: time.Now().Unix(),
	})
	return nil
}

func RemoveShare(user string, id string) utils.Err {
	p := postDb[id]
	if p == nil {
		return utils.NewErr(ErrNotFound, "post")
	}

	// use 2 annoymous func to ensure completely deletion

	err1 := func() utils.Err {
		for i, v := range shareDb {
			if v.ID == id && v.User == user {
				shareDb = slices.Delete(shareDb, i, i+1)
				return nil
			}
		}
		return utils.NewErr(ErrNotFound)
	}()

	err2 := func() utils.Err {
		for i, v := range p.Shares {
			if v == user {
				p.Shares = slices.Delete(p.Shares, i, i+1)
				return nil
			}
		}
		return utils.NewErr(ErrNotFound)
	}()

	if err1 != nil || err2 != nil {
		return utils.NewErr(ErrNotFound, "share")
	} else {
		return nil
	}
}
