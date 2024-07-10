package federal

import (
	"fmt"
	"strings"

	"github.com/kidommoc/gustrody/internal/models"
)

var ctxMap = map[string]string{
	"as":  "https://www.w3.org/ns/activitystreams",
	"sec": "https://w3id.org/security/v1",
}

var (
	ctxKEY       = ctxMap["sec"] + "#Key"
	ctxPUBLIC    = ctxMap["as"] + "#Public"
	ctxSENSITIVE = ctxMap["as"] + "#sensitive"
	ctxLOCK      = ctxMap["as"] + "#manuallyApprovesFollowers"
)

func expandCtx(s string) string {
	parts := strings.Split(s, ":")
	rep := ctxMap[parts[0]]
	if rep == "" {
		return s
	}
	return fmt.Sprintf("%s#%s", rep, strings.Join(parts[1:], ":"))
}

type Object struct {
	Context []interface{} `json:"@context,omitempty"`
	ID      string        `json:"id,omitempty"`
	Type    string        `json:"type,omitempty"`
}

type PublicKeyObj struct {
	Object
	Owner string `json:"owner"`
	Pem   string `json:"publicKeyPem"`
}

type MediaType string

type MediaObj struct {
	Object
	MediaType MediaType `json:"mediaType"`
	Url       string    `json:"url"`
	Name      string    `json:"name,omitempty"`
}

func makeMedia(img *models.Img) MediaObj {
	return MediaObj{
		Object: Object{
			Type: "Document",
			ID:   img.Url,
		},
		MediaType: MediaType(img.Type), // !!need check
		Url:       img.Url,
	}
}

type PersonObj struct {
	Object
	Username  string `json:"preferredUsername"`
	Nickname  string `json:"name"`
	Date      string `json:"published"`
	Inbox     string `json:"inbox"`
	Outbox    string `json:"outbox"`
	Endpoints struct {
		SharedInbox string `json:"sharedInbox"`
	} `json:"endpoints"`
	Following string       `json:"following"`
	Followers string       `json:"followers"`
	Summary   string       `json:"summary"`
	Icon      MediaObj     `json:"icon"`
	PublicKey PublicKeyObj `json:"publicKey"`
}

type NoteObj struct {
	Object
	Url         string        `json:"url"`
	Date        string        `json:"published"`
	Publisher   string        `json:"attributedTo"`
	To          []string      `json:"to"`
	Cc          []string      `json:"cc"`
	Content     string        `json:"content"`
	Attachments []MediaObj    `json:"attachment,omitempty"`
	ReplyTo     string        `json:"inReplyTo,omitempty"`
	Replies     CollectionObj `json:"replies,omitempty"`
}

type CollectionObj struct {
	Object
	Length int64         `json:"totalItems"`
	First  interface{}   `json:"first,omitempty"`
	Next   interface{}   `json:"next,omitempty"`
	Of     string        `json:"partof,omitempty"`
	Items  []interface{} `json:"items"`
}
