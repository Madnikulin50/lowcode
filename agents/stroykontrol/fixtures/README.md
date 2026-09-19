# Fixtures — running stroykontrol-web with no Compose at all

`--fixtures=<dir>` swaps the agent's data backend from a live Compose
instance (`--api`/`--token`/`--namespace`) to this directory. No Compose
server, no network call, no API token — the binary is fully self-contained.
Useful for UI development on the viewer itself, demos, and screenshots.

```bash
go run . --listen=:8092 --fixtures=fixtures
# open http://localhost:8092/?recordID=demo-1&namespaceID=x
```

`recordID` in the URL/query is just the case's directory name under
`--fixtures`; `namespaceID` is accepted (the real viewer always sends one)
but ignored.

## Layout

```
<fixtures-dir>/<caseID>/
  meta.json           optional  {"title","status","similarityPercent","comment"}
  discrepancies.json  optional  [{"type","severity","description","pageNumber"}, ...]
  pd.<ext>            optional  any extension — content is sniffed (see IsPDF/IsDOCX
  rd.<ext>            optional  in rasterize.go), so pd.pdf or pd.docx both work
```

Everything is optional: a case with no `meta.json` still renders (title
falls back to the caseID); a case with no `pd.*`/`rd.*` just reports that
side as absent (`hasPdFile`/`hasRdFile: false`) instead of erroring — same
as a real comparison record whose attachment hasn't been uploaded yet.

`meta.json` fields map straight onto `pd_rd_comparisons`: `status` is
whatever string the viewer's status badge expects (`new`/`processing`/
`done`/`failed` — see `NamespaceSidebar`/viewer JS for the exact set),
`similarityPercent` is a plain string like `"92.5"`.

`discrepancies.json` entries map onto `pd_rd_discrepancies`: `pageNumber` is
a string; a value that doesn't resolve within the rendered page range shows
as a non-clickable "стр. N ⚠" in the UI (see main README's UI section) —
same fallback a real out-of-range check result gets.

## `demo-1`

Ships a ready-to-run example: two one-page PDFs (`pd.pdf`/`rd.pdf`,
generated via `ps2pdf`, see git history for the source `.ps` if you want to
regenerate them) with one deliberate numeric difference, plus two seeded
discrepancies pointing at it. Good for a quick smoke-test of the viewer or
for taking screenshots without touching any real data.

## Snapshotting a real comparison into a fixture

[`compose/export_fixture.mjs`](../compose/export_fixture.mjs) pulls one real
`pd_rd_comparisons` record — its title/status/similarity/comment, its
`pd_rd_discrepancies` rows, and its `pd_file`/`rd_file` attachments — out of
a live Compose instance straight into this layout:

```bash
cd agents/stroykontrol/compose
COMPOSE_DSN=postgres://postgres:Zse45rdx@127.0.0.1:5432/test11?sslmode=disable \
  node export_fixture.mjs <namespaceID> <recordID>
```

(same `mintToken()`/`detectBase()` auth path as `seed.mjs` — see its own
doc-comment for the DSN/TOKEN options). It writes into
`fixtures/<recordID>/` by default; pass a third argument to pick a
different `caseID`/dirname. Once exported, that record is viewable fully
offline:

```bash
go run . --fixtures=fixtures
# http://localhost:8092/?recordID=<recordID>&namespaceID=x
```
