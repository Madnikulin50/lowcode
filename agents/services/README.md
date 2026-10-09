# Lowcode skills

Small HTTP services: one ability, Compose contract (`agents/sdk`), no own UI and no writes to records. A namespace (cmdb, invest) is the product; a skill is a node in a rule chain.

| Service | Port | Kind | Call |
|---|---|---|---|
| `calc-evm` | `:8088` | sync | `POST /api/call/evm` body `{items, facts, projectID, now}` |
| `scan-cidr` | `:8089` | async | `POST /api/jobs` body `{cidr, callbackUrl, recordID}` |

```bash
go run ./agents/services/calc-evm/cmd/calc-evm
go run ./agents/services/scan-cidr/cmd/scan-cidr
```

`GET /api/meta` publishes the palette descriptor. `scan-cidr` posts an `Envelope` (`items` = devices) to `callbackUrl`; ingest stays a chain (`cmdb-ingest-scan`). Invest EVM buttons search WBS/facts, POST `/api/call/evm`, then PATCH records — engine is not on that path. CMDB trigger defaults to `:8089` (`CMDB_AGENT_URL` still overrides; `:8085` fat agent keeps classifier/vulns).
