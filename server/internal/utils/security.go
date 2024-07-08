package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

func SHA256Hash(s string) []byte {
	sha := sha256.New()
	sha.Write([]byte(s))
	return sha.Sum(nil)
}

func NewKeyPair() (pub string, pri string) {
	// assume that everything ok
	priKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubKey := &priKey.PublicKey

	pub = GetPublicKeyPem(pubKey)

	priBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priKey),
	}
	pri = string(pem.EncodeToMemory(priBlock))

	return pub, pri
}

func GetPublicKey(s string) *rsa.PublicKey {
	b, _ := pem.Decode([]byte(s))
	if b.Type != "PUBLIC KEY" {
		// handle error
		fmt.Println("type error: ", b.Type)
		return nil
	}
	k, err := x509.ParsePKIXPublicKey(b.Bytes)
	if err == nil {
		// should cast to a specified type
		key, ok := k.(*rsa.PublicKey)
		if ok {
			return key
		}
		fmt.Println("failed to cast to RSA PublicKey")
		return nil
	}
	// handle error
	fmt.Println(err)
	return nil
}

func GetPublicKeyPem(key *rsa.PublicKey) string {
	b, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		// handle error
		fmt.Println(err)
	}
	pubBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: b,
	}
	return string(pem.EncodeToMemory(pubBlock))
}

func GetPrivateKey(s string) *rsa.PrivateKey {
	b, _ := pem.Decode([]byte(s))
	if b.Type != "RSA PRIVATE KEY" {
		// handle error
		fmt.Println("type error: ", b.Type)
		return nil
	}
	k, err := x509.ParsePKCS1PrivateKey(b.Bytes)
	if err == nil {
		return k
	}
	// handle error
	fmt.Println(err)
	return nil
}

func Sign(pri *rsa.PrivateKey, msg string) string {
	hashed := SHA256Hash(msg)
	signed, err := rsa.SignPKCS1v15(nil, pri, crypto.SHA256, hashed)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	encoded := base64.StdEncoding.EncodeToString(signed)

	return string(encoded)
}

func Verify(pub *rsa.PublicKey, signed, compare string) error {
	hashed := SHA256Hash(compare)
	decoded, err := base64.StdEncoding.DecodeString(signed)
	if err != nil {
		return err
	}
	return rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed, []byte(decoded))
}
