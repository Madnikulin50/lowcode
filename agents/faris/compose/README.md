# Al-Faris Procurement — Compose namespace

Demo space for **Vendor Management and Procurement** at a fictional Saudi holding company, **Al-Faris Holding Group** (eight subsidiaries). English UI. ERP vendor master and budgets are **mocked** (SAP-style tables inside Corta).

Pattern matches CMDB / Invest: `apply.mjs` is idempotent.

## Apply

Server on `:3333`. JWT from a live refresh token, or `TOKEN=…`.

```bash
cd agents/faris/compose
TOKEN=$(node /tmp/opencode/mint-token.mjs | head -1) \
  node apply.mjs
```

| Variable | Default |
|---|---|
| `COMPOSE_API` | auto-detect `http://localhost:3333/api/compose` |
| `TOKEN` | refresh exchange from Postgres |
| `COMPOSE_DSN` | `postgres://postgres:@127.0.0.1:5432/test9?sslmode=disable` |
| `FARIS_RESEED` | `1` to seed again even if subsidiaries exist |

Open: `/ns/faris` (page **Group dashboard**).

Seed is skipped if subsidiaries already have records. Re-seed: `FARIS_RESEED=1 node apply.mjs`.

## Model

| Module | Role |
|---|---|
| **Subsidiaries** | Eight operating companies + ERP company code |
| **Spend categories** | IT, Facilities, Professional services, Materials, Fleet, Marketing |
| **ERP vendor master** | Mock SAP vendor (code, CR, VAT, status, last sync) |
| **ERP budgets** | Annual / committed / remaining SAR by subsidiary × category |
| **Vendors** | Onboarding requests, document pack, status pipeline |
| **Purchase requests** | Item, qty, estimate, vendor, budget check flags |
| **Approval log** | Step, decision, actor, days in step |

## Workflows (rule chains)

Edit live under **Admin → Rule chains** (add/reorder steps, no code).

**Vendor:** Submit (CR+VAT check) → Incomplete / Procurement → Compliance → Finance → Approve / Reject.

**Purchase request:** Submit → Procurement → Check ERP budget → Finance approve / Hold over budget / Reject.

## Demo story

1. Group dashboard — onboarding in flight, PR mix, spend, stalled list.
2. Incomplete vendor **Asir Digital Media** — missing CR.
3. Approve path on a submitted vendor.
4. PR **PR-2026-0214** / **PR-2026-0220** — over budget, on hold.
5. Admin → Rule chains: add or move the Compliance step.

`data_model/` describes the same handles (not live IDs).
