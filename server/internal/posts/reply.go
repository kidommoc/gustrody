package posts

import (
	"time"
	"unicode/utf8"

	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/utils"
)

func (service *PostService) Reply(username string, postID string, vsb string, content string, attachments []AttachImg) utils.Error {
	if !service.user.IsUserExist(username) {
		return newErr(ErrUserNotFound, username)
	}
	if content == "" {
		return newErr(ErrContent, "empty")
	}
	if utf8.RuneCountInString(content) > service.maxContentLength {
		return newErr(ErrContent, "too long")
	}

	id := service.newID()
	for service.db.IsPostExist(id) {
		id = service.newID()
	}
	url := service.getUrl(id)

	v, ok := utils.GetVsb(vsb)
	if !ok {
		// get default
	}

	imgs := []models.Img{}
	for i, v := range attachments {
		if i >= service.maxImgInPost {
			break
		}
		imgs = append(imgs, ToModelImg(v))
	}

	if e := service.db.SetPost(
		id, url, username, time.Now(),
		postID, v, content, imgs,
	); e != nil {
		switch {
		case e.Code() == models.ErrNotFound && e.Error() == "post":
			return newErr(ErrPostNotFound, postID)
			// default:
		}
	}

	return nil
}

// set replyings and replies of a post
func (service *PostService) setReplies(post *Post) utils.Error {
	logger := logging.Get()
	rt, rs, e := service.db.QueryPostReplies(post.ID)
	if e != nil {
		logger.Error("[Posts.Reply] Cannot query replyings and replies of post", e)
		return newErr(ErrPostNotFound, post.ID)
	}

	// replying to. as list
	rt = rt[1:]
	for _, v := range rt {
		p, e := service.makePost(v)
		if e != nil {
			logger.Error("[Posts.Reply] Cannot make post", e)
			continue
		}
		post.Replyings = append(post.Replyings, &p)
	}

	// replies. as tree
	lrs := len(rs)
	if lrs == 0 {
		post.Replies = make([]*Post, 0)
		return nil
	}
	m := make(map[string]*Post)
	m[post.ID] = post
	lev := rs[0].Level
	maxLev := rs[len(rs)-1].Level
	i := 1
	for lev < maxLev && len(m) != 0 {
		m2 := make(map[string]*Post)
		for ; i < lrs && rs[i].Level == lev+1; i += 1 {
			p, e := service.makePost(rs[i])
			if e != nil {
				logger.Error("[Posts.Reply] Cannot make post", e)
				continue
			}
			p.Replies = make([]*Post, 0)
			r := rs[i].Replying
			m[r].Replies = append(m[r].Replies, &p)
			m2[p.ID] = &p
		}
		m = m2
		lev += 1
	}

	return nil
}
