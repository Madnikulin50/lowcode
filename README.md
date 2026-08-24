<h1 align="center">
  Lowcode
  <br />
  <br />
  <abbr>Low-code platform for data apps, automation, and AI</abbr>
  <br />
  <br />
  <div align="center">

  [![License](https://img.shields.io/github/license/Madnikulin50/lowcode?style=for-the-badge)](LICENSE)
  [![Go Report Card](https://goreportcard.com/badge/github.com/madnikulin50/lowcode?style=for-the-badge)](https://goreportcard.com/report/github.com/madnikulin50/lowcode)

  </div>
</h1>

Lowcode is a platform to build CRM and operations apps from modules, pages and charts, run workflows and rule chains, and talk to records through an AI assistant.

This repository is a fork of [Corteza](https://github.com/cortezaproject/corteza).

Docker image: [`madnikulin50/pnp-lowcode`](https://hub.docker.com/r/madnikulin50/pnp-lowcode).

## Features

* **Vue 3 webapps** — Compose, Admin, One, Workflow, Reporter, Privacy and Discovery are Vite + Vue 3 + Bootstrap 5 + Pinia (`client3/`). Legacy Vue 2 lives in `client/` and is not the production UI.
* **AI assistant** — in-app chat with Ollama (`qwen3:8b` by default). Static CRUD tools for modules/pages/charts plus per-module record tools. Create/delete actions ask for confirmation. Chat export to PDF and DOCX.
* **MCP** — the same Compose tools over STDIO (`MCP_STDIO=true`) or SSE (`MCP_SSE_ADDR=:9090`). The CMDB agent exposes its own MCP tools when started with `--mcp`.
* **Pages RAG** — index page blocks (with translations) for retrieval; record-bound blocks are skipped.
* **Rule chains** — in-namespace graphs (HTTP, CRUD, foreach, AI nodes). Admin UI under **Rule chains**; used by CMDB and Invest packs.
* **ETL** — per-module jobs (REST / MCP / SMB sources), run from the module editor.
* **Charts** — ECharts: pie/bar/line/doughnut/funnel/gauge/radar/scatter, stacked bar/line, `stackBy` second dimension, fill gradient, data-table toolbox, save-as-image.
* **i18n** — English and Russian resource translations (`locale/` and embedded `server/pkg/locale/src`).
* **Agents** — standalone Go services that scan or calculate, then write Compose records.

## Architecture

```
┌─────────────┐     REST / WS      ┌──────────────────┐
│  Vue 3 UI   │ ◄────────────────► │  Go server       │
│  client3/   │                    │  server/         │
└─────────────┘                    │  Compose, Auth,  │
                                   │  Chat, MCP, RAG, │
┌─────────────┐    Ollama HTTP     │  Rule chains,ETL │
│  Ollama     │ ◄────────────────► └────────┬─────────┘
│  :11434     │                             │
└─────────────┘                             │ Compose API
                                   ┌────────▼─────────┐
                                   │  agents/         │
                                   │  cmdb  :8085     │
                                   │  invest :8086    │
                                   └──────────────────┘
```

| Path | Role |
|---|---|
| `server/` | API, Compose, auth, chat, MCP, RAG, rule chains, ETL, locale embed |
| `client3/` | Production webapps (Vite). Bundles go to `webapp-assets/` (not `assets/` — that path is reserved for branding CSS) |
| `lib/` | Shared JS/Vue libraries |
| `locale/` | YAML translations loaded in development (`LOCALE_DEVELOPMENT_MODE`) |
| `agents/cmdb` | Network discovery agent + Compose namespace pack |
| `agents/invest` | EVM / critical-path engine + Invest namespace pack |
| `manual/` | AsciiDoc manuals |

## Applications

| App | URL (typical) | Purpose |
|---|---|---|
| One | `/` | App launcher |
| Compose | `/compose` | Low-code namespaces, pages, records, chat |
| Admin | `/admin` | Users, roles, settings, AI models, automation |
| Workflow | `/workflow` | Visual workflows |
| Reporter | `/reporter` | Reports |
| Privacy | `/privacy` | Privacy requests |
| Discovery | `/discovery` | Search |

## Getting started

### Docker

```bash
# .env must set VERSION, HTTP_PORT, HTTP_DOC_PORT, LOCAL
docker compose up -d
```

`docker-compose.yml` runs `madnikulin50/pnp-lowcode:${VERSION}` on `${HTTP_PORT}` and optional docs on `${HTTP_DOC_PORT}`. Data is stored in `./data/server`.

Build a local image (after compiling server + `client3`):

```bash
make ddebug          # tag pnp-lowcode:2026.08.20
# or
make drelease        # build, tag madnikulin50/pnp-lowcode:2026.08.20, push
```

### Local development

Typical layout in this repo:

| Process | Address |
|---|---|
| Go server | `:3333` (`server/.env` / `DB_DSN`) |
| Compose Vite | `:8080` (proxies API to the server) |
| Ollama | `http://localhost:11434` |

1. PostgreSQL with `DB_DSN` pointing at your database.
2. From `server/`: build and `serve-api` (or run the existing debug binary).
3. From `client3/web/compose`: `yarn && yarn dev` (same pattern for other webapps).
4. Optional: start Ollama and pull `qwen3:8b` (override with `CHAT_MODEL` or **System → AI models**).

Chat model resolution: admin `ai.roles.*` → `CHAT_MODEL` → `<model>…</model>` in the prompt → default `qwen3:8b`. Map tiles: setting `ui.map.tileURL` (empty = public OSM).

### AI / MCP

```bash
# Compose MCP over SSE
MCP_SSE_ADDR=:9090 ./lowcode-server serve-api

# or STDIO
MCP_STDIO=true ./lowcode-server serve-api
```

Warmup and models:

* `GET /compose/chat/models`
* `POST /compose/chat/warmup`
* `POST /compose/pages-rag/reindex`

## Agents

### CMDB (`agents/cmdb`)

Network inventory: ICMP/TCP scan of CIDR ranges, mDNS/DNS-SD, OUI vendor lookup, rule + LLM classification, optional vulnerability pass. Writes devices (and, via the Compose pack, services / findings / scan runs).

```bash
cd agents/cmdb
make build    # vite UI + go build → bin/cmdb-agent
./bin/cmdb-agent --db=embedded --listen=:8085
# or persist into Lowcode:
./bin/cmdb-agent --db=lowcode \
  --api=http://localhost:3333/api \
  --namespace=<namespaceID> \
  --token="$TOKEN" \
  --listen=:8085
```

| Flag | Default | Meaning |
|---|---|---|
| `--listen` | `:8085` | HTTP (API + embedded UI) |
| `--db` | `embedded` | `embedded` (SQLite) or `lowcode` |
| `--db-path` | `cmdb.db` | SQLite file |
| `--llm-url` / `--llm-model` | Ollama / `deepseek-v2` | Classification fallback |
| `--mcp` | off | MCP listen address or `stdio` |
| `--scan-interval` | `20m` | Periodic scan (`0` = off) |
| `--auto-cidrs` | empty | CIDRs for periodic scan |
| `--status-interval` | `5m` | Online/offline ping |

```bash
curl -s -X POST http://localhost:8085/api/scan \
  -H 'Content-Type: application/json' \
  -d '{"cidr":"192.168.1.0/24"}'
```

**Discovery signals**

* TCP ports including 5555 (adb), 5223 (apns), 7000 (airplay); mDNS adds 5353/udp.
* mDNS browse + unicast probes (AirPlay, RAOP, HiSuite, Mi PCS, Google Cast, Miracast, …). TXT `model=` / `deviceid=` fill **Model** and hostname.
* IEEE OUI table for mobile vendors (`tools/gen_oui.py` → `agent/oui_mobile.go`) — supporting signal only.
* Classifier: strong rules (model, mDNS service, hostname) → heuristic (OUI + empty ports) → LLM. Types: `router`, `switch`, `server`, `workstation`, `printer`, `camera`, `firewall`, `phone`, `tablet`, `iot`, `unknown`.

mDNS is link-local (multicast). Remote subnets only get unicast probes. OUI alone never decides type (Apple also makes Macs; Samsung/LG also make TVs).

Provision the Compose space: see [`agents/cmdb/compose/README.md`](agents/cmdb/compose/README.md).

### Invest (`agents/invest`)

EVM (PV/EV/AC/SPI/CPI/EAC), critical path, and threshold alerts for the investment-projects namespace.

```bash
cd agents/invest
go run . --api=http://localhost:3333 --namespace=<id> --token="$TOKEN" --listen=:8086
```

Pack and model: [`agents/invest/compose/README.md`](agents/invest/compose/README.md).

## Building webapps

From `client3/`:

```bash
make build    # all webapps under client3/web/*/
```

Vite must emit `webapp-assets/` (not `assets/`). The server already serves `/assets/*` from branding/embed.

Shared `lib/js` and `lib/vue` need a dist rebuild after type or component changes (`yarn build` in each lib). Keep `@tiptap/*` versions aligned between `lib/vue` and compose.

## Docs

Modules, pages, workflows and the REST API are covered in this README and under `agents/*/compose/`. Upstream 2024.9 integrator and DevOps guides still apply for those core concepts.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Issues: [github.com/Madnikulin50/lowcode](https://github.com/Madnikulin50/lowcode).

## License

Apache-2.0. See [LICENSE](LICENSE).
