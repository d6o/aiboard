---
metadata.openclaw:
  emoji: "\U0001F4CB"
  requires:
    bins:
      - curl
    env:
      - AIBOARD_URL
    config:
      - aiboard_user_id
  install: |
    echo "Verify AIBoard is reachable..."
    curl -sf "${AIBOARD_URL:-http://localhost:8080}/api/users" > /dev/null || {
      echo "AIBoard is not running. Start it with: make run"
      exit 1
    }
    echo "AIBoard connection verified."
---

# AIBoard

AIBoard is a kanban board designed for AI agents and humans to manage work together. You use it to create tasks, track progress, coordinate with teammates, and stay informed about changes happening across the board. Think of it as your team's shared workspace where every card, comment, and status change is driven through a REST API that you call directly.

Use this skill when the user asks you to manage tasks, track work items, update project status, coordinate with teammates, triage backlogs, break down features into subtasks, report progress, or interact with a kanban board. Also use it when the user mentions cards, tickets, todos, or anything related to project management on AIBoard.

## How AIBoard Works

The board has three columns: **Todo**, **Doing**, and **Done**. Cards move freely between them in any direction. There is no enforced linear flow, so you can move a card from Done back to Todo if work needs to be revisited.

There is no authentication. You identify yourself by passing a `user_id` field with every write request. This means your first action should always be figuring out who you are on the board.

Every response from the API follows the same shape. Successful responses wrap the result in a `data` field. Errors wrap the problem in an `error` field with a machine-readable code, a human-readable message, and for validation failures, a list of which fields had problems. This consistency means you can always check for `error` in the response to know if something went wrong.

All resource IDs are UUIDs. All timestamps are ISO 8601 with timezone.

## Configuration

The base URL defaults to `http://localhost:8080`. Override it by setting the `AIBOARD_URL` environment variable. Your agent identity is stored in the `aiboard_user_id` config key. If no user ID is configured, list all users first and select one.

## Usage Examples

Here are the kinds of requests a user might make, and how you should handle them:

**"Show me what's on the board"** — Fetch all cards with `GET /api/cards` and summarize them grouped by column. Mention the card count per column, highlight high-priority items, and flag any cards where all child cards are done but the parent has not moved to Done yet.

**"Create a card for implementing dark mode"** — Call `POST /api/cards` with the title, a sensible priority, column set to `todo`, and both reporter and assignee set to the current user unless the user specifies otherwise. Always include an `Idempotency-Key` header when creating resources.

**"Move the auth refactor card to done"** — First search for the card by listing cards and matching the title. Then call `PATCH /api/cards/{id}/move` with `column: done`. Tell the user that moving to Done will notify the reporter.

**"Break down the API migration into subtasks"** — Find the card, then create child cards with `POST /api/cards` setting `parent_id` to the parent card's ID. Each child card is a full card with its own assignee, priority, and column. You can assign different team members to each child.

**"What happened on the board today"** — Query the activity log with `GET /api/activity` and summarize recent actions. Group by card if that makes the output clearer.

**"Tag the search card as urgent"** — List available tags with `GET /api/tags`, find the urgent tag's ID, then attach it with `POST /api/cards/{card_id}/tags`.

**"Leave a comment telling Alice the PR is ready"** — Post a comment containing `@Alice` in the text. The system parses mentions automatically and sends notifications. You do not need to trigger notifications yourself.

**"Check my notifications"** — Call `GET /api/users/{user_id}/notifications?unread=true` and present unread items. Offer to mark them as read.

**"Start fresh"** or **"Reset the board"** — Call `POST /api/board/reset`. Warn the user this wipes everything: all users, cards, tags, activity, and notifications. After reset the board is completely empty. Default tags will be re-seeded on the next server restart.

**"Remove that user"** — Call `DELETE /api/users/{id}`. This only works if the user is not a reporter or assignee on any card. If they are, tell the user which cards reference them and suggest reassigning first.

## Implementation Details

### Identifying Yourself

Before any write operation, you need a user ID. Call `GET /api/users` to get the list of users. The board starts with no users. If the list is empty, you must create a user first with `POST /api/users` passing `name` and `avatar_color` (a hex color string like `#3B82F6`). If the user has told you who they are or if a user ID is configured, use that. Otherwise, ask.

Every request that creates, updates, or deletes something requires a `user_id` field in the JSON body or as a query parameter on DELETE requests. Never omit it.

To delete a user, call `DELETE /api/users/{id}`. This fails with `USER_IN_USE` if the user is a reporter or assignee on any card. Reassign or delete those cards first.

### Cards

Cards are the core unit of work. Each card has a title, optional description, priority from 1 (highest) to 5 (lowest), a column, a reporter, an assignee, an optional parent card, and optional tags, child cards, and comments.

To list cards, call `GET /api/cards`. You can filter with query parameters that combine freely: `column`, `assignee_id`, `reporter_id`, `tag_id`, `priority`, `parent_id`. For example, `GET /api/cards?column=doing&priority=1` returns only high-priority cards currently in progress. Use `parent_id` to list children of a specific card.

To get a single card with all its details including children, comments, and tags, call `GET /api/cards/{id}`. The response includes `child_total` and `child_completed` counts, plus a `children` array with the child card objects.

To create a card, `POST /api/cards` with this shape:

```json
{
  "title": "Implement user search",
  "description": "Add search endpoint filtering users by name prefix",
  "priority": 2,
  "column": "todo",
  "reporter_id": "uuid",
  "assignee_id": "uuid",
  "parent_id": "uuid (optional)",
  "user_id": "uuid"
}
```

Title, priority, column, reporter_id, and assignee_id are all required. The API rejects requests missing any of these. Description defaults to empty. Set `parent_id` to create a child card under another card. Child cards appear on the board like any other card and can have their own children (recursive nesting).

To update a card, `PUT /api/cards/{id}` with all fields. This is a full replace. Always read the card first so you can preserve fields the user did not ask to change.

To move a card between columns without touching other fields, `PATCH /api/cards/{id}/move` with `column` and `user_id`. This is the preferred way to change a card's column because it records the move in the activity log with the old and new column names.

To delete a card, `DELETE /api/cards/{id}?user_id={user_id}`.

When a card moves to Done, the system automatically notifies the card's reporter. This only fires on the transition into Done, not on further updates to a card already in Done.

**Important constraint:** A parent card cannot move to Done until all of its child cards are in Done. Attempting to do so returns a `CHILDREN_NOT_DONE` error. When the last child card of a parent moves to Done, the system sends a notification to the parent card's assignee that all children are complete.

### Tags

Tags are global labels with a name and a color. Seven default tags ship with the board: bug, feature, enhancement, urgent, design, backend, and frontend.

List all tags with `GET /api/tags`. Create a new one with `POST /api/tags` passing `name`, `color` (hex string like `#9333ea`), and `user_id`. Delete one with `DELETE /api/tags/{id}?user_id={user_id}`.

Attach a tag to a card with `POST /api/cards/{card_id}/tags` passing `tag_id` and `user_id`. Detach with `DELETE /api/cards/{card_id}/tags/{tag_id}?user_id={user_id}`.

Tag names must be unique. Attaching a tag that is already on a card returns a `TAG_ALREADY_ATTACHED` error. Handle it gracefully by telling the user the tag is already there rather than reporting a failure.

### Comments and Mentions

Comments are text entries on a card. List them with `GET /api/cards/{card_id}/comments`. Create one with `POST /api/cards/{card_id}/comments` passing `content` and `user_id`. Delete one with `DELETE /api/cards/{card_id}/comments/{id}?user_id={user_id}`.

When a comment contains `@Username` (for example `@Alice`), the system matches against existing user names and sends a notification to each mentioned user. You do not need to do anything beyond including the mention in the comment text. The matching is case-insensitive on the first word-character sequence after the `@`.

When composing comments that mention users, always use the exact user name as it appears in the user list. Misspelled mentions are silently ignored.

### Notifications

Notifications are messages delivered to a specific user. Three events create them:

1. Someone mentions you with `@YourName` in a comment.
2. All child cards of a card assigned to you move to Done.
3. A card you reported moves to the Done column.

Fetch notifications with `GET /api/users/{user_id}/notifications`. Add `?unread=true` to see only unread ones.

Mark one as read with `PATCH /api/users/{user_id}/notifications/{id}/read`. Mark all as read with `PATCH /api/users/{user_id}/notifications/read-all`.

Check notifications periodically when working on multi-step tasks. They tell you when teammates respond to your work or when tasks you are tracking reach completion.

### Activity Log

Every mutation on the board is recorded. Query it with `GET /api/activity`. Filter with `card_id`, `user_id`, or `action` query parameters. All filters are optional and combinable.

Actions recorded: `created`, `updated`, `moved`, `deleted`, `attached`, `detached`.

Each entry includes what changed, which resource was affected, who did it, when, and a details string with context like "moved from todo to doing" or "card created".

Use the activity log to answer questions like "what did Bob do today" (`GET /api/activity?user_id={bob_id}`) or "what happened to this card" (`GET /api/activity?card_id={card_id}`).

### Board Reset

To wipe all data and start fresh, call `POST /api/board/reset`. This truncates every table: users, cards, tags, comments, notifications, activity log, and idempotency keys. The board will be completely empty afterward. Default tags are re-seeded on the next server restart but not immediately after reset.

This is a destructive operation. Always confirm with the user before calling it. There is no undo.

### Idempotency

All POST endpoints support an `Idempotency-Key` header. When you include this header and the server has already processed a request with that key, it returns the original response instead of creating a duplicate resource.

Always include an idempotency key when creating cards, comments, tags, or users. Use a descriptive key that ties to the logical operation, like `create-card-dark-mode-v1` or `create-child-{parent_id}-write-tests`. This protects against accidental duplicates when retrying after network errors or timeouts.

## Handling Errors

When something goes wrong, the API returns a structured error you can reason about:

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

The error codes you will encounter and what to do about them:

**VALIDATION_ERROR** — You sent bad input. Read the `fields` array to see exactly which fields failed and why. Fix each one and retry. Common causes: blank title, priority outside 1-5 range, missing reporter or assignee.

**NOT_FOUND** — The resource ID does not exist. Verify you have the correct ID. If you were searching by title, re-list the resources to find the current ID.

**DUPLICATE_NAME** — A user or tag with that name already exists. If creating a user, fetch the existing one instead. If creating a tag, find and use the existing tag.

**CHILDREN_NOT_DONE** — You tried to move a parent card to Done but some of its child cards are not in Done yet. List the children, move them to Done first, then retry.

**TAG_ALREADY_ATTACHED** — The tag is already on this card. This is not a real failure. Tell the user the tag was already there and move on.

**USER_IN_USE** — You tried to delete a user who is a reporter or assignee on existing cards. Reassign or delete those cards first, then retry the delete.

**INVALID_JSON** — The request body was not valid JSON. Check your payload structure.

**INTERNAL_ERROR** — Something broke server-side. Retry the request with an idempotency key. If it persists, report the issue to the user.

## Edge Cases and Preferences

When searching for a card by title, always list cards first and do a case-insensitive match. Never guess card IDs.

When updating a card, read it first to get current values for any fields the user did not mention. The update endpoint is a full replace, so sending a blank description when the user only asked to change the priority would erase the existing description.

When breaking work into child cards, keep each title short, specific, and distinct. Assign different team members to child cards when the user mentions delegation. Child cards appear on the board alongside other cards.

When the user asks to "finish" or "complete" a card, move it to Done rather than deleting it. Only delete cards when explicitly asked.

When posting comments on behalf of the user, write in a natural, professional tone. Include `@mentions` when the user asks to notify someone, but do not add unsolicited mentions.

Prefer the move endpoint over update when only the column is changing. It produces cleaner activity log entries.

When the user asks broad questions about project status, combine card listing, child card progress, and activity log data to give a complete picture rather than dumping raw API responses.
