package posts

import (
	"time"
	"unicode/utf8"

	"github.com/kidommoc/gustrody/internal/database"
	"github.com/kidommoc/gustrody/internal/users"
	"github.com/kidommoc/gustrody/internal/utils"
)

func (service *PostService) makePost(p *database.Post) (post Post, err utils.Err) {
	u, e := service.user.GetInfo(p.User)
	if e != nil {
		return post, newErr(ErrUserNotFound)
	}

	post = Post{
		ID:          p.ID,
		User:        &u,
		PublishedAt: time.Unix(p.Date, 0).Format(time.RFC3339),
		Content:     p.Content,
		Likes:       len(p.Likes),
		Shares:      len(p.Shares),
	}

	return post, nil
}

func (service *PostService) Get(postID string) (post Post, err utils.Err) {
	postID = service.fullID(postID)
	result, e := service.db.QueryPostByID(postID)
	if e != nil {
		switch {
		case e.Code() == database.ErrNotFound && e.Error() == "post":
			return post, newErr(ErrPostNotFound, postID)
			// default:
		}
	}

	post, e = service.makePost(&result)
	if e != nil {
		switch {
		case e.Code() == database.ErrNotFound && e.Error() == "user":
			return post, newErr(ErrOwner, "not found")
			// default:
		}
	}
	service.setReplies(&post)

	return post, nil
}

func (service *PostService) GetLikes(postID string) (list []*users.UserInfo, err utils.Err) {
	postID = service.fullID(postID)
	result, e := service.db.QueryPostByID(postID)
	if e != nil {
		switch {
		case e.Code() == database.ErrNotFound && e.Error() == "post":
			return list, newErr(ErrPostNotFound, postID)
			// default:
		}
	}

	list = make([]*users.UserInfo, 0, len(result.Likes))
	for _, u := range result.Likes {
		info, e := service.user.GetInfo(u)
		if e == nil {
			list = append(list, &info)
		}
	}

	return list, nil
}

func (service *PostService) GetShares(postID string) (list []*users.UserInfo, err utils.Err) {
	postID = service.fullID(postID)
	result, e := service.db.QueryPostByID(postID)
	if e != nil {
		switch {
		case e.Code() == database.ErrNotFound && e.Error() == "post":
			return list, newErr(ErrPostNotFound, postID)
			// default:
		}
	}

	list = make([]*users.UserInfo, 0, len(result.Shares))
	for _, u := range result.Shares {
		info, e := service.user.GetInfo(u)
		if e == nil {
			list = append(list, &info)
		}
	}

	return list, nil
}

func (service *PostService) GetByUser(username string) (list []*Post, err utils.Err) {
	user, e := service.user.GetInfo(username)
	if e != nil {
		switch {
		case e.Code() == database.ErrNotFound && e.Error() == "user":
			return list, newErr(ErrUserNotFound, username)
			// default:
		}
	}

	posts, _ := service.db.QueryPostsByUser(username, true) // ascending by date
	shares, _ := service.db.QuerySharesByUser(username, true)
	iP := len(posts) - 1 // start from end (newest)
	iS := len(shares) - 1

	for iP >= 0 || iS >= 0 {
		var p *database.Post
		var flag bool // true: post, false: share
		var actor *users.UserInfo
		var sharedBy *users.UserInfo

		if iP < 0 {
			flag = false
		} else if iS < 0 {
			flag = true
		} else {
			if posts[iP].Date > shares[iS].Date {
				flag = true
			} else {
				flag = false
			}
		}

		if flag {
			p = posts[iP]
			actor = &user
			sharedBy = nil
			iP = iP - 1
		} else {
			shared, e := service.db.QueryPostByID(shares[iS].ID)
			if e != nil {
				continue
			}
			p = &shared
			u, e := service.user.GetInfo(p.User)
			if e != nil {
				continue
			}
			actor = &u
			sharedBy = &user
			iS = iS - 1
		}

		post, _ := service.makePost(p)
		post.User = actor
		post.SharedBy = sharedBy
		list = append(list, &post)
	}

	return list, nil
}

func (service *PostService) New(username string, content string) utils.Err {
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
	// note: should not return any error here
	service.db.SetPost(id, username, content)

	return nil
}

func (service *PostService) Edit(username string, postID string, content string) utils.Err {
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
		case e.Code() == database.ErrNotFound && e.Error() == "post":
			return newErr(ErrPostNotFound, postID)
			// default:
		}
	}
	if post.User != username {
		return newErr(ErrOwner, "not "+username)
	}

	if e := service.db.UpdatePost(postID, content); e != nil {
		switch {
		// default:
		}
	}

	return nil
}

func (service *PostService) Remove(username string, postID string) utils.Err {
	postID = service.fullID(postID)
	post, e := service.db.QueryPostByID(postID)
	if e != nil {
		switch {
		case e.Code() == database.ErrNotFound && e.Error() == "post":
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
