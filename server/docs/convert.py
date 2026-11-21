#!/usr/bin/env python3
"""Generate OpenAPI 3 YAML in this directory from */rest.yaml."""

import copy
import sys
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[1]
DOCS = Path(__file__).resolve().parent

NAMESPACES = [
    ("system", "System"),
    ("compose", "Compose"),
    ("automation", "Automation"),
    ("federation", "Federation"),
    ("discovery", "Discovery"),
]

SCALARS = {
    "bool": {"type": "boolean"},
    "boolean": {"type": "boolean"},
    "int": {"type": "integer"},
    "int32": {"type": "integer"},
    "int64": {"type": "string"},
    "uint": {"type": "string"},
    "uint32": {"type": "string"},
    "uint64": {"type": "string"},
    "float32": {"type": "number"},
    "float64": {"type": "number"},
    "string": {"type": "string"},
    "time.Time": {"type": "string", "format": "date-time"},
    "multipart.FileHeader": {"type": "string", "format": "binary"},
    "json.RawMessage": {"type": "string", "format": "json"},
    "sqlxTypes.JSONText": {"type": "string", "format": "json"},
    "id.ID": {"type": "string"},
    "types.UserKind": {"type": "string"},
    "types.ChannelMembershipPolicy": {"type": "string"},
}

SCHEMAS = {
    "types.RecordValueSet": {
        "type": "array",
        "items": {
            "type": "object",
            "properties": {
                "name": {"type": "string"},
                "value": {"type": "string"},
            },
        },
    },
    "types.RecordBulkSet": {
        "type": "array",
        "items": {
            "type": "object",
            "properties": {
                "recordID": {"type": "string"},
                "moduleID": {"type": "string"},
                "namespaceID": {"type": "string"},
                "values": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "properties": {
                            "name": {"type": "string"},
                            "value": {"type": "string"},
                        },
                    },
                },
            },
        },
    },
    "types.SettingValueSet": {
        "type": "array",
        "items": {
            "type": "object",
            "properties": {
                "name": {"type": "string"},
                "value": {"type": "string"},
            },
        },
    },
    "permissions.RuleSet": {
        "type": "array",
        "items": {
            "type": "object",
            "properties": {
                "roleID": {"type": "string"},
                "resource": {"type": "string"},
                "operation": {"type": "string"},
                "access": {"type": "string"},
            },
        },
    },
    "types.ModuleFieldSet": {
        "type": "array",
        "items": {
            "type": "object",
            "properties": {
                "name": {"type": "string"},
                "kind": {"type": "string"},
                "label": {"type": "string"},
                "isRequired": {"type": "boolean"},
                "isMulti": {"type": "boolean"},
                "options": {"type": "object"},
            },
        },
    },
}


def normalize_type(raw):
    if not isinstance(raw, str) or not raw.strip():
        return "string"
    value = raw.strip().strip("'\"")
    if value.startswith("*"):
        value = value[1:]
    return value


def schema_for(raw, name):
    if name == "password":
        return {"type": "string", "format": "password"}

    value = normalize_type(raw)
    if value.startswith("[]"):
        return {"type": "array", "items": schema_for(value[2:], name)}
    if value.startswith("map["):
        return {"type": "object", "additionalProperties": True}
    if value in SCHEMAS:
        return copy.deepcopy(SCHEMAS[value])
    if value in SCALARS:
        return copy.deepcopy(SCALARS[value])
    if value.startswith("types.") or value.startswith("labelTypes."):
        return {"type": "object"}
    return {"type": "string"}


def as_params(value):
    if value is None:
        return []
    if isinstance(value, dict):
        return [value] if value.get("name") else []
    if isinstance(value, list):
        return [item for item in value if isinstance(item, dict) and item.get("name")]
    return []


def merge_params(endpoint, api):
    merged = {}
    for source in (endpoint.get("parameters") or {}, api.get("parameters") or {}):
        if not isinstance(source, dict):
            continue
        for key, value in source.items():
            merged.setdefault(key, [])
            merged[key].extend(as_params(value))
    return merged


def operation(endpoint, api, params):
    method = str(api.get("method") or "").lower()
    op = {
        "tags": [endpoint.get("title") or "default"],
        "summary": api.get("title") or api.get("name") or method.upper(),
        "operationId": ".".join(
            part for part in [
                endpoint.get("entrypoint") or "",
                api.get("name") or "",
            ] if part
        ) or None,
        "responses": {"200": {"description": "OK"}},
    }
    if not op["operationId"]:
        del op["operationId"]

    for key, items in params.items():
        if not items:
            continue
        if key == "get":
            op.setdefault("parameters", [])
            for item in items:
                op["parameters"].append({
                    "in": "query",
                    "name": item["name"],
                    "description": item.get("title") or "",
                    "required": bool(item.get("required")),
                    "schema": schema_for(item.get("type"), item["name"]),
                })
        elif key == "path":
            op.setdefault("parameters", [])
            for item in items:
                op["parameters"].append({
                    "in": "path",
                    "name": item["name"],
                    "description": item.get("title") or "",
                    "required": True,
                    "schema": schema_for(item.get("type"), item["name"]),
                })
        elif key == "post":
            properties = {}
            required = []
            for item in items:
                properties[item["name"]] = schema_for(item.get("type"), item["name"])
                if item.get("title"):
                    properties[item["name"]]["description"] = item["title"]
                if item.get("required"):
                    required.append(item["name"])
            body = {
                "content": {
                    "application/json": {
                        "schema": {"type": "object", "properties": properties},
                    },
                    "application/x-www-form-urlencoded": {
                        "schema": {"type": "object", "properties": copy.deepcopy(properties)},
                    },
                },
            }
            if required:
                body["content"]["application/json"]["schema"]["required"] = required
                body["content"]["application/x-www-form-urlencoded"]["schema"]["required"] = list(required)
            op["requestBody"] = body
    return method, op


def build(namespace, class_name, spec):
    doc = {
        "openapi": "3.0.0",
        "info": {
            "title": f"Lowcode {class_name} API",
            "description": f"Lowcode {namespace} REST API. Generated from {namespace}/rest.yaml.",
            "version": "2026.9",
            "license": {
                "name": "Apache 2.0",
                "url": "http://www.apache.org/licenses/LICENSE-2.0.html",
            },
        },
        "servers": [{"url": "/api", "description": "Lowcode API"}],
        "paths": {},
    }

    count = 0
    for endpoint in spec:
        base = f"/{namespace}{endpoint.get('path') or ''}"
        for api in endpoint.get("apis") or []:
            method, op = operation(endpoint, api, merge_params(endpoint, api))
            if not method:
                continue
            extra = api.get("path") or ""
            if not isinstance(extra, str):
                extra = ""
            full = base + extra
            slot = doc["paths"].setdefault(full, {})
            if method in slot:
                print(f"duplicate {method.upper()} {full}", file=sys.stderr)
            slot[method] = op
            count += 1
    return doc, count


def main():
    for namespace, class_name in NAMESPACES:
        source = ROOT / namespace / "rest.yaml"
        if not source.exists():
            print(f"skip {namespace}: {source} not found", file=sys.stderr)
            continue
        spec = yaml.safe_load(source.read_text()) or {}
        endpoints = spec.get("endpoints") or []
        doc, count = build(namespace, class_name, endpoints)
        target = DOCS / f"{namespace}.yaml"
        text = yaml.safe_dump(
            doc,
            sort_keys=False,
            allow_unicode=True,
            default_flow_style=False,
            width=100,
        )
        target.write_text(
            f"# Generated from {namespace}/rest.yaml by docs/convert.py. Do not edit by hand.\n" + text
        )
        print(f"{namespace}: {count} operations -> {target.name}")


if __name__ == "__main__":
    main()
