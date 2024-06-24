package posts

import (
	"fmt"
	"strings"
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
	logger := logging.Get()
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
		Url:         p.Url,
		User:        u,
		Date:        p.Date.Format(time.RFC3339),
		Visibility:  string(p.Vsb),
		Content:     p.Content,
		Likes:       p.Likes,
		Shares:      p.Shares,
		Attachments: make([]AttachImg, 0, len(p.Media)),
	}

	for _, v := range p.Media {
		a := strings.Split(v.Url, ".")
		if len(a) < 2 {
			logger.Warning("[Post] Wrong image url: no extension",
				"url", v.Url,
			)
			continue
		}
		img := AttachImg{
			Url: v.Url,
			Alt: v.Alt,
		}
		ext := a[len(a)-1]
		switch ext {
		case "jpeg":
		case "png":
			img.Type = "image/" + ext
		default:
			logger.Warning("[Post] Wrong image url: wrong extension",
				"url", v.Url,
			)
			continue
		}
		post.Attachments = append(post.Attachments, img)
	}

	return post, nil
}

func (service *PostService) Get(user string, postID string) (post Post, err utils.Error) {
	result, e := service.db.QueryPostByID(postID)
	if e != nil {
		switch {
		case e.Code() == models.ErrNotFound && e.Error() == "post":
			return post, newErr(ErrPostNotFound, postID)
			// default:
		}
	}

	if !service.checkPermission(user, result.User, result.ID, result.Vsb) {
		return post, newErr(
			ErrNotPermitted,
			fmt.Sprintf("%s is not allowed to visit %s", user, postID),
		)
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

func (service *PostService) GetByUser(username string, target string) (list []*Post, err utils.Error) {
	user, e := service.user.GetInfo(target)
	if e != nil {
		switch {
		case e.Code() == models.ErrNotFound && e.Error() == "user":
			return list, newErr(ErrUserNotFound, target)
			// default:
		}
	}
	us := make(map[string]*users.UserInfo)
	us[target] = &user

	logger := logging.Get()
	posts, e := service.db.QueryPostsAndSharesByUser(target, false) // descending by date
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
	fo_only := service.checkPermission(username, target, "", utils.Vsb_FOLLOWER)
	for _, v := range posts {
		switch v.Vsb {
		case utils.Vsb_FOLLOWER:
			if !fo_only {
				continue
			}
		case utils.Vsb_DIRECT:
			if !service.checkPermission(username, target, v.ID, v.Vsb) {
				continue
			}
		}
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

func (service *PostService) New(username string, vsb string, content string, attachments []AttachImg) utils.Error {
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
	url := service.getUrl(id)

	imgs := []models.Img{}
	for i, v := range attachments {
		if i >= service.maxImgInPost {
			break
		}
		imgs = append(imgs, ToModelImg(v))
	}

	v, ok := utils.GetVsb(vsb)
	if !ok {
		// get default
	}

	if e := service.db.SetPost(
		id, url, username, time.Now(),
		"", v, content, imgs,
	); e != nil {
		logger.Error("[Post] Cannot set post", e)
		return newErr(ErrInternal, e.CodeString()+" "+e.Error())
	}

	return nil
}

func (service *PostService) Edit(username string, postID string, content string, attachments []AttachImg) utils.Error {
	if content == "" {
		return newErr(ErrContent, "empty")
	}
	if utf8.RuneCountInString(content) > service.maxContentLength {
		return newErr(ErrContent, "too long")
	}

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

	imgs := []models.Img{}
	for i, v := range attachments {
		if i >= service.maxImgInPost {
			break
		}
		imgs = append(imgs, ToModelImg(v))
	}

	if e := service.db.UpdatePost(postID, time.Now(), content, imgs); e != nil {
		switch {
		// default:
		}
	}

	return nil
}

func (service *PostService) Remove(username string, postID string) utils.Error {
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
