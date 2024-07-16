package federal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/utils"
)

func (service *FederalService) udToID(ud models.UD) string {
	logger := service.lg
	if ud.Domain != "" {
		user, _, err := service.GetForeignUserByUD(ud)
		if err != nil {
			logger.Error("[Federal.udToID] Failed to get foreign user.", nil)
			return ""
		}
		return user.ID
	} else {
		return utils.GenerateUserID(ud.Username, service.scheme, service.domain)
	}
}

func (service *FederalService) idToUD(id string) models.UD {
	logger := service.lg
	match := service.uidReg.FindStringSubmatch(id)
	if match != nil {
		return models.UD{Username: match[1]}
	} else {
		user, _, err := service.GetForeignUserByID(id)
		if err != nil {
			logger.Error("[Federal.idToUD] Failed to get foreign user.", nil)
			return models.UD{}
		}
		return user.Username
	}
}

func (service *FederalService) GetPerson(username string) (p PersonObj, err error) {
	logger := service.lg
	user, err := service.db.UserInfo.QueryUser(username)
	if err != nil {
		return p, service.errDb("GetPerson", "query user", err)
	}
	id := utils.GenerateUserID(username, service.scheme, service.domain)
	p = PersonObj{
		Object: Object{
			Context: []interface{}{
				ctxMap["as"],
				ctxMap["sec"],
				map[string]string{
					"Key": ctxKEY,
				},
			},
			Type: "Person",
			ID:   id,
		},
		Username:  user.Username.Username,
		Nickname:  user.Nickname,
		Inbox:     id + "/inbox",
		Outbox:    id + "/outbox",
		Following: id + "/following",
		Followers: id + "/followers",
		Summary:   user.Summary,
	}
	p.Endpoints.SharedInbox = service.domain + "/inbox"
	p.Icon = service.makeImage(&user.Avatar)

	pub, _, err := service.db.UserAccount.QueryUserKeys(username)
	if err != nil {
		return p, service.errDb("GetPerson", "query user keys", err)
	}
	p.PublicKey = PublicKeyObj{
		Object: Object{Type: "Key", ID: id + "#main-key"},
		Owner:  id,
		Pem:    pub,
	}

	logger.Debug(fmt.Sprintf("%+v\n", p))
	return p, nil
}

func (service *FederalService) GetForeignPerson(id string) (person PersonObj, err error) {
	logger := service.lg
	client := service.net.HttpClient()
	defer client.Close()
	u, err := url.Parse(id)
	if err != nil {
		logger.Error("[Federal.getForeignPerson] Failed to parse url.", err)
		return person, ErrSyntax
	}

	header := http.Header{}
	header.Add("Accept", "application/activity+json")
	req := http.Request{
		Method: "GET", URL: u, Header: header,
	}
	res, err := client.Do(&req, nil)
	if err != nil {
		return person, service.errReq("getForeignPerson", id, err)
	}

	if err = json.Unmarshal(res.Body, &person); err != nil {
		return person, service.errJsonUnmarshal("getForeignPerson", err)
	}

	fu := models.ForeignUser{
		Username: models.UD{Username: person.Username, Domain: u.Host},
		ID:       person.ID,
		AvtUrl:   person.Icon.Url,
		PubKey:   person.PublicKey.Pem,
	}
	if person.Endpoints.SharedInbox != "" {
		fu.Inbox = person.Endpoints.SharedInbox
	} else {
		fu.Inbox = person.Inbox
	}

	if fu.AvtUrl != "" {
		ofu, err := service.db.UserForeign.GetForeignUserByID(id)
		if err == nil {
			fu.Avatar = ofu.Avatar
		}
		if err == nil && ofu.AvtUrl != fu.AvtUrl || err == models.ErrNotFound {
			u, err := url.Parse(fu.AvtUrl)
			if err != nil {
				logger.Error("[Federal.getForeignPerson] Failed to parse url.", err)
				return person, ErrSyntax
			}
			req := http.Request{
				Method: "GET", URL: u,
			}
			res, err := client.Do(&req, nil)
			if err != nil {
				return person, service.errReq("getForeignPerson", id, err)
			}
			localUrl, mediaType, err := service.file.StoreImage(res.Body)
			if err != nil {
				logger.Error("[Federal.getForeignPerson] Failed to store image.", err)
				return person, ErrFile
			}
			fu.Avatar = localUrl
			person.Icon.Url = localUrl
			person.Icon.MediaType = MediaType(mediaType)
		} else {
			return person, service.errDb("getForeignUser", "query foreign user", err)
		}
	}

	if err = service.db.UserForeign.SetForeignUser(&fu); err != nil {
		return person, service.errDb("getForeignUser", "set foreign user", err)
	}
	person.Context = nil
	return person, nil
}

func (service *FederalService) GetForeignUserByUD(ud models.UD) (user models.ForeignUser, person *PersonObj, err error) {
	logger := service.lg
	user, err = service.db.UserForeign.GetForeignUserByUD(ud)
	if err == models.ErrNotFound {
		id, err := service.requestWebfinger(ud)
		if err != nil {
			logger.Error("[Federal.GetForeignUser] Failed to request webfinger.", err)
			return user, nil, ErrRequest
		}
		p, err := service.GetForeignPerson(id)
		if err != nil {
			logger.Error("[Federal.GetForeignUser] Failed to request foreign person.", err)
			return user, nil, ErrRequest
		}
		person = &p
		user, err = service.db.UserForeign.GetForeignUserByUD(ud)
	}
	if err != nil {
		return user, nil, service.errDb("GetForeignUser", "query foreign user", err)
	}
	return user, person, nil
}

func (service *FederalService) GetForeignUserByID(id string) (user models.ForeignUser, person *PersonObj, err error) {
	logger := service.lg
	user, err = service.db.UserForeign.GetForeignUserByID(id)
	if err == models.ErrNotFound {
		p, err := service.GetForeignPerson(id)
		if err != nil {
			logger.Error("[Federal.GetForeignUser] Failed to request foreign person.", err)
			return user, nil, ErrRequest
		}
		person = &p
		user, err = service.db.UserForeign.GetForeignUserByID(id)
	}
	if err != nil {
		return user, nil, service.errDb("GetForeignUser", "query foreign user", err)
	}
	return user, person, nil
}

func (service *FederalService) RequestFollow(user, target models.UD) error {
	logger := service.lg
	usrID := service.udToID(user)
	tgtID := service.udToID(target)
	if usrID == "" || tgtID == "" {
		logger.Error("[Federal.Follow] Failed to convert to ud or id.", ErrSyntax)
		return ErrSyntax
	}

	fo := Activity{
		Object: Object{
			Context: []interface{}{ctxMap["as"], ctxMap["sec"]},
			Type:    "Follow",
			ID:      fmt.Sprintf("%s#follow/%s", usrID, target.String()),
		},
		Actor:  usrID,
		Target: tgtID,
	}
	body, err := json.Marshal(fo)
	if err != nil {
		return service.errJsonMarshal("RequestFollow", err)
	}

	inboxes, err := service.getInboxes([]models.UD{target})
	if err != nil {
		return service.errDb("RequestFollow", "get inboxes", err)
	}
	if err = service.sendActivity(user, body, inboxes); err != nil {
		return service.errSendActivity("RequestFollow", err)
	}
	return nil
}

func (service *FederalService) UndoFollow(user, target models.UD) error {
	logger := service.lg
	uid := service.udToID(user)
	if uid == "" {
		logger.Error("[Federal.UndoFollow] Failed to convert to ud or id.", ErrSyntax)
		return ErrSyntax
	}
	fid := fmt.Sprintf("%s#follow/%s", uid, target.String())

	undo := Activity{
		Object: Object{
			Context: []interface{}{ctxMap["as"], ctxMap["sec"]},
			ID:      service.tempActivityID(),
			Type:    "Undo",
		},
		Actor:  uid,
		Target: fid,
	}
	body, err := json.Marshal(undo)
	if err != nil {
		return service.errJsonMarshal("UndoFollow", err)
	}

	inboxes, err := service.getInboxes([]models.UD{target})
	if err != nil {
		return service.errDb("UndoFollow", "get inboxes", err)
	}
	if err = service.sendActivity(user, body, inboxes); err != nil {
		return service.errSendActivity("UndoFollow", err)
	}

	if err = service.db.UserFollow.RemoveFollow(user, target); err != nil {
		return service.errDb("UndoFollow", "remove follow", err)
	}

	return nil
}

func (service *FederalService) ParseFollowRequest(req *Activity) (actor, target models.UD, err error) {
	logger := service.lg
	if req == nil {
		logger.Error("[Federal.ParseFollowRequest] Empty request.", nil)
		return actor, target, ErrSyntax
	}

	usrUD := service.idToUD(req.Actor)
	var tgtID string
	switch req.Target.(type) {
	case string:
		tgtID, _ = req.Target.(string)
	case map[string]string:
		m, _ := req.Target.(map[string]string)
		if m["type"] == "Person" && m["id"] != "" {
			tgtID = m["id"]
		} else {
			msg := fmt.Sprintf("[Federal.] ")
			logger.Warning(msg, nil)
			return actor, target, ErrSyntax
		}
	default:
		msg := fmt.Sprintf("[Federal.ParseFollowRequest] ")
		logger.Warning(msg, nil)
		return actor, target, ErrSyntax
	}
	tgtUD := service.idToUD(tgtID)

	if usrUD.Username == "" || tgtUD.Username == "" {
		msg := fmt.Sprintf("[Federal.ParseFollowRequest] Failed to convert to ud.\nActor: %s\nObject: %s", req.Actor, tgtID)
		logger.Warning(msg, nil)
		return actor, target, ErrSyntax
	}
	// object should be local user
	if tgtUD.Domain != "" {
		msg := fmt.Sprintf("[Federal.ParseFollowRequest] Received a follow request to foreign user. request: %+v", req)
		logger.Warning(msg, nil)
		return actor, target, ErrSyntax
	}

	return usrUD, tgtUD, nil
}

// reqID is the ID of request activity.
//
// when manually approving following, ID should be stored in notification.
func (service *FederalService) AcceptFollow(user, target models.UD, reqID string) error {
	logger := service.lg

	if user.Domain != "" {
		msg := fmt.Sprintf("[Federal.AcceptFollow] Actor is foreign user: %s", user)
		logger.Error(msg, nil)
		return ErrSyntax
	}
	usrID := service.udToID(user)
	if usrID == "" {
		msg := fmt.Sprintf("[Federal.AcceptFollow] Failed to convert to id: %s", user)
		logger.Error(msg, nil)
		return ErrSyntax
	}

	acp := Activity{
		Object: Object{
			Context: []interface{}{ctxMap["as"], ctxMap["sec"]},
			ID:      service.tempActivityID(),
			Type:    "Accept",
		},
		Actor:  usrID,
		Target: reqID,
	}
	body, err := json.Marshal(acp)
	if err != nil {
		return service.errJsonMarshal("AcceptFollow", err)
	}

	// in case target is local user.
	if target.Domain != "" {
		inboxes, err := service.getInboxes([]models.UD{target})
		if err != nil {
			return service.errDb("AcceptFollow", "get inboxes", err)
		}
		if err = service.sendActivity(user, body, inboxes); err != nil {
			return service.errSendActivity("AcceptFollow", err)
		}
	}

	if err := service.db.UserFollow.SetFollow(target, user); err != nil {
		return service.errDb("AcceptFollow", "set follow", err)
	}

	return nil
}

// reqID is the ID of request activity.
//
// when manually approving following, ID should be stored in notification.
func (service *FederalService) RejectFollow(user, target models.UD, reqID string) error {
	logger := service.lg

	if user.Domain != "" {
		msg := fmt.Sprintf("[Federal.RejectFollow] Actor is foreign user: %s", user)
		logger.Error(msg, nil)
		return ErrSyntax
	}
	usrID := service.udToID(user)
	if usrID == "" {
		msg := fmt.Sprintf("[Federal.RejectFollow] Failed to convert to id: %s", user)
		logger.Error(msg, nil)
		return ErrSyntax
	}

	rjt := Activity{
		Object: Object{
			Context: []interface{}{ctxMap["as"], ctxMap["sec"]},
			ID:      service.tempActivityID(),
			Type:    "Reject",
		},
		Actor:  usrID,
		Target: reqID,
	}
	body, err := json.Marshal(rjt)
	if err != nil {
		return service.errJsonMarshal("RejectFollow", err)
	}

	inboxes, err := service.getInboxes([]models.UD{target})
	if err != nil {
		return service.errDb("RejectFollow", "get inboxes", err)
	}
	if err = service.sendActivity(user, body, inboxes); err != nil {
		return service.errSendActivity("RejectFollow", err)
	}

	return nil
}
