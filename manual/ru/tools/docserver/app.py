import http.server, os, re, html, json, mimetypes
from pathlib import Path
from urllib.parse import urlparse, unquote

MD_DIR = Path(os.environ.get('DOCS_DIR', '/docs'))
PORT = int(os.environ.get('PORT', '5000'))

HEAD = '''<!DOCTYPE html>
<html lang="ru">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>{title} — Lowcode Docs</title>
<style>
*{{box-sizing:border-box;margin:0;padding:0}}
body{{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;font-size:16px;line-height:1.6;color:#1a1a1a;background:#fafafa}}
header{{background:#1a73e8;color:#fff;padding:0 24px;position:sticky;top:0;z-index:100}}
.header-inner{{max-width:1400px;margin:0 auto;display:flex;align-items:center;height:52px;justify-content:space-between}}
.header-inner a{{color:#fff;text-decoration:none;font-weight:600;font-size:17px}}
.header-lang{{font-size:12px;opacity:.7}}
.layout{{display:flex;max-width:1400px;margin:0 auto;min-height:calc(100vh - 52px)}}
.sidebar{{width:300px;flex-shrink:0;background:#fff;border-right:1px solid #e0e0e0;padding:16px 0;overflow-y:auto;max-height:calc(100vh - 52px);position:sticky;top:52px}}
.sidebar ul{{list-style:none;padding:0;margin:0}}
.sidebar li{{margin:0}}
.sidebar li.dir>ul{{display:none;padding-left:16px}}
.sidebar a,.sidebar .dir-label{{display:block;padding:4px 20px;font-size:14px;color:#444;text-decoration:none;cursor:pointer;border-left:3px solid transparent;transition:.12s}}
.sidebar .dir-label{{font-weight:600;color:#1a73e8}}
.sidebar a:hover{{background:#f0f7ff;border-left-color:#1a73e8;color:#1a73e8}}
.sidebar li.active>a{{background:#e8f0fe;border-left-color:#1a73e8;font-weight:600;color:#1a73e8}}
.content{{flex:1;padding:32px 48px;max-width:960px}}
.content h1{{font-size:2rem;margin:0 0 .5em;border-bottom:2px solid #e0e0e0;padding-bottom:.3em}}
.content h2{{font-size:1.5rem;margin:1.2em 0 .5em}}
.content h3{{font-size:1.2rem;margin:1em 0 .4em}}
.content p{{margin:.6em 0}}
.content a{{color:#1a73e8}}
.content code{{background:#f0f0f0;padding:2px 6px;border-radius:3px;font-size:.9em}}
.content pre{{background:#1e1e2e;color:#cdd6f4;padding:16px;border-radius:8px;overflow-x:auto;font-size:14px;margin:.8em 0}}
.content pre code{{background:none;padding:0}}
.content blockquote{{border-left:4px solid #1a73e8;padding:.5em 1em;margin:.8em 0;background:#f8faff}}
.content table{{border-collapse:collapse;width:100%;margin:1em 0}}
.content th,.content td{{border:1px solid #ddd;padding:8px 12px;text-align:left}}
.content th{{background:#f5f5f5;font-weight:600}}
.content img{{max-width:100%;height:auto;margin:1em 0;border-radius:4px}}
.content ul,.content ol{{margin:.4em 0;padding-left:1.5em}}
@media(max-width:900px){{.sidebar{{display:none}}.content{{padding:16px 20px}}}}
</style></head>
<body><header><div class="header-inner">
<a href="/">📖 Lowcode Documentation</a><span class="header-lang">RU</span>
</div></header><div class="layout">
<nav class="sidebar"><ul>{nav}</ul></nav>
<main class="content"><article>{content}</article></main>
</div></body></html>'''

def build_tree(base: Path) -> list:
    items = []
    for e in sorted(base.iterdir(), key=lambda p: (p.is_file(), p.name.lower())):
        if e.name.startswith('.'): continue
        if e.is_dir():
            kids = build_tree(e)
            if kids:
                items.append((e, True, kids))
        elif e.suffix == '.md':
            items.append((e, False, None))
    return items

def render_nav(tree, active_path: str) -> str:
    html = ''
    for entry, is_dir, kids in tree:
        if is_dir:
            rel = entry.relative_to(MD_DIR).as_posix()
            kids_html = render_nav(kids, active_path) if kids else ''
            expanded = ' dir' if active_path.startswith(rel) else ''
            html += f'<li class="dir{expanded}"><span class="dir-label">{entry.name.replace("-"," ").title()}</span>'
            if kids_html:
                html += f'<ul>{"open" if expanded else ""}">{kids_html}</ul>'
            html += '</li>'
        else:
            rel = entry.relative_to(MD_DIR).as_posix()
            label = entry.stem.replace('-', ' ').replace('_', ' ').title()
            if label.lower() == 'index':
                label = entry.parent.name.replace('-', ' ').title() or 'Overview'
            active = ' active' if rel == active_path else ''
            html += f'<li{active}><a href="/{rel.replace(".md",".html")}">{label}</a></li>'
    return html

def render_md(text: str) -> str:
    lines = text.split('\n')
    out = []
    in_code = False
    in_table = False
    table_rows = []
    
    for line in lines:
        # Code blocks
        if line.startswith('```'):
            if in_code:
                out.append('</code></pre>')
                in_code = False
            else:
                lang = line[3:].strip()
                out.append(f'<pre><code class="language-{lang}">')
                in_code = True
            continue
        if in_code:
            out.append(html.escape(line))
            continue
        
        # Admonition
        m = re.match(r'^!!!\s*(note|important|warning|caution|tip)', line)
        if m:
            out.append(f'<div class="admonition {m.group(1)}">')
            continue
        if line.strip() == '' and out and out[-1].startswith('<div class="admonition'):
            out.append('')
            continue
        
        # Headings
        m = re.match(r'^(#{1,6})\s+(.+)$', line)
        if m:
            level = len(m.group(1))
            out.append(f'<h{level}>{m.group(2).strip()}</h{level}>')
            continue
        
        # Horizontal rule
        if re.match(r'^---+\s*$', line) or re.match(r'^\*\*\*+\s*$', line):
            out.append('<hr>')
            continue
        
        # Bullet list
        if re.match(r'^\s*[-*]\s', line):
            text = re.sub(r'^\s*[-*]\s', '', line)
            text = inline_md(text)
            out.append(f'<li>{text}</li>')
            continue
        
        # Numbered list
        if re.match(r'^\s*\d+\.\s', line):
            text = re.sub(r'^\s*\d+\.\s', '', line)
            text = inline_md(text)
            out.append(f'<li>{text}</li>')
            continue
        
        # Table
        if '|' in line and line.strip().startswith('|') and not line.strip().startswith('|---'):
            cells = [cell.strip() for cell in line.strip().split('|')[1:-1]]
            tag = 'th' if out and '<tr>' in out[-1] and '---' in lines[max(0,lines.index(line)-1)] else 'td'
            out.append(f'<tr>{"".join(f"<{tag}>{html.escape(c)}</{tag}>" for c in cells)}</tr>')
            in_table = True
            continue
        if in_table and not line.strip():
            in_table = False
            continue
        
        # Paragraph
        stripped = line.strip()
        if stripped:
            out.append(f'<p>{inline_md(stripped)}</p>')
        else:
            out.append('')
    
    html_content = '\n'.join(out)
    # Fix .md links to .html
    html_content = re.sub(r'href="([^"]+?)\.md"', r'href="\1.html"', html_content)
    return html_content

def inline_md(text: str) -> str:
    text = html.escape(text)
    # Code
    text = re.sub(r'`([^`]+)`', r'<code>\1</code>', text)
    # Bold
    text = re.sub(r'\*\*(.+?)\*\*', r'<strong>\1</strong>', text)
    # Italic
    text = re.sub(r'\*(.+?)\*', r'<em>\1</em>', text)
    # Links [text](url)
    text = re.sub(r'\[([^\]]+)\]\(([^)]+)\)', r'<a href="\2">\1</a>', text)
    return text

def find_md(path: str) -> Path:
    if not path or path == '/':
        # Try common starting pages
        for candidate in ['modules/ROOT/pages/index', 'index', 'README', 'modules/end-user-guide/pages/index']:
            p = MD_DIR / candidate
            if p.with_suffix('.md').exists():
                return p.with_suffix('.md')
        return None
    p = MD_DIR / path
    if p.exists() and p.is_file():
        return p
    if p.with_suffix('.md').exists():
        return p.with_suffix('.md')
    if (p / 'index.md').exists():
        return p / 'index.md'
    if (p / 'README.md').exists():
        return p / 'README.md'
    return None

class Handler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        path = unquote(urlparse(self.path).path).lstrip('/')
        
        # Static files
        if '.' in path.split('/')[-1] and not path.endswith('.html'):
            static_path = str(MD_DIR / path)
            if os.path.exists(static_path) and os.path.isfile(static_path):
                with open(static_path, 'rb') as f:
                    data = f.read()
                ct, _ = mimetypes.guess_type(path)
                self.send_response(200)
                self.send_header('Content-Type', ct or 'application/octet-stream')
                self.end_headers()
                self.wfile.write(data)
                return
            self.send_error(404)
            return
        
        # Remove .html extension for finding the .md file
        if path.endswith('.html'):
            path = path[:-5]
        if path == '' or path == 'index':
            md_file = find_md('')
            if md_file:
                rel = md_file.relative_to(MD_DIR).as_posix()
                self.send_response(302)
                self.send_header('Location', '/' + rel.replace('.md', '.html'))
                self.end_headers()
                return
            self.send_error(404)
            return
        
        md_file = find_md(path)
        if not md_file:
            self.send_error(404)
            return
        
        try:
            text = md_file.read_text('utf-8')
        except Exception:
            self.send_error(500)
            return
        
        rel = md_file.relative_to(MD_DIR).as_posix()
        title = md_file.stem.replace('-', ' ').replace('_', ' ').title()
        first_line = text.strip().split('\n')[0] if text.strip() else ''
        if first_line.startswith('# '):
            title = first_line[2:].strip()
        
        tree = build_tree(MD_DIR)
        nav_html = render_nav(tree, rel)
        content_html = render_md(text)
        
        page = HEAD.format(title=title, nav=nav_html, content=content_html)
        self.send_response(200)
        self.send_header('Content-Type', 'text/html; charset=utf-8')
        self.end_headers()
        self.wfile.write(page.encode('utf-8'))

if __name__ == '__main__':
    server = http.server.HTTPServer(('0.0.0.0', PORT), Handler)
    print(f'Docs server on http://0.0.0.0:{PORT}')
    server.serve_forever()