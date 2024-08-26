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
		uu, e := service.user.GetInfo(p.User.String())
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
		Date:        utils.DateString(p.Date),
		Visibility:  p.Vsb.String(),
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
			Type: v.Type,
			Url:  service.domain + v.Url,
			Alt:  v.Alt,
		}
		ext := a[len(a)-1]
		if ext == "jpeg" || ext == "png" {
			img.Type = "image/" + ext
		} else {
			logger.Warning("[Post] Wrong image url: wrong extension",
				"url", v.Url,
			)
			continue
		}
		post.Attachments = append(post.Attachments, img)
	}

	return post, nil
}

func (service *PostService) Get(user, postID string) (post *Post, err error) {
	logger := service.lg
	result, e := service.db.Query.QueryPost(postID)
	if e != nil {
		switch e {
		case models.ErrNotFound:
			// !!try foreign
			return post, ErrPostNotFound
		default:
			msg := fmt.Sprintf("[Posts] ")
			logger.Error(msg, e)
			return post, ErrInternal
		}
	}

	if !service.checkPermission(user, &result) {
		return post, ErrNotPermitted
	}

	p, e := service.makePost(&result)
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
	service.setReplies(&p)

	return &p, nil
}

func (service *PostService) GetByUser(username, target string, maxDate time.Time) (list []*Post, err error) {
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

	if utils.UdReg.MatchString(target) {
		// may update foreign
	}
	posts, e := service.db.Query.QueryUserContent(target, maxDate)
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
	fo_only := service.checkPermission(username, &models.Post{User: models.NewUD(target), Vsb: utils.Vsb_FOLLOWER})
	for _, v := range posts {
		switch v.Vsb {
		case utils.Vsb_FOLLOWER:
			if !fo_only {
				continue
			}
		case utils.Vsb_DIRECT:
			if !service.checkPermission(username, &v) {
				continue
			}
		}
		u := gu(v.User.String())
		if u == nil {
			continue
		}
		p, e := service.makePost(&v, u)
		if e != nil {
			continue
		}
		p.ReplyTo = gu(v.ReplyTo.String())
		p.SharedBy = gu(v.SharedBy.String())
		list = append(list, &p)
	}

	return list, nil
}

// only used with local user
func (service *PostService) New(username, vsb, content string, date time.Time, attachments []AttachImg) error {
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

	id := utils.NewUUID()
	for service.db.Query.IsPostExist(id) {
		id = utils.NewUUID()
	}
	url := utils.GeneratePostID(id, service.scheme, service.domain)

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
		ID: id, Url: url, User: models.NewUD(username),
		Date: date, Replying: "", Vsb: v, Content: content,
	}
	for i, v := range attachments {
		if i >= service.maxImgInPost {
			break
		}
		p.Media = append(p.Media, ToModelImg(v))
	}
	if e := service.db.Set.SetPost(&p); e != nil {
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

	post, e := service.db.Query.QueryPost(postID)
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
	if post.User.String() != username {
		return ErrOwner
	}

	p := models.Post{
		ID: postID, Date: time.Now(), Content: content,
	}
	for i, v := range attachments {
		if i >= service.maxImgInPost {
			break
		}
		p.Media = append(p.Media, ToModelImg(v))
	}
	if e := service.db.Set.UpdatePost(&p); e != nil {
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
	post, e := service.db.Query.QueryPost(postID)
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
	if post.User.String() != username {
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
