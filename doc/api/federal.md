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

Content type: `application/activity+json` or `application/ld+json; profile="https://www.w3.org/ns/activitystreams"`

For `GET`, request `Accept` should be one of these. For `POST`, request `Content-Type` should be one of these.

### GET `/users/<username>`

RESPONSE: 200, 404

```
A Person Object presenting the specified user.
```

### GET `/users/<username>/followers[?from=?]`

RESPONSE: 200, 404

```
A Collection Object presenting the followers of the specified user.
```

### GET `/users/<username>/followings[?from=?]`

RESPONSE: 200, 404

```
A Collection Object presenting the followings of the specified user.
```

### GET `/posts/<postID>`

RESPONSE: 200, 403, 404

```
A Note Object presenting the specified post.
```

## Activity

## POST `/inbox`, `/users/<username>/inbox`

REQUEST: require [Mastodon style header signature](https://docs.joinmastodon.org/spec/security/)

```
Any Activity.
```

RESPONSE: 200

## GET `/<username>/outbox[?from=?]`

RESPONSE: 200, 404

```
A Collection Object presenting Actor's Outbox (Activities).
```