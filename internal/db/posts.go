package db

import (
	"slices"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/kidommoc/gustrody/internal/utils"
)

type Post struct {
	ID       string   `json:"id"`
	User     string   `json:"user"`
	Date     int64    `json:"date"`
	Content  string   `json:"content"`
	Replying string   `json:"replying"` // post id
	Likes    []string `json:"likes"`    // user id
	Shares   []string `json:"shares"`   // user id
	Level    int      `json:"level"`    // temporary field, used in replying and replies
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
	tmp1 := id()
	SetPost(tmp1, "u1", "p:u1-1")
	time.Sleep(time.Second)
	tmp2 := id()
	SetPost(tmp2, "u2", "p:u2-1")
	time.Sleep(time.Second)
	SetShare("u1", tmp2)
	time.Sleep(time.Second)
	SetPost(id(), "u1", "p:u1-2")
	time.Sleep(time.Second)
	SetPost(id(), "u3", "p:u3-1")
	time.Sleep(time.Second)
	tmp4 := id()
	SetReply("u2", tmp4, tmp1, "r:u1-1")
	time.Sleep(time.Second)
	SetReply("u1", id(), tmp4, "r:u2-u1-1")
}

func checkPostOwner(id string, user string) bool {
	if p := postDb[id]; p != nil && p.User == user {
		return true
	} else {
		return false
	}
}

func now() int64 {
	return time.Now().Unix()
}

func IsPostExsit(id string) bool {
	if postDb[id] != nil {
		return true
	} else {
		return false
	}
}

func QueryPostByID(id string) (post Post, err utils.Err) {
	if !IsPostExsit(id) {
		return post, utils.NewErr(ErrNotFound, "post")
	}
	return *postDb[id], nil
}

func QueryPostReplies(id string) (replyings []*Post, replies []*Post, err utils.Err) {
	if !IsPostExsit(id) {
		return nil, nil, utils.NewErr(ErrNotFound, "post")
	}
	p := postDb[id]

	replyings = make([]*Post, 0)
	r := postDb[p.Replying]
	lev := 1
	for r != nil {
		p := *r
		p.Level = lev
		lev += 1
		replyings = append(replyings, &p)
		r = postDb[r.Replying]
	}

	replies = make([]*Post, 0)
	working := make(map[string]*Post)
	next := make(map[string]*Post)
	working[id] = postDb[id]
	lev = 1
	for len(working) > 0 {
		for _, v := range postDb {
			if working[v.Replying] != nil {
				p := *v
				p.Level = lev
				next[v.ID] = &p
				replies = append(replies, &p)
			}
		}
		working = next
		next = make(map[string]*Post)
		lev += 1
	}

	sort.Slice(replyings, func(i, j int) bool {
		if replyings[i].Level < replyings[j].Level {
			return true
		} else {
			return false
		}
	})
	sort.Slice(replies, func(i, j int) bool {
		if replies[i].Level < replies[j].Level {
			return true
		} else {
			return false
		}
	})
	return replyings, replies, nil
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
	if IsPostExsit(id) {
		return utils.NewErr(ErrDunplicate, "post")
	}

	p := &Post{
		ID:      id,
		User:    user,
		Date:    now(),
		Content: content,
		Likes:   make([]string, 0),
		Shares:  make([]string, 0),
	}
	postDb[id] = p
	return nil
}

func UpdatePost(id string, content string) utils.Err {
	if !IsPostExsit(id) {
		return utils.NewErr(ErrNotFound, "post")
	}

	p := postDb[id]
	p.Content = content
	p.Date = now()
	return nil
}

func RemovePost(id string) utils.Err {
	if !IsPostExsit(id) {
		return utils.NewErr(ErrNotFound, "post")
	}

	delete(postDb, id)
	return nil
}

func SetLike(user string, id string) utils.Err {
	if !IsPostExsit(id) {
		return utils.NewErr(ErrNotFound, "post")
	}

	p := postDb[id]
	for _, v := range p.Likes {
		if v == user {
			return nil
		}
	}
	p.Likes = append(p.Likes, user)
	return nil
}

func RemoveLike(user string, id string) utils.Err {
	if !IsPostExsit(id) {
		return utils.NewErr(ErrNotFound, "post")
	}

	p := postDb[id]
	for i, v := range p.Likes {
		if v == user {
			p.Likes = slices.Delete(p.Likes, i, i+1)
			return nil
		}
	}
	return utils.NewErr(ErrNotFound, "like")
}

func SetShare(user string, id string) utils.Err {
	if !IsPostExsit(id) {
		return utils.NewErr(ErrNotFound, "post")
	}

	p := postDb[id]
	for _, v := range p.Shares {
		if v == user {
			return nil
		}
	}
	p.Shares = append(p.Likes, user)
	shareDb = append(shareDb, &Share{
		ID:   id,
		User: user,
		Date: now(),
	})
	return nil
}

func RemoveShare(user string, id string) utils.Err {
	if !IsPostExsit(id) {
		return utils.NewErr(ErrNotFound, "post")
	}
	p := postDb[id]

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

func SetReply(user string, id string, replying string, content string) utils.Err {
	if IsPostExsit(id) {
		return utils.NewErr(ErrDunplicate, "post")
	}
	if !IsPostExsit(replying) {
		return utils.NewErr(ErrNotFound, "post")
	}

	r := &Post{
		ID:       id,
		User:     user,
		Date:     now(),
		Content:  content,
		Replying: replying,
		Likes:    make([]string, 0),
		Shares:   make([]string, 0),
	}
	postDb[id] = r
	return nil
}
