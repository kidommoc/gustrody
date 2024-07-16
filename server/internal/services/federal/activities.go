package federal

import (
	"fmt"
	"net/http"
	_url "net/url"
	"time"

	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/utils"
)

type Activity struct {
	Object
	Actor  string      `json:"actor"`
	Target interface{} `json:"object"`
}

type NoteActivity struct {
	Activity
	Date string   `json:"published"`
	To   []string `json:"to,omitempty"`
	Cc   []string `json:"cc,omitempty"`
}

func (service *FederalService) tempActivityID() string {
	return fmt.Sprintf("%s://%s/activity/%s", service.scheme, service.domain, utils.NewUUID())
}

func (service *FederalService) errSendActivity(loc string, err error) error {
	msg := fmt.Sprintf("[Federal.%s] Failed to send activity.", loc)
	service.lg.Error(msg, err)
	return nil
}

func (service *FederalService) sendActivity(actor models.UD, body []byte, dst []string) error {
	logger := service.lg
	client := service.net.HttpClient()
	defer client.Close()

	for _, u := range dst {
		url, err := _url.Parse(u)
		if err != nil {
			logger.Error("[Federal.sendActivity] Failed to parse url.", err)
			continue
		}

		header := http.Header{}
		header.Add("Content-Type", "application/activity+json")
		header.Add("Host", service.domain)
		d := utils.HttpDateString(time.Now())
		header.Add("Date", d)

		headerMap := map[string]string{
			"method": "post", "path": url.Path,
			"host": service.domain, "date": d,
		}
		signature, err := service.Sign(actor.String(), headerMap)
		if err != nil {
			msg := fmt.Sprintf("[Federal.sendActivity] Fail to sign for %s.", u)
			logger.Error(msg, err)
			continue
		}
		header.Add("Signature", signature)
		if len(body) != 0 {
			digest, err := service.Digest(actor.String(), string(body))
			if err != nil {
				msg := fmt.Sprintf("[Federal.sendActivity] Fail to digest for %s.", u)
				logger.Error(msg, err)
				continue
			}
			header.Add("Digest", digest)
		}

		req := http.Request{
			Method: "POST", URL: url,
			Header: header,
		}
		_, err = client.Do(&req, body)
		if err != nil {
			logger.Error("[Federal.sendActivity] Failed to send.", err)
		}
	}
	return nil
}

func (service *FederalService) getInboxes(tgt []models.UD) (inboxes []string, err error) {
	logger := service.lg
	for _, user := range tgt {
		if user.Domain != "" && !service.db.UserForeign.IsForeignExist(user) {
			id, err := service.requestWebfinger(user)
			if err != nil {
				logger.Warning("[Federal.getInboxes] Failed to request webfinger.", "user-domain", user)
				continue
			}
			if _, err = service.GetForeignPerson(id); err != nil { // not implemented yet
				logger.Warning("[Federal.getInboxes] Failed to get foreign person.", "user id", id)
			}
		}
	}
	inboxes, err = service.db.UserForeign.GetInboxes(tgt)
	if err != nil {
		return nil, service.errDb("UndoLike", "get inboxes", err)
	}
	return inboxes, nil
}
