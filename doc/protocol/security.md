# Security

## Public Key

When user's account is created, an `RSA` public/private key pair is generated automatically. The public key is part of `Person`, ready for being retrieved, and will be used for verify request signature.

```json
{
  "id": "https://id.of/user#main-key",
  "owner": "https://id.of/user",
  "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nGENERATED_PUBLIC_KEY\n-----END PUBLIC KEY-----\n"
}
```

## Http-Header Signature

```
Signature:
  keyId="https://id.of/user#main-key"
  headers="(request-target) host date"
  signature="signed_signature_string"
```

### Generate

First using `headers` to generate signature string to be signed. For example, if `headers` is `"(request-target) host date"` and requet header is

```
POST /inbox
Host: antarctica.sns
Date: date string
```

, it should generate signature string

```
(request-target): post /inbox
host: antarctica.sns
date: date string
```

Next the signature string is hashed and signed by `RSA-SHA256` (`RSASSA-PKCS1-v1_5` with `SHA-256`) with user's private key. The result is encoded in `Base64`.

### Verify

1. compose signature string (#1)
2. decrypt signatrue with user's public key (#2).
3. compare #1 and #2.

## Digest

When making POST request, a body-digest header is required. The digest is also hashed and signed by `RSA-SHA256`, encoded in `Base64`.