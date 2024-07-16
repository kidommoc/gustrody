package federal

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/services/net"
	"github.com/kidommoc/gustrody/internal/test"
)

var wfcfg = config.Config{
	Scheme: "http",
	Domain: "austrody.sns",
}

func TestServeWebfinger(t *testing.T) {
	logger := test.NewMockingLogger(t)
	mUInfDb := newMockingUserInfoDb()
	dbs := FederalDbs{
		UserInfo: mUInfDb,
	}
	service := NewService(nil, nil, dbs, wfcfg, logger)

	wf, err := service.Webfinger("penguin@" + wfcfg.Domain)
	test.AssertNoError(t, err, "when serve: %v")
	result, _ := json.Marshal(wf)
	t.Logf("%s", result)
}

var wfTable = map[string]string{
	"fu1@127.0.0.1:7999": `{"subject":"acct:fu1@127.0.0.1:7999","aliases":["http://127.0.0.1:7999/@fu1","https://127.0.0.1:7999/users/fu1"],"links":[{"rel":"http://webfinger.net/rel/profile-page","type":"text/html","href":"http://127.0.0.1:7999/@fu1"},{"rel":"self","type":"application/activity+json","href":"http://127.0.0.1:7999/users/fu1"},{"rel":"http://ostatus.org/schema/1.0/subscribe","template":"http://127.0.0.1:7999/authorize_interaction?uri={uri}"},{"rel":"http://webfinger.net/rel/avatar","type":"image/gif","href":"http://img.127.0.0.1:7999/accounts/avatars/109/409/091/728/348/948/original/429408d7ea454238.gif"}]}`,
	"fu2@127.0.0.1:7999": `{"subject":"acct:fu2@127.0.0.1:7999","links":[{"rel":"self","type":"application/activity+json","href":"http://127.0.0.1:7999/users/fu2"}]}`,
}

var wfInputTable = map[models.UD]string{
	{Username: "fu1", Domain: "127.0.0.1:7999"}: "http://127.0.0.1:7999/users/fu1",
	{Username: "fu2", Domain: "127.0.0.1:7999"}: "http://127.0.0.1:7999/users/fu2",
}

func startWFServer() {
	app := fiber.New()

	app.Use("/.well-known/webfinger", func(c *fiber.Ctx) error {
		acct := c.Query("resource")
		if acct == "" || !strings.HasPrefix(acct, "acct:") {
			c.Status(fiber.StatusBadRequest)
			return c.SendString("Wrong resource.")
		}
		username := strings.TrimPrefix(acct, "acct:")
		if len(strings.Split(username, "@")) < 2 {
			c.Status(fiber.StatusBadRequest)
			return c.SendString("Wrong resource.")
		}

		c.Set("Content-Type", "application/jrd+json")
		return c.Send([]byte(wfTable[username]))
	})

	app.Listen("127.0.0.1:7999")
}

func TestRequestWebfinger(t *testing.T) {
	logger := test.NewMockingLogger(t)
	dbs := FederalDbs{}
	netService := net.NewNetService(logger)
	service := NewService(netService, nil, dbs, actcfg, logger)

	go startWFServer()
	time.Sleep(2 * time.Second)
	t.Log("...starts.")

	for k, v := range wfInputTable {
		url, err := service.requestWebfinger(k)
		test.AssertNoError(t, err, "when request "+k.String()+": %v")
		test.AssertEqual(t, v, url)
	}
}
