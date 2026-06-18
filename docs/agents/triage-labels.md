# Triage Labels

The skills speak in terms of five canonical triage roles. This file maps those roles to beans statuses and body annotations.

| Canonical role | Beans mapping | Meaning |
|---|---|---|
| `needs-triage` | New bean (no status set, or status not yet assigned) | Maintainer needs to evaluate this issue |
| `needs-info` | `draft` status | Waiting on reporter for more information |
| `ready-for-agent` | `todo` status + body tag `Triage: ready-for-agent` | Fully specified, ready for an AFK agent |
| `ready-for-human` | `todo` status + body tag `Triage: ready-for-human` | Requires human implementation |
| `wontfix` | `scrapped` status | Will not be actioned |

## How to apply

- **Set `needs-info`**: `beans update <id> -s draft`
- **Set `ready-for-agent`**: `beans update <id> -s todo` and `beans update <id> --body-append "Triage: ready-for-agent"`
- **Set `ready-for-human`**: `beans update <id> -s todo` and `beans update <id> --body-append "Triage: ready-for-human"`
- **Set `wontfix`**: `beans update <id> -s scrapped`

## How to query

- **Find beans needing triage**: `beans query --json '{ beans(filter: { status: [] }) { id title } }'` (beans without a status)
- **Find beans needing info**: `beans list --json -s draft`
- **Find ready-for-agent**: `beans query --json '{ beans(filter: { status: ["todo"], search: "Triage: ready-for-agent" }) { id title } }'`
- **Find ready-for-human**: `beans query --json '{ beans(filter: { status: ["todo"], search: "Triage: ready-for-human" }) { id title } }'`
- **Find wontfix**: `beans list --json -s scrapped`

## Transitioning between states

When moving a bean from one triage state to another, remove the old body tag and add the new one:

```bash
beans update <id> --body-replace-old "Triage: needs-triage" --body-replace-new "Triage: ready-for-agent"
```
