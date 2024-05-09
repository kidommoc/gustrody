# API Document

## Status Code

- *200*: succeed
- *401*: require login / failed login
- *403*: access not allowed
- *404*: not found
- *500*: server error

## Authentication and Authorization

### POST `/api/auth/login`

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

### POST `/api/auth/token`

SEND:
```json
{
  "session": "string",
  "refresh": "string"
}
```

RETURN: 200, 401, 500  
```json
{
  "token": "string",
  "refresh": "string"
}
```

## Timeline and Notification

### GET `/api/home[?from=<?>]`

SEND:
```json
REQUIRE PROVIDING SESSION IN HTTP HEADER: Session 
REQUIRE PROVIDING TOKEN IN HTTP HEADER: Authorization(Bearer)
```

RETURN:  
```json
NOT IMPLEMENTED
```

### GET `/api/public[?from=<?>]`

SEND:
```json
REQUIRE PROVIDING SESSION IN HTTP HEADER: Session 
REQUIRE PROVIDING TOKEN IN HTTP HEADER: Authorization(Bearer)
```

RETURN:  
```json
NOT IMPLEMENTED
```

### GET `/api/notification[?from=<?>]`

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

### GET `/api/users/<username>`

RETURN: 200, 404, 500  
```json
{
  "id": "string(url)",
  "username": "string",
  "bio": "string",
  "avatar": "string(url)", // NOT IMPLEMENTED
  "follows": "number(count)",
  "followings": "string(url)",
  "followed": "number(count)",
  "followers": "string(url)"
}
```

### GET `/api/users/<username>/posts`

SEND:
```json
OPTIONALLY PROVIDE SESSION IN HTTP HEADER: Session 
OPTIONALLY PROVIDE TOKEN IN HTTP HEADER: Authorization(Bearer)
```

RETURN: 200, 401, 403, 404, 500  
```json
{
  "list": [
    {
      "id": "string(url)",
      "user": {
        "id": "string(url)",
        "username": "string",
        "avatar": "string(url)" // NOT IMPLEMENETED
      },
      "publishedAt": "string(utc)",
      "content": "string"
    }, ...
  ]
}
```

### GET `/api/users/<username>/followings`

SEND:
```json
OPTIONALLY PROVIDE SESSION IN HTTP HEADER: Session 
OPTIONALLY PROVIDE TOKEN IN HTTP HEADER: Authorization(Bearer)
```

RETURN: 200, 401, 403, 404, 500  
```json
{
  "list": [
    {
      "id": "string(url)",
      "username": "string",
      "avatar": "string(url)" // NOT IMPLEMENETED
    }, ...
  ]
}
```

### GET `/api/users/<username>/followers`

SEND:
```json
OPTIONALLY PROVIDE SESSION IN HTTP HEADER: Session 
OPTIONALLY PROVIDE TOKEN IN HTTP HEADER: Authorization(Bearer)
```

RETURN: 200, 401, 403, 404, 500  
```json
{
  "list": [
    {
      "id": "string(url)",
      "username": "string",
      "avatar": "string(url)" // NOT IMPLEMENETED
    }, ...
  ]
}
```

### PUT `/api/users/follow/<username>`

SEND:
```json
REQUIRE PROVIDING SESSION IN HTTP HEADER: Session 
REQUIRE PROVIDING TOKEN IN HTTP HEADER: Authorization(Bearer)
```

RETURN: 200, 401, 404, 500  
```json
{
  "auth": {
    "token": "string",
    "refresh": "string"
  }
}
```

### DELETE `/api/users/follow/<username>`

SEND:
```json
REQUIRE PROVIDING SESSION IN HTTP HEADER: Session 
REQUIRE PROVIDING TOKEN IN HTTP HEADER: Authorization(Bearer)
```

RETURN: 200, 401, 404, 500  
```json
{
  "auth": {
    "token": "string",
    "refresh": "string"
  }
}
```

## Posts

### GET `/api/posts/<postId>`

SEND:
```json
OPTIONALLY PROVIDE SESSION IN HTTP HEADER: Session 
OPTIONALLY PROVIDE TOKEN IN HTTP HEADER: Authorization(Bearer)
```

RETURN: 200, 401, 403, 404, 500  
```json
{
  "auth": {
    "token": "string",
    "refresh": "string"
  }, // ONLY WHEN PROVIDED SESSION AND TOKEN
  "id": "string(url)",
  "user": {
    "id": "string(url)",
    "username": "string",
    "avatar": "string(url)" // NOT IMPLEMENETED
  },
  "publishedAt": "string(utc)",
  "content": "string"
}
```

### PUT `/api/posts`

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
{
  "auth": {
    "token": "string",
    "refresh": "string"
  }
}
```

### POST `/api/posts/<postId>`

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
{
  "auth": {
    "token": "string",
    "refresh": "string"
  },
}
```

### DELETE `/api/posts/<postId>`

SEND:
```json
REQUIRE PROVIDING SESSION IN HTTP HEADER: Session 
REQUIRE PROVIDING TOKEN IN HTTP HEADER: Authorization(Bearer)
```

RETURN: 200, 401, 404, 500  
```json
{
  "auth": {
    "token": "string",
    "refresh": "string"
  }
}
```

### PUT `/api/posts/<postId>/like`

SEND:
```json
REQUIRE PROVIDING SESSION IN HTTP HEADER: Session 
REQUIRE PROVIDING TOKEN IN HTTP HEADER: Authorization(Bearer)
```

RETURN: 200, 401, 403, 404, 500  
```json
{
  "auth": {
    "token": "string",
    "refresh": "string"
  }
}
```

### DELETE `/api/posts/<postId>/like`

SEND:
```json
REQUIRE PROVIDING SESSION IN HTTP HEADER: Session 
REQUIRE PROVIDING TOKEN IN HTTP HEADER: Authorization(Bearer)
```

RETURN: 200, 401, 403, 404, 500  
```json
{
  "auth": {
    "token": "string",
    "refresh": "string"
  }
}
```