# Ovek Workflow Example

A tiny Go example with two OCI images:

- `ghcr.io/massivemoose/ovek-workflow-example-app:latest`
- `ghcr.io/massivemoose/ovek-workflow-example-digest:latest`

The app keeps the original PocketBase-backed signup behavior. The `digest` workflow is a one-shot image that reads signup records, calculates total and new signups since the previous successful digest, and writes the workflow result back to PocketBase. The homepage renders the signup form first and a server-rendered "Recent workflow results" panel below it.

The browser never receives PocketBase credentials.

## What This Demonstrates

- Running a normal Ovek app capsule with `ovek run`.
- Running an async one-shot workflow image with `ovek workflow run`.
- Optionally scheduling that workflow with `ovek workflow set --schedule`.
- Sharing Ovek-injected PocketBase app secrets between app and workflow containers.
- Persisting workflow results in PocketBase so the app can display them later.

Both images read:

- `POCKETBASE_URL`
- `PB_SUPERUSER_EMAIL`
- `PB_SUPERUSER_PASSWORD`

The app also reads `PORT`, default `8080`. The workflow also reads `OVEK_WORKFLOW_RUN_ID`.

## Images

The app image is built from `Dockerfile`:

```text
ghcr.io/massivemoose/ovek-workflow-example-app:latest
```

The digest workflow image is built from `Dockerfile.workflow`:

```text
ghcr.io/massivemoose/ovek-workflow-example-digest:latest
```

GitHub Actions publishes both images as multi-platform OCI-compatible images for `linux/amd64` and `linux/arm64`.

## Manual Workflow Run

Start from an Ovek server where the CLI is authenticated.

1. Initialize the project's PocketBase sidecar and app secrets:

```bash
ovek db init workflow-demo --app-secrets
```

2. Run the app capsule:

```bash
ovek run workflow-demo ghcr.io/massivemoose/ovek-workflow-example-app:latest
```

3. Open the app:

```text
http://workflow-demo.localhost/
```

4. Submit 2-3 email signups.

5. Register the digest workflow image:

```bash
ovek workflow set workflow-demo digest --image ghcr.io/massivemoose/ovek-workflow-example-digest:latest
```

6. Run the workflow:

```bash
ovek workflow run workflow-demo digest
```

7. Check workflow status:

```bash
ovek workflow status workflow-demo digest
```

8. Read logs after substituting the run ID from status:

```bash
ovek workflow logs workflow-demo <RUN_ID> --no-follow
```

9. Refresh the app page:

```text
http://workflow-demo.localhost/
```

Success criteria:

- The workflow run reaches `succeeded`.
- Logs show total and new signup counts.
- The app UI shows a recent digest result.
- Running the workflow again without new signups creates another result with `new_signups=0`.
- Adding another signup and rerunning creates a result with `new_signups=1`.

## Optional Scheduled Workflow

Register the same workflow with an hourly schedule:

```bash
ovek workflow set workflow-demo digest \
  --image ghcr.io/massivemoose/ovek-workflow-example-digest:latest \
  --schedule '@hourly'
```

The app does not trigger workflows directly. That is intentional for Ovek V1. See [docs/future-workflow-triggers.md](docs/future-workflow-triggers.md) for the future "Run digest now" button plan once Ovek has app-safe trigger tokens and idempotency.

## PocketBase Collections

Both the app and the workflow idempotently ensure the required collections exist:

- `signups`
  - `email`
  - `source`
- `workflow_results`
  - `workflow`
  - `run_id`
  - `status`
  - `started_at`
  - `finished_at`
  - `total_signups`
  - `new_signups`
  - `latest_signup_created_at`
  - `summary`

The app creates signup records. The workflow creates `workflow_results` records. The homepage reads the latest 5 workflow results server-side.

## Local Development

Run PocketBase locally at `http://127.0.0.1:8090`, then provide either `PB_SUPERUSER_TOKEN` or both `PB_SUPERUSER_EMAIL` and `PB_SUPERUSER_PASSWORD`.

Run the app:

```bash
PB_SUPERUSER_EMAIL="admin@example.com" \
PB_SUPERUSER_PASSWORD="your-password" \
go run ./cmd/app
```

Run the digest workflow locally:

```bash
PB_SUPERUSER_EMAIL="admin@example.com" \
PB_SUPERUSER_PASSWORD="your-password" \
OVEK_WORKFLOW_RUN_ID="local-test" \
go run ./cmd/digest
```

Run tests:

```bash
go test ./...
```

## Local Container Checks

Build the app image for `linux/amd64`:

```bash
podman build --platform linux/amd64 -t ovek-workflow-example-app:local -f Dockerfile .
```

Build the digest image for `linux/amd64`:

```bash
podman build --platform linux/amd64 -t ovek-workflow-example-digest:local -f Dockerfile.workflow .
```

Run the app image against a local PocketBase instance:

```bash
podman run --rm -p 8080:8080 \
  -e PORT=8080 \
  -e POCKETBASE_URL=http://host.containers.internal:8090 \
  -e PB_SUPERUSER_EMAIL=<email> \
  -e PB_SUPERUSER_PASSWORD=<password> \
  ovek-workflow-example-app:local
```

Run the digest image against a local PocketBase instance:

```bash
podman run --rm \
  -e POCKETBASE_URL=http://host.containers.internal:8090 \
  -e PB_SUPERUSER_EMAIL=<email> \
  -e PB_SUPERUSER_PASSWORD=<password> \
  -e OVEK_WORKFLOW_RUN_ID=local-container-test \
  ovek-workflow-example-digest:local
```

Docker works too. Use `http://host.docker.internal:8090` for the local PocketBase URL when running on Docker Desktop.

## Publishing

To publish manually:

1. Open the repo on GitHub.
2. Go to **Actions**.
3. Run **Publish OCI images to GHCR** from `main`.

After the first publish, make both GHCR packages public in package settings so Ovek servers can pull them without registry credentials.

Verify anonymous pulls:

```bash
podman logout ghcr.io
```

```bash
podman pull ghcr.io/massivemoose/ovek-workflow-example-app:latest
```

```bash
podman pull ghcr.io/massivemoose/ovek-workflow-example-digest:latest
```
