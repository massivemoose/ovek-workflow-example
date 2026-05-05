# Ovek Signup Capsule

A tiny Go signup app packaged as an Ovek capsule. It runs behind Ovek, writes email signups to a per-project PocketBase sidecar, and keeps persistent data local to the server running your app.

![Signup page](images/signup.png)

## What This Demonstrates

- Publishing a Go app as an OCI-compatible capsule image on GHCR.
- Running that image with `ovek run`.
- Using Ovek-managed persistent data through a per-project PocketBase sidecar.
- Injecting PocketBase credentials into the app as project secrets.
- Keeping the app portable: it only depends on runtime env vars.

The app reads:

- `PORT`, default `8080`
- `POCKETBASE_URL`
- `PB_SUPERUSER_EMAIL`
- `PB_SUPERUSER_PASSWORD`

On startup, it authenticates to PocketBase, ensures the `signups` collection exists, and serves the signup form.

## Capsule Image

Published image:

```text
ghcr.io/massivemoose/ovek-signup-example:latest
```

This repo uses a `Dockerfile` as the build recipe. The published artifact is an OCI-compatible multi-platform image for `linux/amd64` and `linux/arm64` that Ovek can pull and run.

## Run With Ovek

Start from an Ovek server where the CLI is authenticated.

Initialize the project's PocketBase sidecar and app secrets:

```bash
ovek pb init signup-demo --app-secrets
```

Run the capsule:

```bash
ovek run signup-demo ghcr.io/massivemoose/ovek-signup-example:latest
```

Check status:

```bash
ovek status signup-demo
```

Open the app:

```text
http://signup-demo.localhost/
```

Submit an email address. A successful signup redirects to `/success`.

## Inspect PocketBase

Check the managed sidecar:

```bash
ovek pb status signup-demo
```

Open a local tunnel:

```bash
ovek pb tunnel signup-demo --listen 127.0.0.1:8091
```

Then open:

```text
http://127.0.0.1:8091/_/
```

Expected: after the app starts, the `signups` collection exists and submitted emails appear as records. Use your `PB_SUPERUSER_` credentials to log in.

## Publish The Image

The GitHub Actions workflow publishes:

```text
ghcr.io/massivemoose/ovek-signup-example:latest
ghcr.io/massivemoose/ovek-signup-example:<git-sha>
```

Each tag points to a multi-platform manifest with `linux/amd64` and `linux/arm64` images.

To publish manually:

1. Open the repo on GitHub.
2. Go to **Actions**.
3. Run **Publish OCI image to GHCR** from `main`.

After the first publish, make the GHCR package public in package settings so Ovek servers can pull it without registry credentials.

Verify anonymous pull:

```bash
podman logout ghcr.io
```

```bash
podman pull ghcr.io/massivemoose/ovek-signup-example:latest
```

To verify a specific platform explicitly:

```bash
podman pull --platform linux/amd64 ghcr.io/massivemoose/ovek-signup-example:latest
```

```bash
podman pull --platform linux/arm64 ghcr.io/massivemoose/ovek-signup-example:latest
```

Docker uses the same pattern:

```bash
docker logout ghcr.io
```

```bash
docker pull ghcr.io/massivemoose/ovek-signup-example:latest
```

```bash
docker pull --platform linux/amd64 ghcr.io/massivemoose/ovek-signup-example:latest
```

```bash
docker pull --platform linux/arm64 ghcr.io/massivemoose/ovek-signup-example:latest
```

## Local Development

Run PocketBase locally at `http://127.0.0.1:8090`, then provide either `PB_SUPERUSER_TOKEN` or both `PB_SUPERUSER_EMAIL` and `PB_SUPERUSER_PASSWORD`.

```bash
PB_SUPERUSER_EMAIL="admin@example.com" \
PB_SUPERUSER_PASSWORD="your-password" \
go run ./...
```

Open:

```text
http://localhost:8080
```

Run tests:

```bash
go test ./...
```

## Local Container Check

Build the local image:

```bash
podman build -t ovek-signup-example:local .
```

To test a specific published platform locally:

```bash
podman build --platform linux/amd64 -t ovek-signup-example:local .
```

```bash
podman build --platform linux/arm64 -t ovek-signup-example:local .
```

Run it against a local PocketBase instance:

```bash
podman run --rm -p 8080:8080 \
  -e PORT=8080 \
  -e POCKETBASE_URL=http://host.containers.internal:8090 \
  -e PB_SUPERUSER_EMAIL=<email> \
  -e PB_SUPERUSER_PASSWORD=<password> \
  ovek-signup-example:local
```

Expected: the app starts, authenticates to PocketBase, ensures the `signups` collection exists, and serves the form at `http://localhost:8080`.

Docker works too:

```bash
docker build -t ovek-signup-example:local .
```

```bash
docker run --rm -p 8080:8080 \
  -e PORT=8080 \
  -e POCKETBASE_URL=http://host.docker.internal:8090 \
  -e PB_SUPERUSER_EMAIL=<email> \
  -e PB_SUPERUSER_PASSWORD=<password> \
  ovek-signup-example:local
```
