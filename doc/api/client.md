# Client API Document

## Status Code

- *200*: succeed
- *401*: require login / failed login
- *403*: access not allowed
- *404*: not found
- *500*: server error

## Shorthands

```json
"image" := {
  "type": "image/jpeg or imag/png",
  "url": "string(url)",
  "alt": "string"
}

"user-info" := {
  "id": "string(url)",
  "username": "string",
  "nickname": "string",
  "avatar": "image"
}

"vsb" := "public" | "follower" | "direct"
```

## Authentication and Authorization

### POST `/auth/login`

Login to get oauth token.

- REQUEST:

```json
[HEADER]Content-Type: application/json
{
  "username": "string",
  "password": "string(encrypted)"
}
```

- RESPONSE: 200, 401, 500  

```json
[HEADER]Content-Type: application/json
{
  "session": "string",
  "token": "string",
  "refresh": "string(for refreshing token)"
}
```

### POST `/auth/token`

Refresh token.

- REQUEST:

```
[HEADER]Session: (REQUIRED)
[HEADER]Authorization: Bearer REFRESH (REQUIRED)
```

- RESPONSE: 200, 401, 500  

```
[HEADER]Token:
[HEADER]Refresh:
```

## Users

### PUT `/users`

Register new user.

- REQUEST:

```json
[HEADER]Content-Type: application/json
{
  "username": "string",
  "nickname": "string",
  "password": "string(encrypted)"
}
```

- RESPONSE: 200, 400, 500

### POST `/users/password`

Edit *my* password.

- REQUEST:

```json
[HEADER]Content-Type: application/json
[HEADER]Session: (REQUIRED)
[HEADER]Authorization: Bearer (REQUIRED)
{
  "password": "string(encrypted)"
}
```

- RESPONSE: 200, 400, 401, 500

```
[HEADER]Token:
[HEADER]Refresh:
```

### GET `/users/profile`

Get *my* profile to edit.

- REQUEST:

```
[HEADER]Accept: application/json
[HEADER]Session: (REQUIRED)
[HEADER]Authorization: Bearer (REQUIRED)
```

- RESPONSE: 200, 401, 500

```json
[HEADER]Content-Type: application/json
[HEADER]Token:
[HEADER]Refresh:
{
  "nickname": "string",
  "summary": "string",
  "avatar": {
  "type": "image/jpeg or imag/png",
  "url": "string(url)"
}
}
```

### POST `/users/profile`

Edit *my* profile.

- REQUEST:

```json
[HEADER]Content-Type: application/json
[HEADER]Session: (REQUIRED)
[HEADER]Authorization: Bearer (REQUIRED)
{
  "nickname": "string",
  "summary": "string",
  "avatar": {
  "type": "image/jpeg or imag/png",
  "url": "string(url)"
}
}
```

- RESPONSE: 200, 400, 401, 500

```
[HEADER]Token:
[HEADER]Refresh:
```

### GET `/users/settings`

Get *my* user settings to edit.

- REQUEST:

```
[HEADER]Accept: application/json
[HEADER]Session: (REQUIRED)
[HEADER]Authorization: Bearer (REQUIRED)
```

- RESPONSE:

```json
[HEADER]Content-Type: application/json
[HEADER]Token:
[HEADER]Refresh:
{
  "locked": true,
  "postVsb": "string(enum)",
  "shareVsb": "string(enum)"
}
```

### POST `/users/settings`

Edit *my* user settings.

- REQUEST

```json
[HEADER]Content-Type: application/json
[HEADER]Session: (REQUIRED)
[HEADER]Authorization: Bearer (REQUIRED)
{
  "locked": true,
  "postVsb": "string(enum)",
  "shareVsb": "string(enum)"
}
```

- RESPONSE: 200, 400, 401, 500

### GET `/users/<username>`

Get a user's profile.

- REQUEST:

```
[HEADER]Accept: application/json
```

- RESPONSE: 200, 404, 500  

```json
[HEADER]Content-Type: application/json
{
  "id": "string(url)",
  "username": "string",
  "nickname": "string",
  "summary": "string",
  "avatar": "image",
  "follows": "number(count)",
  "followed": "number(count)",
}
```

### GET `/users/<username>/posts`

- REQUEST:

```
[HEADER]Accept: application/json
[HEADER]Session: (OPTIONAL)
[HEADER]Authorization: Bearer (OPTIONAL)
```

- RESPONSE: 200, 401, 403, 404, 500  

```json
[HEADER]Content-Type: application/json
[HEADER]Token: (ONLY WHEN PROVIDED SESSION AND TOKEN)
[HEADER]Refresh: (ONLY WHEN PROVIDED SESSION AND TOKEN)
{
  "list": [
    {
      "id": "string(url)",
      // note: when the post is shared from others,
      // "user" should be the publisher of content
      // and "sharedBy" should be one shares it.
      "user": "user-info",
      "replyTo": "user-info", // may be null
      "sharedBy": "user-info", // may be null
      "date": "string(rfc3339)",
      "content": "string",
      "attachments": [
        "image", ... // max 4
      ]
    }, ...
  ]
}
```

### GET `/users/<username>/followings`

Get a user's following list.

- REQUEST:

```
[HEADER]Accept: application/json
[HEADER]Session: (OPTIONAL)
[HEADER]Authorization: Bearer (OPTIONAL)
```

- RESPONSE: 200, 401, 403, 404, 500  

```json
[HEADER]Content-Type: application/json
[HEADER]Token: (ONLY WHEN PROVIDED SESSION AND TOKEN)
[HEADER]Refresh: (ONLY WHEN PROVIDED SESSION AND TOKEN)
[
  "user-info", ...
]
```

### GET `/users/<username>/followers`

Get a user's follower list.

- REQUEST:

```
[HEADER]Accept: application/json
[HEADER]Session: (OPTIONAL)
[HEADER]Authorization: Bearer (OPTIONAL)
```

- RESPONSE: 200, 401, 403, 404, 500  

```json
[HEADER]Content-Type: application/json
[HEADER]Token: (ONLY WHEN PROVIDED SESSION AND TOKEN)
[HEADER]Refresh: (ONLY WHEN PROVIDED SESSION AND TOKEN)
[
  "user-info", ...
]
```

### PUT `/users/<username>/follow`

Follow a user.

- REQUEST:

```
[HEADER]Session: (REQUIRED)
[HEADER]Authorization: Bearer (REQUIRED)
```

- RESPONSE: 200, 401, 404, 500  

```
[HEADER]Token:
[HEADER]Refresh:
```

### DELETE `/users/<username>/follow`

Unfollow a user.

- REQUEST:

```
[HEADER]Session: (REQUIRED)
[HEADER]Authorization: Bearer (REQUIRED)
```

- RESPONSE: 200, 401, 404, 500  

```
[HEADER]Token:
[HEADER]Refresh:
```

## Posts

### GET `/posts/<postID>`

Get a post.

- REQUEST:

```
[HEADER]Accept: application/json
[HEADER]Session: (OPTIONAL)
[HEADER]Authorization: Bearer (OPTIONAL)
```

- RESPONSE: 200, 401, 403, 404, 500  

```json
[HEADER]Content-Type: application/json
[HEADER]Token: (ONLY WHEN PROVIDED SESSION AND TOKEN)
[HEADER]Refresh: (ONLY WHEN PROVIDED SESSION AND TOKEN)
{
  "id": "string(url)",
  "user": "user-info",
  "date": "string(rfc3339)",
  "content": "string",
  "attachments": [
    "image", ... // max 4
  ],
  "replyings": [
    // posts list
  ],
  "replies": [
    // posts tree
  ]
}
```

### PUT `/posts`

Post a new post.

- REQUEST:

```json
[HEADER]Content-Type: application/json
[HEADER]Session: (REQUIRED)
[HEADER]Authorization: Bearer (REQUIRED)
{
  "visibility": "vsb",
  "content": "string",
  "attachments": [
    "image", ... // max 4
  ]
}
```

- RESPONSE: 200, 400, 401, 500  

```
[HEADER]Token:
[HEADER]Refresh:
```

### PUT `/posts/<postID>/reply`

Reply a post.

- REQUEST:

```json
[HEADER]Content-Type: application/json
[HEADER]Session: (REQUIRED)
[HEADER]Authorization: Bearer (REQUIRED)
{
  "content": "string",
  "attachments": [
    "image", ... // max 4
  ]
}
```

- RESPONSE: 200, 400, 401, 404, 500  

```
[HEADER]Token:
[HEADER]Refresh:
```

### POST `/posts/<postID>`

Edit a post of *me*.

- REQUEST:

```json
[HEADER]Content-Type: application/json
[HEADER]Session: (REQUIRED)
[HEADER]Authorization: Bearer (REQUIRED)
{
  "content": "string",
  "attachments": [
    "image", ... // max 4
  ]
}
```

- RESPONSE: 200, 400, 401, 404, 500  

```
[HEADER]Token:
[HEADER]Refresh:
```

### DELETE `/posts/<postID>`

Remove a post of *me*.

- REQUEST:

```
[HEADER]Session: (REQUIRED)
[HEADER]Authorization: Bearer (REQUIRED)
```

- RESPONSE: 200, 401, 404, 500  

```
[HEADER]Token:
[HEADER]Refresh:
```

### GET `/posts/<postID>/likes`

Get likes of a post.

- REQUEST:

```
[HEADER]Accept: application/json
[HEADER]Session: (OPTIONAL)
[HEADER]Authorization: Bearer (OPTIONAL)
```

- RESPONSE: 200, 401, 403, 404, 500  

```json
[HEADER]Content-Type: application/json
[HEADER]Token: (ONLY WHEN PROVIDED SESSION AND TOKEN)
[HEADER]Refresh: (ONLY WHEN PROVIDED SESSION AND TOKEN)
[
  "user-info", ...
]
```

### PUT `/posts/<postID>/like`

Like a post.

- REQUEST:

```
[HEADER]Session: (REQUIRED)
[HEADER]Authorization: Bearer (REQUIRED)
```

- RESPONSE: 200, 401, 403, 404, 500  

```
[HEADER]Token:
[HEADER]Refresh:
```

### DELETE `/posts/<postID>/like`

Unlike a post.

- REQUEST:

```
[HEADER]Session: (REQUIRED)
[HEADER]Authorization: Bearer (REQUIRED)
```

- RESPONSE: 200, 401, 403, 404, 500  

```
[HEADER]Token
[HEADER]Refresh
```

### GET `/posts/<postID>/shares`

Get shares of a post.

- REQUEST:

```
[HEADER]Accept: application/json
[HEADER]Session: (OPTIONAL)
[HEADER]Authorization: Bearer (OPTIONAL)
```

- RESPONSE: 200, 401, 403, 404, 500  

```json
[HEADER]Content-Type: application/json
[HEADER]Token: (ONLY WHEN PROVIDED SESSION AND TOKEN)
[HEADER]Refresh: (ONLY WHEN PROVIDED SESSION AND TOKEN)
[
  "user-info", ...
]
```

### PUT `/posts/<postID>/share`

Share a post.

- REQUEST:

```
[HEADER]Session: (REQUIRED)
[HEADER]Authorization: Bearer (REQUIRED)
```

- RESPONSE: 200, 401, 403, 404, 500  

```
[HEADER]Token:
[HEADER]Refresh:
```

### DELETE `/posts/<postID>/share`

Unshare a post.

- REQUEST:

```
[HEADER]Session: (REQUIRED)
[HEADER]Authorization: Bearer (REQUIRED)
```

- RESPONSE: 200, 401, 403, 404, 500  

```
[HEADER]Token:
[HEADER]Refresh:
```

## Files

### GET `/images/<filename>`

Get an image.

- RESPONSE: 200, 404, 500

```
[HEADER]Content-Type: image/jpeg or image/png
```

### PUT `/images`

Upload an image.

- REQUEST:

```
[HEADER]Content-Type: multipart/form-data
[HEADER]Session: (REQUIRED)
[HEADER]Authorization: Bearer (REQUIRED)
Form data...
```

- RESPONSE: 200, 401, 500  

```
[HEADER]Content-Type: text/plain
[HEADER]Token:
[HEADER]Refresh:
https://url.to/image
```

## Timeline and Notification

### GET `/home[?from=<?>]`

- REQUEST:

```
[HEADER]Session: (REQUIRED)
[HEADER]Authorization: Bearer (REQUIRED)
```

- RESPONSE:  

```json
NOT IMPLEMENTED
```

### GET `/public[?from=<?>]`

- REQUEST:

```
[HEADER]Session: (REQUIRED)
[HEADER]Authorization: Bearer (REQUIRED)
```

- RESPONSE:  

```json
NOT IMPLEMENTED
```

### GET `/notification[?from=<?>]`

- REQUEST:

```
[HEADER]Session: (REQUIRED)
[HEADER]Authorization: Bearer (REQUIRED)
```

- RESPONSE:  

```json
NOT IMPLEMENTED
```