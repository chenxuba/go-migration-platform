# Architecture Map

## Backend

The repo is organized around three Go services:

- `services/iam`: authentication, SSO compatibility, users, roles, menus.
- `services/platform`: tenant features, dictionaries, modules, customization summaries.
- `services/education`: institution-side education business capabilities.

Shared packages live under `pkg/`:

- `pkg/tenant`: tenant context and middleware.
- `pkg/customization`: tenant configuration, feature flags, rule packages.
- `pkg/httpx`: unified HTTP responses.
- `pkg/logx`: structured logging.
- `pkg/search`: search integration.
- `pkg/messaging`: event/messaging support.

## Frontends

- `institution-admin`: institution operations UI. Business-heavy pages generally use Ant Design Vue.
- `platform-admin`: platform operations UI. Platform modal work should reuse shared platform modal shells where available.
- `assessment_pad_app`: Flutter pad app. Existing pad design references are in `assessment_pad_app/designs/`.

## Local Development

Useful startup context from `README.md`:

- `./scripts/restart.sh` runs local infrastructure checks and starts the three Go services.
- NATS JetStream defaults to `nats://127.0.0.1:4222`.
- Meilisearch defaults to `http://127.0.0.1:7700`.

## Design Principle

Tenant-specific behavior should be expressed through tenant config, feature flags, rule packages, workflow schemes, and integrations. Avoid hard-coding tenant-specific branches in service logic.
