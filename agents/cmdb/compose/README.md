# CMDB Corteza namespace

Low-code space that matches the CMDB agent in `agents/cmdb`: scan CIDR ranges, inventory hosts, list open services, and triage vulnerabilities.

## Apply to a running server

Server must be up (`:3333` in this repo). JWT is minted with `/tmp/opencode/mint-token.mjs` unless `TOKEN` is set.

```bash
cd agents/cmdb/compose
TOKEN=$(node /tmp/opencode/mint-token.mjs | head -1) \
  node apply.mjs
```

Optional env:

| Variable | Default |
|---|---|
| `COMPOSE_API` | auto-detect `http://localhost:3333/api/compose` |
| `TOKEN` | mint-token.mjs |
| `CMDB_AGENT_URL` | `http://localhost:8085/api` |

The script is idempotent (updates by handle). Writes `applied.json` with IDs.

Re-running **does not reset Page Builder layout**. Existing block sizes (`xywh`), extra blocks, titles, styles, icons and layout buttons stay. New blocks from the template are appended. To force the YAML layout:

```bash
APPLY_RESET_PAGES=1 node apply.mjs
# or
node apply.mjs --reset-pages
```

Open compose at `/ns/cmdb` (page **Dashboard**).

## Model

| Module (`handle`) | Role |
|---|---|
| **Networks** (`networks`) | CIDR ranges to scan (`name`, `cidr`, `enabled`, `last_scan`) |
| **Devices** (`devices`) | Configuration items — same fields the agent writes (`ip_address`, `mac_address`, `hostname`, `vendor`, `device_type`, `os`, `domain`, ports/services/shares/vulns JSON, `last_seen`, `status`) |
| **Services** (`services`) | Open ports, Record → devices |
| **Vulnerabilities** (`vulnerabilities`) | Findings (`severity` CRITICAL/HIGH/MEDIUM/LOW/INFO, `cve`, `status` open/acknowledged/fixed/false_positive) |
| **Scans** (`scans`) | Scan runs, Record → networks |

Device types match the agent: router, switch, server, workstation, printer, camera, firewall, iot, unknown.

## Pages

- **Dashboard** — hero metrics, scan / stale-device buttons, running-scan progress, doughnut charts, open vulns and recent devices
- **Devices / Networks / Vulnerabilities / Services / Scans** — list pages (vulns also has a status kanban)
- Record pages are hidden from the sidebar. Device / network / finding cards use field roles (title, badges, meta) plus related lists and rule-chain buttons.

## Rule chains

In-memory on the server (re-apply after a process restart):

| ID | Where | What it does |
|---|---|---|
| `cmdb-trigger-scan` | Dashboard, **Scans**, Network card | HTTP POST `{cidr, namespaceID}` to the CMDB agent, then create a `scans` record |
| `cmdb-stale-devices` | Dashboard | Search devices (pass Corteza QL in `query`, default `status = 'online'`) |
| `cmdb-flag-insecure-ports` | Device card | HIGH finding if ports 21/23 or telnet/ftp show up on the host |
| `cmdb-insecure-service` | Service card | HIGH finding when the service name is telnet/ftp |
| `cmdb-high-vuln-alert` | Finding card | Mail when `severity` is HIGH or CRITICAL |
| `cmdb-ack-finding` | Finding card | Set status to `acknowledged` |
| `cmdb-close-finding` | Finding card | Set status to `fixed` |

Test from admin **Rule chains**, or click the buttons on the pages. `POST /api/compose/admin/rulechain/{id}/test` still works.

## Point the agent at this namespace

```bash
cd agents/cmdb
# after apply.mjs, namespace ID is in applied.json
go run . --db=corteza \
  --api=http://localhost:3333/api \
  --namespace=<namespaceID> \
  --token="$TOKEN" \
  --listen=:8085
```

`EnsureModule` reuses handle `devices` and will not create a second module.

Scan:

```bash
curl -s -X POST http://localhost:8085/api/scan \
  -H 'Content-Type: application/json' \
  -d '{"cidr":"192.168.1.0/24"}'
```

`data_model/` is the envoy-style description of the same model (handles, not live IDs).
