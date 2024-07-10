package federal

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/services/net"
	"github.com/kidommoc/gustrody/internal/test"
	"github.com/kidommoc/gustrody/internal/utils"
)

var actcfg = config.Config{
	Site: "127.0.0.1:7999",
}

var actTable = []struct {
	u models.UD
	b string
	p string
}{
	{models.UD{Username: "u1"}, "", "/inbox"},
	{models.UD{Username: "u2"}, "foobar", "/users/u2/inbox"},
}

func startActServer(t *testing.T, dbs FederalDbs) {
	service := NewService(nil, dbs, actcfg, test.NewMockingLogger(t))

	app := fiber.New()

	app.Use("/", func(c *fiber.Ctx) error {
		t.Logf("Handle %s at %s", c.Method(), c.Path())
		headers := c.GetReqHeaders()
		headerMap := make(map[string]string, len(headers)+2)
		headerMap["method"] = strings.ToLower(c.Method())
		headerMap["path"] = strings.ToLower(c.Path())
		for k, v := range headers {
			headerMap[strings.ToLower(k)] = strings.Join(v, ",")
		}

		body := c.Body()
		if !service.Verify(headerMap, body) {
			t.Errorf("Fail to verify:\nheaders: %+v\nbody: %s", headerMap, body)
		}

		return c.SendStatus(fiber.StatusOK)
	})

	app.Listen(actcfg.Site)
}

func TestSendActivity(t *testing.T) {
	logger := test.NewMockingLogger(t)
	mUAccDb := newMockingUserAccountDb()
	dbs := FederalDbs{
		UserAccount: mUAccDb,
	}
	netService := net.NewNetService(logger)
	service := NewService(netService, dbs, actcfg, logger)

	go startActServer(t, dbs)
	time.Sleep(2 * time.Second) // wait for mocking foreign server starting
	t.Log("...starts.")

	for _, v := range actTable {
		pub, pri := utils.NewKeyPair()
		mUAccDb.data[v.u.String()] = struct {
			pub string
			pri string
		}{pub, pri}
		dst := "http://" + actcfg.Site + v.p
		err := service.sendActivity(v.u, []byte(v.b), []string{dst})
		test.AssertNoError(t, err,
			fmt.Sprintf("when %s sends to %s: ", v.u.String(), v.p)+"%v",
		)
	}
}
