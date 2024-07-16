package federal

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kidommoc/gustrody/internal/utils"
)

// header: http header as map[string]string. keys are lowercase. HTTP method `method` and request path `path` should be included.
//
// Examples:
//
//	header["method"] -> "post"
//	header["path"] -> "/inbox"
func (service *FederalService) Sign(username string, header map[string]string) (result string, err error) {
	logger := service.lg

	keyId := utils.GenerateUserID(username, service.scheme, service.domain) + "#main-key"
	headers := "(request-target) host date"
	ns := strings.Split(headers, " ")
	for i, v := range ns {
		var m string
		if v == "(request-target)" {
			method := header["method"]
			path := header["path"]
			if method == "" || path == "" {
				logger.Error("[Federal.Sign] Header missing method or path.", nil)
				return "", ErrSyntax
			}
			m = fmt.Sprintf("%s %s", method, path)
		} else {
			m = header[v]
			if m == "" {
				msg := fmt.Sprintf("[Federal.Sign] Header missing %s", v)
				logger.Error(msg, nil)
				return "", ErrSyntax
			}
		}
		ns[i] = fmt.Sprintf("%s: %s", ns[i], m)
	}
	signature := strings.Join(ns, "\n")

	signature, err = service.Digest(username, signature)
	if err != nil {
		return "", err
	}

	result = fmt.Sprintf(`keyId="%s",headers="%s",signature="%s"`,
		keyId, headers, signature,
	)

	return result, nil
}

func (service *FederalService) Digest(username string, data string) (result string, err error) {
	logger := service.lg
	_, pri, err := service.db.UserAccount.QueryUserKeys(username)
	if err != nil {
		return "", service.errDb("Digest", "query user keys", err)
	}
	priK := utils.GetPrivateKey(pri)
	if priK == nil {
		logger.Error("[Federal.Digest] Failed to convert pem to private key.", nil)
		return "", ErrSyntax
	}
	result = utils.Sign(priK, data)
	return result, nil
}

// header: http header as map[string]string. keys are lowercase. HTTP method `method` and request path `path` should be included.
//
// Examples:
//
//	header["method"] -> "post"
//	header["path"] -> "/inbox"
func (service *FederalService) Verify(header map[string]string, body []byte) bool {
	logger := service.lg

	s := header["signature"]
	if s == "" {
		logger.Error("[Federal.Verify] Missing Signature in header.", nil)
		return false
	}
	hs := make(map[string]string)
	reg := regexp.MustCompile(`([A-z]+)="(.+)"`)
	for _, v := range strings.Split(s, ",") {
		matched := reg.FindStringSubmatch(v)
		if len(matched) < 3 {
			logger.Warning("[Federal.Verify] Part of Signature not fit.", v)
			continue
		}
		hs[matched[1]] = matched[2]
	}

	keyId := hs["keyId"]
	headers := hs["headers"]
	signature := hs["signature"]
	if keyId == "" || headers == "" || signature == "" {
		logger.Error("[Fedral.Verify] Missing keyId or headers or signature in Signature.", nil)
		return false
	}

	ns := strings.Split(headers, " ")
	for i, v := range ns {
		var m string
		if v == "(request-target)" {
			method := header["method"]
			path := header["path"]
			if method == "" || path == "" {
				logger.Error("[Federal.Verify] Missing method or path in header.", nil)
				return false
			}
			m = fmt.Sprintf("%s %s", method, path)
		} else {
			m = header[v]
			if m == "" {
				msg := fmt.Sprintf("[Federal.Verify] Missing %s in header.", v)
				logger.Error(msg, nil)
				return false
			}
		}
		ns[i] = fmt.Sprintf("%s: %s", ns[i], m)
	}
	compare := strings.Join(ns, "\n")

	uid := strings.Split(keyId, "#")[0]
	user, _, err := service.GetForeignUserByID(uid)
	if err != nil {
		logger.Error("[Federal.Verify] Failed to get Foreign user.", err)
		return false
	}
	pubK := utils.GetPublicKey(user.PubKey)

	if pubK == nil {
		logger.Error("[Federal.Verify] Failed to convert pem to public key.", nil)
		return false
	}
	resSignature := true
	if err = utils.Verify(pubK, signature, compare); err != nil {
		logger.Error("[Federal.Verify] Failed to verify signature.", err)
		resSignature = false
	}

	resDigest := true
	if len(body) != 0 {
		s = header["digest"]
		if s == "" {
			logger.Error("[Federal.Verify] Missing Digest in header.", nil)
			return false
		}
		if err = utils.Verify(pubK, s, string(body)); err != nil {
			logger.Error("[Federal.Verify] Failed to verify digest.", err)
			resDigest = false
		}
	}

	return resSignature && resDigest
}
