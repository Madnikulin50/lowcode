#!/usr/bin/env python3
"""
AsciiDoc → Markdown converter for Antora docs (fixed).
"""

import os, re, sys
from pathlib import Path
from functools import lru_cache

RU_SRC = Path('/home/maxim/work/lowcode/manual/ru/src')
OUT_DIR = Path('/home/maxim/work/lowcode/manual/ru/markdown')

IGNORED_ATTRS = {'page-aliases', 'description', 'keywords', 'page-noindex', 'page-latest',
                 'set:page-noindex', 'experimental', 'PAGE-VERSION', 'VERSION',
                 'CLI_CMD_VIA_DOCKER', 'CLI_CMD_PREFIX', 'CLI_CMD', 'CLI_CMD_SYSTEM',
                 'CLI_CMD_COMPOSE', 'API_DOMAIN', 'API_AUTH_BASE_URL', 'API_SYSTEM_BASE_URL',
                 'API_COMPOSE_BASE_URL', 'API_FEDERATION_BASE_URL',
                 'APP_DOMAIN', 'APP_MESSAGING_BASE_URL', 'APP_COMPOSE_BASE_URL',
                 'APP_ADMIN_BASE_URL', 'GIT_REPO_LINK_PREFIX', 'GIT_MONOREPO_LINK',
                 'GIT_MONOREPO_ROOT_LINK_PREFIX', 'LOCAL_HOSTNAME',
                 'LOWCODE_PULL_BASE', 'LOWCODE_COMMIT_BASE',
                 'SERVER_COMMIT_BASE', 'SERVER_PULL_BASE',
                 'SERVER_DISCOVERY_COMMIT_BASE', 'SERVER_CORREDOR_COMMIT_BASE',
                 'LOCALE_COMMIT_BASE', 'WEBAPP_COMPOSE_COMMIT_BASE',
                 'WEBAPP_COMPOSE_PULL_BASE', 'WEBAPP_ONE_COMMIT_BASE',
                 'WEBAPP_ADMIN_COMMIT_BASE', 'WEBAPP_WORKFLOW_COMMIT_BASE',
                 'WEBAPP_REPORTER_COMMIT_BASE', 'WEBAPP_DISCOVERY_COMMIT_BASE',
                 'WEBAPP_PRIVACY_COMMIT_BASE', 'WEBAPP_JS_COMMIT_BASE',
                 'JS_COMMIT_BASE', 'JS_PULL_BASE', 'WEBAPP_VUE_COMMIT_BASE',
                 'VUE_COMMIT_BASE', 'APP_NAME_SHELL', 'APP_NAME_AUTH',
                 'APP_NAME_ADMIN', 'APP_NAME_COMPOSE', 'APP_NAME_REPORTER',
                 'APP_NAME_WORKFLOW', 'APP_NAME_FEDERATION', 'APP_NAME_DISCOVERY',
                 'APP_AUTOMATION', 'PRODUCT_NAME', 'DOMAIN'}

@lru_cache(None)
def load_attributes():
    attrs = {}
    for vf in RU_SRC.rglob('variables.adoc'):
        for line in open(vf, encoding='utf-8'):
            m = re.match(r'^:([A-Z_][A-Z_0-9]*):\s*(.*)', line)
            if m:
                attrs[m.group(1)] = m.group(2).strip()
    return attrs

ATTRS = load_attributes()

def find_module_file(module, kind, filepath):
    """Find file in module structure: kind = partial, pages, images etc."""
    # Try exact kind
    c1 = RU_SRC / 'modules' / module / kind / filepath
    if c1.exists():
        return c1
    # Try with 's' (partials, pages, images)
    c2 = RU_SRC / 'modules' / module / f'{kind}s' / filepath
    if c2.exists():
        return c2
    # Try without trailing 's'  (pages→page)
    if kind.endswith('s'):
        c3 = RU_SRC / 'modules' / module / kind[:-1] / filepath
        if c3.exists():
            return c3
    return None

def resolve_include(inc_directive, current_file):
    """Resolve include:: directive to file path."""
    m = re.match(r'include::(.+?)\[.*?\]', inc_directive)
    if not m:
        return None
    raw = m.group(1).strip()
    
    # Antora style: module:partial$path
    if ':' in raw and '$' in raw:
        mod, rest = raw.split(':', 1)
        if '$' in rest:
            kind, path = rest.split('$', 1)
            return find_module_file(mod, kind, path)
    
    # Relative: ./path or just path
    if raw.startswith('{'):
        return None
    cand = (current_file.parent / raw).resolve()
    if cand.exists():
        return cand
    
    # Module relative: partial$path (from current module)
    if '$' in raw:
        kind, path = raw.split('$', 1)
        # Find current module
        for mod_root in [RU_SRC / 'modules' / d for d in os.listdir(RU_SRC / 'modules')]:
            if str(current_file).startswith(str(mod_root)):
                return find_module_file(mod_root.name, kind, path)
    
    return None

def expand_attrs(text):
    """Replace {ATTRIBUTE} with values, keep unknown ones as-is."""
    def repl(m):
        key = m.group(1)
        if key in ATTRS:
            return ATTRS[key]
        return m.group(0)
    return re.sub(r'\{(\w+)\}', repl, text)

def convert_xref(match, current_file):
    """Convert xref:path[text] to markdown link."""
    target = match.group(1).strip()
    raw_text = match.group(2).strip()
    
    text = raw_text
    anchor = ''
    if '#' in target:
        target, anchor = target.split('#', 1)
    
    # Resolve module:page style
    if ':' in target:
        parts = target.split(':')
        if len(parts) == 2:
            mod, page = parts
            tgt = RU_SRC / 'modules' / mod / 'pages' / page
            if tgt.exists():
                rel = tgt.relative_to(RU_SRC)
                url = rel.with_suffix('.md').as_posix()
            else:
                url = f'{mod}/{page}'
                url = url.replace('.adoc', '.md')
        else:
            url = target.replace('.adoc', '.md')
    else:
        tgt = (current_file.parent / target).resolve()
        try:
            rel = tgt.relative_to(RU_SRC)
            url = rel.with_suffix('.md').as_posix()
        except ValueError:
            url = target.replace('.adoc', '.md')
    
    if anchor:
        url += f'#{anchor.lower().replace(".", "")}'
    
    if text:
        return f'[{text}]({url})'
    else:
        # Use the page filename as name; for Antora module:page format use page part
        last = target.rsplit(':', 1)[-1] if ':' in target else target
        name = last.split('/')[-1].replace('.adoc', '').replace('-', ' ').title()
        if name.lower() == 'index':
            # Try to use parent dir name
            pp = target.replace(':', '/')
            ps = pp.split('/')
            if len(ps) > 1:
                name = ps[-2].replace('-', ' ').title()
            else:
                name = 'Overview'
        return f'[{name}]({url})'

def convert_file(adoc_path):
    rel = adoc_path.relative_to(RU_SRC)
    out_path = OUT_DIR / rel.with_suffix('.md')
    out_path.parent.mkdir(parents=True, exist_ok=True)
    
    lines = _do_convert(adoc_path, out_path)
    text = '\n'.join(lines)
    text = re.sub(r'\n{4,}', '\n\n\n', text)
    text = text.strip() + '\n'
    
    open(out_path, 'w', encoding='utf-8').write(text)
    return out_path

def _do_convert(adoc_path, out_path, depth=0):
    """Recursive converter. Returns list of output lines."""
    if depth > 15:
        return []
    
    text = open(adoc_path, encoding='utf-8').read()
    lines = text.split('\n')
    out = []
    
    in_source = False
    in_table = False
    admonition_queue = []  # (type, start_idx, lines)
    in_admon = False
    admon_type = None
    admon_lines = []
    admon_got_fence = False
    admon_got_fence = False
    
    i = 0
    while i < len(lines):
        line = lines[i]
        ls = line.strip()
        
        if ls.startswith('include::') and not in_admon:
            fp = resolve_include(ls, adoc_path)
            if fp and fp.exists():
                included = _do_convert(fp, out_path, depth+1)
                out.extend(included)
            else:
                out.append(f'<!-- include: {ls} -->')
            i += 1
            continue
        
        if re.match(r'^\[source(?:,\w+)?\]$', ls):
            lang = ''
            m = re.match(r'\[source(?:,(\w+))?\]', ls)
            if m and m.group(1):
                lang = m.group(1)
            out.append(f'```{lang}')
            in_source = True
            i += 1
            continue
        
        if in_source:
            if ls == '----':
                out.append('```')
                in_source = False
            else:
                out.append(line.rstrip())
            i += 1
            continue
        
        # Admonition start
        admon_start = re.match(r'^\[(NOTE|IMPORTANT|CAUTION|WARNING|TIP)\]$', ls)
        if not in_admon and admon_start:
            admon_type = admon_start.group(1)
            in_admon = True
            admon_lines = []
            i += 1
            continue
        
        if in_admon:
            if ls == '====':
                if not admon_got_fence:
                    admon_got_fence = True
                    i += 1
                    continue
                in_admon = False
                admon_got_fence = False
                md_type = {'NOTE': 'note', 'IMPORTANT': 'important',
                           'CAUTION': 'caution', 'WARNING': 'warning',
                           'TIP': 'tip'}[admon_type]
                out.append(f'!!! {md_type}')
                for al in admon_lines:
                    al2 = expand_attrs(al)
                    al2 = re.sub(r'xref:([^\[]+)\[(.*?)\]', lambda m: convert_xref(m, adoc_path), al2)
                    out.append(f'    {al2}')
                out.append('')
                i += 1
                continue
            if ls.startswith('include::'):
                fp = resolve_include(ls, adoc_path)
                if fp and fp.exists():
                    inc_lines = _do_convert(fp, out_path, depth+1)
                    for il in inc_lines:
                        admon_lines.append(il)
                else:
                    admon_lines.append(f'<!-- unresolved include: {ls} -->')
                i += 1
                continue
            admon_lines.append(line.rstrip())
            i += 1
            continue
        
        # Skip attribute definitions (including page-aliases, description, keywords)
        if re.match(r'^:(page-aliases|description|keywords|[A-Z_])', ls):
            i += 1
            continue
        
        # Skip stray delimiter fences
        if ls == '====' or ls == '----':
            i += 1
            continue
        
        # Comments
        if ls.startswith('// @todo'):
            out.append(f'<!-- TODO: {ls[7:].strip()} -->')
            i += 1
            continue
        if ls.startswith('//'):
            i += 1
            continue
        
        # Empty lines pass through
        if not ls:
            out.append('')
            i += 1
            continue
        
        # Process line
        pl = line  # processed line
        
        # Expand attributes
        pl = expand_attrs(pl)
        
        # Headings
        hm = re.match(r'^(={1,6})\s+(.+)$', pl)
        if hm:
            level = len(hm.group(1))
            htext = hm.group(2).strip()
            out.append(f"{'#' * level} {htext}")
            i += 1
            continue
        
        # xref: with display text
        pl = re.sub(r'xref:([^\[]+)\[(.*?)\]', lambda m: convert_xref(m, adoc_path), pl)
        
        # <<anchor>>
        pl = re.sub(r'<<([^>]+)>>', r'[\1](#\1)', pl)
        
        # [[anchor]] / [#anchor]
        am = re.match(r'^\[#([\w-]+)\]$', pl.strip()) or re.match(r'^\[\[([\w-]+)\]\]$', pl.strip())
        if am:
            out.append(f'<a id="{am.group(1)}"></a>')
            i += 1
            continue
        
        # image::
        pl = re.sub(r'image::([^\[]+)\[([^\]]*)\]', lambda m: f'![{m.group(2).split(",")[0].strip()}]({m.group(1).strip()})', pl)
        pl = re.sub(r'image:([^\[]+)\[([^\]]*)\]', lambda m: f'![{m.group(2).split(",")[0].strip()}]({m.group(1).strip()})', pl)
        
        # link:
        pl = re.sub(r'link:([^\[]+)\[([^\]]*)\]', lambda m: f'[{m.group(2).strip()}]({m.group(1).strip()})', pl)
        
        # URL[text]
        pl = re.sub(r'(https?://[^\s<\[\]]+)\[([^\]]*)\]', lambda m: f'[{m.group(2).strip()}]({m.group(1).strip()})', pl)
        
        # Bold/italic
        pl = re.sub(r'\*(\S.*?\S)\*', r'**\1**', pl)
        pl = re.sub(r'_(\S.*?\S)_', r'*\1*', pl)
        pl = re.sub(r'``([^`]+)``', r'`\1`', pl)
        
        # Numbered lists: . item → 1. item
        pl = re.sub(r'^\.(\s+)', r'1.\1', pl)
        
        # Bullet lists
        pl = re.sub(r'^(\s*)\*\s', r'\1- ', pl)
        
        out.append(pl)
        i += 1
    
    return out

def main():
    total = 0
    errors = 0
    
    for adoc_file in sorted(RU_SRC.rglob('*.adoc')):
        if 'archive' in str(adoc_file):
            continue
        try:
            out = convert_file(adoc_file)
            total += 1
            if total % 100 == 0:
                print(f'  {total}...', file=sys.stderr)
        except Exception as e:
            errors += 1
            print(f'ERROR: {adoc_file.relative_to(RU_SRC)}: {e}', file=sys.stderr)
    
    print(f'Done: {total} files converted, {errors} errors', file=sys.stderr)

if __name__ == '__main__':
    main()