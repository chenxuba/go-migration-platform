# LLM Wiki Index

This is the first page to read when using the project wiki as LLM context.

## Project Snapshot

`go-migration-platform` is a Go migration platform for a SaaS education/rehabilitation business system.

Core services:

- `iam-service`: login, users, tenants, roles, permissions, SSO compatibility.
- `platform-service`: tenant configuration, feature flags, dictionaries, menus, branding, rules.
- `education-service`: institution-side business domains such as students, orders, approvals, assessments, reports, IEP, and training.

Frontends:

- `institution-admin`: institution backend, Vue + Ant Design Vue.
- `platform-admin`: platform backend, Vue + Ant Design Vue.
- `government-admin` and `government-screen`: government-side admin/screen apps.
- `assessment_pad_app`: Flutter pad app for assessment assistant workflows.

## Read Next

- `wiki/architecture.md`: service and frontend map.
- `wiki/business-domains.md`: business domains and important workflows.
- `wiki/operating-playbooks.md`: common task playbooks for Codex.
- `sources.md`: evidence registry and external reference links.

## High-Value Context

- Keep tenant differences in configuration/rules, not hard-coded tenant branches.
- Prefer existing frontend component patterns before introducing new abstractions.
- Institution backend IEP generation is centered around assessment records, AI generated IEP total plans, monthly plans, weekly plans, draft/confirmed states, and Word export.
- Pad work often has static design artifacts under `assessment_pad_app/designs/` and Flutter source under `assessment_pad_app/lib/`.
- The repo may contain unrelated user changes; do not revert files unless explicitly asked.

## Open Questions

- `待验证`: Which docs should become canonical for deployment and production operations.
- `待验证`: Whether all assessment and IEP APIs are fully migrated to Go or still proxied/compatible with legacy services.
