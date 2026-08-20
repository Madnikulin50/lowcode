# CMDB Lowcode namespace

Low-code space that matches the CMDB agent in `agents/cmdb`: scan CIDR ranges, inventory hosts (including phones and tablets), list open services, and triage vulnerabilities.

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
| **Devices** (`devices`) | Configuration items — same fields the agent writes (`ip_address`, `mac_address`, `hostname`, `vendor`, `model`, `device_type`, `os`, `domain`, ports/services/shares/vulns JSON, `last_seen`, `status`) |
| **Services** (`services`) | Open ports, Record → devices |
| **Vulnerabilities** (`vulnerabilities`) | Findings (`severity` CRITICAL/HIGH/MEDIUM/LOW/INFO, `cve`, `status` open/acknowledged/fixed/false_positive) |
| **Scans** (`scans`) | Scan runs, Record → networks |

Device types match the agent: router, switch, server, workstation, printer, camera, firewall, **phone**, **tablet**, iot, unknown.

`model` is the hardware/software string when the agent can see it (AirPlay TXT `model=iPhone14,2`, similar mDNS keys). With `--db=lowcode`, `EnsureModule` creates the field and the phone/tablet select options if the Devices module does not exist yet. The `cmdb-ingest-scan` chain upserts devices from the agent payload (plus nested services and findings).

## Discovery (agent)

The namespace is filled by `cmdb-agent`, not by Compose itself. Current scan pipeline:

1. **ICMP + TCP** over the CIDR (ports include 5555 adb, 5223 apns, 7000 airplay).
2. **mDNS / DNS-SD** on the local link (`224.0.0.251:5353`): browse `_services._dns-sd._udp.local` plus AirPlay, RAOP, companion-link, sleep-proxy, HiSuite, Mi PCS, Google Cast, Miracast, Android TV remote, etc. Hosts not in the browse set get a unicast probe (QU bit). PTR/SRV/TXT/A are merged by IP; TXT `model` / `deviceid` become **Model** / hostname; announced types go to `Services`; 5353/udp `mdns` is added to open ports.
3. **OUI** — IEEE-derived mobile vendor table (`agent/oui_mobile.go`, regenerated with `tools/gen_oui.py`). Supporting signal only.
4. **Classifier** — strong rules (model, mDNS service, hostname patterns) → heuristic (mobile OUI + no open ports → phone) → Ollama. Network-gear vendors (Cisco/Huawei/Juniper) become `switch` only when 22/23/161 are open.

Limitations: multicast mDNS does not cross routers; remote subnets get unicast probes only. UDP service scan (53/123/161/1900) per host is not implemented yet.

Agent UI (`:8085`) shows a **Model** column on the device list.

## Pages

- **Dashboard** — hero metrics, scan / stale-device buttons, running-scan progress, doughnut charts, open vulns and recent devices
- **Devices / Networks / Vulnerabilities / Services / Scans** — list pages (vulns also has a status kanban)
- Record pages are hidden from the sidebar. Device / network / finding cards use field roles (title, badges, meta) plus related lists and rule-chain buttons.

## Rule chains

In-memory on the server (re-apply after a process restart):

| ID | Where | What it does |
|---|---|---|
| `cmdb-trigger-scan` | Dashboard, **Scans**, Network card | HTTP POST `{cidr, namespaceID}` to the CMDB agent, then create a `scans` record |
| `cmdb-ingest-scan` | Agent callback / poll | Update the scan row, upsert **devices**, then upsert **services** (each open port) and **vulnerabilities** (each finding) linked to the device |
| `cmdb-stale-devices` | Dashboard | Search devices (pass Compose QL in `query`, default `status = 'online'`) |
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
go run . --db=lowcode \
  --api=http://localhost:3333/api \
  --namespace=<namespaceID> \
  --token="$TOKEN" \
  --listen=:8085
```

`EnsureModule` reuses handle `devices` and will not create a second module. It will add `model` and phone/tablet options on a freshly created Devices module.

Scan:

```bash
curl -s -X POST http://localhost:8085/api/scan \
  -H 'Content-Type: application/json' \
  -d '{"cidr":"192.168.1.0/24"}'
```

Embedded UI: `http://localhost:8085/` (dashboard + device list). MCP: `--mcp=:9091` or `--mcp=stdio` (`scan_network`, `list_devices`, `get_device`, …).

`data_model/` is the envoy-style description of the same model (handles, not live IDs).
