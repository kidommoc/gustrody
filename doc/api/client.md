# API Document

## Status Code

- *200*: succeed
- *401*: require login / failed login
- *403*: access not allowed
- *404*: not found
- *500*: server error

## Authentication and Authorization

### POST `/auth/login`

SEND:
```json
{
  "username": "string",
  "password": "string(encrypted)"
}
```

RETURN: 200, 401, 500  
```json
{
  "session": "string",
  "token": "string",
  "refresh": "string(for refreshing token)"
}
```

### POST `/auth/token`

SEND:
```json
REQUIRE PROVIDING SESSION IN HTTP HEADER: Session
REQUIRE PROVIDING REFRESH TOKEN IN HTTP HEADER: Authorization(Bearer)
```

RETURN: 200, 401, 500  
```json
HEADER: Token
HEADER: Refresh
```

## Timeline and Notification

### GET `/home[?from=<?>]`

SEND:
```json
REQUIRE PROVIDING SESSION IN HTTP HEADER: Session
REQUIRE PROVIDING TOKEN IN HTTP HEADER: Authorization(Bearer)
```

RETURN:  
```json
NOT IMPLEMENTED
```

### GET `/public[?from=<?>]`

SEND:
```json
REQUIRE PROVIDING SESSION IN HTTP HEADER: Session
REQUIRE PROVIDING TOKEN IN HTTP HEADER: Authorization(Bearer)
```

RETURN:  
```json
NOT IMPLEMENTED
```

### GET `/notification[?from=<?>]`

SEND:
```json
REQUIRE PROVIDING SESSION IN HTTP HEADER: Session
REQUIRE PROVIDING TOKEN IN HTTP HEADER: Authorization(Bearer)
```

RETURN:  
```json
NOT IMPLEMENTED
```

## Users

### GET `/users/<username>`

RETURN: 200, 404, 500  
```json
{
  "id": "string(url)",
  "preferredUsername": "string",
  "name": "string",
  "summary": "string",
  "icon": "string(url)", // NOT IMPLEMENTED
  "follows": "number(count)",
  "followings": "string(url)",
  "followed": "number(count)",
  "followers": "string(url)"
}
```

### GET `/users/<username>/posts`

SEND:
```json
OPTIONALLY PROVIDE SESSION IN HTTP HEADER: Session
OPTIONALLY PROVIDE TOKEN IN HTTP HEADER: Authorization(Bearer)
```

RETURN: 200, 401, 403, 404, 500  
```json
// ONLY WHEN PROVIDED SESSION AND TOKEN
HEADER: Token
HEADER: Refresh
{
  "list": [
    {
      "id": "string(url)",
      "publisher": {
        "id": "string(url)",
        "preferredUsername": "string",
        "name": "string",
        "icon": "string(url)" // NOT IMPLEMENETED
      },
      "publishedAt": "string(rfc3339)",
      "content": "string"
    }, ...
  ]
}
```

### GET `/users/<username>/followings`

SEND:
```json
OPTIONALLY PROVIDE SESSION IN HTTP HEADER: Session
OPTIONALLY PROVIDE TOKEN IN HTTP HEADER: Authorization(Bearer)
```

RETURN: 200, 401, 403, 404, 500  
```json
// ONLY WHEN PROVIDED SESSION AND TOKEN
HEADER: Token
HEADER: Refresh
{
  "list": [
    {
      "id": "string(url)",
      "preferredUsername": "string",
      "name": "string",
      "icon": "string(url)" // NOT IMPLEMENETED
    }, ...
  ]
}
```

### GET `/users/<username>/followers`

SEND:
```json
OPTIONALLY PROVIDE SESSION IN HTTP HEADER: Session
OPTIONALLY PROVIDE TOKEN IN HTTP HEADER: Authorization(Bearer)
```

RETURN: 200, 401, 403, 404, 500  
```json
// ONLY WHEN PROVIDED SESSION AND TOKEN
HEADER: Token
HEADER: Refresh
{
  "list": [
    {
      "id": "string(url)",
      "preferredUsername": "string",
      "name": "string",
      "icon": "string(url)" // NOT IMPLEMENETED
    }, ...
  ]
}
```

### PUT `/users/<username>/follow`

SEND:
```json
REQUIRE PROVIDING SESSION IN HTTP HEADER: Session
REQUIRE PROVIDING TOKEN IN HTTP HEADER: Authorization(Bearer)
```

RETURN: 200, 401, 404, 500  
```json
HEADER: Token
HEADER: Refresh
```

### DELETE `/users/<username>/follow`

SEND:
```json
REQUIRE PROVIDING SESSION IN HTTP HEADER: Session
REQUIRE PROVIDING TOKEN IN HTTP HEADER: Authorization(Bearer)
```

RETURN: 200, 401, 404, 500  
```json
HEADER: Token
HEADER: Refresh
```

## Posts

### GET `/posts/<postId>`

SEND:
```json
OPTIONALLY PROVIDE SESSION IN HTTP HEADER: Session
OPTIONALLY PROVIDE TOKEN IN HTTP HEADER: Authorization(Bearer)
```

RETURN: 200, 401, 403, 404, 500  
```json
// ONLY WHEN PROVIDED SESSION AND TOKEN
HEADER: Token
HEADER: Refresh
{
  "id": "string(url)",
  "publisher": {
    "id": "string(url)",
    "preferredUsername": "string",
    "name": "string",
    "icon": "string(url)" // NOT IMPLEMENETED
  },
  "publishedAt": "string(rfc3339)",
  "content": "string"
}
```

### PUT `/posts`

SEND:
```json
REQUIRE PROVIDING SESSION IN HTTP HEADER: Session
REQUIRE PROVIDING TOKEN IN HTTP HEADER: Authorization(Bearer)
{
  "content": "string"
}
```

RETURN: 200, 401, 500  
```json
HEADER: Token
HEADER: Refresh
```

### POST `/posts/<postId>`

SEND:
```json
REQUIRE PROVIDING SESSION IN HTTP HEADER: Session
REQUIRE PROVIDING TOKEN IN HTTP HEADER: Authorization(Bearer)
{
  "content": "string"
}
```

RETURN: 200, 401, 404, 500  
```json
HEADER: Token
HEADER: Refresh
```

### DELETE `/posts/<postId>`

SEND:
```json
REQUIRE PROVIDING SESSION IN HTTP HEADER: Session
REQUIRE PROVIDING TOKEN IN HTTP HEADER: Authorization(Bearer)
```

RETURN: 200, 401, 404, 500  
```json
HEADER: Token
HEADER: Refresh
```

### PUT `/posts/<postId>/like`

SEND:
```json
REQUIRE PROVIDING SESSION IN HTTP HEADER: Session
REQUIRE PROVIDING TOKEN IN HTTP HEADER: Authorization(Bearer)
```

RETURN: 200, 401, 403, 404, 500  
```json
HEADER: Token
HEADER: Refresh
```

### DELETE `/posts/<postId>/like`

SEND:
```json
REQUIRE PROVIDING SESSION IN HTTP HEADER: Session
REQUIRE PROVIDING TOKEN IN HTTP HEADER: Authorization(Bearer)
```

RETURN: 200, 401, 403, 404, 500  
```json
HEADER: Token
HEADER: Refresh
```

### PUT `/posts/<postId>/share`

SEND:
```json
REQUIRE PROVIDING SESSION IN HTTP HEADER: Session
REQUIRE PROVIDING TOKEN IN HTTP HEADER: Authorization(Bearer)
```

RETURN: 200, 401, 403, 404, 500  
```json
HEADER: Token
HEADER: Refresh
```

### DELETE `/posts/<postId>/share`

SEND:
```json
REQUIRE PROVIDING SESSION IN HTTP HEADER: Session
REQUIRE PROVIDING TOKEN IN HTTP HEADER: Authorization(Bearer)
```

RETURN: 200, 401, 403, 404, 500  
```json
HEADER: Token
HEADER: Refresh
```