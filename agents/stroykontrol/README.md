# stroykontrol-web — agent-hosted comparison viewer

> Также хостит **проверку ИД (АОСР ↔ реестр ↔ КС-2)** — пакет
> [`idcheck/`](idcheck), пространство `stroykontrol` (бывшее `stroykontrol-id`, объединено), эндпоинт
> `POST /api/idcheck/run`. См. [compose_id/README.md](compose_id/README.md).

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
XLSX follows the same client-side path: the workbook is unzipped in the
browser (`jszip`) and each sheet is drawn as a grid (raw cell values, column
letters and row numbers), split into pages of 50 rows. Same-named sheets on
both sides share column widths so overlay/diff line up cell-for-cell; the
text mode uses the backend extractor's line shape (one row per line, cells
tab-separated, a `Лист: <name>` header per sheet). Legacy binary `.xls` is
not supported. Try it offline with the `demo-xlsx` fixture.
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

## Сравнение комплектов (ПД — эталон, РД)

The comparison is a **package** of PDFs per stage: one file, several files, a
pack against a pack or against a single file. Files are rows of the Compose
module `cmp_documents` (side `ref`/`target`); the old `pd_file`/`rd_file`
fields of a comparison keep working as a one-file package.

The engine is the Go package [`cmp/`](cmp) — deterministic, no LLM decides
whether something differs. The rule chain `stroykontrol-compare-pd-rd` only
starts it (`POST /api/cmp/run?recordID=&namespaceID=`, an `http` node, like
`idcheck`); the agent writes status, metrics, discrepancies and detected
volumes back into the namespace itself.

1. **Extract** (`cmp/extract.go`): `pdftotext -bbox` → words with positions →
   visual rows with a box in page fractions. Every sheet is classified:
   - *title block* rows (основная надпись, margin stamps, signature rows) are recognised
     **by their content** (Изм. Кол.уч. / Стадия Лист Листов / Взам. инв. / Проверил …), not
     by position alone — body text near the page bottom stays content; on drawing sheets
     the bottom-right corner counts as title block too;
   - *cover / front matter* (title page, approval sheet, contents register, first 6
     sheets) is skipped: dates, signatures and edition marks change by nature;
   - *drawing sheets* (A2+ or mostly scattered short labels, or no text layer) are never
     text-diffed — their scattered labels are noise.
2. **Volumes** (`cmp/volume.go`): **one file is one volume** unless its title block
   shows ≥2 different шифры (strict shape: digit, ≥2 letters, ≥2 parts, read from the
   title block's own row, ≥2 sheets each; shorter stretches are absorbed). A manual code on a
   `cmp_documents` row (`detected_by = manual`) overrides the detection.
3. **Matching** (`cmp/match.go`): one volume against one volume is paired as given (a
   file against its next edition needs no content overlap); otherwise volumes are paired
   by shared content (+ шифр / series), so a ПД volume may match several РД volumes, and
   a РД volume may cover only part of a ПД volume (*fragment*). Unpaired ПД volume →
   `not_released`, unpaired РД volume → `scope_diff`; both informational (`low`).
4. **Diff** (`cmp/diff.go`, `compare.go`): patience line diff over the whole volume
   (page shifts do not misalign it), similar lines are paired as an *edit*.
   Changed numbers → `high`, rewording → `medium`, a numbering-only change →
   `low`. A ПД line missing from РД is `medium` (not reported for fragments);
   extra РД lines are `low` (more detailed РД is normal). With no reference text
   (drawings-only volume) no text findings are made.
5. **Key parameters** (`cmp/params.go`): concrete grade, water/frost resistance,
   rebar class are extracted by regex; a conflicting value is `high`.
6. **Drawings** (`cmp/raster.go`): drawing sheets are compared as images only when
   both sides have the same number of them with the same proportions (regions become
   boxes); otherwise one informational finding asks for a visual check. No OCR yet.
7. **Similarity**: full pairs — Dice over lines (2·equal/(ПД+РД)); fragment pairs —
   share of РД lines confirmed by ПД. Unknown (empty in Compose) when there was no
   text to compare, never a fake 0/100. The comment lists coverage and drawing counts.

Every finding carries `kind`, the documents/pages/regions on both sides, the
fragment it is anchored to, an `evidence` JSON (what the conclusion rests on)
and a `fingerprint`; the reviewer's decision (`new/confirmed/rejected`) is kept
across re-runs by fingerprint. Remark cards (`remark_cards`) are filled from
selected findings by a template; the local model (`--remark-model`,
`REMARK_MODEL`) only drafts wording when the specialist asks for it.

### Iterations (before/after corrections)

A comparison may point to a previous one (`parent_comparison`, `iteration_no`): the same
pair re-checked after the designer's corrections. The ПД (reference) set is **inherited**
from the nearest previous iteration that has one — only the corrected РД is uploaded.
After a run every finding has an `iter_state` (`cmp/iteration.go`, matched by fingerprint,
which is built from the compared fragments and survives re-pagination):

- `persisting` — still there, **keeps the reviewer's decision**;
- `new` — not in the previous iteration;
- `resolved` — was there, gone now (recreated in the new comparison, struck through in the
  viewer). Findings the reviewer had rejected as false alarms are never reported as "fixed".

The comparison comment and the viewer header show «устранено / не устранено / новых».

### OCR of scans

Sheets with no text layer and not larger than A2 are recognized with tesseract
(`cmp/ocr.go`, `--tesseract`, `--ocr-lang`; pages are rendered at 200 dpi and cached per
page; at most 40 per file). Recognized sheets are compared like text pages, but findings
that rest on them are capped at «средняя», marked «[по распознанному тексту скана]» and carry
the rule `…+ocr`. A recognized sheet with fewer than 20 words, or scattered labels, stays a
drawing.

### ВОР and specifications

`cmp/vor.go`, `cmp/vor_compare.go`. Tables are read from word positions: the header names
the columns (Наименование / Ед. изм. / Кол-во), every number in the quantity column is a
position (numbers split by a thousands space are joined), its name is the text of the name
column nearest to it. The kind (ВОР / specification) comes from the sheet's title; following
sheets inherit it. Pages read as tables are compared as tables, not line by line.

- **Between stages** — positions are paired by name + unit; a changed quantity is
  `high` (>10 %), `medium` (>3 %) or `low`; a fivefold jump is never `high` (more likely two
  different positions or a misread cell). Positions with no counterpart are information only
  (bills are aggregated in one stage and split in the other).
- **Inside a volume** — the bill against its own specification: a position found in both must
  agree; if the specification splits it over several lines their **sum** is compared; a bill
  position with no specification counterpart is not a mismatch (works vs materials).

Limits: only sheets that carry the column header (or follow a titled sheet) are read;
tables whose layout differs per sheet fall back to the line diff; no aggregation between
volumes.

### List of remarks made from a comparison

`cmp/issue.go`, `issue_registry.go`, `issue_export.go`. The specialist confirms findings (✓ in the
viewer; «Подтвердить все выше «низкой»» does it in bulk). From the **confirmed** ones — not rejected,
not information about unreleased volumes or out-of-scope ones — the viewer's «Список замечаний»
builds one remark per *sheet pair* (ПД sheet + РД sheet): a numbered list of the findings, «Ссылка на
материалы» (files and sheets of both stages), the section (volume), the basis (stage П is the standard),
criticality (`critical` when a high finding is among them). The same list goes three ways:

- **into the registry** of the object's remarks (`POST /api/issue/register`, button/chain
  `stroykontrol-issue-remarks`): numbers `ПТО-<batch>.<n>` (one batch per comparison), status «Ответ не
  предоставлен», `origin = comparison`, linked to the findings and the documents. The sheet pair is the
  remark's key, so pressing it again **updates** the remarks and never duplicates them, keeping the
  specialist's status, assignee and notes; when a later iteration of the comparison has fixed every
  finding of a sheet the remark is **closed** (history entry), when the difference comes back it is
  **reopened**. These remarks are left alone by an export import and by the document check;
- **Excel** (`GET /api/issue/export?format=xlsx`) with the headings of an expertise export — it reads
  back through the same importer, so the designer's answers can come in on the same format;
- **Word** (`format=docx`): a letter with a table of the remarks (№, references, remark, criticality).

### Expertise remarks: workflow

The remarks of an expertise live in the **registry of the object** (module `expertise_remarks`),
not in a comparison. The input is the expertise system's export («Выгрузка данных»: remark,
references to volumes, section, basis, answer, rejection reason, previous answer and reason,
НД, expert, criticality, proposals, iterations, last change, status, №), attached to the
object (`objects.remarks_file`); the agent reads it by its headings (`cmp/expertise_export.go`;
any other table with a remark column goes through the generic parser; a PDF is read by `cmp/expertise_pdf.go`, see below).

The cycle the registry follows: remark → designer's answer → the expertise accepts
(«Устранено») or rejects with a reason → the remark comes back, iteration + 1, the export
carrying the previous answer and reason.

- **Import** (`POST /api/expertise/import?objectID=`, button/chain `stroykontrol-import-remarks`,
  `Runner.ImportObjectRemarks`): remarks are matched by their number (a repeated number gets
  `#2`, a missing one a hash of the text); an export that was imported before is skipped.
  Every import writes a **journal** record (`expertise_imports`: new, changed, status changes,
  returned, dropped out + the list) and, per remark, **history** (`expertise_remark_history`):
  the previous iteration is built from the export's «предыдущий ответ / причина», then one entry
  per change of answer, iteration or status. Remarks missing from a later export are flagged, not deleted.
- **One status, edited by hand** (Ответ не предоставлен / Устраняется заявителем / Требуется
  рассмотрение ответа / Устранено): an import overwrites it **only when the expertise's own
  status changed since the previous export** (its decision is newer); the overwrite is logged and
  lands in the history. Assignee, due date and a note are the specialist's own (`cmp/expertise_workflow.go`).
- **Check against the documents**: «Ссылка на материалы» is read into volumes (`ПЗ5.1`, `ТКР5.4.1`,
  sheet, page) and linked to the loaded documents of the object's comparisons by шифр
  (`ТКР5.1` ~ `100720-1/П-3-ТКР5-1`). A remark whose volumes are loaded is matched to the
  findings of those documents (shared words, and more strongly codes/values) — `still_differs`
  (a difference remains; if the answer claims a correction: «ответ ≠ документы»), `resolved`
  (fixed in an iteration), `no_match`; volumes that are not loaded give `no_documents`.
  `POST /api/expertise/check` repeats it after more documents were compared.
- **Where a remark leads** — `expertise_remarks.links_json` (`cmp/remark_links.go`): the places the remark's references name (volume → page; a sheet number is turned into a page by the document's `sheet_map`, «страница:лист,…», read from the title blocks during a run and a sets check) and the regions (`bbox`) of the findings it was matched to. The card shows «Где в документах»; `?recordID=&findingID=` opens the comparison on that finding with the region marked, `GET /api/sheets` gives the page → sheet numbers of a PDF so the viewer writes «лист N» beside the page.
- **Workbench** — `?view=remarks&objectID=` (embedded in the object card): KPIs by status,
  filters (status, criticality, expert, НД, section, assignee, iteration, check result, overdue),
  the remark with the answer next to the previous one (changed words highlighted), the history,
  quick status / assignee / due date / note, the journal. The comparison viewer lists the remarks that concern
  that comparison and links to the workbench. In `--fixtures` mode a case's `remarks*.xlsx` is the registry.

Not done: aggregation of ВОР across volumes; estimates.

### Report, versions, manual pairing, formatting and cross-volume checks, AI reading, dialog

- **Automatic statement of volumes** — «Предложить ведомость томов» in the object overview (`POST /api/volumes/suggest`, `cmp/volumes_auto.go`): volumes named in the contents tables of the documents («Состав проектной документации», «Содержание», «Перечень томов»; the volume itself, its parts and stationing are not proposed) and, for a stage with no statement, the volumes of its earliest set. Proposals are `object_volumes` rows with `proposed = yes`: they do not count for completeness until accepted (`POST /api/volumes/resolve`: accept / reject / accept all); a confirmed row is never deleted by a reject. On the sample documents explicit contents tables rarely list volumes, so the earliest set is the usual source.
- **Report** — `GET /api/report/export?recordID=&namespaceID=&format=xlsx|docx` (`cmp/report.go`): summary, findings per document pair (severity, decision, iteration, number of the remark they went into), a separate block for «not released / out of scope», formatting checks and the table of key parameters. A reference for the specialist, not the list of remarks (that is «List of remarks made from a comparison»).
- **Versions of one stage** — comparison field `kind` = `versions`: the same engine, but the sides are «До»/«После» everywhere (viewer, texts of findings, remarks, report, label «ПД v1 → ПД v2»). `cmp.SideNames`, `cmp.Relabel`.
- **Manual pairing** — comparison field `pairing` (JSON `[{ref:{doc,from,to}, targets:[…]}]`, edited in the viewer panel «Сопоставление»; `GET/POST /api/pairing`). The listed page ranges are compared as told, the rest of the volumes is matched automatically; the result of the last run is kept in `pairs_json` (`cmp/pairing.go`).
- **Formatting checks** (`cmp/format.go`, kind `format`, switch off with `Options.NoFormat`): sheet numbering in the title block (a restart from 1 is a new document, not an error), «стр. N» references of the contents, a шифр of the same project but another volume, identical sheets; references to «Таблица N» / «Приложение X» / «раздел N» that have no heading in the file (`format-ref-*`; a line that speaks of another volume, СП/ГОСТ or a contract is skipped, so is the line after a norm number), appendices the contents names but the file lacks (`format-appendix`), numbered sections with no text (`format-empty-section`, only for files with ≥5 numbered headings, table pages skipped), sheets the «Ведомость рабочих чертежей» lists that no title block has (`format-sheet-list`, only when ≥60 % of the sheets have a readable number) — `cmp/format_refs.go`. Low severity, always «к проверке». A finding can be confirmed, rejected or sent to a clarification (`clarify`, the comment is required: «?» in the viewer, `POST /api/discrepancy/status` with `comment`); the decision and the comment survive re-runs and iterations by fingerprint. A set also gets the note «наименование не совпадает с ведомостью» when a volume has the right шифр but no word of the statement's title (`SetInfo.Renamed`).
- **Cross-volume consistency** (`cmp/crossvol.go`, kind `cross_volume`, `Options.NoCross`): the same parameter (concrete grade, W, F, rebar class, and the values of the dictionary of `cmp/value_params.go`) for the same structural element named differently in two volumes of one side; parameters without an element only go to the table of parameters. Low severity — it may well be two different elements.
- **AI reading of an expertise remark** — `POST /api/remarks/ai` (`cmp/remark_ai.go`): the model sees the remark, the designer's answer and the fragments of the linked volumes that share the most with them, and says fixed / partial / not fixed / unclear. Every quotation is checked against the fragment text, a firm verdict without a valid quotation becomes «unclear». Writes only `ai_verdict`, `ai_note`, `ai_at`; the working status is never changed. Model: `--remark-model` / `REMARK_MODEL`.
- **Assistant for any document** — `POST /api/assist` (`assist.go`, `cmp/dialog.go`, panel `AssistantPanel.vue`): questions about the sheet that is on screen. Sources: a comparison (`source=comparison`: the open ПД/РД sheets and the selected finding), a Compose attachment (`attachment`), a Compose document (`document`), a file of a set (`cmp_document`), a file of a comparison (`comparison_doc` + `recordID`: opened from the «↗ документ» link next to each pager of the comparison viewer on the page on screen, with «← к сравнению» back), an ИД act (`act`). The model gets the text of the sheet (title block included) and, for a drawing or a scan (almost no text layer) or on request, the picture of the page (`--vision-model`, `deepseek-r15` by default; if it is missing or fails the answer is made from the text and says so) together with the method of the `read-drawings` skill (`AI_SKILLS_DIR` or `./skills`). `GET /api/assist/doc` and `GET /api/assist/page` serve the standalone viewer `?view=doc&source=…&id=…&namespaceID=…&page=` that Compose opens from the ✦ next to a file name and from the floating «Ассистент» button (client3: `useAssistantContext.js`, `AssistantLauncher.vue`; registered by the attachment lightbox, file links and the PDF view of Docs). In the agent the panel sits under the comparison viewer («✦ Ассистент» in the mode bar), under the ИД viewer and in the document viewer; the remarks workbench and the object overview link to it from every volume. Markdown documents of Compose have no file and are not covered. The question is started as a job (`"async": true` → `{jobID}`) and the answer is collected from `GET /api/assist/result?job=`, because a local model needs minutes while the API gateway waits 30 s for one request (setting `apigw.proxy.outbound-timeout`, not stored by default); the AI reading of remarks works the same way (background, progress in `X-AI-Status` / `X-AI-Error` of the board). Answers are facts-only, nothing is changed, the specialist decides.
- **Editions** (`cmp/edition.go`) — a volume is told by its шифр and its **change number**, not by the file name: «(Изм. N)» on the title page, else «изм. N» in the name of a one-volume file; the part, book and subsection come from the title pages. Stored on the document row (`change_no`, `part`, `book`, `subsection`, `file_hash`; typed values stay while the file hash is the same, `structure_object` is typed by hand). Of several editions of one volume on one side only the newest is compared (the rest is named in the comment); a «до → после» pair whose «после» is older, or has the same number, is warned about.
- **Scans and pages** (`cmp/pagekind.go`, `cmp/ocr.go`) — every recognized page keeps tesseract's mean confidence; below 85 % its findings are «низкая» and say how sure the reading is, weak and not-trusted scans are listed in the comment. A page is raster when it carries a full-size image (`pdfimages -list`, optional), and the format/orientation of a sheet is named («A3 альбомный», «A4×3 альбомный») where drawings of different proportions cannot be paired.
- **Excluded areas** (`cmp/exclude.go`, comparison field `excluded_regions`, `POST /api/exclusions`, panel «Сопоставление» → «Исключаемые области») — a rectangle of a page (side, file, page or all, box in page fractions) is left out: its text lines are not compared, drawing differences centred in it are not reported. The files are never changed; the title block is always out.
- **Reports and registry** — `GET /api/remarks/export?objectID=&…` (`cmp/registry_report.go`, `remarks_report.go`): the registry of the remarks as Excel — «Реестр» (the expertise's own status, the working status, the check against the documents, the AI reading, the system's result and the specialist's decision, assignee, due date, places in the documents), «История итераций», «Проверка устранения», «Динамика», «Сводка» (what the rows were selected by). The filters are those of the workbench (`status`, `crit`, `expert`, `discipline`, `section`, `assignee` — «—» is nobody, `iter`, `verdict`, `outcome`, `overdue`, `claim`, `archived`, `q`) plus `from`/`to`: the remarks with any activity in the period (a change in the expertise or a history entry). `?view=reports` is the «Отчёты» page: per object the registry export (with the filters above) and, per comparison, the report (xlsx/docx) and the list of remarks; `GET /api/reports/index` feeds it. The history of a remark keeps assignments, due dates and notes too (`kind`: import | status | assignee | due | note | decision | auto | comparison).
- **Correction packages** (`cmp/correction_package.go`, module `correction_packages`) — an iteration of a comparison (a corrected РД against the earlier one) becomes a package: the corrected files, the remarks of the object they concern, the result (resolved / persisting / new) and one history entry per run. Making it re-links and re-checks the remarks, so a remark whose difference the new edition fixed (and whose designer answered) is handed over to the control. Findings of a comparison that a later iteration continues are not used for matching any more.
- **Drawings and elements** (`cmp/raster.go`, `cmp/compare.go`, `cmp/elements.go`) — drawing sheets are paired by the sheet numbers of their title blocks (when ≥60 % have a readable, unique one), else by what is drawn (the most alike sheet of the other side, at most 30 % of the grid different; a sheet with none stays unpaired and is named), else by position for equal counts. Before the diff the contents of the two renderings are aligned (the bounding box of everything drawn is taken as the sheet, so a shift and a uniform rescale cancel out; a brightness range of the neighbouring cells tolerates the edges that two renderings never place alike; a region thinner than two cells is the edge of a frame, not a change). Findings carry the region on both pages (`raster-diff+aligned`). Design marks («Ф-2», «Пл-1») are read from the text layer of text pages and drawing sheets: a mark the ПД has and the РД has not (`marks-missing`, средняя; not for fragments), a mark only the РД has (`marks-added`, низкая), and within a side a mark the specification lists that no text or sheet mentions and vice versa (`spec-not-on-drawings`, `drawings-not-in-spec`, низкая). Drawings whose text is outlined (a common CAD export) have no marks to read. `Options.NoElements` switches the marks off.
- **Change registration** (`cmp/changereg.go`, kind `change_reg`, ТЗ 3.3.7–3.3.8) — in a «до → после» comparison whose «после» is a newer edition (change number of the volume, else the last change of its table) the agent reads, in the new edition, the table of the change registration (columns «изменённых / заменённых / новых / аннулированных» found by their heads in the words of the page; a row starting with a number is a change), the change rows of the title blocks («1 - Зам. 15-24», marked as title block so they are not compared) and the permission to make changes if it is in the volume. Every sheet (number from the title block) gets a fact from the comparison — same, changed (has findings), new, annulled — beside what the table and the title block declare; disagreements are findings: changed but not declared (high); new / annulled not declared, annulled but still there without «Аннулирован», not in the permission, table ≠ title block, no row for the change (medium); declared changed with no difference, declared new but there before, no table at all (low). The statuses go to `pairs_json` (`sheetChanges`), the panel «Изменения листов» under the viewer (a row opens the sheet) and the sheet «Изменения листов» of the Excel report. Fixture: `fixtures/changes-1` (`fixtures/gen_change_fixtures.mjs`).
- **Requirements of a remark** (`cmp/requirements.go`, `expertise_remarks.requirements_json`, ТЗ 3.4.3, 3.6.6, 3.9.5) — a remark is split into its requirements («1) … 2) …», markers, «необходимо / следует / представить»); each is matched to the findings, read by the AI and decided by the specialist on its own; the result of the remark is made from them. A list edited by hand is never overwritten (it is flagged when the remark's text changes). The card shows them, the registry export has the sheet «Требования».
- **Materials declared in an answer** (`cmp/materials.go`, `cmp/materials_object.go`, ТЗ 3.4.9–3.4.11) — the designer's answer is parsed for what it says it supplied: volumes («том КР5.2», «ТКР5.4.1»), sheets and pages, appendices («Приложение Б»), documents (calculation, letter, certificate, protocol…); norms and contracts are skipped. Each is looked for in the loaded volumes of the object (шифр, `sheet_map`, the headings of appendices and documents inside a volume — `cmp_documents.headings_json`, read during a run) and in the supporting files attached to the remark or to a correction package (`materials_files`, PDF headings are read too). States: found / not in the volume / not found / volume not loaded; per remark `materials_state` and the list `materials_json`. A remark whose declared materials are missing gets «Требует проверки» instead of «Устранено». The card has the block «Заявленные в ответе материалы», the workbench the KPI/filter «материалы не найдены» (`materials=1` in the export), the registry the column and the sheet «Материалы»; a correction package keeps a snapshot of the answers and the count of remarks with missing materials. The content of a material is never judged.
- **Notifications** (`cmp/notify.go`, `notify.go`, ТЗ 3.7.5) — the agent looks for the events itself: an assignment and a change of the working status from the history of a remark (an entry is one event; the status of an entry against the one before), a near due date (`--notify-soon-days`, 3) and a broken one from the registry. A pass runs every `--notify-every` (15 min) over the objects of the default namespace, and at once for an object after a hand-made change of the status, assignee or due date; `POST /api/notifications/scan?objectID=` runs it on request. The recipients are the directory `notify_recipients` (name as written in «Ответственный», e-mail, role `assignee`/`manager`, object or all, events — empty: an assignee gets all four, a manager the overdue ones); an assignee without a row gets the feed only (or mail, if the field holds an address). Each notification has a key (history entry, or remark + due date + event) and is never made twice; history older than 3 days is not announced, so the first run does not flood. Notifications go to the module `notifications` — the feed «🔔 Уведомления» of the workbench (`GET /api/notifications?objectID=&recipient=`, `POST /api/notifications/read`), a click opens the remark — and, with SMTP set (`--smtp host:port`, `--smtp-from`, `--smtp-user`, `SMTP_PASSWORD`, `--smtp-tls starttls|tls|none`; or `SMTP_ADDR`, `SMTP_FROM`, `SMTP_USER`, `SMTP_TLS`), as a plain-text letter with a link to the remark (`--public-url`, opens `?view=remarks&objectID=&remark=`); a failed letter is tried again on the next passes, at most 5 times. In `--fixtures` mode the feed is computed from the registry in memory (managers from `<case>/recipients.json`), nothing is sent.
- **Where a remark points inside a volume** (`cmp/remark_fragment.go`, ТЗ 3.4.2, 3.5.1, 3.4.7) — the text the remark quotes («…») or the lines that share most of its words and values (a window of 1–3 lines; pages the remark names count more) are located in the linked volumes on both sides; `links_json` gets a `fragment` place with its box, the card opens the document viewer on that page with the box highlighted (`?view=doc&…&page=&box=x0,y0,x1,y1`). The quotations of the AI reading keep their page and box too (`ai_cites_json`) and are links.
- **Roles** (`cmp/roles.go`, `identity.go`, `compose/roles.mjs`, ТЗ 3.8, 1.2.6) — `node compose/roles.mjs` creates the platform roles `stroykontrol_mgr` (руководитель проекта), `stroykontrol_qc` (специалист по контролю качества), `stroykontrol_des` (проектировщик), `stroykontrol_view` (наблюдатель) with read access to every module of the namespace and write access by role (`DRY_RUN=1` prints the rules); administrators (`admins`) can do everything. With `--roles=on` (`ROLES=on`) the agent checks every POST: the interface, opened inside the platform, passes the token of the signed-in user (`X-Platform-Token`, `web/src/platform.js`), the agent validates it at the authorization service (`--auth-url`, default `<api root>/auth`, `/oauth2/info`), maps the roles and refuses with 403 what the role may not do: upload (sets, volumes, runs), review (findings, pairing, remark cards), work on remarks, assign, confirm the result, ИД checks. A confirmation by the assignee of the remark is refused. The history and the decisions are signed with the user's name instead of «специалист ПТО». `GET /api/me` says who is signed in and what is allowed; the interface hides the rest. `--roles=off` (default) and `--fixtures` check nothing.
- **Documents of an object** (`cmp/doc_registry.go`, ТЗ 2.2.4, 3.1.2, 3.1.4, 3.1.7) — `cmp_documents.section` (from the volume code, «2024-15-КЖ2» → «КЖ», manual value kept) and `structure` (a record of the new module `structures` — the structures and sections of the route of the object; matched by code or name in the field «Сооружение / участок» or the title of the volume, a manual choice is kept). Reading a file for the first time writes the upload journal `upload_log` (who and when attached it — the system fields of the record, file, hash, pages, and whether it is a new volume, a new edition of a volume or a repeat of a loaded file). The object overview shows the tree «сооружение → раздел → том → редакции» and the last uploads.
- **Completeness beyond volumes** (`cmp/set_units.go`, ТЗ 3.2.1, 3.2.4, 3.2.5, 3.2.8) — a row of the statement of volumes has a kind (`object_volumes.kind`: том, раздел, приложение, ссылочный документ; `parent_code` — the volume an appendix belongs to) and counts for completeness when required: a section is present when a volume of its mark is in the set, an appendix when its heading is in the volume, a reference when a volume with its code or title is loaded; a missing reference counts only when the row is required. References to other volumes of the same project inside the text («см. 2024-15-КЖ3») are read (`volume_refs_json`) and the ones the object lacks are named in the note of the set. The titles of the volumes are checked against the statement, the contents of other volumes and between files of one volume (`contents_json`).
- **Tables cell by cell** (`cmp/tables.go`, kind `table_cell`, ТЗ 3.3.3) — the tables of text pages other than ВОР and specifications are read from the words of the page (rows by height, cells by gaps, columns by the header; wrapped cells are joined), the tables of two editions are matched by their header, rows by the key column (the first unique non-running column), cells by the column name. A changed cell, a removed and an added row are findings with the box of the row on both pages; numbers changed by more than 10 % are «высокая». The line findings of the same rows are dropped. Drawings and tables without a text layer are not read.
- **Values between drawings and text** (`cmp/value_params.go`, ТЗ 4.2.7) — besides the key parameters a dictionary of values: thickness, depth of foundation, protective layer, spacing, diameter, height, span, elevations, absolute elevation of zero, loads (snow, wind, live), slope, seismicity, fire resistance, responsibility level; units are brought to the unit of the dictionary. Within a side the values the text pages name for an element are compared with the ones its drawing sheets name (text layer of the sheets); a drawing that names other values only is a `cross_volume` finding «Чертёж ↔ текст», low, «к проверке».
- **Remarks from PDF** (`cmp/expertise_pdf.go`, ТЗ 2.2.3, 3.4.1) — an attached PDF is read without a fixed template: a table with a «Замечание» column (from the words of the pages; continued over pages, the repeated header skipped) goes through the same parsers as Excel (the expertise export, else the generic one); without a table, numbered items («1.», «2.3)», «Замечание № 4») with «Ответ: …» paragraphs and «Раздел …» headings. PDFs without a text layer are refused. Samples: `cmp/testdata/remarks-pdf` (`fixtures/gen_remarks_pdf.mjs`). Tuned on generated samples only — a real expertise PDF may need adjustments.
- **Dialog** — `POST /api/dialog` (`cmp/dialog.go`, panel «Диалог о расхождениях»): the model gets the facts (selected finding, neighbours, the text around it, expertise remarks), no opinions about design decisions; sheet markers `[[ПД:12]]` in its answer become «go to the sheet» buttons only for sheets named in the facts.

### Backup

Everything the agent keeps lives in the platform: the records of the namespace and the attached files (the database of Compose and its attachment store) — back them up with the platform (database dump plus the storage of attachments, on the schedule of the platform). The agent's own directory `--cache` (`var/stroykontrol-raster-cache`: rendered pages, text layers, OCR results, keyed by the hash of the file) is a cache: it can be deleted and is rebuilt on the next run, so it is not backed up. The journal of actions is the history of the records in the platform plus the agent's own journals (`expertise_remark_history`, `expertise_imports`, `upload_log`, `notifications`).

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
| `GET /api/pages?recordID=&namespaceID=&side=pd\|rd` | `{supported, kind, pageCount}` — `kind` is `pdf`/`docx`/`dxf`/`xlsx`/`other`; PDF is rasterized server-side on first call, DOCX/DXF/XLSX report `pageCount:0` (the browser determines it after rendering) |
| `GET /api/page?recordID=&namespaceID=&side=pd\|rd&n=` | One PDF page as JPEG (PDF only) |
| `GET /api/file?recordID=&namespaceID=&side=pd\|rd` | Raw attachment bytes with its Content-Type (used by the browser to render DOCX client-side) |
| `GET /api/text?recordID=&namespaceID=&side=pd\|rd` | `{pages: [...]}` — per-page extracted text, PDF only (via `pdftotext -layout`, cached); DOCX text comes from the client's own render, no call needed |
| `GET /api/report/export` · `GET/POST /api/pairing` · `POST /api/dialog` · `POST /api/remarks/ai` · `GET /api/sets/overview`, `POST /api/sets/check` | See the sections above and «Документы объекта» in [compose/README.md](compose/README.md) |
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
