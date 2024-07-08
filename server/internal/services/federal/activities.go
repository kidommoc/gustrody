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
	return fmt.Sprintf("https://%s/activity/%s", service.site, utils.NewUUID())
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
		header.Add("Host", service.site)
		d := utils.DateString(time.Now())
		header.Add("Date", d)

		headerMap := map[string]string{
			"method": "post", "path": url.Path,
			"host": service.site, "date": d,
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
