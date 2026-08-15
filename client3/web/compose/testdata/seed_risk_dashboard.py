#!/usr/bin/env python3
"""Seed Risk_slice module, charts and the Риски summary page (stores / stock / products / incidents)."""
import json
import os
import subprocess
import sys

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "../../../.."))
ENV_PATH = os.path.join(ROOT, ".env")
SQL_PATH = os.path.join(os.path.dirname(__file__), "seed_risk_dashboard.sql")

NS = "495727984893558785"
CONN = "495279171887497217"

MOD_INC = 495727984904372225
MOD_FACT = 510020000000100002
MOD_SLICE = 510030000000100001

CH_CRIT = 510030000000120001
CH_TYPE = 510030000000120002
CH_STORE_INC = 510030000000120003
CH_STORE_RISK = 510030000000120004
CH_STOCK = 510030000000120005
CH_PRODUCT = 510030000000120006

PAGE = 510030000000130001
LAYOUT = 510030000000130002
FID0 = 510030000000110001


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


def num_opts(suffix="", precision=2):
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


def lit(obj):
    return "'" + json.dumps(obj, ensure_ascii=False).replace("'", "''") + "'::jsonb"


def field_sql(fid, module_id, place, kind, name, label, options):
    return (
        f"INSERT INTO compose_module_field "
        f"(id, rel_module, place, kind, options, name, label, config, is_required, is_multi, default_value, expressions) "
        f"VALUES ({fid}, {module_id}, {place}, '{kind}', {lit(options)}, '{name}', '{label}', "
        f"{lit(field_config())}, false, false, '[]'::jsonb, '{{}}'::jsonb);"
    )


def chart_cfg(module_id, filter_ql, dim_field, metric_field, metric_type, label, aggregate="MAX",
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
                "isHidden": metric_type == "doughnut",
                "isScrollable": True,
                "orientation": "horizontal",
                "position": {"isDefault": True},
            },
            "offset": {
                "bottom": "20",
                "isDefault": False,
                "left": "140" if horizontal else "40",
                "right": "20",
                "top": "40",
            },
            "renderer": {},
        }],
        "colorScheme": "lowcode",
        "toolbox": {"saveAsImage": False, "showDataTable": False, "timeline": ""},
    }


def metric_item(label, field, suffix, number_format, color="black"):
    return {
        "label": label,
        "filter": "slice_kind = 'kpi'",
        "prefix": "",
        "suffix": suffix,
        "moduleID": str(MOD_SLICE),
        "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
        "operation": "sum",
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


def main():
    env = load_env(ENV_PATH)
    dsn = dsn_local(env["DB_DSN"])

    fact_exists = psql_at(dsn, "SELECT to_regclass('public.stock_reorder_fact') IS NOT NULL;")
    if fact_exists != "t":
        sys.stderr.write(
            "stock_reorder_fact missing — run seed_stock_reorder.py first\n"
        )
        raise SystemExit(1)

    print("Building risk_slice…")
    psql(dsn, open(SQL_PATH).read())

    print("Replacing Compose metadata…")
    psql(dsn, f"""
DELETE FROM compose_page_layout WHERE id = {LAYOUT} OR (page_id = {PAGE} AND handle = 'primary');
DELETE FROM compose_page WHERE id = {PAGE} OR handle = 'risks';
DELETE FROM compose_chart WHERE id IN ({CH_CRIT},{CH_TYPE},{CH_STORE_INC},{CH_STORE_RISK},{CH_STOCK},{CH_PRODUCT})
  OR handle IN (
    'risk-incident-criticality','risk-incident-type','risk-incident-store',
    'risk-store-score','risk-stock-health','risk-product-top'
  );
DELETE FROM compose_module_field WHERE rel_module = {MOD_SLICE};
DELETE FROM compose_module WHERE id = {MOD_SLICE} OR handle = 'Risk_slice';
""")

    psql(dsn, (
        "INSERT INTO compose_module (id, rel_namespace, handle, name, meta, config) "
        f"VALUES ({MOD_SLICE}, {NS}, 'Risk_slice', 'risk_slice', "
        f"{lit({'ui': {'admin': {'fields': []}}})}, {lit(dal_cfg('risk_slice'))});"
    ))

    fields = [
        (0, "String", "slice_kind", "Срез"),
        (1, "String", "slice_key", "Ключ"),
        (2, "String", "slice_label", "Название"),
        (3, "Number", "risk_score", "Риск-балл"),
        (4, "Number", "incident_count", "Инцидентов"),
        (5, "Number", "open_incident_count", "Открытых инцидентов"),
        (6, "Number", "critical_incident_count", "Критических инцидентов"),
        (7, "Number", "escalated_count", "Эскалаций"),
        (8, "Number", "stock_critical_count", "Критических остатков"),
        (9, "Number", "stock_understock_count", "Недосток SKU"),
        (10, "Number", "stock_risk_count", "Рисковых SKU"),
        (11, "Number", "store_at_risk_count", "Магазинов в риске"),
        (12, "Number", "sku_count", "SKU"),
        (13, "Number", "days_of_cover", "DOC, дн."),
        (14, "Number", "reorder_qty", "К заказу"),
        (15, "Number", "order_value", "Сумма заказа"),
    ]
    fields_sql = []
    slice_fields = []
    for i, (place, kind, name, label) in enumerate(fields):
        opts = str_opts() if kind == "String" else num_opts("" if name != "order_value" else " ₽", 0 if name != "days_of_cover" else 1)
        if name == "risk_score":
            opts = num_opts("", 1)
        fid = FID0 + i
        fields_sql.append(field_sql(fid, MOD_SLICE, place, kind, name, label, opts))
        slice_fields.append((fid, name, kind, label, opts))
    psql(dsn, "\n".join(fields_sql))

    charts = [
        (CH_CRIT, "risk-incident-criticality", "Инциденты по важности",
         chart_cfg(MOD_SLICE, "slice_kind = 'incident_criticality'", "slice_label", "incident_count",
                   "doughnut", "Инциденты", aggregate="SUM", fmt="0.")),
        (CH_TYPE, "risk-incident-type", "Инциденты по типу",
         chart_cfg(MOD_SLICE, "slice_kind = 'incident_type'", "slice_label", "open_incident_count",
                   "bar", "Открытые", aggregate="SUM", horizontal=True, fmt="0.")),
        (CH_STORE_INC, "risk-incident-store", "Открытые инциденты по магазинам",
         chart_cfg(MOD_SLICE, "slice_kind = 'store_incident'", "slice_label", "open_incident_count",
                   "bar", "Открытые", aggregate="SUM", horizontal=True, fmt="0.")),
        (CH_STORE_RISK, "risk-store-score", "Риск по магазинам",
         chart_cfg(MOD_SLICE, "slice_kind = 'store'", "slice_label", "risk_score",
                   "bar", "Риск-балл", aggregate="MAX", horizontal=True, fmt="0.0")),
        (CH_STOCK, "risk-stock-health", "Риски остатков",
         chart_cfg(MOD_SLICE, "slice_kind = 'stock_health'", "slice_label", "sku_count",
                   "doughnut", "SKU", aggregate="SUM", fmt="0.")),
        (CH_PRODUCT, "risk-product-top", "Товары в риске (DOC)",
         chart_cfg(MOD_SLICE, "slice_kind = 'product'", "slice_label", "days_of_cover",
                   "bar", "DOC, дн.", aggregate="MIN", horizontal=True, fmt="0.0")),
    ]
    ch_sql = []
    for cid, handle, name, cfg in charts:
        cfg_txt = json.dumps(cfg, ensure_ascii=False).replace("'", "''")
        ch_sql.append(
            "INSERT INTO compose_chart (id, handle, rel_namespace, name, config) "
            f"VALUES ({cid}, '{handle}', {NS}, '{name}', '{cfg_txt}');"
        )
    psql(dsn, "\n".join(ch_sql))

    inc_rl = load_module_fields(dsn, MOD_INC, [
        "dt", "store_id", "criticality", "incident_status", "incident_type", "body",
    ])
    fact_rl = load_module_fields(dsn, MOD_FACT, [
        "product_name", "store_id", "health_level", "days_of_cover",
        "stock_quantity", "reorder_qty", "order_value",
    ])
    store_rl = []
    for fid, name, kind, label, options in slice_fields:
        if name in (
            "slice_label", "risk_score", "open_incident_count", "critical_incident_count",
            "stock_critical_count", "stock_understock_count", "days_of_cover", "reorder_qty",
        ):
            store_rl.append(rl_field(fid, name, kind, label, options))

    blocks = [
        wrap_block("Metric", "Сводка рисков", 1, [0, 0, 48, 14], {
            "metrics": [
                metric_item("Открытые инциденты", "open_incident_count", "", "0.", "#e74a3b"),
                metric_item("Крит. инциденты", "critical_incident_count", "", "0.", "#e74a3b"),
                metric_item("Магазинов в риске", "store_at_risk_count", "", "0.", "#f6c23e"),
                metric_item("Рисковых SKU", "stock_risk_count", "", "0.", "#f6c23e"),
                metric_item("Крит. остатки", "stock_critical_count", "", "0.", "#e74a3b"),
            ],
            "refreshRate": 0,
            "showRefresh": False,
            "magnifyOption": "",
            "likeRecordList": True,
            "recordFieldLayoutOption": "default",
            "horizontalFieldLayoutEnabled": True,
        }),
        wrap_block("Chart", "Инциденты по важности", 2, [0, 14, 24, 32], {
            "chartID": str(CH_CRIT),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0,
            "showRefresh": False,
            "magnifyOption": "modal",
            "liveFilterEnabled": False,
        }),
        wrap_block("Chart", "Риски остатков", 3, [24, 14, 24, 32], {
            "chartID": str(CH_STOCK),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0,
            "showRefresh": False,
            "magnifyOption": "modal",
            "liveFilterEnabled": False,
        }),
        wrap_block("Chart", "Открытые инциденты по магазинам", 4, [0, 46, 24, 34], {
            "chartID": str(CH_STORE_INC),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0,
            "showRefresh": False,
            "magnifyOption": "modal",
            "liveFilterEnabled": False,
        }),
        wrap_block("Chart", "Риск по магазинам (остатки + инциденты)", 5, [24, 46, 24, 34], {
            "chartID": str(CH_STORE_RISK),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0,
            "showRefresh": False,
            "magnifyOption": "modal",
            "liveFilterEnabled": False,
        }),
        wrap_block("Chart", "Инциденты по типу", 6, [0, 80, 24, 34], {
            "chartID": str(CH_TYPE),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0,
            "showRefresh": False,
            "magnifyOption": "modal",
            "liveFilterEnabled": False,
        }),
        wrap_block("Chart", "Товары в риске", 7, [24, 80, 24, 34], {
            "chartID": str(CH_PRODUCT),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0,
            "showRefresh": False,
            "magnifyOption": "modal",
            "liveFilterEnabled": False,
        }),
        wrap_block("RecordList", "Магазины: рейтинг риска", 8, [0, 114, 48, 28], {
            "moduleID": str(MOD_SLICE),
            "fields": store_rl,
            "prefilter": "slice_kind = 'store'",
            "presort": "risk_score DESC",
            "perPage": 20,
            "hideAddButton": True,
            "hideImportButton": True,
            "hideHeader": False,
            "hidePaging": False,
            "hideSearch": False,
            "hideSorting": False,
            "hideFiltering": False,
            "editable": False,
            "enableRecordPageNavigation": False,
            "fullPageNavigation": False,
            "inlineRecordEdit": False,
            "recordSelectorDisplay": "",
            "magnifyOption": "modal",
            "liveFilterEnabled": False,
        }),
        wrap_block("RecordList", "Открытые и эскалированные инциденты", 9, [0, 142, 48, 32], {
            "moduleID": str(MOD_INC),
            "fields": inc_rl,
            "prefilter": "incident_status = 'Open' OR incident_status = 'In Progress' OR incident_status = 'Escalated'",
            "presort": "dt DESC",
            "perPage": 25,
            "hideAddButton": True,
            "hideImportButton": True,
            "hideHeader": False,
            "hidePaging": False,
            "hideSearch": False,
            "hideSorting": False,
            "hideFiltering": False,
            "editable": False,
            "enableRecordPageNavigation": False,
            "fullPageNavigation": False,
            "inlineRecordEdit": False,
            "recordSelectorDisplay": "",
            "magnifyOption": "modal",
            "liveFilterEnabled": False,
        }),
        wrap_block("RecordList", "Товары / остатки в риске", 10, [0, 174, 48, 32], {
            "moduleID": str(MOD_FACT),
            "fields": fact_rl,
            "prefilter": "health_level = 'understock' OR health_level = 'critical'",
            "presort": "days_of_cover",
            "perPage": 25,
            "hideAddButton": True,
            "hideImportButton": True,
            "hideHeader": False,
            "hidePaging": False,
            "hideSearch": False,
            "hideSorting": False,
            "hideFiltering": False,
            "editable": False,
            "enableRecordPageNavigation": False,
            "fullPageNavigation": False,
            "inlineRecordEdit": False,
            "recordSelectorDisplay": "",
            "magnifyOption": "modal",
            "liveFilterEnabled": False,
        }),
    ]

    page_cfg = {
        "prompt": "Ты риск-аналитик розницы. По этой странице разбери риски сети: "
                  "инциденты (открытые/эскалации/критичность/типы), остатки "
                  "(critical/understock, DOC), сводный риск-балл магазинов и товары в дефиците. "
                  "Предлагай приоритеты действий по магазинам и SKU."
    }
    page_meta = {"notifications": {"enabled": True}, "allowPersonalLayouts": False}
    psql(dsn, f"""
INSERT INTO compose_page (id, title, handle, self_id, rel_module, rel_namespace, meta, config, blocks, visible, weight, description)
VALUES ({PAGE}, 'Риски', 'risks', 0, 0, {NS},
        {lit(page_meta)}, {lit(page_cfg)}, {lit(blocks)},
        true, 4, 'Сводка рисков: магазины, товары/остатки, инциденты');
""")

    layout_blocks = [
        {
            "blockID": str(b["blockID"]),
            "xywh": b["xywh"],
            "meta": {"hidden": False, "visibility": {"roles": [], "expression": ""}},
        }
        for b in blocks
    ]
    layout_meta = {"title": "Риски", "description": ""}
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
        "SELECT open_incident_count||'/'||store_at_risk_count||'/'||stock_risk_count "
        "FROM risk_slice WHERE slice_kind='kpi' LIMIT 1;",
    )
    print("OK")
    print("  module Risk_slice → risk_slice")
    print("  charts risk-incident-*, risk-store-score, risk-stock-health, risk-product-top")
    print("  page /ns/loop/pages/%s (Риски)" % PAGE)
    print("  kpi open_incidents/stores_at_risk/stock_risk_sku = %s" % kpi)
    print("  restart GoLand Server_RU_test_translations for DalModelReload")


if __name__ == "__main__":
    main()
