# Activities

Activities are *actions* of actors. In version `1.0`, the only actor is `Person`.

## On Person

### Follow

Follow a user.

```json
{
  "@context": [],
  "id": "https://id.of/actor/activity/uuid",
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
  "id": "https://id.of/actor/activity/uuid",
  "type": "Accept",
  "actor": "https://id.of/actor",
  "object": {
    "id": "https://id.of/following",
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
  "id": "https://id.of/actor/activity/uuid",
  "type": "reject",
  "actor": "https://id.of/actor",
  "object": {
    "id": "https://id.of/following",
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
  "id": "https://id.of/actor/activity/uuid",
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
  "id": "https://instance.url/users/actor/status/noteID/activity",
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
  "id": "",
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
  "id": "",
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
  "id": "https://instance.url/users/actorID#likes/id",
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
  "id": "https://instance.url/users/actorID/status/forwardID/activity",
  "type": "Announce",
  "actor": "https://id.of/actor",
  "published": "utc-date",
  "to": [],
  "cc": [],
  "object": "https://id.of/noteToForward"
}
```

### Undo

`Undo` is supported for `Like` and `Announce`.

```json
{
  "@context": [],
  "id": "https://instance.url/tempID",
  "type": "Undo",
  "actor": "https://id.of/actor",
  "object": "https://id.of/activityIDToUndo"
}
```