#!/usr/bin/env python3
"""Seed HR modules, charts and the Персонал summary page."""
import json
import os
import subprocess
import sys

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "../../../.."))
ENV_PATH = os.path.join(ROOT, ".env")
SQL_PATH = os.path.join(os.path.dirname(__file__), "seed_hr_personnel.sql")

NS = "495727984893558785"
CONN = "495279171887497217"

MOD_SLICE = 510040000000100001
MOD_EMP = 510040000000100002
MOD_LEAVE = 510040000000100003
MOD_VAC = 510040000000100004

CH_DEPT = 510040000000120001
CH_STATUS = 510040000000120002
CH_LOC = 510040000000120003
CH_TENURE = 510040000000120004
CH_VAC_DEPT = 510040000000120005
CH_LEAVE = 510040000000120006

PAGE = 510040000000130001
LAYOUT = 510040000000130002
FID0 = 510040000000110001


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


def select_opts(options, badge=True):
    return {
        "hint": {"view": ""},
        "prefix": "",
        "suffix": "",
        "options": options,
        "selectType": "default",
        "description": {"view": ""},
        "displayType": "badge" if badge else "default",
        "badgeGradient": True,
        "multiDelimiter": "\n",
        "isUniqueMultiValue": False,
    }


def date_opts():
    return {
        "hint": {"view": ""},
        "format": "",
        "description": {"view": ""},
        "onlyFutureDates": False,
        "onlyPastDates": False,
        "multiDelimiter": "\n",
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
              horizontal=False, rotate=0, suffix="", fmt="0."):
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

    print("Building HR tables + hr_slice…")
    psql(dsn, open(SQL_PATH).read())

    print("Replacing Compose metadata…")
    psql(dsn, f"""
DELETE FROM compose_page_layout WHERE id = {LAYOUT} OR (page_id = {PAGE} AND handle = 'primary');
DELETE FROM compose_page WHERE id = {PAGE} OR handle = 'hr-personnel';
DELETE FROM compose_chart WHERE id IN ({CH_DEPT},{CH_STATUS},{CH_LOC},{CH_TENURE},{CH_VAC_DEPT},{CH_LEAVE})
  OR handle IN (
    'hr-headcount-dept','hr-headcount-status','hr-headcount-location',
    'hr-tenure','hr-vacancy-dept','hr-leave-type'
  );
DELETE FROM compose_module_field WHERE rel_module IN ({MOD_SLICE},{MOD_EMP},{MOD_LEAVE},{MOD_VAC});
DELETE FROM compose_module WHERE id IN ({MOD_SLICE},{MOD_EMP},{MOD_LEAVE},{MOD_VAC})
  OR handle IN ('Hr_slice','Hr_employee','Hr_leave','Hr_vacancy');
""")

    insert_module(dsn, MOD_SLICE, "Hr_slice", "hr_slice", "hr_slice")
    insert_module(dsn, MOD_EMP, "Hr_employee", "hr_employee", "hr_employee")
    insert_module(dsn, MOD_LEAVE, "Hr_leave", "hr_leave", "hr_leave")
    insert_module(dsn, MOD_VAC, "Hr_vacancy", "hr_vacancy", "hr_vacancy")

    status_opts = select_opts([
        {"text": "Работает", "value": "active", "style": {"textColor": "light", "backgroundColor": "success"}},
        {"text": "Испытательный", "value": "probation", "style": {"textColor": "dark", "backgroundColor": "warning"}},
        {"text": "В отпуске", "value": "on_leave", "style": {"textColor": "light", "backgroundColor": "info"}},
        {"text": "Уволен", "value": "terminated", "style": {"textColor": "light", "backgroundColor": "danger"}},
    ])
    emp_type_opts = select_opts([
        {"text": "Полная", "value": "full_time", "style": {"textColor": "light", "backgroundColor": "primary"}},
        {"text": "Частичная", "value": "part_time", "style": {"textColor": "dark", "backgroundColor": "light"}},
        {"text": "Договор", "value": "contract", "style": {"textColor": "light", "backgroundColor": "secondary"}},
        {"text": "Сменный", "value": "shift", "style": {"textColor": "dark", "backgroundColor": "warning"}},
    ], badge=True)
    leave_type_opts = select_opts([
        {"text": "Отпуск", "value": "vacation", "style": {"textColor": "light", "backgroundColor": "primary"}},
        {"text": "Больничный", "value": "sick", "style": {"textColor": "light", "backgroundColor": "danger"}},
        {"text": "Без содержания", "value": "unpaid", "style": {"textColor": "dark", "backgroundColor": "light"}},
        {"text": "Декрет", "value": "parental", "style": {"textColor": "light", "backgroundColor": "info"}},
    ])
    leave_status_opts = select_opts([
        {"text": "Согласовано", "value": "approved", "style": {"textColor": "light", "backgroundColor": "success"}},
        {"text": "Ожидает", "value": "pending", "style": {"textColor": "dark", "backgroundColor": "warning"}},
        {"text": "Отклонено", "value": "rejected", "style": {"textColor": "light", "backgroundColor": "danger"}},
    ])
    priority_opts = select_opts([
        {"text": "Высокий", "value": "high", "style": {"textColor": "light", "backgroundColor": "danger"}},
        {"text": "Средний", "value": "medium", "style": {"textColor": "dark", "backgroundColor": "warning"}},
        {"text": "Низкий", "value": "low", "style": {"textColor": "light", "backgroundColor": "success"}},
    ])

    # --- slice fields ---
    slice_fields_def = [
        ("String", "slice_kind", "Срез"),
        ("String", "slice_key", "Ключ"),
        ("String", "slice_label", "Название"),
        ("Number", "headcount", "Численность"),
        ("Number", "active_count", "Работают"),
        ("Number", "probation_count", "Испытательный"),
        ("Number", "on_leave_count", "В отпуске"),
        ("Number", "terminated_count", "Уволены"),
        ("Number", "open_positions", "Вакансии"),
        ("Number", "pending_leave", "Заявки на отпуск"),
        ("Number", "currently_away", "Сейчас отсутствуют"),
        ("Number", "turnover_pct", "Текучесть, %"),
        ("Number", "avg_tenure", "Средний стаж"),
        ("Number", "openings", "Открыто ставок"),
        ("String", "extra_label", "Доп."),
    ]
    slice_sql = []
    slice_meta = []
    for i, (kind, name, label) in enumerate(slice_fields_def):
        opts = num_opts("%" if name == "turnover_pct" else "", 1 if name in ("avg_tenure", "turnover_pct") else 0) if kind == "Number" else str_opts()
        fid = FID0 + i
        slice_sql.append(field_sql(fid, MOD_SLICE, i, kind, name, label, opts))
        slice_meta.append((fid, name, kind, label, opts))
    psql(dsn, "\n".join(slice_sql))

    # --- employee fields ---
    emp_fields_def = [
        ("String", "emp_code", "Таб. №", str_opts()),
        ("String", "full_name", "ФИО", str_opts()),
        ("String", "department", "Отдел", str_opts()),
        ("String", "title", "Должность", str_opts()),
        ("Select", "status", "Статус", status_opts),
        ("Select", "employment_type", "Занятость", emp_type_opts),
        ("Date", "hire_date", "Дата найма", date_opts()),
        ("Number", "tenure_years", "Стаж, лет", num_opts("", 1)),
        ("String", "salary_band", "Грейд", str_opts()),
        ("String", "manager_name", "Руководитель", str_opts()),
        ("String", "location", "Локация", str_opts()),
        ("String", "email", "Email", str_opts()),
    ]
    emp_sql = []
    for i, (kind, name, label, opts) in enumerate(emp_fields_def):
        emp_sql.append(field_sql(FID0 + 100 + i, MOD_EMP, i, kind, name, label, opts))
    psql(dsn, "\n".join(emp_sql))

    # --- leave fields ---
    leave_fields_def = [
        ("String", "emp_code", "Таб. №", str_opts()),
        ("String", "full_name", "ФИО", str_opts()),
        ("String", "department", "Отдел", str_opts()),
        ("Select", "leave_type", "Тип", leave_type_opts),
        ("Date", "start_date", "С", date_opts()),
        ("Date", "end_date", "По", date_opts()),
        ("Number", "days", "Дней", num_opts("", 0)),
        ("Select", "status", "Статус", leave_status_opts),
    ]
    leave_sql = []
    for i, (kind, name, label, opts) in enumerate(leave_fields_def):
        leave_sql.append(field_sql(FID0 + 200 + i, MOD_LEAVE, i, kind, name, label, opts))
    psql(dsn, "\n".join(leave_sql))

    # --- vacancy fields ---
    vac_fields_def = [
        ("String", "position_title", "Должность", str_opts()),
        ("String", "department", "Отдел", str_opts()),
        ("String", "location", "Локация", str_opts()),
        ("Select", "employment_type", "Занятость", emp_type_opts),
        ("Number", "openings", "Ставок", num_opts("", 0)),
        ("Select", "priority", "Приоритет", priority_opts),
        ("Number", "days_open", "Дней открыта", num_opts("", 0)),
        ("String", "hiring_manager", "Заказчик", str_opts()),
    ]
    vac_sql = []
    for i, (kind, name, label, opts) in enumerate(vac_fields_def):
        vac_sql.append(field_sql(FID0 + 300 + i, MOD_VAC, i, kind, name, label, opts))
    psql(dsn, "\n".join(vac_sql))

    charts = [
        (CH_DEPT, "hr-headcount-dept", "Численность по отделам",
         chart_cfg(MOD_SLICE, "slice_kind = 'department'", "slice_label", "headcount",
                   "bar", "Сотрудники", aggregate="SUM", horizontal=True, fmt="0.")),
        (CH_STATUS, "hr-headcount-status", "Состав по статусу",
         chart_cfg(MOD_SLICE, "slice_kind = 'status'", "slice_label", "headcount",
                   "doughnut", "Сотрудники", aggregate="SUM", fmt="0.")),
        (CH_LOC, "hr-headcount-location", "Численность по городам",
         chart_cfg(MOD_SLICE, "slice_kind = 'location'", "slice_label", "headcount",
                   "bar", "Сотрудники", aggregate="SUM", horizontal=True, fmt="0.")),
        (CH_TENURE, "hr-tenure", "Стаж работы",
         chart_cfg(MOD_SLICE, "slice_kind = 'tenure'", "slice_label", "headcount",
                   "bar", "Сотрудники", aggregate="SUM", fmt="0.")),
        (CH_VAC_DEPT, "hr-vacancy-dept", "Вакансии по отделам",
         chart_cfg(MOD_SLICE, "slice_kind = 'vacancy_dept'", "slice_label", "openings",
                   "bar", "Ставок", aggregate="SUM", horizontal=True, fmt="0.")),
        (CH_LEAVE, "hr-leave-type", "Отсутствия по типу",
         chart_cfg(MOD_SLICE, "slice_kind = 'leave_type'", "slice_label", "headcount",
                   "doughnut", "Заявки", aggregate="SUM", fmt="0.")),
    ]
    ch_sql = []
    for cid, handle, name, cfg in charts:
        cfg_txt = json.dumps(cfg, ensure_ascii=False).replace("'", "''")
        ch_sql.append(
            "INSERT INTO compose_chart (id, handle, rel_namespace, name, config) "
            f"VALUES ({cid}, '{handle}', {NS}, '{name}', '{cfg_txt}');"
        )
    psql(dsn, "\n".join(ch_sql))

    emp_rl = load_module_fields(dsn, MOD_EMP, [
        "emp_code", "full_name", "department", "title", "status",
        "employment_type", "location", "tenure_years", "salary_band",
    ])
    leave_rl = load_module_fields(dsn, MOD_LEAVE, [
        "full_name", "department", "leave_type", "start_date", "end_date", "days", "status",
    ])
    vac_rl = load_module_fields(dsn, MOD_VAC, [
        "position_title", "department", "location", "openings", "priority", "days_open", "hiring_manager",
    ])

    blocks = [
        wrap_block("Metric", "Сводка по персоналу", 1, [0, 0, 48, 14], {
            "metrics": [
                metric_item("Численность", "headcount", "", "0.", "#2e59d9", "hero"),
                metric_item("Испытательный", "probation_count", "", "0.", "#f6c23e", "badge"),
                metric_item("Сейчас отсутствуют", "currently_away", "", "0.", "#36b9cc", "badge"),
                metric_item("Открытых вакансий", "open_positions", "", "0.", "#e74a3b", "badge"),
                metric_item("Текучесть, %", "turnover_pct", "%", "0.0", "#858796", "meta"),
            ],
            "refreshRate": 0,
            "showRefresh": False,
            "magnifyOption": "",
            "likeRecordList": True,
            "density": "comfortable",
            "recordFieldLayoutOption": "default",
            "horizontalFieldLayoutEnabled": True,
        }),
        wrap_block("Chart", "Численность по отделам", 2, [0, 14, 24, 34], {
            "chartID": str(CH_DEPT),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0,
            "showRefresh": False,
            "magnifyOption": "modal",
            "liveFilterEnabled": False,
        }),
        wrap_block("Chart", "Состав по статусу", 3, [24, 14, 24, 34], {
            "chartID": str(CH_STATUS),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0,
            "showRefresh": False,
            "magnifyOption": "modal",
            "liveFilterEnabled": False,
        }),
        wrap_block("Chart", "Численность по городам", 4, [0, 48, 24, 34], {
            "chartID": str(CH_LOC),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0,
            "showRefresh": False,
            "magnifyOption": "modal",
            "liveFilterEnabled": False,
        }),
        wrap_block("Chart", "Стаж работы", 5, [24, 48, 24, 34], {
            "chartID": str(CH_TENURE),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0,
            "showRefresh": False,
            "magnifyOption": "modal",
            "liveFilterEnabled": False,
        }),
        wrap_block("Chart", "Вакансии по отделам", 6, [0, 82, 24, 34], {
            "chartID": str(CH_VAC_DEPT),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0,
            "showRefresh": False,
            "magnifyOption": "modal",
            "liveFilterEnabled": False,
        }),
        wrap_block("Chart", "Отсутствия по типу", 7, [24, 82, 24, 34], {
            "chartID": str(CH_LEAVE),
            "drillDown": {"blockID": "", "enabled": False, "recordListOptions": {"fields": []}},
            "refreshRate": 0,
            "showRefresh": False,
            "magnifyOption": "modal",
            "liveFilterEnabled": False,
        }),
        wrap_block("RecordList", "Сотрудники", 8, [0, 116, 48, 32], {
            "moduleID": str(MOD_EMP),
            "fields": emp_rl,
            "prefilter": "status = 'active' OR status = 'probation' OR status = 'on_leave'",
            "presort": "department, full_name",
            "perPage": 20,
            **rl_look(
                signalField="status",
                rowHighlightField="status",
                showRowSignal=True,
                groupByField="department",
                sparklineField="tenure_years",
                sparklineMax=10,
                displayMode="responsive",
            ),
        }),
        wrap_block("RecordList", "Открытые вакансии", 9, [0, 148, 48, 26], {
            "moduleID": str(MOD_VAC),
            "fields": vac_rl,
            "prefilter": "",
            "presort": "priority, days_open DESC",
            "perPage": 15,
            **rl_look(
                signalField="priority",
                rowHighlightField="priority",
                showRowSignal=True,
                groupByField="department",
                sparklineField="days_open",
                sparklineMax=60,
                displayMode="responsive",
            ),
        }),
        wrap_block("RecordList", "Отсутствия и заявки", 10, [0, 174, 48, 28], {
            "moduleID": str(MOD_LEAVE),
            "fields": leave_rl,
            "prefilter": "status = 'approved' OR status = 'pending'",
            "presort": "start_date DESC",
            "perPage": 15,
            **rl_look(
                signalField="leave_type",
                rowHighlightField="status",
                showRowSignal=True,
                groupByField="leave_type",
                sparklineField="days",
                sparklineMax=28,
                displayMode="responsive",
            ),
        }),
    ]

    page_cfg = {
        "prompt": "Ты HR-аналитик. По этой странице разбери численность по отделам и городам, "
                  "статусы (работает / испытательный / отпуск / уволен), открытые вакансии и приоритеты найма, "
                  "текущие отсутствия и заявки на отпуск, стаж и текучесть. "
                  "Предлагай приоритеты: где нехватка людей, где риск ухода, кого ускорить в найме."
    }
    page_meta = {"notifications": {"enabled": True}, "allowPersonalLayouts": False}
    psql(dsn, f"""
INSERT INTO compose_page (id, title, handle, self_id, rel_module, rel_namespace, meta, config, blocks, visible, weight, description)
VALUES ({PAGE}, 'Персонал', 'hr-personnel', 0, 0, {NS},
        {lit(page_meta)}, {lit(page_cfg)}, {lit(blocks)},
        true, 5, 'Управление персоналом: численность, вакансии, отсутствия');
""")

    layout_blocks = [
        {
            "blockID": str(b["blockID"]),
            "xywh": b["xywh"],
            "meta": {"hidden": False, "visibility": {"roles": [], "expression": ""}},
        }
        for b in blocks
    ]
    layout_meta = {"title": "Персонал", "description": ""}
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
        "SELECT headcount||'/'||open_positions||'/'||currently_away||'/'||turnover_pct "
        "FROM hr_slice WHERE slice_kind='kpi' LIMIT 1;",
    )
    print("OK")
    print("  modules Hr_slice / Hr_employee / Hr_leave / Hr_vacancy")
    print("  charts hr-headcount-*, hr-tenure, hr-vacancy-dept, hr-leave-type")
    print("  page /ns/loop/pages/%s (Персонал)" % PAGE)
    print("  kpi headcount/vacancies/away/turnover%% = %s" % kpi)
    print("  restart GoLand Server_RU_test_translations for DalModelReload")


if __name__ == "__main__":
    main()
