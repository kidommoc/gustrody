package posts

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/utils"
)

func (service *PostService) Reply(username, postID, vsb, content string, attachments []AttachImg) error {
	logger := service.lg
	if !service.user.IsUserExist(username) {
		return ErrUserNotFound
	}
	if content == "" {
		return ErrContentEmpty
	}
	if utf8.RuneCountInString(content) > service.maxContentLength {
		return ErrContentTooLong
	}

	id := service.newID()
	for service.db.Query.IsPostExist(id) {
		id = service.newID()
	}
	url := service.getUrl(id)

	v, ok := utils.GetVsb(vsb)
	if !ok {
		pf, err := service.user.GetPreferences(username)
		if err != nil {
			logger.Error("[Posts.Reply] Cannot get user preferences.", err)
			return ErrInternal
		}
		v = pf.PostVsb
	}

	imgs := []models.Img{}
	for i, v := range attachments {
		if i >= service.maxImgInPost {
			break
		}
		imgs = append(imgs, ToModelImg(v))
	}

	p := models.Post{
		ID: id, Url: url, User: username, Date: time.Now(),
		Replying: postID, Vsb: v, Content: content,
	}
	if e := service.db.Set.SetPost(&p, imgs); e != nil {
		switch e {
		case models.ErrNotFound:
			return ErrPostNotFound
		default:
			msg := fmt.Sprintf("[Posts.Reply] Cannot have %s reply to %s", username, postID)
			logger.Error(msg, e)
			return ErrInternal
		}
	}

	return nil
}

// set replyings and replies of a post
func (service *PostService) setReplies(post *Post) error {
	logger := service.lg
	rt, rs, e := service.db.Query.QueryPostReplies(post.ID)
	if e != nil {
		msg := fmt.Sprintf("[Posts.Reply] Cannot get replyings and replies of %s", post.ID)
		logger.Error(msg, e)
		return ErrPostNotFound
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
