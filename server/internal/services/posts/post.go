package posts

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/services/users"
	"github.com/kidommoc/gustrody/internal/utils"
)

// ERRORS
//
//   - UserNotFound
func (service *PostService) makePost(p *models.Post, us ...*users.UserInfo) (post Post, err error) {
	logger := service.lg
	var u *users.UserInfo
	if len(us) == 0 {
		uu, e := service.user.GetInfo(p.User)
		if e != nil {
			return post, ErrUserNotFound
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
		Visibility:  p.Vsb.String(),
		Content:     p.Content,
		Likes:       p.Likes,
		Shares:      p.Shares,
		Attachments: make([]AttachImg, 0, len(p.Media.Data())),
	}

	for _, v := range p.Media.Data() {
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

func (service *PostService) Get(user, postID string) (post Post, err error) {
	logger := service.lg
	result, e := service.db.Query.QueryPostByID(postID)
	if e != nil {
		switch e {
		case models.ErrNotFound:
			return post, ErrPostNotFound
		default:
			msg := fmt.Sprintf("[Posts] ")
			logger.Error(msg, e)
			return post, ErrInternal
		}
	}

	if !service.checkPermission(user, result.User, result.ID, result.Vsb) {
		return post, ErrNotPermitted
	}

	post, e = service.makePost(&result)
	if e != nil {
		switch e {
		case models.ErrNotFound:
			return post, ErrOwner
		default:
			msg := fmt.Sprintf("[Posts] ")
			logger.Error(msg, e)
			return post, ErrInternal
		}
	}
	service.setReplies(&post)

	return post, nil
}

func (service *PostService) GetByUser(username, target string) (list []*Post, err error) {
	logger := service.lg
	user, e := service.user.GetInfo(target)
	if e != nil {
		switch e {
		case models.ErrNotFound:
			return nil, ErrUserNotFound
		default:
			msg := fmt.Sprintf("[Posts] Cannot get info of %s", target)
			logger.Error(msg, e)
			return nil, ErrInternal
		}
	}
	us := make(map[string]*users.UserInfo)
	us[target] = &user

	posts, e := service.db.Query.QueryPostsAndSharesByUser(target, false) // descending by date
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
			msg := fmt.Sprintf("[Posts] Cannot get info of %s", u)
			logger.Error(msg, e)
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

func (service *PostService) New(username, vsb, content string, attachments []AttachImg) error {
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

	imgs := []models.Img{}
	for i, v := range attachments {
		if i >= service.maxImgInPost {
			break
		}
		imgs = append(imgs, ToModelImg(v))
	}

	v, ok := utils.GetVsb(vsb)
	if !ok {
		pf, err := service.user.GetPreferences(username)
		if err != nil {
			logger.Error("[Posts] Cannot get user preferences.", err)
			return ErrInternal
		}
		v = pf.PostVsb
	}

	p := models.Post{
		ID: id, Url: url, User: username, Date: time.Now(),
		Replying: "", Vsb: v, Content: content,
	}
	if e := service.db.Set.SetPost(&p, imgs); e != nil {
		logger.Error("[Post] Cannot set post", e)
		return ErrInternal
	}

	return nil
}

func (service *PostService) Edit(username, postID, content string, attachments []AttachImg) error {
	logger := service.lg
	if content == "" {
		return ErrContentEmpty
	}
	if utf8.RuneCountInString(content) > service.maxContentLength {
		return ErrContentTooLong
	}

	post, e := service.db.Query.QueryPostByID(postID)
	if e != nil {
		switch e {
		case models.ErrNotFound:
			return ErrPostNotFound
		default:
			msg := fmt.Sprintf("[Posts] Cannot get %s", postID)
			logger.Error(msg, e)
			return ErrInternal
		}
	}
	if post.User != username {
		return ErrOwner
	}

	imgs := []models.Img{}
	for i, v := range attachments {
		if i >= service.maxImgInPost {
			break
		}
		imgs = append(imgs, ToModelImg(v))
	}

	p := models.Post{
		ID: postID, Date: time.Now(), Content: content,
	}
	if e := service.db.Set.UpdatePost(&p, imgs); e != nil {
		switch e {
		case models.ErrNotFound:
			return ErrPostNotFound
		default:
			msg := fmt.Sprintf("[Posts] Cannot edit %s", postID)
			logger.Error(msg, e)
			return ErrInternal
		}
	}

	return nil
}

func (service *PostService) Remove(username, postID string) error {
	logger := service.lg
	post, e := service.db.Query.QueryPostByID(postID)
	if e != nil {
		switch e {
		case models.ErrNotFound:
			return ErrPostNotFound
		default:
			msg := fmt.Sprintf("[Posts] Cannot get %s", postID)
			logger.Error(msg, e)
			return ErrInternal
		}
	}
	if post.User != username {
		return ErrOwner
	}

	if e := service.db.Set.RemovePost(postID); e != nil {
		switch e {
		case models.ErrNotFound:
			return ErrPostNotFound
		default:
			msg := fmt.Sprintf("[Posts] Cannot remove %s", postID)
			logger.Error(msg, e)
			return ErrInternal
		}
	}

	return nil
}
