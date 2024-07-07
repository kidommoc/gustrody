# Activities

Activities are *actions* of actors. In version `1.0`, the only actor is `Person`.

## On Person

### Follow

Follow a user.

```json
{
  "@context": [],
  "id": "https://id.of/actor#follow/user@domain",
  "type": "Follow",
  "actor": "https://id.of/actor",
  "object": "https://id.of/actorToFollow"
}
```

### Accept

Accept a follow request.

```json
{
  "@context": [],
  "id": "https://instance.url/activity/tempID",
  "type": "Accept",
  "actor": "https://id.of/actor",
  "object": {
    "id": "https://id.of/follow",
    "type": "Follow",
    "andOther": "properties"
  }
}
```

### Reject

Reject a follow request.

```json
{
  "@context": [],
  "id": "https://instance.url/activity/tempID",
  "type": "reject",
  "actor": "https://id.of/actor",
  "object": {
    "id": "https://id.of/follow",
    "type": "Follow",
    "andOther": "properties"
  }
}
```

### Undo

`Undo` is supported for `Follow`.

```json
{
  "@context": [],
  "id": "https://instance.url/activity/tempID",
  "type": "Undo",
  "actor": "https://id.of/actor",
  "object": "https://id.of/activityToUndo"
}
```

### Future Supporting

- manully accept/reject follow request.

- `Block` on `Person` and `Undo` on `Block`.

## On Note

### Create

Publish a new note.

```json
{
  "@context": [],
  "id": "https://instance.url/activity/tempID",
  "type": "Create",
  "actor": "https://id.of/actor",
  "published": "utc-date",
  "to": [],
  "cc": [],
  "object": {
    "id": "https://id.of/note",
    "type": "Note",
    "andOther": "properties"
  }
}
```

### Update

Update a existing note.

```json
{
  "@context": [],
  "id": "https://instance.url/activity/tempID",
  "type": "Update",
  "actor": "https://id.of/actor",
  "published": "utc-date",
  "object": {
    "id": "https://id.of/noteToUpdate",
    "type": "Note",
    "andOther": "properties"
  }
}
```

### Delete

Delete a existing note.

```json
{
  "@context": [],
  "id": "https://instance.url/activity/tempID",
  "type": "Delete",
  "actor": "https://id.of/actor",
  "published": "utc-date",
  "object": "https://id.of/noteToDelete"
}
```

### Like

Like a note.

```json
{
  "@context": [],
  "id": "https://id.of/actor#likes/id",
  "type": "Like",
  "actor": "https://id.of/actor",
  "object": "https://id.of/noteToLike"
}
```

### Announce

Share a note.

```json
{
  "@context": [],
  "id": "https://id.of/actor#shares/id",
  "type": "Announce",
  "actor": "https://id.of/actor",
  "published": "utc-date",
  "to": [],
  "cc": [],
  "object": "https://id.of/noteToShare"
}
```

### Undo

`Undo` is supported for `Like` and `Announce`.

```json
{
  "@context": [],
  "id": "https://instance.url/activity/tempID",
  "type": "Undo",
  "actor": "https://id.of/actor",
  "object": "https://id.of/activityIDToUndo"
}
```