package federal

import (
	"encoding/json"
	"fmt"
	"net/http"
	_url "net/url"
	"regexp"

	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/services/net"
	"github.com/kidommoc/gustrody/internal/utils"
)

type FederalDbs struct {
	UserInfo    models.IUserInfo
	UserAccount models.IUserAccount
	UserForeign models.IUserForeign
	UserFollow  models.IUserFollow
	PostQuery   models.IPostQuery
	PostSet     models.IPostSet
	PostLike    models.IPostLike
	PostShare   models.IPostShare
}

type FederalService struct {
	lg     logging.Logger
	site   string
	net    *net.NetService
	db     FederalDbs
	uidReg *regexp.Regexp
	pidReg *regexp.Regexp
}

func NewService(net *net.NetService, dbs FederalDbs, cfg config.Config, lg logging.Logger) *FederalService {
	uidReg := utils.UserIDReg(cfg.Site)
	pidReg := utils.PostIDReg(cfg.Site)
	if uidReg == nil || pidReg == nil {
		panic("Failed to generate user or post id regexp.")
	}
	return &FederalService{
		lg:     lg,
		site:   cfg.Site,
		net:    net,
		db:     dbs,
		uidReg: uidReg,
		pidReg: pidReg,
	}
}

type WFLink struct {
	Rel  string `json:"rel"`
	Type string `json:"type"`
	Href string `json:"href"`
}

type WF struct {
	Subject string   `json:"subject"`
	Links   []WFLink `json:"links"`
}

func (service *FederalService) Webfinger(username, site string) (wf WF, err error) {
	if site != service.site || !service.db.UserInfo.IsUserExist(username) {
		return wf, ErrNotFound
	}

	return WF{
		Subject: fmt.Sprintf("acct:%s@%s", username, site),
		Links: []WFLink{
			{
				Rel:  "self",
				Type: "application/activity+json",
				Href: utils.GenerateUserID(username, service.site),
			},
		},
	}, nil
}

func (service *FederalService) requestWebfinger(user models.UD) (url string, err error) {
	logger := service.lg
	client := service.net.HttpClient()

	// try http
	rawUrl := fmt.Sprintf("http://%s/.well-known/webfinger/?resource=acct:%s",
		user.Domain, user.String(),
	)
	u, err := _url.Parse(rawUrl)
	if err != nil {
		logger.Error("[Federal.reqWebfinger] Failed to parse url.", err)
		return "", ErrSyntax
	}

	header := http.Header{}
	req := http.Request{Method: "GET", URL: u, Header: header}

	res, err := client.Do(&req, nil)
	if err != nil {
		// try https
		rawUrl := fmt.Sprintf("https://%s/.well-known/webfinger/?resource=acct:%s",
			user.Domain, user.String(),
		)
		u, err := _url.Parse(rawUrl)
		if err != nil {
			logger.Error("[Federal.reqWebfinger] Failed to parse url.", err)
			return "", ErrSyntax
		}
		req.URL = u
		res, err = client.Do(&req, nil)
		if err != nil {
			logger.Error("[Federal.reqWebfinger] Failed to send.", err)
			return "", ErrRequest
		}
	}

	var wf WF
	err = json.Unmarshal(res.Body, &wf)
	if err != nil {
		service.errJsonUnmarshal("reqWebfinger", err)
		return "", ErrSyntax
	}
	for _, link := range wf.Links {
		if link.Rel == "self" && link.Href != "" {
			return link.Href, nil
		}
	}

	logger.Warning("[Federal.reqWebfinger] link `self` not find.", "response", wf)
	return "", nil
}

func (service *FederalService) errJsonMarshal(loc string, err error) error {
	msg := fmt.Sprintf("[Federal.%s] Failed to marshal to json.", loc)
	service.lg.Error(msg, err)
	return ErrSyntax
}

func (service *FederalService) errJsonUnmarshal(loc string, err error) error {
	msg := fmt.Sprintf("[Federal.%s] Failed to unmarshal to json.", loc)
	service.lg.Error(msg, err)
	return ErrSyntax
}

func (service *FederalService) errDb(loc, dbAct string, err error) error {
	msg := fmt.Sprintf("[Federal.%s] Failed to %s.", loc, dbAct)
	service.lg.Error(msg, err)
	return ErrDbInternal
}
