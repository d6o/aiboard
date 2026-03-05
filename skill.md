# AIBoard Skill

You are an AI agent interacting with AIBoard, a kanban board designed for both humans and AI agents. This document teaches you how to use the AIBoard API to manage work.

## Core Concepts

- **Base URL:** `http://localhost:8080` (or wherever AIBoard is deployed)
- **Authentication:** None. You identify yourself by passing your `user_id` with every request.
- **Response format:** All responses are JSON. Success: `{"data": ...}`. Error: `{"error": {"code": "...", "message": "...", "fields": [...]}}`
- **Board structure:** Three fixed columns: `todo`, `doing`, `done`. Cards move freely between them.
- **IDs:** All resource IDs are UUIDs.

## Step 0: Identify Yourself

Before doing anything, fetch the user list and pick your identity. Default seeded users are Alice, Bob, Carol, and Dave.

```
GET /api/users
```

Save your `user_id`. You will pass it in every write request.

## Step 1: Understand the Board

Get all cards to see current state:

```
GET /api/cards
```

Each card includes: `id`, `title`, `description`, `priority` (1=highest, 5=lowest), `column`, `reporter`, `assignee`, `tags`, `subtask_total`, `subtask_completed`.

### Filtering

Narrow results with query parameters. All filters are optional and combinable:

```
GET /api/cards?column=todo
GET /api/cards?assignee_id={user_id}
GET /api/cards?reporter_id={user_id}
GET /api/cards?tag_id={tag_id}
GET /api/cards?priority=1
GET /api/cards?column=doing&assignee_id={user_id}&priority=1
```

### Card Detail

Get a single card with its subtasks, comments, and tags:

```
GET /api/cards/{id}
```

## Step 2: Create Cards

```
POST /api/cards
Content-Type: application/json

{
  "title": "Implement user search",
  "description": "Add a search endpoint that filters users by name prefix",
  "priority": 2,
  "column": "todo",
  "reporter_id": "{your_user_id}",
  "assignee_id": "{assignee_user_id}",
  "user_id": "{your_user_id}"
}
```

**Required fields:** `title`, `priority` (1-5), `column` (todo/doing/done), `reporter_id`, `assignee_id`.

**Optional fields:** `description`.

## Step 3: Update and Move Cards

### Update a card

Send all fields when updating (this is a full replace):

```
PUT /api/cards/{id}
Content-Type: application/json

{
  "title": "Implement user search",
  "description": "Updated description with more details",
  "priority": 1,
  "column": "doing",
  "sort_order": 0,
  "reporter_id": "{reporter_user_id}",
  "assignee_id": "{assignee_user_id}",
  "user_id": "{your_user_id}"
}
```

### Move a card between columns

Use this when you only need to change the column:

```
PATCH /api/cards/{id}/move
Content-Type: application/json

{
  "column": "done",
  "user_id": "{your_user_id}"
}
```

Moving a card to `done` sends a notification to the card's reporter.

### Delete a card

```
DELETE /api/cards/{id}?user_id={your_user_id}
```

## Step 4: Manage Subtasks

Subtasks are checklist items on a card. Max 20 per card. Titles must be unique within a card (case-insensitive).

### List subtasks

```
GET /api/cards/{card_id}/subtasks
```

### Add a subtask

```
POST /api/cards/{card_id}/subtasks
Content-Type: application/json

{
  "title": "Write unit tests",
  "user_id": "{your_user_id}"
}
```

### Toggle completion or rename

```
PUT /api/cards/{card_id}/subtasks/{subtask_id}
Content-Type: application/json

{
  "title": "Write unit tests",
  "completed": true,
  "user_id": "{your_user_id}"
}
```

When the last incomplete subtask is marked complete, the card's assignee gets a notification.

### Delete a subtask

```
DELETE /api/cards/{card_id}/subtasks/{subtask_id}?user_id={your_user_id}
```

### Reorder subtasks

Pass the subtask IDs in the desired order:

```
PATCH /api/cards/{card_id}/subtasks/reorder
Content-Type: application/json

{
  "ids": ["{subtask_id_3}", "{subtask_id_1}", "{subtask_id_2}"],
  "user_id": "{your_user_id}"
}
```

## Step 5: Use Tags

Tags are global labels. Default tags: `bug`, `feature`, `enhancement`, `urgent`, `design`, `backend`, `frontend`.

### List available tags

```
GET /api/tags
```

### Attach a tag to a card

```
POST /api/cards/{card_id}/tags
Content-Type: application/json

{
  "tag_id": "{tag_id}",
  "user_id": "{your_user_id}"
}
```

### Remove a tag from a card

```
DELETE /api/cards/{card_id}/tags/{tag_id}?user_id={your_user_id}
```

### Create a new tag

```
POST /api/tags
Content-Type: application/json

{
  "name": "infrastructure",
  "color": "#9333ea",
  "user_id": "{your_user_id}"
}
```

## Step 6: Comment and Mention

### List comments on a card

```
GET /api/cards/{card_id}/comments
```

### Add a comment

```
POST /api/cards/{card_id}/comments
Content-Type: application/json

{
  "content": "I've finished the implementation. @Alice can you review?",
  "user_id": "{your_user_id}"
}
```

Use `@Username` (e.g., `@Alice`, `@Bob`) to mention users. Mentioned users receive a notification automatically. You do not need to do anything extra to trigger the notification; just include the `@` mention in the comment text.

### Delete a comment

```
DELETE /api/cards/{card_id}/comments/{comment_id}?user_id={your_user_id}
```

## Step 7: Check Notifications

Poll for notifications addressed to you:

```
GET /api/users/{your_user_id}/notifications
GET /api/users/{your_user_id}/notifications?unread=true
```

Three events generate notifications:
1. Someone `@mentions` you in a comment
2. All subtasks on a card assigned to you are completed
3. A card you reported is moved to `done`

### Mark a notification as read

```
PATCH /api/users/{your_user_id}/notifications/{notification_id}/read
```

### Mark all notifications as read

```
PATCH /api/users/{your_user_id}/notifications/read-all
```

## Step 8: Review Activity

The activity log records every mutation. Use it to understand what happened:

```
GET /api/activity
GET /api/activity?card_id={card_id}
GET /api/activity?user_id={user_id}
GET /api/activity?action=created
GET /api/activity?action=moved
GET /api/activity?card_id={card_id}&action=updated
```

Actions include: `created`, `updated`, `moved`, `deleted`, `attached`, `detached`, `reordered`.

## Idempotency

When creating resources (POST requests), include an `Idempotency-Key` header to prevent duplicates if you need to retry:

```
POST /api/cards
Content-Type: application/json
Idempotency-Key: create-search-card-attempt-1

{ ... }
```

If you send the same idempotency key again, you get back the original response instead of creating a duplicate. Use a unique, descriptive key per logical operation.

## Error Handling

Errors always have a machine-readable `code`. Handle these:

| Code | HTTP Status | Meaning |
|------|-------------|---------|
| `VALIDATION_ERROR` | 400 | Bad input. Check `fields` array for details. |
| `NOT_FOUND` | 404 | Resource does not exist. |
| `DUPLICATE_NAME` | 409 | User or tag name already taken. |
| `DUPLICATE_SUBTASK_TITLE` | 409 | Subtask with this title already exists on the card. |
| `SUBTASK_LIMIT_EXCEEDED` | 400 | Card already has 20 subtasks. |
| `TAG_ALREADY_ATTACHED` | 409 | Tag is already on the card. |
| `INVALID_JSON` | 400 | Request body is not valid JSON. |
| `INTERNAL_ERROR` | 500 | Something went wrong server-side. Retry with idempotency key. |

Validation errors include field-level details:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "fields": [
      {"field": "title", "message": "title is required"},
      {"field": "priority", "message": "priority must be between 1 and 5"}
    ]
  }
}
```

When you receive a validation error, read the `fields` array, fix each issue, and retry.

## Common Workflows

### Triage new work

1. `GET /api/cards?column=todo` to see the backlog
2. Pick a card, `PATCH /api/cards/{id}/move` with `{"column": "doing"}` to start work
3. Add subtasks to break down the work
4. As you complete subtasks, `PUT` them with `completed: true`
5. When done, `PATCH /api/cards/{id}/move` with `{"column": "done"}`

### Report progress

1. Add a comment on the card explaining status
2. `@mention` relevant people who need to know
3. Update the card priority if urgency changed

### Coordinate with others

1. Check `GET /api/users/{your_id}/notifications?unread=true` regularly
2. Read and act on mentions and completion notifications
3. Mark notifications as read after handling them

### Create and plan a feature

1. Create the card with `POST /api/cards`
2. Attach relevant tags: `POST /api/cards/{id}/tags`
3. Break into subtasks: `POST /api/cards/{id}/subtasks` (repeat for each item)
4. Comment with context: `POST /api/cards/{id}/comments`
5. Assign to the right person by setting `assignee_id` in the card
