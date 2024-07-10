package federal

import (
	"testing"
	"time"

	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/test"
	"github.com/kidommoc/gustrody/internal/utils"
)

var seccfg = config.Config{
	Site: "austrody.sns",
}

func TestSignAndVerify(t *testing.T) {
	logger := test.NewMockingLogger(t)
	mUAccDb := newMockingUserAccountDb()
	dbs := FederalDbs{
		UserAccount: mUAccDb,
	}
	service := NewService(nil, dbs, seccfg, logger)

	table := []struct {
		u string
		h map[string]string
		b string
	}{
		{"u1", map[string]string{
			"method": "get", "path": "/users/u1",
			"host": seccfg.Site, "date": time.Now().Format(time.RFC822),
		}, ""},
		{"u2", map[string]string{
			"method": "post", "path": "/inbox",
			"host": seccfg.Site, "date": time.Now().Add(-1 * time.Hour).Format(time.RFC822),
		}, "foobar"},
	}
	for _, v := range table {
		pub, pri := utils.NewKeyPair()
		mUAccDb.data[v.u] = struct {
			pub string
			pri string
		}{pub, pri}

		signed, err := service.Sign(v.u, v.h)
		test.AssertNoError(t, err, "Error when sign: %v")
		v.h["signature"] = signed

		if len(v.b) != 0 {
			digest, err := service.Digest(v.u, v.b)
			test.AssertNoError(t, err, "Error when digest: %v")
			v.h["digest"] = digest
		}

		t.Logf("%+v\n", v.h)

		if !service.Verify(v.h, []byte(v.b)) {
			t.Errorf(v.u + " Verification failed.")
		}
	}

}
