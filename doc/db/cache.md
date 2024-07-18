# Cache Database

Use Redis. Max memory policy: Least frequently used.

## Regular cache

`user:<USERNAME>`: the profile of the specified user (not fedral person). Type: `STRING` of `json`.

`post:<POST_ID>`: the specified post (not federal note). Type: `STRING` of `json`.

`posts:<USERNAME>`: the posts and shares of the specified user. Type: `LIST` of `json` array as cached page.

> page syntax:
>
> ```json
> [
>   // ordered by date descendingly
>   {"id": "<type>:<post_id>", "date": "datetime string"}, ...
> ]
> ```