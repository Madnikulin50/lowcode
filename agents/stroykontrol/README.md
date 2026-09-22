# stroykontrol-web — agent-hosted comparison viewer

Standalone service (like `agents/cmdb`'s `web/`) that owns the **visual**
side of "Сравнить документы": Compose only needed a working `IFrame` page
block (see below) — everything document-comparison-specific (fetching
files, PDF rasterization, the overlay UI) lives here, outside the lowcode
platform.

## Run

**Local dev, quickest path:** `./run.sh` — mints a fresh Compose token
itself (`mintToken()`, same as `seed.mjs`) and starts the agent against the
namespace it's set up for by default. No manual token, no fixtures.

```bash
cd agents/stroykontrol
./run.sh                    # defaults: api=localhost:3333/compose, the stroykontrol dev namespace
./run.sh --listen=:9000      # extra args pass straight to the binary
TOKEN=<token> ./run.sh       # skip minting, use a token you already have
NAMESPACE_ID=<id> ./run.sh   # point at a different namespace
```

**Manual / non-dev / a real service token:**

```bash
cd agents/stroykontrol
go build -o bin/stroykontrol-web .
TOKEN=<compose api token> ./bin/stroykontrol-web \
  --listen=:8092 \
  --api=http://localhost:3333/compose \
  --namespace=<stroykontrol namespaceID>
```

## Debugging in GoLand

`run.sh` is a shell wrapper, so GoLand can't attach a real breakpoint
debugger to what it starts. Use a native **Go Build** run/debug
configuration instead, with the token refresh split out as a "Before
launch" step — that way hitting the Debug button always gets a fresh token,
never a manually copy-pasted one that quietly expires mid-session (the JWT
is short-lived, ~2h).

A best-effort configuration named **"stroykontrol-web (debug)"** is
committed under `.idea/runConfigurations/` — open Run/Debug Configurations
and it should already be there with the right module/working
directory/program arguments (`--token-file=var/dev-token`, see below);
adjust `--namespace` if you're pointed at a different one. `.idea/` itself
is gitignored, so this doesn't sync between machines — every dev sets up
their own once. If it doesn't show up or the module name doesn't resolve,
just create one by hand:

1. **Run → Edit Configurations → + → Go Build.**
   - Run kind: **Directory**, pointing at `agents/stroykontrol`.
   - Working directory: `agents/stroykontrol`.
   - Program arguments: `--api=http://localhost:3333/compose --namespace=512312736219004929 --token-file=var/dev-token`
2. **Before launch → + → Run External tool** (create one if you don't have
   it yet: Program `bash`, Arguments `mint-token.sh`, Working directory
   `agents/stroykontrol`, pointing at [mint-token.sh](mint-token.sh) —
   it just mints a token and writes it to `var/dev-token`, which `--token-file`
   above reads on every startup).
3. Set your breakpoints, hit **Debug** (not Run) on this configuration.

`--token-file` (main.go) is only consulted when both `--token` and `TOKEN`
are empty, so this doesn't interfere with `run.sh` or a real service token
in CI/prod.

`--namespace` is only the *default* (used if a request omits `?namespaceID=`;
the IFrame block always sends it explicitly). Rasterized PDF pages are
cached on disk under `--cache` (default `var/stroykontrol-raster-cache`),
keyed by file content hash — same PDF is never re-rasterized.

Needs `pdftoppm` (poppler-utils) on PATH for PDF pages. DOCX pages don't need
it: there's no LibreOffice on this host to convert DOCX→PDF server-side, so
DOCX is instead rendered **client-side** — the browser fetches the raw file
(`/api/file`) and paginates it with `docx-preview` + `html2canvas` (npm
deps of the Vue 3 + Vite frontend under `web/`, built via `make web` into
`web/dist` and embedded into the binary), producing the same kind of
page images the PDF side gets from the server. Both PDF and DOCX therefore
share one viewer (side-by-side / opacity overlay / diff / **text**).
Anything else (other formats) falls back to the text-only discrepancy list.

**Text mode** compares each page's actual wording, not its image: a
word-level LCS diff (plain JS, no library) redlines PD against RD —
matching text stays plain, words only in PD are struck through in red,
words only in RD are underlined in green. Page text comes from `pdftotext
-layout` for PDF (cached alongside the rasterized images) and from
docx-preview's own paginated DOM for DOCX (`section.textContent`, no extra
round trip). Useful when a discrepancy is wording/numbers rather than
layout, and pairs with the visual modes rather than replacing them.

### Running without Compose

`--fixtures=<dir>` runs the whole agent against a local directory instead
of a live Compose instance — no server, no network call, no API token.
Useful for developing/demoing the viewer itself:

```bash
go run . --listen=:8092 --fixtures=fixtures
# http://localhost:8092/?recordID=demo-1&namespaceID=x
```

See [fixtures/README.md](fixtures/README.md) for the directory layout;
`fixtures/demo-1` is a ready-to-run example. `ComposeClient` and
`FixtureStore` both implement the same `store` interface (`store.go`), so
main.go's HTTP handlers don't know or care which backend is active.

**Token stability**: the token needs to stay valid for as long as the agent
runs. The `mintToken()` dev helper in `compose/helpers.mjs` mints from a
live refresh token in `auth_oa2tokens` and can rotate/invalidate whatever
was minted just before it (observed during testing) — fine for one-off
scripts, not for a long-running service. For real use, give this agent its
own dedicated API token/service account rather than a token shared with
interactive dev scripts.

## UI

The header shows the real attachment names (not just "ПД"/"РД"), a status
badge with an icon (spinning while `processing`), and similarity as a
colored progress bar. While status is `new`/`processing` the page polls
`/api/comparison` every 4s and reloads once it changes — no manual refresh
needed for the ~minutes-long analysis run. The mode switcher has icons and a
one-line hint for whichever mode is selected, plus zoom controls (50–200%,
contained within the page-image stage, not the whole page) for the three
visual modes. The discrepancies list opens with severity-count chips, is
sorted by severity, and shows each finding's source (⚙ детерминированно /
🤖 ИИ, parsed back out of the `[детерм.]`/`[ИИ] ` prefix the rule chain
bakes into `description` — chains.mjs) and type as separate tags instead of
inline text. A discrepancy with a page number inside the rendered range is
clickable — jumps the viewer to that page; one outside that range (a check
that flagged a page the viewer never loaded) shows a dashed, non-clickable
"стр. N ⚠" instead, so a page reference that doesn't resolve reads as
informational, never as a broken link. A discrepancy with no `description`
at all (seed/demo data, or a check that only flags a page) falls back to
naming its type instead of rendering an empty row, with a one-line note
that there's no further detail from the check itself.

`compose/seed.mjs`'s synthetic discrepancies used to set `page_number` up
to a fictional `total_pages_pd` (15–120) that has nothing to do with the
actual placeholder docx (a few lines of text — always renders as page 1)
— every seeded discrepancy pointed at a page that could never exist in the
viewer. Fixed to always seed `page_number: 1`, the only page that's real.

## Endpoints

| Route | What |
|---|---|
| `GET /api/comparison?recordID=&namespaceID=` | Title/status/similarity/comment/file names + `pd_rd_discrepancies` for the record |
| `GET /api/pages?recordID=&namespaceID=&side=pd\|rd` | `{supported, kind, pageCount}` — `kind` is `pdf`/`docx`/`other`; PDF is rasterized server-side on first call, DOCX reports `pageCount:0` (the browser determines it after rendering) |
| `GET /api/page?recordID=&namespaceID=&side=pd\|rd&n=` | One PDF page as JPEG (PDF only) |
| `GET /api/file?recordID=&namespaceID=&side=pd\|rd` | Raw attachment bytes with its Content-Type (used by the browser to render DOCX client-side) |
| `GET /api/text?recordID=&namespaceID=&side=pd\|rd` | `{pages: [...]}` — per-page extracted text, PDF only (via `pdftotext -layout`, cached); DOCX text comes from the client's own render, no call needed |
| `GET /*` | The viewer UI (Vue 3 + Vite app in `web/`, built to `web/dist` and embedded) |

## Wired into Compose

[`compose/apply.mjs`](compose/apply.mjs) adds an `IFrame` block to the
`pd_rd_comparison` card:
```
src: 'http://localhost:8092/?recordID=${recordID}&namespaceID=${namespaceID}'
```
Override the base with `STROYKONTROL_WEB_URL` before running `apply.mjs` if
the agent runs somewhere other than `localhost:8092`.

## Platform fix this relied on

`IFrameBase.vue` in the compose webapp rendered `<img>` unconditionally —
the `displayAsImage` config checkbox existed in the configurator but had no
effect, so a real `<iframe>` could never be embedded. Fixed to branch on
that option; also added `${namespaceID}`/`${moduleID}` to the URL
interpolation variables (`record-filter.js`), which only had
record/user/recordID/ownerID/userID before.
