---
name: openapi-first-delivery
description: Enforces an OpenAPI-first implementation workflow for ticketing-system changes. Use when adding or modifying features that touch backend/frontend behavior: update openapi.yaml first, regenerate backend/frontend via Makefile, implement code, add E2E tests after implementation, run tests through Makefile targets, and update .documentation files.
---

# OpenAPI First Delivery

Follow this workflow for all feature work in `ticketing-system`.

## Non-negotiable order
1. Update API contract first:
   - Edit `ticketing-system/openapi.yaml` before code changes.
2. Regenerate backend and frontend types from OpenAPI:
   - `make -C ticketing-system generate`
3. Implement backend and frontend behavior:
   - Backend handlers/store/migrations.
   - Frontend store/views/components using generated types.
4. Add or update E2E tests only after implementation is complete:
   - Update contract keys/selectors and routes as needed.
   - Prefer contract-driven E2E patterns.
5. Run tests via Makefile:
   - `make -C ticketing-system e2e-contract`
   - `make -C ticketing-system e2e`
6. Update documentation:
   - `./.documentation/current_features.md` for shipped capabilities.
   - `./.documentation/feature_roadmap.md` and `./.documentation/ticket_drafts.md` when roadmap/ticket plans change.

## Required commands
- Regenerate code after OpenAPI edits:
  - `make -C ticketing-system generate`
- Refresh E2E contract:
  - `make -C ticketing-system e2e-contract`
- Run E2E suite:
  - `make -C ticketing-system e2e`

## Implementation guardrails
- Do not hand-edit generated OpenAPI outputs unless absolutely necessary; regenerate instead.
- Keep frontend API usage aligned with generated `api.schema.ts` and wrappers.
- Keep E2E tests behavior-focused and contract-key driven, not raw selectors.
- Update `.documentation` in the same change set as code.

## Done checklist
- OpenAPI spec updated and committed.
- `make generate` outputs included.
- Backend/frontend implementation complete.
- E2E tests added/updated after implementation.
- Makefile-driven E2E commands executed.
- `.documentation` updated to reflect delivered behavior.
