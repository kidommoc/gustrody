package posts

import (
	"unicode/utf8"

	"github.com/kidommoc/gustrody/internal/db"
	"github.com/kidommoc/gustrody/internal/utils"
)

func Reply(username string, postID string, content string) utils.Err {
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
	if e := db.SetReply(username, id, fullID(postID), content); e != nil {
		switch {
		case e.Code() == db.ErrNotFound && e.Error() == "post":
			return utils.NewErr(ErrPostNotFound)
			// default:
		}
	}

	return nil
}

// set replyings and replies of a post
func setReplies(p *Post) utils.Err {
	rt, rs, e := db.QueryPostReplies(p.ID)
	if e != nil {
		return utils.NewErr(ErrPostNotFound)
	}

	// replying to. as (linked) list
	for _, v := range rt {
		p, e := makePost(v)
		if e != nil {
			continue // ?
		}
		p.Replyings = append(p.Replyings, &p)
	}

	// replies. as tree
	lrs := len(rs)
	if lrs == 0 {
		return nil
	}
	m := make(map[string]*Post)
	m[p.ID] = p
	lev := rs[0].Level
	maxLev := rs[len(rs)-1].Level
	i := 0
	for lev <= maxLev {
		m2 := make(map[string]*Post)
		for ; i < lrs && rs[i].Level == lev; i += 1 {
			p, e := makePost(rs[i])
			if e != nil {
				continue // ?
			}
			r := rs[i].Replying
			m[r].Replies = append(m[r].Replies, &p)
			m2[p.ID] = &p
		}
		m = m2
		lev += 1
	}

	return nil
}
