#!/usr/bin/env python3
"""Seed Compose modules, charts and the Остатки и автозаказ page on top of stock_reorder tables."""
import json
import os
import subprocess
import sys

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "../../../.."))
ENV_PATH = os.path.join(ROOT, ".env")
SQL_PATH = os.path.join(os.path.dirname(__file__), "seed_stock_reorder.sql")

NS = "495727984893558785"
CONN = "495279171887497217"

MOD_POLICY = 510020000000100001
MOD_FACT = 510020000000100002
MOD_SLICE = 510020000000100003
MOD_PO = 510020000000100004
MOD_POLINE = 510020000000100005

CH_HEALTH = 510020000000120001
CH_STORE = 510020000000120002
CH_CAT = 510020000000120003
CH_DOC = 510020000000120004

PAGE = 510020000000130001
LAYOUT = 510020000000130002
FID0 = 510020000000110001


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


def dt_opts():
    return {
        "hint": {"view": ""},
        "format": "",
        "prefix": "",
        "suffix": "",
        "onlyDate": False,
        "onlyTime": False,
        "description": {"view": ""},
        "multiDelimiter": "\n",
        "onlyPastValues": False,
        "outputRelative": False,
        "onlyFutureValues": False,
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
              horizontal=False, rotate=0, suffix="", fmt="0.0", time_labels=False,
              dim_mod="(no grouping / buckets)"):
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
                "modifier": dim_mod,
                "rotateLabel": rotate,
                "timeLabels": time_labels,
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
                "left": "12",
                "right": "20",
                "top": "40",
            },
            "renderer": {},
        }],
        "colorScheme": "lowcode",
        "toolbox": {"saveAsImage": False, "showDataTable": False, "timeline": ""},
    }


def metric_item(label, module_id, field, suffix, number_format, color="black", operation="sum"):
    return {
        "label": label,
        "filter": "slice_kind = 'kpi'" if module_id == MOD_SLICE else "",
        "prefix": "",
        "suffix": suffix,
        "moduleID": str(module_id),
        "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
        "operation": operation,
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
        "options": options,
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


def main():
    env = load_env(ENV_PATH)
    dsn = dsn_local(env["DB_DSN"])

    print("Building stock reorder tables…")
    psql(dsn, open(SQL_PATH).read())

    print("Replacing Compose metadata…")
    psql(dsn, f"""
DELETE FROM compose_page_layout WHERE id = {LAYOUT} OR (page_id = {PAGE} AND handle = 'primary');
DELETE FROM compose_page WHERE id = {PAGE} OR handle = 'stock-reorder';
DELETE FROM compose_chart WHERE id IN ({CH_HEALTH},{CH_STORE},{CH_CAT},{CH_DOC})
  OR handle IN ('stock-health-dist','stock-understock-store','stock-order-category','stock-doc-store');
DELETE FROM compose_module_field WHERE rel_module IN ({MOD_POLICY},{MOD_FACT},{MOD_SLICE},{MOD_PO},{MOD_POLINE});
DELETE FROM compose_module WHERE id IN ({MOD_POLICY},{MOD_FACT},{MOD_SLICE},{MOD_PO},{MOD_POLINE})
  OR handle IN ('Stock_policy','Stock_reorder_fact','Stock_health_slice','Purchase_order','Purchase_order_line');
""")

    modules = [
        (MOD_POLICY, "Stock_policy", "stock_policy", "stock_policy"),
        (MOD_FACT, "Stock_reorder_fact", "stock_reorder_fact", "stock_reorder_fact"),
        (MOD_SLICE, "Stock_health_slice", "stock_health_slice", "stock_health_slice"),
        (MOD_PO, "Purchase_order", "purchase_order", "purchase_order"),
        (MOD_POLINE, "Purchase_order_line", "purchase_order_line", "purchase_order_line"),
    ]
    sql = []
    for mid, handle, name, ident in modules:
        sql.append(
            "INSERT INTO compose_module (id, rel_namespace, handle, name, meta, config) "
            f"VALUES ({mid}, {NS}, '{handle}', '{name}', "
            f"{lit({'ui': {'admin': {'fields': []}}})}, {lit(dal_cfg(ident))});"
        )
    psql(dsn, "\n".join(sql))

    fid = FID0
    fields_sql = []
    fact_fields = []
    slice_fields = []
    po_fields = []

    def add(module_id, place, kind, name, label, options, collect=None):
        nonlocal fid
        fields_sql.append(field_sql(fid, module_id, place, kind, name, label, options))
        rec = (fid, name, kind, label, options)
        if collect is not None:
            collect.append(rec)
        fid += 1
        return rec

    # stock_policy
    add(MOD_POLICY, 0, "Number", "policy_id", "Policy ID", num_opts("", 0))
    add(MOD_POLICY, 1, "Number", "store_id", "Магазин", num_opts("", 0))
    add(MOD_POLICY, 2, "Number", "product_id", "Товар", num_opts("", 0))
    add(MOD_POLICY, 3, "Number", "category_id", "Категория", num_opts("", 0))
    add(MOD_POLICY, 4, "Number", "lead_time_days", "Срок поставки, дн.", num_opts("", 0))
    add(MOD_POLICY, 5, "Number", "review_period_days", "Период обзора, дн.", num_opts("", 0))
    add(MOD_POLICY, 6, "Number", "target_doc_days", "Целевой DOC, дн.", num_opts("", 0))
    add(MOD_POLICY, 7, "Number", "moq", "MOQ", num_opts("", 0))
    add(MOD_POLICY, 8, "Number", "pack_size", "Кратность упаковки", num_opts("", 0))
    add(MOD_POLICY, 9, "Number", "service_z", "Z сервиса", num_opts("", 3))
    add(MOD_POLICY, 10, "Number", "min_qty", "Мин. кол-во", num_opts("", 0))
    add(MOD_POLICY, 11, "Number", "max_qty", "Макс. кол-во", num_opts("", 0))

    # stock_reorder_fact
    add(MOD_FACT, 0, "Number", "store_id", "Магазин ID", num_opts("", 0), fact_fields)
    add(MOD_FACT, 1, "Number", "product_id", "Товар ID", num_opts("", 0), fact_fields)
    add(MOD_FACT, 2, "String", "ean", "EAN/SKU", str_opts(), fact_fields)
    add(MOD_FACT, 3, "String", "product_name", "Товар", str_opts(), fact_fields)
    add(MOD_FACT, 4, "Number", "category_id", "Категория", num_opts("", 0), fact_fields)
    add(MOD_FACT, 5, "Number", "supplier_id", "Поставщик", num_opts("", 0), fact_fields)
    add(MOD_FACT, 6, "Number", "stock_quantity", "Остаток", num_opts("", 3), fact_fields)
    add(MOD_FACT, 7, "Number", "stock_sum", "Сумма остатка", num_opts(" ₽"), fact_fields)
    add(MOD_FACT, 8, "Number", "avg_daily_qty", "Ср. дневной спрос", num_opts("", 3), fact_fields)
    add(MOD_FACT, 9, "Number", "demand_std_daily", "σ спроса", num_opts("", 3), fact_fields)
    add(MOD_FACT, 10, "Number", "days_of_cover", "DOC, дн.", num_opts("", 1), fact_fields)
    add(MOD_FACT, 11, "Number", "lead_time_days", "Lead time", num_opts("", 0), fact_fields)
    add(MOD_FACT, 12, "Number", "review_period_days", "Review", num_opts("", 0), fact_fields)
    add(MOD_FACT, 13, "Number", "target_doc_days", "Целевой DOC", num_opts("", 0), fact_fields)
    add(MOD_FACT, 14, "Number", "moq", "MOQ", num_opts("", 0), fact_fields)
    add(MOD_FACT, 15, "Number", "pack_size", "Упаковка", num_opts("", 0), fact_fields)
    add(MOD_FACT, 16, "Number", "service_z", "Z", num_opts("", 3), fact_fields)
    add(MOD_FACT, 17, "Number", "unit_cost", "Себест. ед.", num_opts(" ₽"), fact_fields)
    add(MOD_FACT, 18, "Number", "safety_stock", "Страх. запас", num_opts("", 3), fact_fields)
    add(MOD_FACT, 19, "Number", "reorder_point", "Точка заказа", num_opts("", 3), fact_fields)
    add(MOD_FACT, 20, "Number", "target_qty", "Целевой остаток", num_opts("", 3), fact_fields)
    add(MOD_FACT, 21, "Number", "reorder_qty", "К заказу", num_opts("", 0), fact_fields)
    add(MOD_FACT, 22, "Number", "order_value", "Сумма заказа", num_opts(" ₽"), fact_fields)
    add(MOD_FACT, 23, "String", "health_level", "Здоровье", str_opts(), fact_fields)
    add(MOD_FACT, 24, "DateTime", "stock_date", "Дата остатка", dt_opts(), fact_fields)

    # stock_health_slice
    add(MOD_SLICE, 0, "String", "slice_kind", "Срез", str_opts(), slice_fields)
    add(MOD_SLICE, 1, "String", "slice_key", "Ключ", str_opts(), slice_fields)
    add(MOD_SLICE, 2, "String", "slice_label", "Название", str_opts(), slice_fields)
    add(MOD_SLICE, 3, "Number", "sku_count", "SKU", num_opts("", 0), slice_fields)
    add(MOD_SLICE, 4, "Number", "stock_quantity", "Остаток", num_opts("", 3), slice_fields)
    add(MOD_SLICE, 5, "Number", "stock_sum", "Сумма остатка", num_opts(" ₽"), slice_fields)
    add(MOD_SLICE, 6, "Number", "avg_daily_qty", "Ср. спрос", num_opts("", 3), slice_fields)
    add(MOD_SLICE, 7, "Number", "days_of_cover", "DOC, дн.", num_opts("", 1), slice_fields)
    add(MOD_SLICE, 8, "Number", "reorder_qty", "К заказу", num_opts("", 0), slice_fields)
    add(MOD_SLICE, 9, "Number", "order_value", "Сумма заказа", num_opts(" ₽"), slice_fields)
    add(MOD_SLICE, 10, "Number", "critical_count", "Критических", num_opts("", 0), slice_fields)
    add(MOD_SLICE, 11, "Number", "understock_count", "Недосток", num_opts("", 0), slice_fields)

    # purchase_order
    add(MOD_PO, 0, "String", "po_number", "Номер ЗП", str_opts(), po_fields)
    add(MOD_PO, 1, "Number", "supplier_id", "Поставщик", num_opts("", 0), po_fields)
    add(MOD_PO, 2, "Number", "store_id", "Магазин", num_opts("", 0), po_fields)
    add(MOD_PO, 3, "String", "status", "Статус", str_opts(), po_fields)
    add(MOD_PO, 4, "DateTime", "order_date", "Дата заказа", dt_opts(), po_fields)
    add(MOD_PO, 5, "DateTime", "expected_date", "Ожидаемая дата", dt_opts(), po_fields)
    add(MOD_PO, 6, "Number", "total_qty", "Кол-во", num_opts("", 0), po_fields)
    add(MOD_PO, 7, "Number", "total_sum", "Сумма", num_opts(" ₽"), po_fields)
    add(MOD_PO, 8, "String", "source", "Источник", str_opts(), po_fields)
    add(MOD_PO, 9, "Number", "line_count", "Строк", num_opts("", 0), po_fields)

    # purchase_order_line
    add(MOD_POLINE, 0, "Number", "po_id", "ЗП ID", num_opts("", 0))
    add(MOD_POLINE, 1, "Number", "product_id", "Товар ID", num_opts("", 0))
    add(MOD_POLINE, 2, "String", "ean", "EAN/SKU", str_opts())
    add(MOD_POLINE, 3, "String", "product_name", "Товар", str_opts())
    add(MOD_POLINE, 4, "Number", "store_id", "Магазин", num_opts("", 0))
    add(MOD_POLINE, 5, "Number", "qty_ordered", "Заказано", num_opts("", 0))
    add(MOD_POLINE, 6, "Number", "qty_suggested", "Рекоменд.", num_opts("", 0))
    add(MOD_POLINE, 7, "Number", "unit_cost", "Себест. ед.", num_opts(" ₽"))
    add(MOD_POLINE, 8, "Number", "line_sum", "Сумма", num_opts(" ₽"))
    add(MOD_POLINE, 9, "Number", "reorder_point", "Точка заказа", num_opts("", 3))
    add(MOD_POLINE, 10, "Number", "days_of_cover", "DOC", num_opts("", 1))
    add(MOD_POLINE, 11, "String", "health_level", "Здоровье", str_opts())
    add(MOD_POLINE, 12, "Number", "rule_score", "Score", num_opts("", 2))

    psql(dsn, "\n".join(fields_sql))

    charts = [
        (CH_HEALTH, "stock-health-dist", "Распределение здоровья остатков",
         chart_cfg(MOD_SLICE, "slice_kind = 'health'", "slice_label", "sku_count", "doughnut",
                   "SKU", aggregate="SUM", suffix="", fmt="0.")),
        (CH_STORE, "stock-understock-store", "Недосток по магазинам",
         chart_cfg(MOD_SLICE, "slice_kind = 'store'", "slice_label", "understock_count", "bar",
                   "Недосток SKU", aggregate="SUM", horizontal=True, suffix="", fmt="0.")),
        (CH_CAT, "stock-order-category", "Сумма заказа по категориям",
         chart_cfg(MOD_SLICE, "slice_kind = 'category_order_top'", "slice_label", "order_value", "bar",
                   "Сумма заказа", aggregate="SUM", horizontal=True, suffix=" ₽", fmt="0.")),
        (CH_DOC, "stock-doc-store", "DOC по магазинам",
         chart_cfg(MOD_SLICE, "slice_kind = 'store'", "slice_label", "days_of_cover", "bar",
                   "DOC, дн.", aggregate="AVG", horizontal=True, suffix="", fmt="0.0")),
    ]
    ch_sql = []
    for cid, handle, name, cfg in charts:
        cfg_txt = json.dumps(cfg, ensure_ascii=False).replace("'", "''")
        ch_sql.append(
            "INSERT INTO compose_chart (id, handle, rel_namespace, name, config) "
            f"VALUES ({cid}, '{handle}', {NS}, '{name}', '{cfg_txt}');"
        )
    psql(dsn, "\n".join(ch_sql))

    by_fact = {f[1]: f for f in fact_fields}
    by_po = {f[1]: f for f in po_fields}

    fact_rl = []
    for key in ("product_name", "store_id", "health_level", "days_of_cover",
                "stock_quantity", "avg_daily_qty", "reorder_qty", "order_value"):
        fid_, name, kind, label, options = by_fact[key]
        fact_rl.append(rl_field(fid_, name, kind, label, options))

    po_rl = []
    for key in ("po_number", "store_id", "supplier_id", "status", "total_qty", "total_sum", "line_count", "order_date"):
        fid_, name, kind, label, options = by_po[key]
        po_rl.append(rl_field(fid_, name, kind, label, options))

    blocks = [
        wrap_block("RuleChain", "Автозаказ", 8, [0, 0, 48, 8], {
            "chainID": "stock_reorder_batch",
            "label": "Запустить автозаказ",
            "variant": "primary",
            "icon": "play",
            "size": "",
            "context": {},
        }),
        wrap_block("Metric", "Остатки и автозаказ", 1, [0, 8, 48, 14], {
            "metrics": [
                metric_item("Недосток SKU", MOD_SLICE, "understock_count", "", "0."),
                metric_item("Критических", MOD_SLICE, "critical_count", "", "0."),
                metric_item("К заказу", MOD_SLICE, "reorder_qty", "", "0."),
                metric_item("Сумма заказа", MOD_SLICE, "order_value", " ₽", "0."),
            ],
            "refreshRate": 0,
            "showRefresh": False,
            "magnifyOption": "",
            "likeRecordList": True,
            "recordFieldLayoutOption": "default",
            "horizontalFieldLayoutEnabled": True,
        }),
        wrap_block("Chart", "Здоровье остатков", 2, [0, 22, 24, 34], {
            "chartID": str(CH_HEALTH),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0,
            "showRefresh": False,
            "magnifyOption": "modal",
            "liveFilterEnabled": False,
        }),
        wrap_block("Chart", "Недосток по магазинам", 3, [24, 22, 24, 34], {
            "chartID": str(CH_STORE),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0,
            "showRefresh": False,
            "magnifyOption": "modal",
            "liveFilterEnabled": False,
        }),
        wrap_block("Chart", "Сумма заказа по категориям", 4, [0, 56, 24, 28], {
            "chartID": str(CH_CAT),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0,
            "showRefresh": False,
            "magnifyOption": "modal",
            "liveFilterEnabled": False,
        }),
        wrap_block("Chart", "DOC по магазинам", 5, [24, 56, 24, 28], {
            "chartID": str(CH_DOC),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0,
            "showRefresh": False,
            "magnifyOption": "modal",
            "liveFilterEnabled": False,
        }),
        wrap_block("RecordList", "Недосток и критические SKU", 6, [0, 84, 48, 36], {
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
        wrap_block("RecordList", "Заказы поставщикам (submitted)", 7, [0, 120, 48, 28], {
            "moduleID": str(MOD_PO),
            "fields": po_rl,
            "prefilter": "status = 'submitted'",
            "presort": "order_date DESC",
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
        "prompt": "Ты аналитик запасов. По этой странице разбери здоровье остатков сети: "
                  "DOC, critical/understock, автозаказ по магазинам и поставщикам. "
                  "Спрос считай только по чекам (receipt_positions), политика — из stock_policy."
    }
    page_meta = {"notifications": {"enabled": True}, "allowPersonalLayouts": False}
    psql(dsn, f"""
INSERT INTO compose_page (id, title, handle, self_id, rel_module, rel_namespace, meta, config, blocks, visible, weight, description)
VALUES ({PAGE}, 'Остатки и автозаказ', 'stock-reorder', 0, 0, {NS},
        {lit(page_meta)}, {lit(page_cfg)}, {lit(blocks)},
        true, 3, 'Здоровье остатков + автозаказ (receipt_positions × stock_policy)');
""")

    layout_blocks = [
        {
            "blockID": str(b["blockID"]),
            "xywh": b["xywh"],
            "meta": {"hidden": False, "visibility": {"roles": [], "expression": ""}},
        }
        for b in blocks
    ]
    layout_meta = {"title": "Остатки и автозаказ", "description": ""}
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

    psql(dsn, f"""
DELETE FROM dal_schema_alterations
 WHERE resource LIKE '%/51002000000010000%';
""")

    print("Seeded:")
    print("  modules Stock_policy / Stock_reorder_fact / Stock_health_slice / Purchase_order / Purchase_order_line")
    print("  charts stock-health-dist, stock-understock-store, stock-order-category, stock-doc-store")
    print("  page /ns/loop/pages/%s (Остатки и автозаказ)" % PAGE)


if __name__ == "__main__":
    main()
