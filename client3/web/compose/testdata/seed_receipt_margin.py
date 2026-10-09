#!/usr/bin/env python3
"""Seed Compose modules, charts and the Маржинальность page on top of receipt_margin MVs."""
import json
import os
import subprocess
import sys
import urllib.parse

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "../../../.."))
ENV_PATH = os.path.join(ROOT, ".env")
SQL_PATH = os.path.join(os.path.dirname(__file__), "seed_receipt_margin.sql")

NS = "495727984893558785"
CONN = "495279171887497217"

MOD_FACT = 510010000000100001
MOD_SLICE = 510010000000100002
MOD_KPI = 510010000000100003
CH_CAT = 510010000000120001
CH_BRAND = 510010000000120002
CH_MONTH = 510010000000120003
CH_PROFIT = 510010000000120004
CH_BRIDGE = 510010000000120005
PAGE = 510010000000130001
LAYOUT = 510010000000130002
FID0 = 510010000000110001


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
              horizontal=False, rotate=0, suffix="", fmt="0.0", time_labels=False, dim_mod="(no grouping / buckets)"):
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
                "isHidden": True,
                "isScrollable": True,
                "orientation": "horizontal",
                "position": {"isDefault": True},
            },
            "offset": {
                "bottom": "20" if not horizontal else "20",
                "isDefault": False,
                "left": "12",
                "right": "20",
                "top": "40",
            },
            "renderer": {},
        }],
        "colorScheme": "lowcode",
        "toolbox": {"saveAsImage": False, "showDataTable": True, "timeline": ""},
    }


def metric_item(label, module_id, field, suffix, number_format, color="black"):
    return {
        "label": label,
        "filter": "",
        "prefix": "",
        "suffix": suffix,
        "moduleID": str(module_id),
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

    print("Refreshing materialized views…")
    psql(dsn, open(SQL_PATH).read())

    print("Replacing Compose metadata…")
    psql(dsn, f"""
DELETE FROM compose_page_layout WHERE id = {LAYOUT} OR (page_id = {PAGE} AND handle = 'primary');
DELETE FROM compose_page WHERE id = {PAGE} OR handle = 'margin';
DELETE FROM compose_chart WHERE id IN ({CH_CAT},{CH_BRAND},{CH_MONTH},{CH_PROFIT},{CH_BRIDGE})
  OR handle IN ('margin-by-category','margin-by-brand','margin-by-month','margin-profit-category','margin-bridge');
DELETE FROM compose_module_field WHERE rel_module IN ({MOD_FACT},{MOD_SLICE},{MOD_KPI});
DELETE FROM compose_module WHERE id IN ({MOD_FACT},{MOD_SLICE},{MOD_KPI})
  OR handle IN ('Receipt_margin','Receipt_margin_slice','Receipt_margin_kpi');
""")

    modules = [
        (MOD_FACT, "Receipt_margin", "receipt_margin", "receipt_margin"),
        (MOD_SLICE, "Receipt_margin_slice", "receipt_margin_slice", "receipt_margin_slice"),
        (MOD_KPI, "Receipt_margin_kpi", "receipt_margin_kpi", "receipt_margin_kpi"),
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
    slice_fields = []

    def add(module_id, place, kind, name, label, options, collect=None):
        nonlocal fid
        fields_sql.append(field_sql(fid, module_id, place, kind, name, label, options))
        rec = (fid, name, kind, label, options)
        if collect is not None:
            collect.append(rec)
        fid += 1
        return rec

    # fact
    add(MOD_FACT, 0, "String", "store_name", "Магазин", str_opts())
    add(MOD_FACT, 1, "String", "store_type", "Тип", str_opts())
    add(MOD_FACT, 2, "String", "product_name", "Товар", str_opts())
    add(MOD_FACT, 3, "String", "brand", "Бренд", str_opts())
    add(MOD_FACT, 4, "String", "category_name", "Категория", str_opts())
    add(MOD_FACT, 5, "DateTime", "dt", "Дата", dt_opts())
    add(MOD_FACT, 6, "Number", "quantity", "Кол-во", num_opts("", 3))
    add(MOD_FACT, 7, "Number", "revenue", "Выручка", num_opts(" ₽"))
    add(MOD_FACT, 8, "Number", "cogs", "Себестоимость", num_opts(" ₽"))
    add(MOD_FACT, 9, "Number", "gross_profit", "Валовая прибыль", num_opts(" ₽"))
    add(MOD_FACT, 10, "Number", "margin_pct", "Маржа %", num_opts(" %"))
    add(MOD_FACT, 11, "Number", "unit_cost", "Себест. ед.", num_opts(" ₽"))
    add(MOD_FACT, 12, "Number", "vat", "НДС", num_opts(" ₽"))
    add(MOD_FACT, 13, "Number", "discount", "Скидка", num_opts(" ₽"))
    add(MOD_FACT, 14, "Number", "store_id", "store id", num_opts("", 0))
    add(MOD_FACT, 15, "Number", "product_id", "product id", num_opts("", 0))
    add(MOD_FACT, 16, "Number", "category_id", "category id", num_opts("", 0))

    add(MOD_SLICE, 0, "String", "slice_kind", "Срез", str_opts(), slice_fields)
    add(MOD_SLICE, 1, "String", "slice_name", "Имя", str_opts(), slice_fields)
    add(MOD_SLICE, 2, "Number", "slice_order", "Порядок", num_opts("", 0), slice_fields)
    add(MOD_SLICE, 3, "Number", "revenue", "Выручка", num_opts(" ₽"), slice_fields)
    add(MOD_SLICE, 4, "Number", "cogs", "Себестоимость", num_opts(" ₽"), slice_fields)
    add(MOD_SLICE, 5, "Number", "gross_profit", "Валовая прибыль", num_opts(" ₽"), slice_fields)
    add(MOD_SLICE, 6, "Number", "margin_pct", "Маржа %", num_opts(" %"), slice_fields)
    add(MOD_SLICE, 7, "Number", "line_count", "Строк", num_opts("", 0), slice_fields)

    add(MOD_KPI, 0, "Number", "revenue", "Выручка", num_opts(" ₽"))
    add(MOD_KPI, 1, "Number", "cogs", "Себестоимость", num_opts(" ₽"))
    add(MOD_KPI, 2, "Number", "gross_profit", "Валовая прибыль", num_opts(" ₽"))
    add(MOD_KPI, 3, "Number", "margin_pct", "Маржа %", num_opts(" %"))
    add(MOD_KPI, 4, "Number", "vat", "НДС", num_opts(" ₽"))
    add(MOD_KPI, 5, "Number", "discount", "Скидка", num_opts(" ₽"))
    add(MOD_KPI, 6, "Number", "line_count", "Строк", num_opts("", 0))
    add(MOD_KPI, 7, "Number", "store_count", "Магазинов", num_opts("", 0))
    add(MOD_KPI, 8, "Number", "sku_count", "SKU", num_opts("", 0))

    psql(dsn, "\n".join(fields_sql))

    charts = [
        (CH_CAT, "margin-by-category", "Маржа по категориям",
         chart_cfg(MOD_SLICE, "slice_kind = 'category'", "slice_name", "margin_pct", "bar",
                   "Маржа %", horizontal=True, suffix=" %")),
        (CH_BRAND, "margin-by-brand", "Маржа по брендам",
         chart_cfg(MOD_SLICE, "slice_kind = 'brand_top'", "slice_name", "margin_pct", "bar",
                   "Маржа %", horizontal=True, suffix=" %")),
        (CH_MONTH, "margin-by-month", "Маржа по месяцам",
         chart_cfg(MOD_SLICE, "slice_kind = 'month'", "slice_name", "margin_pct", "line",
                   "Маржа %", suffix=" %")),
        (CH_PROFIT, "margin-profit-category", "Прибыль по категориям",
         chart_cfg(MOD_SLICE, "slice_kind = 'category'", "slice_name", "gross_profit", "bar",
                   "Валовая прибыль", aggregate="SUM", horizontal=True, suffix=" ₽", fmt="0,0")),
        (CH_BRIDGE, "margin-bridge", "Выручка → себестоимость → прибыль",
         chart_cfg(MOD_SLICE, "slice_kind = 'bridge'", "slice_name", "revenue", "bar",
                   "Сумма", aggregate="SUM", suffix=" ₽", fmt="0,0")),
    ]
    ch_sql = []
    for cid, handle, name, cfg in charts:
        cfg_txt = json.dumps(cfg, ensure_ascii=False).replace("'", "''")
        ch_sql.append(
            "INSERT INTO compose_chart (id, handle, rel_namespace, name, config) "
            f"VALUES ({cid}, '{handle}', {NS}, '{name}', '{cfg_txt}');"
        )
    psql(dsn, "\n".join(ch_sql))

    by_name = {f[1]: f for f in slice_fields}
    rl_fields = []
    for key in ("slice_name", "revenue", "cogs", "gross_profit", "margin_pct"):
        fid_, name, kind, label, options = by_name[key]
        rl_fields.append(rl_field(fid_, name, kind, label, options))

    blocks = [
        wrap_block("Metric", "Валовая маржа", 1, [0, 0, 48, 14], {
            "metrics": [
                # "0." (not "0,0"): Metric Item runs fmt.number/Intl after numeral.
                # numeral "0,0" inserts grouping commas → Intl ru-RU shows "не число".
                # Working loop KPIs (Сводка) use "0." so Intl can add locale grouping.
                metric_item("Выручка", MOD_KPI, "revenue", " ₽", "0."),
                metric_item("Себестоимость", MOD_KPI, "cogs", " ₽", "0."),
                metric_item("Валовая прибыль", MOD_KPI, "gross_profit", " ₽", "0.", "black"),
                metric_item("Маржа", MOD_KPI, "margin_pct", " %", "0.0"),
            ],
            "refreshRate": 0,
            "showRefresh": False,
            "magnifyOption": "",
            "likeRecordList": True,
            "recordFieldLayoutOption": "default",
            "horizontalFieldLayoutEnabled": True,
        }),
        wrap_block("Chart", "Маржа по категориям", 2, [0, 14, 24, 34], {
            "chartID": str(CH_CAT),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0,
            "showRefresh": False,
            "magnifyOption": "modal",
            "liveFilterEnabled": False,
        }),
        wrap_block("Chart", "Маржа по брендам", 3, [24, 14, 24, 34], {
            "chartID": str(CH_BRAND),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0,
            "showRefresh": False,
            "magnifyOption": "modal",
            "liveFilterEnabled": False,
        }),
        wrap_block("Chart", "Маржа по месяцам", 4, [0, 48, 24, 28], {
            "chartID": str(CH_MONTH),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0,
            "showRefresh": False,
            "magnifyOption": "modal",
            "liveFilterEnabled": False,
        }),
        wrap_block("Chart", "Выручка → себестоимость → прибыль", 5, [24, 48, 24, 28], {
            "chartID": str(CH_BRIDGE),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0,
            "showRefresh": False,
            "magnifyOption": "modal",
            "liveFilterEnabled": False,
        }),
        wrap_block("Chart", "Прибыль по категориям", 6, [0, 76, 48, 28], {
            "chartID": str(CH_PROFIT),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0,
            "showRefresh": False,
            "magnifyOption": "modal",
            "liveFilterEnabled": False,
        }),
        wrap_block("RecordList", "SKU с наименьшей маржой", 7, [0, 104, 48, 36], {
            "moduleID": str(MOD_SLICE),
            "fields": rl_fields,
            "prefilter": "slice_kind = 'sku'",
            "presort": "margin_pct",
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
        "prompt": "Ты финансовый аналитик. По этой странице разбери валовую маржу сети: "
                  "выручка, себестоимость, прибыль, срезы категория/бренд/месяц. "
                  "Дай рекомендации по SKU с низкой маржой."
    }
    page_meta = {"notifications": {"enabled": True}, "allowPersonalLayouts": False}
    psql(dsn, f"""
INSERT INTO compose_page (id, title, handle, self_id, rel_module, rel_namespace, meta, config, blocks, visible, weight, description)
VALUES ({PAGE}, 'Маржинальность', 'margin', 0, 0, {NS},
        {lit(page_meta)}, {lit(page_cfg)}, {lit(blocks)},
        true, 2, 'Валовая маржа: чеки × as-of себестоимость');
""")

    layout_blocks = [
        {
            "blockID": str(b["blockID"]),
            "xywh": b["xywh"],
            "meta": {"hidden": False, "visibility": {"roles": [], "expression": ""}},
        }
        for b in blocks
    ]
    layout_meta = {"title": "Маржинальность", "description": ""}
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

    # Pending attributeReType/modelAdd skip DalModelReload cache; tables are
    # already BASE TABLE with typed columns, so leftover alterations must go.
    psql(dsn, f"""
DELETE FROM dal_schema_alterations
 WHERE resource LIKE '%/51001000000010000%';
""")

    print("Seeded:")
    print("  modules Receipt_margin / Receipt_margin_slice / Receipt_margin_kpi")
    print("  charts margin-by-category, margin-by-brand, margin-by-month, margin-profit-category, margin-bridge")
    print("  page /ns/loop/pages/%s (Маржинальность)" % PAGE)


if __name__ == "__main__":
    main()
