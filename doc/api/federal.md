# Federal API Document

## WebFinger

Resolve `username@domain` to an user's id.

### GET `/.well-known/webfinger?resource=acct:<username>@<domain>`

- RESPONSE: 200, 404

```json
[HEADER]Content-Type:application/jrd+json
{
  "subject": "acct:<username>@<domain>",
  "links": [
    {
      "rel": "self",
      "type": "application/activity+json",
      "herf": "https://id.of/user"
    }
  ]
}
```

## Object

### GET `/users/<username>`

### GET `/users/<username>/followers[?from=?]`

### GET `/users/<username>/followings[?from=?]`

### GET `/posts/<postID>`

## Activity

## POST `/inbox`, `/users/<username>/inbox`

## GET `/<username>/outbox[?from=?]`