# Objects

Objects are entities *carrying data* in ActivityPub.

## Person

`Person` is the actor presenting a user. Every activity is acted by an actor.

```json
{
  "@context": [
    "https://www.w3.org/ns/activitystreams",
    "https://w3id.org/security/v1",
    {
      "manuallyApprovesFollowers": "as:manuallyApprovesFollowers"
    }
  ],
  "id": "https://instance.url/users/username",
  "type": "Person",
  "preferredUsername": "username",
  "name": "User's Nickname",
  "published": "utc-date",
  "inbox": "https://id.of/person/inbox",
  "outbox": "https://id.of/person/outbox",
  "endpoints": {
    "sharedInbox": "https://instance.url/inbox"
  },
  "followers": "https://id.of/person/followers",
  "following": "https://id.of/person/following",
  "manuallyApprovesFollowers": true,
  "summary": "User's bio.",
  "icon": { // used as avatar
    "type": "Image",
    "mediaType": "image/jpeg or image/png",
    "url": "https://url.of/avatarImage"
  },
  "publicKey": { // see security.md
    "id": "https://id.of/person#main-key",
    "andOther": "properties"
  }
}
```

### Future Supporting

- `manuallyApprovesFollowers`

## Note

`Note` is the user-publishing content (post), composing the feed user comsumes.

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "https://instance.url/users/publisherID/status/noteID",
  "type": "Note",
  "inReplyTo": "https://id.of/noteToReply",
  "published": "utc-date",
  "url": "https://instance.url/@publisherID/noteID",
  "attributedTo": "https://id.of/publisher",
  "to": [],
  "cc": [],
  "content": "<p>content in html</p>",
  "attachment": [],
  "replies": {
      "id": "https://instance.url/publisher/noteID/replies",
      "type": "Collection",
      "andOther": "properties"
  }
}
```

### Future Supporting

- `Mention` tag

- Custom `Emoji` tag

### Visibility of Note

- Public

```json
{
  "to": [
    "as:Public"
  ],
  "cc": [
    "https://uri.to/users/actorID/followers",
    "https://inbox.of/actorsMentioned" // endpoints first
  ]
}
```

- Followers Only

```json
{
  "to": [
    "https://uri.to/users/actorID/followers",
    "https://inbox.of/actorsMentioned" // endpoints first
  ]
}
```

- Direct Message


```json
{
  "to": [
    "https://inbox.of/actorsMentioned" // endpoints first
  ]
}
```

- when Update

```json
{
    "to": [
        "https://id.of/actorsAnnounced"
    ]
}
```

## Collection

Representing a list, such as followers or outbox content.

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "https://instance.url/users/ownerID/collectionID",
  "type": "Collection",
  "totalItems": 128,
  "first": "https://id.of/collection?page=1"
}
```

- It's ok to embed a `CollectoinPage` in `Collection.first`

### CollectionPage

```json
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "https://id.of/collection?page=PAGE",
  "type": "CollectionPage",
  "totalItems": 128,
  "next": "https://id.of/collection?page=NEXT_PAGE",
  "partof": "https://id.of/collection",
  "items": [
    "https://id.of/item1",
    "https://id.of/item2",
    "https://id.of/item3",
    "https://id.of/item4",
  ]
}
```
## Media

`Media` is the payload `Note.attachment` may carry. In version `1.0`, only `Image` is supported.

### Image

```json
{
  "type": "Document",
  "mediaType": "image/jpeg or image/png",
  "url": "https://instance.url/imgs/filename.ext",
  "name": "alternative text",
  "sensitive": true,
  "blurhash": "blurhash string"
}
```

### Future Supporting

- sensitive and Blurhash

- Audio

- Video