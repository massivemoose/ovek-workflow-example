# Future App-Triggered Workflows

This repo intentionally demonstrates the Ovek V1 async model: Brain triggers workflows manually or from a cron schedule with `ovek workflow run` and `ovek workflow set --schedule`.

Do not add a browser-triggered workflow call yet. The browser must never receive PocketBase credentials, Brain credentials, or any broad workflow-control secret.

When Ovek supports app-safe workflow trigger tokens, add a server-owned `POST /workflows/digest/run` route and wire a "Run digest now" button to that route. The server route should use a project-scoped, run-only token, not browser JavaScript calling Brain directly.

When Ovek supports idempotent workflow triggers, include an idempotency key for the request. If Ovek later supports per-run payloads, the route can pass a reason such as `manual-ui` or a requested date range. Until then, the digest workflow derives work from durable PocketBase state.

Long-running Bun or TypeScript hot workers are out of scope for this example. Keep the repo focused on one-shot OCI workflows.
