# Operating Playbooks

## Add Or Update Institution Backend UI

1. Inspect the relevant page and nearby components.
2. Reuse Ant Design Vue patterns and shared components before adding one-off UI.
3. Keep changes scoped to the requested workflow.
4. Validate with a focused build or lint command when practical.
5. Mention unrelated dirty files without touching them.

## Add Or Update Pad Designs

1. Check `assessment_pad_app/designs/` for existing 1366 x 768 design language.
2. Build static design artifacts in the same directory when the user asks for design output.
3. Use existing Flutter page structure as business context.
4. Export PNG screenshots when producing design drawings.
5. Keep pad UI optimized for scanability, touch targets, and operational workflows.

## Work On IEP

1. Start with `generate-iep-modal.vue` for UI behavior.
2. Check `iep-plan-adapters.ts` to understand PEP-3 vs 儿心量表 behavior.
3. Check API files for payload shape and backend endpoints.
4. Preserve draft/confirmed semantics and execution plan save behavior.
5. Be careful with plan duration changes because loading, fallback, saving, and execution plan navigation all depend on it.

## Update This Wiki

1. Add or update concise facts in `wiki/`.
2. Register supporting sources in `sources.md`.
3. Put raw long materials in `raw/`.
4. Mark uncertain statements as `待验证`.
5. Update `notes/changelog.md`.
