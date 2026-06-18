# Issue tracker: beans

Issues and PRDs for this repo live as "beans" managed by the `beans` CLI — a local-first, markdown-based issue tracker.

## Conventions

- **Create an issue**: `beans create "Title" -t <type> -s <status> -d "Description..."`. Use `--json` for machine-readable output. Always specify `-t` (one of: `milestone`, `epic`, `bug`, `feature`, `task`).
- **Read an issue**: `beans show --json <id>` or `beans query --json '{ bean(id: "<id>") { id title body status type } }'`.
- **List issues**: `beans list --json` with appropriate `-t`, `-s`, `-S` filters. Use `beans query` with GraphQL for complex filtering.
- **Comment on / update an issue**: `beans update <id> --body-append "Text"` to append, `beans update <id> --body-replace-old "old" --body-replace-new "new"` for exact replacements.
- **Change status**: `beans update <id> -s <status>` (one of: `todo`, `in-progress`, `draft`, `completed`, `scrapped`).
- **Close (complete)**: `beans update <id> -s completed`. Append a `## Summary of Changes` section before closing.
- **Close (scrap)**: `beans update <id> -s scrapped`. Append a `## Reasons for Scrapping` section before closing.

## Bean types

| Type | Purpose |
|------|---------|
| `milestone` | Target release or checkpoint; group work that ships together |
| `epic` | Thematic container for related work; should have children |
| `feature` | User-facing capability or enhancement |
| `bug` | Something broken that needs fixing |
| `task` | Concrete piece of work (chore, sub-task) |

## Bean statuses

| Status | Meaning |
|--------|---------|
| `draft` | Needs refinement before it can be worked on |
| `todo` | Ready to be worked on |
| `in-progress` | Currently being worked on |
| `completed` | Finished successfully |
| `scrapped` | Will not be done |

## Relationships

- **Parent**: Set with `beans update <id> --parent <parent-id>` for hierarchy (milestone → epic → feature → task/bug).
- **Blocked-by**: Set with `beans update <id> --blocked-by <other-id>` when this bean can't proceed until another is done.
- **Blocking**: Set with `beans update <id> --blocking <other-id>` when this bean blocks another.

## Concurrency

Use etags with `--if-match` to avoid conflicts:
```bash
ETAG=$(beans show <id> --etag-only)
beans update <id> --if-match "$ETAG" ...
```

## When a skill says "publish to the issue tracker"

Create a bean with `beans create "Title" -t <appropriate-type> -s todo -d "Body..."`.

## When a skill says "fetch the relevant ticket"

Run `beans show --json <id>`.
