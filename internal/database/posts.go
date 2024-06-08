package database

import (
	"slices"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/utils"
)

// models

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
	ID   string `json:"id"` // origin post id
	User string `json:"user"`
	Date int64  `json:"date"`
}

// database

type IPostDb interface {
	checkPostOwner(id string, user string) bool
	IsPostExist(id string) bool
	QueryPostByID(id string) (post Post, err utils.Err)
	QueryPostReplies(id string) (replyings []*Post, replies []*Post, err utils.Err)
	QueryPostsByUser(user string, asec bool) (l []*Post, err utils.Err)
	QuerySharesByUser(user string, asec bool) (l []*Share, err utils.Err)
	SetPost(id string, user string, content string) utils.Err
	UpdatePost(id string, content string) utils.Err
	RemovePost(id string) utils.Err
	SetLike(user string, id string) utils.Err
	RemoveLike(user string, id string) utils.Err
	SetShare(user string, id string) utils.Err
	RemoveShare(user string, id string) utils.Err
	SetReply(user string, id string, replying string, content string) utils.Err
}

// should implemented with Postgre
type PostDb struct {
	site    string
	postDb  map[string]*Post
	shareDb []*Share
}

var postsIns *PostDb = nil

func PostInstance() *PostDb {
	cfg := config.Get()
	if postsIns == nil {
		postsIns = &PostDb{
			site:    cfg.Site,
			postDb:  make(map[string]*Post),
			shareDb: make([]*Share, 0, 100),
		}
	}
	return postsIns
}

// functions

func initPostDb() {
	db := PostInstance()
	id := func() string {
		return db.site + "/posts/" + uuid.New().String()
	}
	tmp1 := id()
	db.SetPost(tmp1, "u1", "p:u1-1")
	time.Sleep(time.Second)
	tmp2 := id()
	db.SetPost(tmp2, "u2", "p:u2-1")
	time.Sleep(time.Second)
	db.SetShare("u1", tmp2)
	time.Sleep(time.Second)
	db.SetPost(id(), "u1", "p:u1-2")
	time.Sleep(time.Second)
	db.SetPost(id(), "u3", "p:u3-1")
	time.Sleep(time.Second)
	tmp4 := id()
	db.SetReply("u2", tmp4, tmp1, "r:u1-1")
	time.Sleep(time.Second)
	db.SetReply("u1", id(), tmp4, "r:u2-u1-1")
}

func now() int64 {
	return time.Now().Unix()
}

func (db *PostDb) checkPostOwner(id string, user string) bool {
	if p := db.postDb[id]; p != nil && p.User == user {
		return true
	} else {
		return false
	}
}

func (db *PostDb) IsPostExist(id string) bool {
	if db.postDb[id] != nil {
		return true
	} else {
		return false
	}
}

func (db *PostDb) QueryPostByID(id string) (post Post, err utils.Err) {
	if !db.IsPostExist(id) {
		return post, utils.NewErr(ErrNotFound, "post")
	}
	return *db.postDb[id], nil
}

func (db *PostDb) QueryPostReplies(id string) (replyings []*Post, replies []*Post, err utils.Err) {
	if !db.IsPostExist(id) {
		return nil, nil, utils.NewErr(ErrNotFound, "post")
	}
	p := db.postDb[id]

	replyings = make([]*Post, 0)
	r := db.postDb[p.Replying]
	lev := 1
	for r != nil {
		p := *r
		p.Level = lev
		lev += 1
		replyings = append(replyings, &p)
		r = db.postDb[r.Replying]
	}

	replies = make([]*Post, 0)
	working := make(map[string]*Post)
	next := make(map[string]*Post)
	working[id] = db.postDb[id]
	lev = 1
	for len(working) > 0 {
		for _, v := range db.postDb {
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

func (db *PostDb) QueryPostsByUser(user string, asec bool) (l []*Post, err utils.Err) {
	l = make([]*Post, 0)
	for _, v := range db.postDb {
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

func (db *PostDb) QuerySharesByUser(user string, asec bool) (l []*Share, err utils.Err) {
	l = make([]*Share, 0)
	for _, v := range db.shareDb {
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

func (db *PostDb) SetPost(id string, user string, content string) utils.Err {
	if db.IsPostExist(id) {
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
	db.postDb[id] = p
	return nil
}

func (db *PostDb) UpdatePost(id string, content string) utils.Err {
	if !db.IsPostExist(id) {
		return utils.NewErr(ErrNotFound, "post")
	}

	p := db.postDb[id]
	p.Content = content
	p.Date = now()
	return nil
}

func (db *PostDb) RemovePost(id string) utils.Err {
	if !db.IsPostExist(id) {
		return utils.NewErr(ErrNotFound, "post")
	}

	delete(db.postDb, id)
	return nil
}

func (db *PostDb) SetLike(user string, id string) utils.Err {
	if !db.IsPostExist(id) {
		return utils.NewErr(ErrNotFound, "post")
	}

	p := db.postDb[id]
	for _, v := range p.Likes {
		if v == user {
			return nil
		}
	}
	p.Likes = append(p.Likes, user)
	return nil
}

func (db *PostDb) RemoveLike(user string, id string) utils.Err {
	if !db.IsPostExist(id) {
		return utils.NewErr(ErrNotFound, "post")
	}

	p := db.postDb[id]
	for i, v := range p.Likes {
		if v == user {
			p.Likes = slices.Delete(p.Likes, i, i+1)
			return nil
		}
	}
	return utils.NewErr(ErrNotFound, "like")
}

func (db *PostDb) SetShare(user string, id string) utils.Err {
	if !db.IsPostExist(id) {
		return utils.NewErr(ErrNotFound, "post")
	}

	p := db.postDb[id]
	for _, v := range p.Shares {
		if v == user {
			return nil
		}
	}
	p.Shares = append(p.Likes, user)
	db.shareDb = append(db.shareDb, &Share{
		ID:   id,
		User: user,
		Date: now(),
	})
	return nil
}

func (db *PostDb) RemoveShare(user string, id string) utils.Err {
	if !db.IsPostExist(id) {
		return utils.NewErr(ErrNotFound, "post")
	}
	p := db.postDb[id]

	// use 2 annoymous func to ensure completely deletion

	err1 := func() utils.Err {
		for i, v := range db.shareDb {
			if v.ID == id && v.User == user {
				db.shareDb = slices.Delete(db.shareDb, i, i+1)
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

func (db *PostDb) SetReply(user string, id string, replying string, content string) utils.Err {
	if db.IsPostExist(id) {
		return utils.NewErr(ErrDunplicate, "post")
	}
	if !db.IsPostExist(replying) {
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
	db.postDb[id] = r
	return nil
}
