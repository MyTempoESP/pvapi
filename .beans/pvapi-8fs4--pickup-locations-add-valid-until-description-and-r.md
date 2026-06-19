---
# pvapi-8fs4
title: 'Pickup locations: add valid_until, description, and rename pickup_at to starts_at'
status: todo
type: feature
priority: normal
created_at: 2026-06-19T02:02:53Z
updated_at: 2026-06-19T02:03:12Z
---

## Problem Statement

Pickup locations (pontos de retirada) currently have a single `datetime` field representing the start of availability, with no end time. Once created, a pickup location stays visible in public listings indefinitely, regardless of whether the pickup window has already passed. Shop admins also have no way to describe a pickup location (no `description` field).

## Solution

Add a `valid_until` timestamp that marks the end of a pickup location's availability window. After this time, the pickup location is excluded from public product listings but remains visible and editable in the shop admin's CRUD interface. Add a `description` text field to capture human-readable details about the pickup location. Rename the internal `pickup_at` database column to `starts_at` for clarity (the JSON API field name `datetime` stays unchanged).

## User Stories

1. As a shop admin, I want to set an end time (`valid_until`) when creating a pickup location, so that the pickup window is bounded and clients see only active pickups.
2. As a shop admin, I want to update the `valid_until` of an existing pickup location, so that I can postpone or extend the availability window.
3. As a shop admin, I want to see the `valid_until` in my pickup location listing, so that I can tell at a glance which pickups have expired.
4. As a shop admin, I want to provide a description when creating a pickup location, so that instructions or notes can accompany the location.
5. As a shop admin, I want to update the description of an existing pickup location, so that I can correct or improve it over time.
6. As a shop admin, I want the system to reject creation of a pickup location where `valid_until` is not after `datetime`, so that I don't accidentally create an invalid time window.
7. As a shop admin, I want update validation to reject a `valid_until` that is not after `datetime` when both fields are provided, so that the time window remains valid after edits.
8. As an anonymous client browsing public products, I should not see pickup locations whose `valid_until` has already passed, so that I only see currently available pickup options.
9. As an anonymous client viewing a specific public product, I should not see packages associated with expired pickup locations, so that I only see actionable purchase options.
10. As a shop admin, I want my own pickup location list to show ALL pickups (including expired ones), so that I can manage them regardless of their current state.

## Implementation Decisions

### Schema change

- Rename column `pickup_at` to `starts_at` (TIMESTAMPTZ NOT NULL).
- Add column `description` (TEXT NOT NULL).
- Add column `valid_until` (TIMESTAMPTZ NOT NULL).
- Migration follows the pattern: add column with a temporary default, then drop the default after backfill.

### JSON API contract

- The create request accepts `validUntil` (string, format `2006-01-02T15:04:05`) and `description` (string).
- The update request accepts `validUntil` and `description` as optional fields (partial update via COALESCE pattern).
- The admin-facing resource includes `validUntil` and `description` in responses.
- The public-facing resource (`ConsignmentPickupResource`) does NOT include `validUntil` or `description` — they are filtered server-side.
- The JSON key for the start time stays `datetime` for backward compatibility.

### Validation

- On create: reject if `valid_until` is not after `datetime` (parse error for bad format, 400 for invalid window).
- On update: if both `valid_until` and `datetime` are provided, reject if `valid_until` is not after the provided `datetime`.

### Public listing filter

- The `PickupLocationsByConsignmentCodeQuery` (used by the public consignment listing) gains `AND pl.valid_until > NOW()`.
- The `PackagesByProductIDs` query (used by the public single-product view) gains `AND pl.valid_until > NOW()`. This means a product whose consignments link only to expired pickups will have zero packages returned.

### Modules modified

- Database migration layer (new migration pair `000052`).
- Pickup location query definitions (CRUD + public listing queries).
- Pickup location model types (struct, blueprint, requests, resources).
- Pickup location database operations (Save, List, Find, Update).
- Pickup location controller (minimal — just passes through new request fields).

No new modules introduced.

## Testing Decisions

- Tests are HTTP-level integration tests following the pattern in `api_pickup_test.go`.
- Tests should cover: create with `validUntil` + `description`, create with `valid_until` not after `datetime` (expect 400), update `validUntil` (full and partial), update validation when both `datetime` and `validUntil` present, list includes new fields, list includes expired pickups for admin.
- Public listing tests: verify expired pickups are excluded from `/produtos?token=...` response.
- Good tests assert on externally visible behavior (HTTP status codes, response body fields) rather than internal state.

## Out of Scope

- Background jobs or cron to clean up expired pickup locations.
- Any UI or frontend changes.
- Changing the time format from `2006-01-02T15:04:05`.
- Any notification or alerting when a pickup location expires.

## Further Notes

- The `starts_at` column rename is internal-only. All JSON keys remain `datetime` for the start time.
- The `valid_until` constraint is enforced only at the application layer (Go validation), not as a database CHECK constraint. This keeps the constraint flexible if future requirements change.

Triage: ready-for-agent
