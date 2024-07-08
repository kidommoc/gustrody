package utils

import (
	"testing"
)

func TestSignAndVerify(t *testing.T) {
	msg := "test message"

	pubKeyPem, priKeyPem := NewKeyPair()
	t.Logf("\npub:%s\npri:\n%s", pubKeyPem, priKeyPem)

	priKey := GetPrivateKey(priKeyPem)
	signed := Sign(priKey, msg)
	t.Log("Signed: ", []byte(signed))

	pubKey := GetPublicKey(pubKeyPem)
	if err := Verify(pubKey, signed, msg); err != nil {
		t.Errorf("Cannot verify: %s", err)
	}
}
