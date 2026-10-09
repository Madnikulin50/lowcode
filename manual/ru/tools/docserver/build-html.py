#!/usr/bin/env python3
"""Build static HTML from Markdown docs for nginx."""
import os, re, html, sys
from pathlib import Path
from urllib.parse import quote

MD_DIR = Path(os.environ.get('DOCS_DIR', '/docs'))
OUT_DIR = Path(os.environ.get('OUT_DIR', '/usr/share/nginx/html'))

HEAD = '''<!DOCTYPE html>
<html lang="ru" data-color-mode="light">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>{title} &mdash; Документация</title>
<link rel="stylesheet" href="/static/docs.css">
<link rel="stylesheet" href="/static/docs-override.css">
<script src="/static/sidebar.js"></script></head>
<body>
<header class="topbar">
<button type="button" class="menu-btn" id="docs-menu" aria-label="Меню"><span></span></button>
<div class="header-inner"><a href="/">Документация</a></div>
<button type="button" class="theme-btn" id="docs-theme" aria-label="Тема"></button>
</header>
<div class="layout">
<nav class="sidebar"><ul>{nav}</ul></nav>
<main class="content"><article class="page-card">{content}</article></main>
</div>
<div class="sidebar-backdrop" id="docs-backdrop"></div>
</body></html>'''

def build_tree(base: Path, prefix='') -> list:
    items = []
    for e in sorted(base.iterdir(), key=lambda p: (p.is_file(), p.name.lower())):
        if e.name.startswith('.'): continue
        if e.is_dir():
            kids = build_tree(e)
            if kids: items.append((e, True, kids))
        elif e.suffix == '.md': items.append((e, False, None))
    return items

def _nav_label(entry):
    label = entry.stem.replace('-',' ').replace('_',' ').title() if entry.suffix=='.md' else entry.name.replace('-',' ').title()
    return label.replace('Lowcode', 'LowCoooode')

def render_nav(tree, active_path: str) -> str:
    h = ''
    for entry, is_dir, kids in tree:
        rel = entry.relative_to(MD_DIR).as_posix()
        label = _nav_label(entry)
        if is_dir:
            kh = render_nav(kids, active_path) if kids else ''
            opened = ' expanded' if (' class="active"' in kh or active_path.startswith(rel + '/')) else ''
            h += f'<li class="dir{opened}"><span class="dir-label"><span class="chevron"></span>{html.escape(label)}</span><ul>{kh}</ul></li>'
        else:
            if label.lower() == 'index':
                parent_label = entry.parent.name.replace('-',' ').title()
                label = parent_label.replace('Lowcode', 'LowCoooode') or 'Overview'
            cls = ' active' if rel == active_path else ''
            h += f'<li{cls}><a href="/{rel.replace(".md",".html")}">{label}</a></li>'
    return h

def _table(rows: list) -> str:
    body = []
    for i, row in enumerate(rows):
        tag = 'th' if i == 0 else 'td'
        cells = ''.join(f'<{tag}>{inline_md(c.strip())}</{tag}>' for c in row)
        body.append(f'<tr>{cells}</tr>')
    return '<table>' + ''.join(body) + '</table>'

def _split_row(line: str) -> list:
    return [c.strip() for c in line.strip().strip('|').split('|')]

def _is_sep(line: str) -> bool:
    return bool(re.match(r'^\s*\|?\s*:?-{3,}:?\s*(\|\s*:?-{3,}:?\s*)+\|?\s*$', line))

def render_md(text: str) -> str:
    lines, out, in_code = text.split('\n'), [], False
    i = 0
    while i < len(lines):
        line = lines[i]
        if line.startswith('```'):
            out.append('</code></pre>' if in_code else '<pre><code>')
            in_code = not in_code
            i += 1
            continue
        if in_code:
            out.append(html.escape(line))
            i += 1
            continue

        if '|' in line and i + 1 < len(lines) and _is_sep(lines[i + 1]):
            rows = [_split_row(line)]
            i += 2
            while i < len(lines) and '|' in lines[i] and lines[i].strip():
                rows.append(_split_row(lines[i]))
                i += 1
            out.append(_table(rows))
            continue

        m = re.match(r'^!!!\s*(note|important|warning|caution|tip)', line)
        if m:
            out.append(f'<div class="admonition {m.group(1)}">')
            i += 1
            continue

        m = re.match(r'^(#{1,6})\s+(.+)$', line)
        if m:
            out.append(f'<h{len(m.group(1))}>{inline_md(m.group(2).strip())}</h{len(m.group(1))}>')
            i += 1
            continue

        if re.match(r'^---+\s*$', line) or re.match(r'^\*\*\*+\s*$', line):
            out.append('<hr>')
            i += 1
            continue

        s = line.strip()
        if not s:
            out.append('')
            i += 1
            continue
        if s.startswith(('- ', '* ')):
            out.append(f'<li>{inline_md(s[2:])}</li>')
            i += 1
            continue
        out.append(f'<p>{inline_md(s)}</p>')
        i += 1
    
    h = '\n'.join(out)
    h = re.sub(r'(?:<li>.*?</li>\n?)+', lambda m: '<ul>' + m.group(0) + '</ul>', h)
    h = re.sub(r'href="([^"]+?)\.md"', r'href="\1.html"', h)
    return h

def inline_md(text: str) -> str:
    t = html.escape(text)
    t = re.sub(r'`([^`]+)`', r'<code>\1</code>', t)
    t = re.sub(r'\*\*(.+?)\*\*', r'<strong>\1</strong>', t)
    t = re.sub(r'\*(.+?)\*', r'<em>\1</em>', t)
    t = re.sub(r'!\[([^\]]*)\]\(([^)]+)\)', lambda m: f'<img src="/{m.group(2)}" alt="{m.group(1)}">' if not m.group(2).startswith(('http://','https://','/')) else f'<img src="{m.group(2)}" alt="{m.group(1)}">', t)
    t = re.sub(r'\[([^\]]+)\]\(([^)]+)\)', lambda m: _fix_link(m), t)
    return t

def _fix_link(m):
    text = m.group(1)
    url = m.group(2)
    if url.startswith(('http://', 'https://', '/', '#', 'mailto:')):
        return f'<a href="{url}">{text}</a>'
    return f'<a href="/{url}">{text}</a>'

def main():
    tree = build_tree(MD_DIR)
    count = 0
    for md_file in sorted(MD_DIR.rglob('*.md')):
        rel = md_file.relative_to(MD_DIR).as_posix()
        text = md_file.read_text('utf-8')
        title = md_file.stem.replace('-',' ').title()
        fl = text.strip().split('\n')[0] if text.strip() else ''
        if fl.startswith('# '): title = fl[2:]
        nav = render_nav(tree, rel)
        content = render_md(text)
        html_text = HEAD.format(title=title, nav=nav, content=content)
        
        out_path = OUT_DIR / rel.replace('.md', '.html')
        out_path.parent.mkdir(parents=True, exist_ok=True)
        out_path.write_text(html_text, 'utf-8')
        
        # index.md → directory redirect
        if md_file.name == 'index.md' and md_file.parent != MD_DIR:
            parent_rel = md_file.parent.relative_to(MD_DIR).as_posix()
            parent_html = OUT_DIR / f'{parent_rel}.html'
            parent_html.parent.mkdir(parents=True, exist_ok=True)
            if not parent_html.exists():
                parent_html.write_text(
                    f'<!DOCTYPE html><html><head><meta charset="utf-8">'
                    f'<meta http-equiv="refresh" content="0;url={parent_rel}/index.html">'
                    f'</head><body></body></html>'
                )
        count += 1
    
    # Copy static assets (CSS, JS, images)
    for ext in ('png', 'jpg', 'jpeg', 'gif', 'svg', 'ico'):
        for sf in MD_DIR.rglob(f'*.{ext}'):
            rel = sf.relative_to(MD_DIR)
            (OUT_DIR / rel).parent.mkdir(parents=True, exist_ok=True)
            (OUT_DIR / rel).write_bytes(sf.read_bytes())
    
    # Copy sidebar.js, docs.css, docs-override.css to /static/
    static_src = Path(__file__).parent / 'static'
    if static_src.exists():
        for f in static_src.iterdir():
            if f.suffix in ('.js', '.css'):
                dest = OUT_DIR / 'static' / f.name
                dest.parent.mkdir(parents=True, exist_ok=True)
                dest.write_bytes(f.read_bytes())
    
    # Root redirect → main doc page
    landing = OUT_DIR / 'modules/ROOT/pages/index.html'
    if landing.exists():
        root_html = OUT_DIR / 'index.html'
        root_html.write_text(
            '<!DOCTYPE html><html><head><meta charset="utf-8">'
            '<meta http-equiv="refresh" content="0;url=modules/ROOT/pages/index.html">'
            '<title>Lowcode Docs</title></head><body></body></html>'
        )
    
    print(f'Generated {count} pages + assets in {OUT_DIR}', file=sys.stderr)

if __name__ == '__main__':
    main()
    sys.stderr.flush()