package posts

import (
	"unicode/utf8"

	"github.com/kidommoc/gustrody/internal/database"
	"github.com/kidommoc/gustrody/internal/utils"
)

func (service *PostService) Reply(username string, postID string, content string) utils.Err {
	if !service.user.IsUserExist(username) {
		return utils.NewErr(ErrUserNotFound)
	}
	if content == "" {
		return utils.NewErr(ErrContent, "empty")
	}
	if utf8.RuneCountInString(content) > service.maxContentLength {
		return utils.NewErr(ErrContent, "long")
	}

	id := service.newID()
	for service.db.IsPostExist(id) {
		id = service.newID()
	}
	if e := service.db.SetReply(username, id, service.fullID(postID), content); e != nil {
		switch {
		case e.Code() == database.ErrNotFound && e.Error() == "post":
			return utils.NewErr(ErrPostNotFound)
			// default:
		}
	}

	return nil
}

// set replyings and replies of a post
func (service *PostService) setReplies(p *Post) utils.Err {
	rt, rs, e := service.db.QueryPostReplies(p.ID)
	if e != nil {
		return utils.NewErr(ErrPostNotFound)
	}

	// replying to. as (linked) list
	for _, v := range rt {
		p, e := service.makePost(v)
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
			p, e := service.makePost(rs[i])
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
