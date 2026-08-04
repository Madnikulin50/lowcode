#!/usr/bin/env python3
"""Build static HTML from Markdown docs for nginx."""
import os, re, html, sys
from pathlib import Path
from urllib.parse import quote

MD_DIR = Path(os.environ.get('DOCS_DIR', '/docs'))
OUT_DIR = Path(os.environ.get('OUT_DIR', '/usr/share/nginx/html'))

HEAD = '''<!DOCTYPE html>
<html lang="ru">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>{title} &mdash; LowCoooode Docs</title>
<link rel="stylesheet" href="/static/docs.css">
<link rel="stylesheet" href="/static/docs-override.css">
<script src="/static/sidebar.js"></script></head>
<body><header><div class="header-inner">
<a href="/">&#128216; LowCoooode Documentation</a>
</div></header><div class="layout">
<nav class="sidebar"><ul>{nav}</ul></nav>
<main class="content"><article>{content}</article></main>
</div></body></html>'''

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
            exp = ' expanded' if kh else ''
            h += f'<li class="dir"><span class="dir-label"><span class="chevron">&#9654;</span>{label}</span><ul style="display:block">{kh}</ul></li>'
        else:
            if label.lower() == 'index':
                parent_label = entry.parent.name.replace('-',' ').title()
                label = parent_label.replace('Lowcode', 'LowCoooode') or 'Overview'
            cls = ' active' if rel == active_path else ''
            h += f'<li{cls}><a href="/{rel.replace(".md",".html")}">{label}</a></li>'
    return h

def render_md(text: str) -> str:
    lines, out, in_code, in_table = text.split('\n'), [], False, False
    for line in lines:
        if line.startswith('```'):
            out.append(('</code></pre>' if in_code else '<pre><code>'))
            in_code = not in_code
            continue
        if in_code: out.append(html.escape(line)); continue
        
        m = re.match(r'^!!!\s*(note|important|warning|caution|tip)', line)
        if m: out.append(f'<div class="admonition {m.group(1)}">'); continue
        
        m = re.match(r'^(#{1,6})\s+(.+)$', line)
        if m: out.append(f'<h{len(m.group(1))}>{m.group(2).strip()}</h{len(m.group(1))}>'); continue
        
        if re.match(r'^---+\s*$', line) or re.match(r'^\*\*\*+\s*$', line):
            out.append('<hr>'); continue
        
        s = line.strip()
        if not s: out.append(''); continue
        out.append(f'<p>{inline_md(s)}</p>')
    
    h = '\n'.join(out)
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