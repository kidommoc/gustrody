## Authentication and Authorization

**POST** `/api/auth/login`

**GET** `/api/auth/token`

## Timeline and Notification

**GET** `/api/home[?from=<?>]`

**GET** `/api/public[?from=<?>]`

**GET** `/api/notification[?from=<?>]`

## Users

**GET** `/api/users/<username>`

**GET** `/api/users/<username>/posts`

## Posts

**GET** `/api/posts/<postId>`

**PUT** `/api/posts`

**POST** `/api/posts/<postId>`

**DELETE** `/api/posts/<postId>`