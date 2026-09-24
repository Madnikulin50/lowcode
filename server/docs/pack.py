#!/usr/bin/env python3
"""Copy user and architecture docs into this package so go:embed can include them.

Sources live outside the Go module (manual/, docs/). The copies under
server/docs/manual and server/docs/architecture are build artifacts.
"""

import html
import re
import shutil
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
PKG = Path(__file__).resolve().parent
MANUAL_SRC = ROOT / "manual" / "ru" / "tools" / "docserver" / "html"
ARCH_SRC = ROOT / "docs"
MANUAL_DST = PKG / "manual"
ARCH_DST = PKG / "architecture"

PAGE = """<!DOCTYPE html>
<html lang="ru">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{title}</title>
  <style>
    :root {{ color-scheme: light; }}
    body {{ margin: 0; font: 16px/1.55 Georgia, "Iowan Old Style", serif; color: #1c1917; background: #fafaf9; }}
    header {{ background: #1c1917; color: #fafaf9; padding: 0.75rem 1.5rem; }}
    header a {{ color: #fafaf9; margin-right: 1rem; text-decoration: none; }}
    header a[aria-current="page"] {{ text-decoration: underline; }}
    main {{ max-width: 52rem; margin: 0 auto; padding: 1.5rem; }}
    pre, code {{ font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }}
    pre {{ background: #f5f5f4; padding: 0.75rem 1rem; overflow: auto; }}
    code {{ background: #f5f5f4; padding: 0.1rem 0.3rem; }}
    pre code {{ background: none; padding: 0; }}
    table {{ border-collapse: collapse; width: 100%; }}
    th, td {{ border: 1px solid #d6d3d1; padding: 0.35rem 0.6rem; vertical-align: top; }}
    th {{ background: #f5f5f4; }}
    h1, h2, h3, h4 {{ font-family: system-ui, sans-serif; line-height: 1.25; }}
    a {{ color: #1d4ed8; }}
  </style>
</head>
<body>
  <header>{nav}</header>
  <main>{body}</main>
</body>
</html>
"""


def rewrite_link(url: str) -> str:
    match = re.match(r"([^)#]+?\.md)(#.*)?$", url)
    if not match:
        return url
    name = Path(match.group(1)).name
    target = "index.html" if name.lower() == "readme.md" else name[:-3] + ".html"
    return target + (match.group(2) or "")


def inline(text: str) -> str:
    parts = []
    pattern = re.compile(r"(`[^`]+`|\*\*[^*]+\*\*|\[[^\]]+\]\([^)]+\))")
    pos = 0
    for match in pattern.finditer(text):
        parts.append(html.escape(text[pos:match.start()]))
        token = match.group(0)
        if token.startswith("`"):
            parts.append("<code>" + html.escape(token[1:-1]) + "</code>")
        elif token.startswith("**"):
            parts.append("<strong>" + html.escape(token[2:-2]) + "</strong>")
        else:
            label, url = re.match(r"\[([^\]]+)\]\(([^)]+)\)", token).groups()
            parts.append(
                f'<a href="{html.escape(rewrite_link(url), quote=True)}">{html.escape(label)}</a>'
            )
        pos = match.end()
    parts.append(html.escape(text[pos:]))
    return "".join(parts)


def slug(text: str) -> str:
    value = re.sub(r"[^\w\s-]", "", text, flags=re.UNICODE).strip().lower()
    return re.sub(r"\s+", "-", value) or "section"


def render_markdown(source: str) -> str:
    lines = source.replace("\r\n", "\n").split("\n")
    out = []
    i = 0

    def flush_paragraph(buf):
        if buf:
            out.append("<p>" + inline(" ".join(buf)) + "</p>")
            buf.clear()

    paragraph = []
    while i < len(lines):
        line = lines[i]
        if line.startswith("```"):
            flush_paragraph(paragraph)
            lang = html.escape(line[3:].strip())
            block = []
            i += 1
            while i < len(lines) and not lines[i].startswith("```"):
                block.append(html.escape(lines[i]))
                i += 1
            klass = f' class="language-{lang}"' if lang else ""
            out.append(f"<pre><code{klass}>" + "\n".join(block) + "</code></pre>")
            i += 1
            continue
        if re.match(r"^\s*\|", line):
            flush_paragraph(paragraph)
            rows = []
            while i < len(lines) and re.match(r"^\s*\|", lines[i]):
                cells = [cell.strip() for cell in lines[i].strip().strip("|").split("|")]
                if not re.match(r"^:?-+:?$", cells[0] if cells else ""):
                    rows.append(cells)
                i += 1
            if rows:
                head, body = rows[0], rows[1:]
                table = ["<table><thead><tr>"]
                table.extend(f"<th>{inline(cell)}</th>" for cell in head)
                table.append("</tr></thead><tbody>")
                for row in body:
                    table.append("<tr>" + "".join(f"<td>{inline(cell)}</td>" for cell in row) + "</tr>")
                table.append("</tbody></table>")
                out.append("".join(table))
            continue
        heading = re.match(r"^(#{1,4})\s+(.*)$", line)
        if heading:
            flush_paragraph(paragraph)
            level = len(heading.group(1))
            title = heading.group(2).strip()
            out.append(f'<h{level} id="{html.escape(slug(title), quote=True)}">{inline(title)}</h{level}>')
            i += 1
            continue
        if re.match(r"^---+\s*$", line):
            flush_paragraph(paragraph)
            out.append("<hr>")
            i += 1
            continue
        if re.match(r"^\s*[-*]\s+", line):
            flush_paragraph(paragraph)
            out.append("<ul>")
            while i < len(lines) and re.match(r"^\s*[-*]\s+", lines[i]):
                item = re.sub(r"^\s*[-*]\s+", "", lines[i])
                out.append("<li>" + inline(item) + "</li>")
                i += 1
            out.append("</ul>")
            continue
        if re.match(r"^\s*\d+\.\s+", line):
            flush_paragraph(paragraph)
            out.append("<ol>")
            while i < len(lines) and re.match(r"^\s*\d+\.\s+", lines[i]):
                item = re.sub(r"^\s*\d+\.\s+", "", lines[i])
                out.append("<li>" + inline(item) + "</li>")
                i += 1
            out.append("</ol>")
            continue
        if not line.strip():
            flush_paragraph(paragraph)
            i += 1
            continue
        paragraph.append(line.strip())
        i += 1
    flush_paragraph(paragraph)
    return "\n".join(out)


NAV_RU = {
    "Modules": "Разделы",
    "Developer Guide": "Руководство разработчика",
    "Devops Guide": "Руководство DevOps",
    "End User Guide": "Руководство пользователя",
    "Integrator Guide": "Руководство интегратора",
    "External": "Внешние материалы",
    "Generated": "Сгенерированное",
    "Root": "Общее",
    "Pages": "Страницы",
    "Partials": "Фрагменты",
    "Examples": "Примеры",
    "Documentation": "Документация",
    "Getting Started": "Начало работы",
    "Interface": "Интерфейс",
    "Records": "Записи",
    "Overview": "Обзор",
    "Changelog": "Журнал изменений",
    "Access Control": "Контроль доступа",
    "Authentication": "Аутентификация",
    "Automation": "Автоматизация",
    "Automation Scripts": "Скрипты автоматизации",
    "Workflows": "Процессы",
    "Compose Configuration": "Настройка Compose",
    "Federation": "Федерация",
    "Reporting": "Отчёты",
    "Security Model": "Модель безопасности",
    "Templates": "Шаблоны",
    "Troubleshooting": "Устранение неполадок",
    "Upgrade": "Обновление",
    "References": "Справочник",
    "Maintenance": "Сопровождение",
    "LowCoooode Documentation": "Документация Lowcode",
    "LowCoooode Server": "Сервер Lowcode",
    "LowCoooode Vue": "Vue Lowcode",
    "LowCoooode Js": "JavaScript Lowcode",
    "Web Applications": "Веб-приложения",
    "Case Management": "Управление обращениями",
    "Data Privacy": "Конфиденциальность данных",
    "Crm": "CRM",
}


def relativize_manual(text: str, depth: int) -> str:
    prefix = "../" * depth

    def repl(match):
        attr, path = match.group(1), match.group(2)
        if path.startswith("//"):
            return match.group(0)
        target = "index.html" if path == "/" else path.lstrip("/")
        target = re.sub(r"\.md(?=#|$)", ".html", target)
        return f'{attr}="{prefix}{target}"'

    text = re.sub(r'(href|src)="(/[^"]*)"', repl, text)
    text = re.sub(
        r"(https?://[^\s<\[\"]+)\[([^\]]+)\]",
        r'<a href="\1">\2</a>',
        text,
    )
    for english, russian in sorted(NAV_RU.items(), key=lambda item: len(item[0]), reverse=True):
        text = text.replace(f">{english}<", f">{russian}<")
    opens = text.count('class="admonition')
    if opens:
        text = text.replace("</article>", "</div>" * opens + "</article>", 1)
    return text


def pack_manual():
    if not (MANUAL_SRC / "index.html").is_file():
        raise SystemExit(f"user manual site not found: {MANUAL_SRC}")
    if MANUAL_DST.exists():
        shutil.rmtree(MANUAL_DST)
    shutil.copytree(MANUAL_SRC, MANUAL_DST)
    count = 0
    for path in MANUAL_DST.rglob("*.html"):
        depth = len(path.relative_to(MANUAL_DST).parts) - 1
        path.write_text(relativize_manual(path.read_text(encoding="utf-8"), depth))
        count += 1
    print(f"manual: {count} html files -> {MANUAL_DST.relative_to(PKG)}")


def pack_architecture():
    sources = sorted(ARCH_SRC.glob("*.md"))
    if not sources:
        raise SystemExit(f"architecture docs not found: {ARCH_SRC}")
    if ARCH_DST.exists():
        shutil.rmtree(ARCH_DST)
    ARCH_DST.mkdir()
    pages = []
    for path in sources:
        name = "index.html" if path.name.lower() == "readme.md" else path.stem + ".html"
        title = path.stem if path.name.lower() != "readme.md" else "Архитектура"
        first = next((line[2:].strip() for line in path.read_text().splitlines() if line.startswith("# ")), title)
        pages.append((name, first, path))
    nav_links = []
    for name, title, _path in pages:
        nav_links.append((name, title))
    for name, title, path in pages:
        nav = " ".join(
            f'<a href="{html.escape(href, quote=True)}"'
            + (' aria-current="page"' if href == name else "")
            + f">{html.escape(label)}</a>"
            for href, label in nav_links
        )
        body = render_markdown(path.read_text())
        (ARCH_DST / name).write_text(PAGE.format(title=html.escape(title), nav=nav, body=body))
    print(f"architecture: {len(pages)} pages -> {ARCH_DST.relative_to(PKG)}")


def main():
    pack_manual()
    pack_architecture()


if __name__ == "__main__":
    main()
