#!/usr/bin/env python3
"""Seed turnover modules, charts and the Оборачиваемость summary page."""
import json
import os
import subprocess
import sys

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "../../../.."))
ENV_PATH = os.path.join(ROOT, ".env")
SQL_PATH = os.path.join(os.path.dirname(__file__), "seed_turnover.sql")

NS = "495727984893558785"
CONN = "495279171887497217"

MOD_SLICE = 510050000000100001
MOD_FACT = 510050000000100002

CH_VEL = 510050000000120001
CH_STORE_TURNS = 510050000000120002
CH_STORE_FROZEN = 510050000000120003
CH_CAT_DOI = 510050000000120004
CH_CAT_TURNS = 510050000000120005
CH_VEL_VALUE = 510050000000120006

PAGE = 510050000000130001
LAYOUT = 510050000000130002
FID0 = 510050000000110001


def load_env(path):
    env = {}
    with open(path) as fh:
        for line in fh:
            line = line.strip()
            if not line or line.startswith("#") or "=" not in line:
                continue
            key, val = line.split("=", 1)
            env[key] = val.strip().strip('"').strip("'")
    return env


def dsn_local(dsn):
    return dsn.replace("host.docker.internal", "127.0.0.1")


def psql(dsn, sql):
    r = subprocess.run(
        ["psql", dsn, "-v", "ON_ERROR_STOP=1", "-q"],
        input=sql,
        text=True,
        capture_output=True,
    )
    if r.returncode != 0:
        sys.stderr.write(r.stderr)
        raise SystemExit(r.returncode)
    if r.stdout:
        sys.stdout.write(r.stdout)
    return r.stdout


def psql_at(dsn, sql):
    r = subprocess.run(
        ["psql", dsn, "-v", "ON_ERROR_STOP=1", "-At"],
        input=sql,
        text=True,
        capture_output=True,
    )
    if r.returncode != 0:
        sys.stderr.write(r.stderr)
        raise SystemExit(r.returncode)
    return r.stdout.strip()


def dal_cfg(ident):
    return {
        "dal": {
            "ident": ident,
            "constraints": {},
            "connectionID": CONN,
            "systemFieldEncoding": {
                "id": None,
                "meta": {"omit": True},
                "ownedBy": {"omit": True},
                "moduleID": {"omit": True},
                "revision": {"omit": True},
                "createdAt": {"omit": True},
                "createdBy": {"omit": True},
                "deletedAt": {"omit": True},
                "deletedBy": {"omit": True},
                "updatedAt": {"omit": True},
                "updatedBy": {"omit": True},
                "namespaceID": {"omit": True},
            },
        },
        "privacy": {"usageDisclosure": ""},
        "recordRevisions": {"ident": "", "enabled": False},
        "recordDeDup": {},
        "discovery": {
            "public": {"result": [{"lang": "", "fields": []}]},
            "private": {"result": [{"lang": "", "fields": []}]},
            "protected": {"result": [{"lang": "", "fields": []}]},
        },
    }


def field_config():
    return {
        "dal": {"encodingStrategy": {"plain": {"ident": ""}}},
        "privacy": {"usageDisclosure": ""},
        "recordRevisions": {"enabled": False},
    }


def num_opts(suffix="", precision=1):
    return {
        "max": 0,
        "min": 0,
        "hint": {"view": ""},
        "step": 1,
        "format": "",
        "prefix": "",
        "suffix": suffix,
        "display": "number",
        "variant": "success",
        "animated": False,
        "precision": precision,
        "showValue": True,
        "thresholds": [],
        "description": {"view": ""},
        "presetFormat": "custom",
        "showProgress": False,
        "showRelative": False,
        "multiDelimiter": "\n",
    }


def str_opts():
    return {
        "hint": {"view": ""},
        "multiLine": False,
        "description": {"view": ""},
        "multiDelimiter": "\n",
        "useRichTextEditor": False,
    }


def select_opts(options):
    return {
        "hint": {"view": ""},
        "prefix": "",
        "suffix": "",
        "options": options,
        "selectType": "default",
        "description": {"view": ""},
        "displayType": "badge",
        "badgeGradient": True,
        "multiDelimiter": "\n",
        "isUniqueMultiValue": False,
    }


def lit(obj):
    return "'" + json.dumps(obj, ensure_ascii=False).replace("'", "''") + "'::jsonb"


def field_sql(fid, module_id, place, kind, name, label, options):
    return (
        f"INSERT INTO compose_module_field "
        f"(id, rel_module, place, kind, options, name, label, config, is_required, is_multi, default_value, expressions) "
        f"VALUES ({fid}, {module_id}, {place}, '{kind}', {lit(options)}, '{name}', '{label}', "
        f"{lit(field_config())}, false, false, '[]'::jsonb, '{{}}'::jsonb);"
    )


def chart_cfg(module_id, filter_ql, dim_field, metric_field, metric_type, label, aggregate="SUM",
              horizontal=False, rotate=0, suffix="", fmt="0.0"):
    return {
        "reports": [{
            "reportID": str(int(module_id) + 9),
            "filter": filter_ql,
            "moduleID": str(module_id),
            "metrics": [{
                "aggregate": aggregate,
                "field": metric_field,
                "fill": False,
                "fixTooltips": True,
                "valueLabelPosition": "top",
                "formatting": {"format": fmt, "prefix": "", "presetFormat": "custom", "suffix": suffix},
                "label": label,
                "metricID": str(int(module_id) + 7),
                "relativeValue": False,
                "rose": False,
                "smooth": True,
                "symbol": "circle",
                "type": metric_type,
            }],
            "dimensions": [{
                "conditions": {},
                "dimensionID": str(int(module_id) + 8),
                "field": dim_field,
                "meta": {"fields": [], "steps": None},
                "modifier": "(no grouping / buckets)",
                "rotateLabel": rotate,
                "timeLabels": False,
            }],
            "yAxis": {
                "axisPosition": "left",
                "axisType": "linear",
                "beginAtZero": True,
                "formatting": {"format": fmt, "prefix": "", "presetFormat": "custom", "suffix": suffix},
                "horizontal": horizontal,
                "labelPosition": "end",
                "rotateLabel": 0,
            },
            "legend": {
                "align": "center",
                "isHidden": metric_type != "doughnut",
                "isScrollable": True,
                "orientation": "horizontal",
                "position": {"isDefault": True},
            },
            "offset": {
                "bottom": "20",
                "isDefault": False,
                "left": "12",
                "right": "20",
                "top": "40",
            },
            "renderer": {},
        }],
        "colorScheme": "lowcode",
        "toolbox": {"saveAsImage": False, "showDataTable": False, "timeline": ""},
    }


def metric_item(label, field, suffix, number_format, color="black", role="default"):
    return {
        "label": label,
        "filter": "slice_kind = 'kpi'",
        "prefix": "",
        "suffix": suffix,
        "moduleID": str(MOD_SLICE),
        "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
        "operation": "max",
        "showLabel": True,
        "bucketSize": "",
        "dateFormat": "",
        "expression": "",
        "valueStyle": {
            "color": color,
            "fontSize": "22",
            "labelColor": "primary",
            "backgroundColor": "#FFFFFF00",
            "colorThresholds": [],
        },
        "metricField": field,
        "transformFx": "",
        "numberFormat": number_format,
        "dimensionField": "",
        "fieldRole": role,
    }


def rl_field(fid, name, kind, label, options):
    return {
        "kind": kind,
        "name": name,
        "label": label,
        "config": field_config(),
        "fieldID": str(fid),
        "isMulti": False,
        "options": options if isinstance(options, dict) else {},
        "isSystem": False,
        "maxLength": 0,
        "isRequired": False,
        "isSortable": True,
        "expressions": {},
        "isQueryable": True,
        "defaultValue": [],
        "isFilterable": True,
        "canReadRecordValue": True,
        "canUpdateRecordValue": False,
    }


def wrap_block(kind, title, block_id, xywh, options):
    return {
        "kind": kind,
        "meta": {"hidden": False, "visibility": {"roles": [], "expression": ""}},
        "xywh": xywh,
        "style": {
            "wrap": {"kind": "card"},
            "border": {"enabled": False},
            "variants": {"headerText": "dark"},
        },
        "title": title,
        "blockID": str(block_id),
        "options": options,
    }


def load_module_fields(dsn, module_id, names):
    out = []
    for name in names:
        row = psql_at(
            dsn,
            f"SELECT id||E'\\t'||kind||E'\\t'||label||E'\\t'||options::text "
            f"FROM compose_module_field WHERE rel_module={module_id} AND name='{name}' "
            f"AND deleted_at IS NULL LIMIT 1;",
        )
        if not row:
            continue
        fid, kind, label, options_txt = row.split("\t", 3)
        try:
            options = json.loads(options_txt)
        except json.JSONDecodeError:
            options = {}
        out.append(rl_field(int(fid), name, kind, label, options))
    return out


def rl_look(**extra):
    base = {
        "hideAddButton": True,
        "hideImportButton": True,
        "hideHeader": False,
        "hidePaging": False,
        "hideSearch": False,
        "hideSorting": False,
        "hideFiltering": True,
        "editable": False,
        "enableRecordPageNavigation": False,
        "fullPageNavigation": False,
        "inlineRecordEdit": False,
        "recordSelectorDisplay": "",
        "magnifyOption": "modal",
        "liveFilterEnabled": False,
        "selectable": False,
        "allowExport": False,
        "compactRows": True,
        "alignNumbers": True,
        "stickyHeader": True,
        "showRowSignal": True,
        "displayMode": "responsive",
        "sparklineMax": 30,
    }
    base.update(extra)
    return base


def insert_module(dsn, mid, handle, name, table):
    psql(dsn, (
        "INSERT INTO compose_module (id, rel_namespace, handle, name, meta, config) "
        f"VALUES ({mid}, {NS}, '{handle}', '{name}', "
        f"{lit({'ui': {'admin': {'fields': []}}})}, {lit(dal_cfg(table))});"
    ))


def main():
    env = load_env(ENV_PATH)
    dsn = dsn_local(env["DB_DSN"])

    fact_exists = psql_at(dsn, "SELECT to_regclass('public.stock_reorder_fact') IS NOT NULL;")
    if fact_exists != "t":
        sys.stderr.write("stock_reorder_fact missing — run seed_stock_reorder.py first\n")
        raise SystemExit(1)
    margin_exists = psql_at(dsn, "SELECT to_regclass('public.receipt_margin') IS NOT NULL;")
    if margin_exists != "t":
        sys.stderr.write("receipt_margin missing — run seed_receipt_margin.py first\n")
        raise SystemExit(1)

    print("Building turnover_fact + turnover_slice…")
    psql(dsn, open(SQL_PATH).read())

    print("Replacing Compose metadata…")
    psql(dsn, f"""
DELETE FROM compose_page_layout WHERE id = {LAYOUT} OR (page_id = {PAGE} AND handle = 'primary');
DELETE FROM compose_page WHERE id = {PAGE} OR handle = 'turnover';
DELETE FROM compose_chart WHERE id IN (
  {CH_VEL},{CH_STORE_TURNS},{CH_STORE_FROZEN},{CH_CAT_DOI},{CH_CAT_TURNS},{CH_VEL_VALUE}
) OR handle IN (
  'turn-velocity','turn-store-turns','turn-store-frozen',
  'turn-cat-doi','turn-cat-turns','turn-velocity-value'
);
DELETE FROM compose_module_field WHERE rel_module IN ({MOD_SLICE},{MOD_FACT});
DELETE FROM compose_module WHERE id IN ({MOD_SLICE},{MOD_FACT})
  OR handle IN ('Turnover_slice','Turnover_fact');
""")

    insert_module(dsn, MOD_SLICE, "Turnover_slice", "turnover_slice", "turnover_slice")
    insert_module(dsn, MOD_FACT, "Turnover_fact", "turnover_fact", "turnover_fact")

    velocity_opts = select_opts([
        {"text": "Дефицит / срочно", "value": "rush", "style": {"textColor": "light", "backgroundColor": "danger"}},
        {"text": "Быстрая", "value": "fast", "style": {"textColor": "light", "backgroundColor": "success"}},
        {"text": "Норма", "value": "normal", "style": {"textColor": "light", "backgroundColor": "info"}},
        {"text": "Медленная", "value": "slow", "style": {"textColor": "dark", "backgroundColor": "warning"}},
        {"text": "Мёртвый остаток", "value": "dead", "style": {"textColor": "light", "backgroundColor": "secondary"}},
    ])

    slice_fields = [
        ("String", "slice_kind", "Срез", str_opts()),
        ("String", "slice_key", "Ключ", str_opts()),
        ("String", "slice_label", "Название", str_opts()),
        ("Number", "sku_count", "SKU", num_opts("", 0)),
        ("Number", "dead_count", "Мёртвый", num_opts("", 0)),
        ("Number", "slow_count", "Медленные", num_opts("", 0)),
        ("Number", "normal_count", "Норма", num_opts("", 0)),
        ("Number", "fast_count", "Быстрые", num_opts("", 0)),
        ("Number", "rush_count", "Дефицит", num_opts("", 0)),
        ("Number", "avg_turns", "Оборачиваемость, x", num_opts("x", 2)),
        ("Number", "avg_doi", "DOI, дн.", num_opts(" дн.", 1)),
        ("Number", "stock_value", "Стоимость остатка", num_opts(" ₽", 0)),
        ("Number", "revenue_30d", "Выручка 30д", num_opts(" ₽", 0)),
        ("Number", "frozen_value", "Заморожено", num_opts(" ₽", 0)),
        ("Number", "problem_count", "Медленные+мёртвые", num_opts("", 0)),
        ("Number", "sort_score", "Сорт.", num_opts("", 2)),
    ]
    psql(dsn, "\n".join(
        field_sql(FID0 + i, MOD_SLICE, i, kind, name, label, opts)
        for i, (kind, name, label, opts) in enumerate(slice_fields)
    ))

    fact_fields = [
        ("String", "store_name", "Магазин", str_opts()),
        ("String", "product_name", "Товар", str_opts()),
        ("String", "category_name", "Категория", str_opts()),
        ("String", "brand", "Бренд", str_opts()),
        ("Number", "stock_quantity", "Остаток, шт", num_opts("", 0)),
        ("Number", "stock_sum", "Остаток, ₽", num_opts(" ₽", 0)),
        ("Number", "qty_30d", "Продано 30д", num_opts("", 1)),
        ("Number", "revenue_30d", "Выручка 30д", num_opts(" ₽", 0)),
        ("Number", "turns_30d", "Оборот 30д, x", num_opts("x", 2)),
        ("Number", "turns_year", "Оборот год, x", num_opts("x", 2)),
        ("Number", "doi_days", "DOI, дн.", num_opts(" дн.", 1)),
        ("Number", "days_of_cover", "DOC, дн.", num_opts(" дн.", 1)),
        ("Select", "velocity", "Оборачиваемость", velocity_opts),
        ("String", "health_level", "Здоровье", str_opts()),
    ]
    psql(dsn, "\n".join(
        field_sql(FID0 + 100 + i, MOD_FACT, i, kind, name, label, opts)
        for i, (kind, name, label, opts) in enumerate(fact_fields)
    ))

    charts = [
        (CH_VEL, "turn-velocity", "SKU по скорости оборачиваемости",
         chart_cfg(MOD_SLICE, "slice_kind = 'velocity'", "slice_label", "sku_count",
                   "doughnut", "SKU", aggregate="SUM", fmt="0.")),
        (CH_VEL_VALUE, "turn-velocity-value", "Стоимость остатка по скорости",
         chart_cfg(MOD_SLICE, "slice_kind = 'velocity'", "slice_label", "stock_value",
                   "bar", "₽", aggregate="SUM", horizontal=True, fmt="0.")),
        (CH_STORE_TURNS, "turn-store-turns", "Оборачиваемость по магазинам",
         chart_cfg(MOD_SLICE, "slice_kind = 'store'", "slice_label", "avg_turns",
                   "bar", "x / год", aggregate="MAX", horizontal=True, fmt="0.0")),
        (CH_STORE_FROZEN, "turn-store-frozen", "Замороженный остаток по магазинам",
         chart_cfg(MOD_SLICE, "slice_kind = 'store_frozen'", "slice_label", "frozen_value",
                   "bar", "₽", aggregate="MAX", horizontal=True, fmt="0.")),
        (CH_CAT_DOI, "turn-cat-doi", "DOI по категориям",
         chart_cfg(MOD_SLICE, "slice_kind = 'cat_doi'", "slice_label", "avg_doi",
                   "bar", "дн.", aggregate="MAX", horizontal=True, fmt="0.0")),
        (CH_CAT_TURNS, "turn-cat-turns", "Оборачиваемость по категориям",
         chart_cfg(MOD_SLICE, "slice_kind = 'category'", "slice_label", "avg_turns",
                   "bar", "x / год", aggregate="MAX", horizontal=True, fmt="0.0")),
    ]
    ch_sql = []
    for cid, handle, name, cfg in charts:
        cfg_txt = json.dumps(cfg, ensure_ascii=False).replace("'", "''")
        ch_sql.append(
            "INSERT INTO compose_chart (id, handle, rel_namespace, name, config) "
            f"VALUES ({cid}, '{handle}', {NS}, '{name}', '{cfg_txt}');"
        )
    psql(dsn, "\n".join(ch_sql))

    fact_rl = load_module_fields(dsn, MOD_FACT, [
        "store_name", "product_name", "category_name", "velocity",
        "stock_quantity", "stock_sum", "qty_30d", "turns_year", "doi_days",
    ])
    slow_rl = load_module_fields(dsn, MOD_FACT, [
        "product_name", "store_name", "category_name", "velocity",
        "doi_days", "stock_sum", "qty_30d", "turns_year",
    ])

    blocks = [
        wrap_block("Metric", "Сводка оборачиваемости", 1, [0, 0, 48, 14], {
            "metrics": [
                metric_item("Средняя оборачиваемость", "avg_turns", "x", "0.0", "#2e59d9", "hero"),
                metric_item("Средний DOI", "avg_doi", " дн.", "0.0", "#36b9cc", "badge"),
                metric_item("Медленные + мёртвые SKU", "problem_count", "", "0.", "#f6c23e", "badge"),
                metric_item("Заморожено", "frozen_value", " ₽", "0.", "#e74a3b", "badge"),
                metric_item("Выручка 30д", "revenue_30d", " ₽", "0.", "#1cc88a", "meta"),
            ],
            "refreshRate": 0,
            "showRefresh": False,
            "magnifyOption": "",
            "likeRecordList": True,
            "density": "comfortable",
            "horizontalFieldLayoutEnabled": True,
        }),
        wrap_block("Chart", "SKU по скорости", 2, [0, 14, 24, 32], {
            "chartID": str(CH_VEL),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0, "showRefresh": False, "magnifyOption": "modal", "liveFilterEnabled": False,
        }),
        wrap_block("Chart", "Стоимость остатка по скорости", 3, [24, 14, 24, 32], {
            "chartID": str(CH_VEL_VALUE),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0, "showRefresh": False, "magnifyOption": "modal", "liveFilterEnabled": False,
        }),
        wrap_block("Chart", "Оборачиваемость по магазинам", 4, [0, 46, 24, 34], {
            "chartID": str(CH_STORE_TURNS),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0, "showRefresh": False, "magnifyOption": "modal", "liveFilterEnabled": False,
        }),
        wrap_block("Chart", "Замороженный остаток по магазинам", 5, [24, 46, 24, 34], {
            "chartID": str(CH_STORE_FROZEN),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0, "showRefresh": False, "magnifyOption": "modal", "liveFilterEnabled": False,
        }),
        wrap_block("Chart", "DOI по категориям", 6, [0, 80, 24, 34], {
            "chartID": str(CH_CAT_DOI),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0, "showRefresh": False, "magnifyOption": "modal", "liveFilterEnabled": False,
        }),
        wrap_block("Chart", "Оборачиваемость по категориям", 7, [24, 80, 24, 34], {
            "chartID": str(CH_CAT_TURNS),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0, "showRefresh": False, "magnifyOption": "modal", "liveFilterEnabled": False,
        }),
        wrap_block("RecordList", "Медленные и мёртвые остатки", 8, [0, 114, 48, 32], {
            "moduleID": str(MOD_FACT),
            "fields": slow_rl,
            "prefilter": "velocity = 'slow' OR velocity = 'dead'",
            "presort": "doi_days DESC, stock_sum DESC",
            "perPage": 25,
            **rl_look(
                signalField="velocity",
                rowHighlightField="velocity",
                showRowSignal=True,
                groupByField="velocity",
                sparklineField="doi_days",
                sparklineMax=180,
            ),
        }),
        wrap_block("RecordList", "Быстрые / дефицитные SKU", 9, [0, 146, 48, 28], {
            "moduleID": str(MOD_FACT),
            "fields": fact_rl,
            "prefilter": "velocity = 'fast' OR velocity = 'rush'",
            "presort": "turns_year DESC",
            "perPage": 20,
            **rl_look(
                signalField="velocity",
                rowHighlightField="velocity",
                showRowSignal=True,
                groupByField="velocity",
                sparklineField="turns_year",
                sparklineMax=24,
            ),
        }),
    ]

    page_cfg = {
        "prompt": "Ты аналитик товарных запасов. По этой странице разбери оборачиваемость: "
                  "средние turns и DOI, долю медленных/мёртвых SKU, замороженную стоимость остатка, "
                  "разрез по магазинам и категориям, дефицитные позиции. "
                  "Предлагай, что распродавать, где снижать заказ, где срочно пополнять."
    }
    page_meta = {"notifications": {"enabled": True}, "allowPersonalLayouts": False}
    psql(dsn, f"""
INSERT INTO compose_page (id, title, handle, self_id, rel_module, rel_namespace, meta, config, blocks, visible, weight, description)
VALUES ({PAGE}, 'Оборачиваемость', 'turnover', 0, 0, {NS},
        {lit(page_meta)}, {lit(page_cfg)}, {lit(blocks)},
        true, 2, 'Анализ оборачиваемости запасов: DOI, turns, медленные/быстрые SKU');
""")

    layout_blocks = [
        {
            "blockID": str(b["blockID"]),
            "xywh": b["xywh"],
            "meta": {"hidden": False, "visibility": {"roles": [], "expression": ""}},
        }
        for b in blocks
    ]
    layout_meta = {"title": "Оборачиваемость", "description": ""}
    layout_cfg = {
        "buttons": {
            "new": {"label": "", "enabled": True},
            "back": {"label": "", "enabled": True},
            "edit": {"label": "", "enabled": True},
            "clone": {"label": "", "enabled": True},
            "delete": {"label": "", "enabled": True},
            "submit": {"label": "", "enabled": True},
        },
        "useTitle": False,
        "validation": {},
        "visibility": {"expression": "", "roles": []},
    }
    psql(dsn, f"""
INSERT INTO compose_page_layout (id, handle, page_id, parent_id, rel_namespace, weight, meta, config, blocks, owned_by)
VALUES ({LAYOUT}, 'primary', {PAGE}, 0, {NS}, 1,
        {lit(layout_meta)}, {lit(layout_cfg)}, {lit(layout_blocks)}, 0);
""")

    kpi = psql_at(
        dsn,
        "SELECT avg_turns||'/'||avg_doi||'/'||problem_count||'/'||frozen_value "
        "FROM turnover_slice WHERE slice_kind='kpi' LIMIT 1;",
    )
    print("OK")
    print("  modules Turnover_slice / Turnover_fact")
    print("  charts turn-velocity*, turn-store-*, turn-cat-*")
    print("  page /ns/loop/pages/%s (Оборачиваемость)" % PAGE)
    print("  kpi avg_turns/avg_doi/problem_sku/frozen = %s" % kpi)
    print("  restart GoLand Server_RU_test_translations for DalModelReload")


if __name__ == "__main__":
    main()
