# Agent State

## Current Architecture

This repo is now `github.com/massivemoose/ovek-workflow-example` and builds two OCI images:

- App image: `ghcr.io/massivemoose/ovek-workflow-example-app:latest`, built from `Dockerfile` and `./cmd/app`.
- Digest workflow image: `ghcr.io/massivemoose/ovek-workflow-example-digest:latest`, built from `Dockerfile.workflow` and `./cmd/digest`.

Shared code lives under `internal/`:

- `internal/config`: app and workflow environment loading for Ovek-injected PocketBase secrets.
- `internal/pocketbase`: superuser auth, collection ensuring, signup writes, signup/result reads, and workflow result writes.
- `internal/workflow`: digest domain types, high-water-mark calculation, summary generation, and PocketBase timestamp decoding.
- `internal/app`: HTTP routes, signup validation, and server-rendered templates.

The app still serves on `PORT` defaulting to `8080`, writes signups to PocketBase, and never sends PocketBase credentials to the browser. The homepage now renders the signup form first and the latest 5 `workflow_results` records below it.

## Key Decisions

- Ovek V1 app-triggered workflow execution is intentionally not implemented.
- Future trigger notes are documented in `docs/future-workflow-triggers.md`; a future "Run digest now" button should call an app server route, not Brain directly from browser JavaScript.
- `workflow_results.latest_signup_created_at` is stored as text so empty/no-signup runs are simple and durable.
- PocketBase date fields are decoded with a custom parser because PocketBase returns timestamps like `2026-05-24 10:00:03.000Z`, not strict RFC3339.
- The PocketBase client does not mutate shared auth state during request handling, avoiding a concurrency footgun.

## Verification Completed

- `GOCACHE=/tmp/ovek-workflow-go-cache go test -count=1 ./...`
- `git -c core.fsmonitor=false diff --check`
- `podman build --platform linux/amd64 -t ovek-workflow-example-app:local -f Dockerfile .`
- `podman build --platform linux/amd64 -t ovek-workflow-example-digest:local -f Dockerfile.workflow .`
- `podman inspect ovek-workflow-example-app:local --format '{{.Architecture}} {{.Os}}'` returned `amd64 linux`.
- `podman inspect ovek-workflow-example-digest:local --format '{{.Architecture}} {{.Os}}'` returned `amd64 linux`.

## Pending Tasks

- Run the documented manual Ovek acceptance flow against published GHCR images after the GitHub Actions publish succeeds.
- Confirm the optional scheduled workflow registration with `ovek workflow set workflow-demo digest --schedule '@hourly'`.
- Make both GHCR packages public after first publish if needed.
