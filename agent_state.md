# Agent State

## Current Architecture

This repo is `github.com/massivemoose/ovek-workflow-example` and builds two OCI images:

- App image: `ghcr.io/massivemoose/ovek-workflow-example-app:latest`, built from `Dockerfile` and `./cmd/app`.
- Digest workflow image: `ghcr.io/massivemoose/ovek-workflow-example-digest:latest`, built from `Dockerfile.workflow` and `./cmd/digest`.

Shared code lives under `internal/`:

- `internal/config`: app and workflow environment loading for Ovek-injected PocketBase secrets.
- `internal/pocketbase`: superuser auth, collection ensuring, signup writes, signup/result reads, and workflow result writes.
- `internal/workflow`: digest domain types, high-water-mark calculation, summary generation, and PocketBase timestamp decoding.
- `internal/app`: HTTP routes, signup validation, and server-rendered templates.

The app serves on `PORT` defaulting to `8080`, writes signups to PocketBase, and never sends PocketBase credentials to the browser. The homepage renders the signup form first and the latest 5 `workflow_results` records below it.

## Key Decisions

- Ovek V1 app-triggered workflow execution is intentionally not implemented.
- Future trigger notes are documented in `docs/future-workflow-triggers.md`; a future "Run digest now" button should call an app server route, not Brain directly from browser JavaScript.
- `workflow_results.latest_signup_created_at` is stored as text so empty/no-signup runs are simple and durable.
- PocketBase date fields are decoded with a custom parser because PocketBase returns timestamps like `2026-05-24 10:00:03.000Z`, not strict RFC3339.
- The PocketBase client does not mutate shared auth state during request handling.
- Digest signup listing no longer sends `sort=created`; the digest calculation sorts returned signup records in Go before computing the high-water mark. This avoids a workflow failure seen as `list signups returned HTTP 400`.
- PocketBase list errors now include up to 64 KiB of response body text so future API errors are visible in workflow logs and failed `workflow_results` summaries.

## Verification Completed

- `GOCACHE=/tmp/ovek-workflow-go-cache go test ./...`
- `git -c core.fsmonitor=false diff --check`
- `podman build --platform linux/amd64 -t ghcr.io/massivemoose/ovek-workflow-example-digest:latest -f Dockerfile.workflow .`

## Pending Tasks

- Publish the updated digest image through the GitHub workflow. Direct local `podman push ghcr.io/massivemoose/ovek-workflow-example-digest:latest` failed here with GHCR `403 Forbidden`.
- Rerun `ovek workflow run workflow-demo digest` after the GitHub workflow publishes the rebuilt digest image.
- If the workflow still fails, inspect `ovek workflow logs workflow-demo <RUN_ID> --no-follow`; PocketBase response bodies should now be included in the error.
