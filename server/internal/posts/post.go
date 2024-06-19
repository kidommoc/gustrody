package posts

import (
	"time"
	"unicode/utf8"

	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/users"
	"github.com/kidommoc/gustrody/internal/utils"
)

// ERRORS
//
//   - UserNotFound
func (service *PostService) makePost(p *models.Post, us ...*users.UserInfo) (post Post, err utils.Error) {
	var u *users.UserInfo
	if len(us) == 0 {
		uu, e := service.user.GetInfo(p.User)
		if e != nil {
			return post, newErr(ErrUserNotFound)
		}
		u = &uu
	} else {
		u = us[0]
	}

	post = Post{
		ID:          p.ID,
		User:        u,
		PublishedAt: p.Date.Format(time.RFC3339),
		Content:     p.Content,
		Likes:       p.Likes,
		Shares:      p.Shares,
		Attachments: p.Media,
	}

	return post, nil
}

func (service *PostService) Get(postID string) (post Post, err utils.Error) {
	postID = service.fullID(postID)
	result, e := service.db.QueryPostByID(postID)
	if e != nil {
		switch {
		case e.Code() == models.ErrNotFound && e.Error() == "post":
			return post, newErr(ErrPostNotFound, postID)
			// default:
		}
	}

	post, e = service.makePost(&result)
	if e != nil {
		switch {
		case e.Code() == models.ErrNotFound && e.Error() == "user":
			return post, newErr(ErrOwner, "not found")
			// default:
		}
	}
	service.setReplies(&post)

	return post, nil
}

func (service *PostService) GetByUser(username string) (list []*Post, err utils.Error) {
	user, e := service.user.GetInfo(username)
	if e != nil {
		switch {
		case e.Code() == models.ErrNotFound && e.Error() == "user":
			return list, newErr(ErrUserNotFound, username)
			// default:
		}
	}
	us := make(map[string]*users.UserInfo)
	us[username] = &user

	logger := logging.Get()
	posts, e := service.db.QueryPostsAndSharesByUser(username, false) // descending by date
	if e != nil {
		logger.Error("[Posts] Error when GetByUser", e)
		return list, nil
	}
	list = make([]*Post, 0, len(posts))
	gu := func(u string) *users.UserInfo {
		if u == "" {
			return nil
		}
		if us[u] != nil {
			return us[u]
		}
		ui, e := service.user.GetInfo(u)
		if e != nil {
			logger.Error("[Posts] Cannot get user info", e)
			return nil
		}
		us[u] = &ui
		return &ui
	}
	for _, v := range posts {
		u := gu(v.User)
		if u == nil {
			continue
		}
		p, e := service.makePost(v, u)
		if e != nil {
			continue
		}
		p.ReplyTo = gu(v.ReplyTo)
		p.SharedBy = gu(v.SharedBy)
		list = append(list, &p)
	}

	return list, nil
}

func (service *PostService) New(username string, content string, attachments []string) utils.Error {
	logger := logging.Get()
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

	if e := service.db.SetPost(id, username, "", content, attachments); e != nil {
		logger.Error("[Post] Cannot set post", e)
		return newErr(ErrInternal, e.CodeString()+" "+e.Error())
	}

	return nil
}

func (service *PostService) Edit(username string, postID string, content string, attachments []string) utils.Error {
	if content == "" {
		return newErr(ErrContent, "empty")
	}
	if utf8.RuneCountInString(content) > service.maxContentLength {
		return newErr(ErrContent, "too long")
	}

	postID = service.fullID(postID)
	post, e := service.db.QueryPostByID(postID)
	if e != nil {
		switch {
		case e.Code() == models.ErrNotFound && e.Error() == "post":
			return newErr(ErrPostNotFound, postID)
			// default:
		}
	}
	if post.User != username {
		return newErr(ErrOwner, "not "+username)
	}

	if e := service.db.UpdatePost(postID, content, attachments); e != nil {
		switch {
		// default:
		}
	}

	return nil
}

func (service *PostService) Remove(username string, postID string) utils.Error {
	postID = service.fullID(postID)
	post, e := service.db.QueryPostByID(postID)
	if e != nil {
		switch {
		case e.Code() == models.ErrNotFound && e.Error() == "post":
			return newErr(ErrPostNotFound, postID)
			// default:
		}
	}
	if post.User != username {
		return newErr(ErrOwner, "not "+username)
	}

	if e := service.db.RemovePost(postID); e != nil {
		switch {
		// default:
		}
	}

	return nil
}
