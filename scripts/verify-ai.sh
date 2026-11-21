#!/usr/bin/env bash
#
# AI change guard - быстрый таргетированный прогон проверок после правок ИИ.
#
#   scripts/verify-ai.sh [--staged|--changed|--ci|--full|--watch[=N]]
#                        [--strict] [--update-baseline] [--verbose]
#
# Режимы выбора файлов:
#   --staged   только проиндексированные файлы (git pre-commit)
#   --changed  staged + unstaged + untracked (по умолчанию)
#   --ci       изменения относительно merge-base с origin/HEAD (или HEAD~1)
#   --full     игнорировать выбор: полная сборка, тесты и адвайзори-проверки
#   --hygiene  только гигиена всего дерева (маркеры/.only/.orig), без сборки
#   --watch    периодически (по умолчанию каждые 5с) повторять --changed
#
# Блокирующие гейты (только НОВЫЕ проблемы, см. scripts/ai-guard-baseline.txt):
#   конфликтные маркеры/.only, сборка server/cmd/corteza, go vet,
#   go test изменённых пакетов, vite build затронутых webapp.
# Предупреждения (не блокируют, если не задан --strict):
#   go build ./... для мёртвых пакетов, eslint, vitest/mocha в lib.
#
#   --update-baseline  записать текущие падения в baseline (принять как известные)
#   --strict           сделать предупреждения блокирующими
#
# Аварийный обход: SKIP_AI_GUARD=1
#
# Подробное описание - в docs/ai-guard.md

set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT" || exit 2

MODE="changed"
STRICT=0
VERBOSE=0
UPDATE_BASELINE=0
WATCH_INTERVAL=5

GO_DIR="server"
WEB_DIR="client3/web"
LIB_DIR="client3/lib"

# Никогда не проверяем сгенерированное/скачанное: vendor, dist и сборка
# документации лежат в дереве незакоммиченными (server/vendor, manual/,
# client3/web/*/dist) и дают тысячи ложных путей.
EXCLUDE_RE='(^|/)(vendor|node_modules|dist|cdist|\.git|\.vite|__pycache__|manual|projectFilesBackup)(/|$)'
LOCKFILE_RE='/(yarn\.lock|package-lock\.json)$'

FAILED=()
WARNED=()
LOG_DIR="$(mktemp -d "${TMPDIR:-/tmp}/ai-guard.XXXXXX")"
trap 'rm -rf "$LOG_DIR"' EXIT

# --- вывод -------------------------------------------------------------------
if [ -t 1 ] && command -v tput >/dev/null 2>&1 && [ "$(tput colors 2>/dev/null || echo 0)" -ge 8 ]; then
  C_RED="$(tput setaf 1)"; C_GREEN="$(tput setaf 2)"; C_YELLOW="$(tput setaf 3)"
  C_BLUE="$(tput setaf 4)"; C_DIM="$(tput setaf 8)"; C_OFF="$(tput sgr0)"
else
  C_RED=""; C_GREEN=""; C_YELLOW=""; C_BLUE=""; C_DIM=""; C_OFF=""
fi

section() { printf '\n%s==> %s%s\n' "$C_BLUE" "$1" "$C_OFF"; }
ok()   { printf '%s  [OK]%s   %s\n' "$C_GREEN" "$C_OFF" "$1"; }
note() { printf '%s  [--]%s   %s\n' "$C_DIM" "$C_OFF" "$1"; }
warn() { printf '%s  [WARN]%s %s\n' "$C_YELLOW" "$C_OFF" "$1"; WARNED+=("$1"); }
fail() { printf '%s  [FAIL]%s %s\n' "$C_RED" "$C_OFF" "$1"; FAILED+=("$1"); }
indent() { sed 's/^/         /'; }

# --- аргументы ---------------------------------------------------------------
while [ $# -gt 0 ]; do
  case "$1" in
    --staged)      MODE="staged" ;;
    --changed)     MODE="changed" ;;
    --ci)          MODE="ci" ;;
    --full)        MODE="full" ;;
    --hygiene)     MODE="hygiene" ;;
    --watch)       MODE="watch" ;;
    --watch=*)     MODE="watch"; WATCH_INTERVAL="${1#*=}" ;;
    --strict)          STRICT=1 ;;
    --update-baseline) UPDATE_BASELINE=1 ;;
    --verbose|-v)      VERBOSE=1 ;;
    -h|--help)     sed -n '3,27p' "$0"; exit 0 ;;
    *) echo "неизвестный аргумент: $1" >&2; exit 2 ;;
  esac
  shift
done

if [ "${SKIP_AI_GUARD:-0}" = "1" ]; then
  echo "SKIP_AI_GUARD=1 - проверки пропущены"
  exit 0
fi

# --- выбор файлов ------------------------------------------------------------
ci_files() {
  local base=""
  base="$(git merge-base HEAD origin/HEAD 2>/dev/null || true)"
  [ -n "$base" ] || base="$(git rev-parse HEAD~1 2>/dev/null || true)"
  if [ -n "$base" ]; then
    git diff --name-only --diff-filter=ACMR "$base"...HEAD
  else
    git ls-files
  fi
}

collect_files() {
  {
    case "$MODE" in
      staged)  git diff --cached --name-only --diff-filter=ACMR ;;
      changed) git diff --cached --name-only --diff-filter=ACMR
               git diff --name-only --diff-filter=ACMR
               git ls-files --others --exclude-standard ;;
      ci)      ci_files ;;
      full)    git ls-files; git ls-files --others --exclude-standard ;;
      hygiene) git ls-files; git ls-files --others --exclude-standard ;;
      watch)   git diff --cached --name-only --diff-filter=ACMR
               git diff --name-only --diff-filter=ACMR
               git ls-files --others --exclude-standard ;;
    esac
  } 2>/dev/null | sort -u | grep -Ev "$EXCLUDE_RE" | grep -v '^$' || true
}

# Отпечаток рабочего дерева - чтобы watch не гонял проверки впустую.
tree_fingerprint() {
  {
    git status --porcelain -uall
    git ls-files -m -o --exclude-standard -z 2>/dev/null | xargs -0 -r stat -c '%n:%s:%Y' 2>/dev/null
  } | md5sum | cut -d' ' -f1
}

# --- проверки -----------------------------------------------------------------

TEXT_RE='\.(go|js|mjs|cjs|ts|tsx|vue|json|ya?ml|md|sh|css|scss|html|sql|toml|template)$'

check_hygiene() {
  local files="$1" f hit=0

  while IFS= read -r f; do
    [ -n "$f" ] && [ -f "$f" ] || continue
    case "$f" in */node_modules/*|*/vendor/*|*/dist/*|*/cdist/*) continue ;; esac
    printf '%s' "$f" | grep -Eq "$TEXT_RE" || continue

    if grep -nE '^<{7}( |$)|^>{7}( |$)' "$f" >/dev/null 2>&1; then
      fail "незавершённый merge (конфликтные маркеры): $f"
      grep -nE '^<{7}( |$)|^>{7}( |$)' "$f" | head -3 | indent
      hit=1
    fi

    case "$f" in
      *.spec.js|*.spec.ts|*.test.js|*.test.ts|*/test/*|*/tests/*)
        if grep -nE '(describe|it|test)\.only\(|(^|[^A-Za-z0-9_])f(describe|it)\(' "$f" >/dev/null 2>&1; then
          fail "забытый .only/fit (отключит остальные тесты): $f"
          hit=1
        fi ;;
    esac
  done <<< "$files"

  while IFS= read -r f; do
    [ -n "$f" ] || continue
    fail "остаток merge/патча: $f"; hit=1
  done < <(git ls-files --others --exclude-standard 2>/dev/null | grep -E '\.(orig|rej)$' || true)

  [ "$hit" -eq 0 ] && ok "гигиена: конфликтных маркеров, .only и .orig/.rej нет"
}

# server/a/b/x.go -> ./a/b ; server/main.go -> .
go_packages() {
  printf '%s\n' "$1" | grep -E "^${GO_DIR}/.*\.go$" \
    | awk -F/ -v pre="${GO_DIR}/" '{
        n = $0; sub("^" pre, "", n);
        if (n ~ /\//) { sub("/[^/]+$", "", n); print "./" n } else { print "." }
      }' | sort -u
}

# --- baseline известных проблем ----------------------------------------------
# Baseline нужен, потому что репозиторий не «зелёный»: часть замечаний go vet и
# часть тестов падают ещё до правок ИИ. Регрессией считаем только то, чего в
# baseline нет, - иначе страж превращается в шум и его начинают игнорировать.

BASELINE_FILE="${ROOT}/scripts/ai-guard-baseline.txt"
NEW_PROBLEMS=()

baseline_has() {
  [ -f "$BASELINE_FILE" ] && grep -qxF -- "$1" "$BASELINE_FILE"
}

# "path/file.go:12:34: msg" -> "path/file.go: msg" (номера строк сдвигаются при правках)
norm_vet() { sed -E 's/^([^:]+\.go):[0-9]+:[0-9]+: /\1: /'; }

vet_problems() {
  local lines
  lines="$(grep -E '^[^ #]' "$1" | sort -u || true)"
  [ -n "$lines" ] || lines="$(head -3 "$1" | tr '\n' ' ')"
  printf '%s\n' "$lines" | grep -v '^$' || true
}

test_problems() {
  local names
  names="$(grep -oE '^--- FAIL: [^ ]+' "$1" | awk '{print $3}' | sort -u || true)"
  [ -n "$names" ] || names="<panic>"
  printf '%s\n' "$names"
}

check_go() {
  local files="$1" pkg_list=() pkgs p key line fresh=0 failed_pkgs=0

  section "Go"

  if ( cd "$GO_DIR" && go build -o "$LOG_DIR/corteza-artifact" ./cmd/corteza ) >"$LOG_DIR/go-build-artifact.log" 2>&1; then
    ok "сборка server/cmd/corteza"
  else
    fail "server/cmd/corteza не собирается"
    head -25 "$LOG_DIR/go-build-artifact.log" | indent
  fi

  pkgs="$(go_packages "$files")"
  while IFS= read -r p; do [ -n "$p" ] && pkg_list+=("$p"); done <<< "$pkgs"

  if [ "${#pkg_list[@]}" -eq 0 ]; then
    note "изменённых Go-пакетов нет"
    return
  fi

  note "затронутых Go-пакетов: ${#pkg_list[@]}"
  [ "$VERBOSE" -eq 1 ] && printf '%s\n' "${pkg_list[@]}" | indent

  if ( cd "$GO_DIR" && go vet "${pkg_list[@]}" ) >"$LOG_DIR/go-vet.log" 2>&1; then
    ok "go vet изменённых пакетов"
  else
    while IFS= read -r line; do
      [ -n "$line" ] || continue
      key="vet $(printf '%s' "$line" | norm_vet)"
      if baseline_has "$key"; then
        note "go vet (уже известное): ${key#vet }"
      else
        fail "go vet (новое): $line"
        NEW_PROBLEMS+=("$key")
        fresh=$((fresh + 1))
      fi
    done < <(vet_problems "$LOG_DIR/go-vet.log")
    [ "$fresh" -gt 0 ] || ok "go vet: новых замечаний нет"
  fi

  for p in "${pkg_list[@]}"; do
    if ( cd "$GO_DIR" && go test -count=1 -timeout 180s "$p" ) >"$LOG_DIR/go-test.log" 2>&1; then
      continue
    fi
    failed_pkgs=$((failed_pkgs + 1))
    while IFS= read -r line; do
      [ -n "$line" ] || continue
      key="test $p $line"
      if baseline_has "$key"; then
        note "go test (уже известное): $p $line"
      else
        fail "go test (новое): $p $line"
        grep -E '^--- FAIL|^panic:|^ +[a-zA-Z0-9_/.]+\.go:[0-9]+' "$LOG_DIR/go-test.log" | head -8 | indent
        NEW_PROBLEMS+=("$key")
      fi
    done < <(test_problems "$LOG_DIR/go-test.log")
  done
  [ "$failed_pkgs" -eq 0 ] && ok "go test изменённых пакетов"

  # Мёртвые/несобираемые пакеты не блокируют, но о них надо знать.
  if [ "$MODE" = "full" ]; then
    if ( cd "$GO_DIR" && go build ./... ) >"$LOG_DIR/go-build-all.log" 2>&1; then
      ok "go build ./... (все пакеты)"
    else
      warn "go build ./...: часть пакетов не собирается (см. ниже)"
      grep -E '^# |undefined:|not found|declared and not used' "$LOG_DIR/go-build-all.log" | head -15 | indent
    fi
  fi
  return 0
}

# --- web ---------------------------------------------------------------------

web_apps() {
  printf '%s\n' "$1" | grep -E "^${WEB_DIR}/[^/]+/" | cut -d/ -f3 | sort -u
}

check_web() {
  local files="$1" apps app dir lint_files log build_log

  section "Web"

  apps="$(web_apps "$files")"
  if [ -z "$apps" ]; then
    note "изменений в ${WEB_DIR} нет"
    return
  fi

  while IFS= read -r app; do
    [ -n "$app" ] || continue
    dir="${WEB_DIR}/${app}"
    if [ ! -f "$dir/package.json" ]; then
      note "$app: не webapp, пропуск"
      continue
    fi

    lint_files="$(printf '%s\n' "$files" | grep -E "^${dir}/.*\.(js|mjs|cjs|ts|vue)$" \
      | sed "s#^${dir}/##" || true)"
    if [ -n "$lint_files" ] && [ -x "$dir/node_modules/.bin/eslint" ]; then
      log="$LOG_DIR/lint-${app}.log"
      if ( cd "$dir" && ./node_modules/.bin/eslint $(printf '%s\n' "$lint_files" | tr '\n' ' ') ) >"$log" 2>&1; then
        ok "$app: eslint изменённых файлов"
      elif grep -qE 'Cannot read config file|ES module scope|No ESLint configuration found|Couldn.t find config file' "$log"; then
        note "$app: eslint-конфиг не загружается, линт пропущен"
        grep -E 'Error:|ReferenceError:' "$log" | head -3 | indent
      elif [ "$STRICT" -eq 1 ]; then
        fail "$app: eslint изменённых файлов"
        head -20 "$log" | indent
      else
        warn "$app: eslint замечания в изменённых файлах"
        head -12 "$log" | indent
      fi
    elif [ -n "$lint_files" ]; then
      note "$app: eslint не установлен, линт пропущен"
    fi

    if [ -z "$(printf '%s\n' "$files" | grep -E "^${dir}/" | grep -Ev "$LOCKFILE_RE" || true)" ]; then
      note "$app: изменились только lock-файлы, сборка не запускается"
      continue
    fi

    if [ ! -d "$dir/node_modules" ]; then
      warn "$app: node_modules отсутствуют, сборка пропущена"
      continue
    fi

    build_log="$LOG_DIR/build-${app}.log"
    if ( cd "$dir" && yarn --silent build ) >"$build_log" 2>&1; then
      ok "$app: vite build"
    else
      fail "$app: vite build не проходит"
      tail -25 "$build_log" | indent
    fi
  done <<< "$apps"
}

check_lib() {
  local files="$1" lib log

  section "Lib"
  local any=0
  for lib in js vue; do
    printf '%s\n' "$files" | grep -q "^${LIB_DIR}/${lib}/" || continue
    [ -f "${LIB_DIR}/${lib}/package.json" ] || continue
    any=1
    log="$LOG_DIR/lib-${lib}.log"
    if ( cd "${LIB_DIR}/${lib}" && yarn --silent test:unit ) >"$log" 2>&1; then
      ok "${lib}: тесты"
    elif [ "$STRICT" -eq 1 ]; then
      fail "${lib}: тесты"
      tail -15 "$log" | indent
    else
      warn "${lib}: тесты падают (базовое состояние репозитория)"
      tail -8 "$log" | indent
    fi
  done
  [ "$any" -eq 1 ] || note "изменений в ${LIB_DIR} нет"
}

# --- прогон ------------------------------------------------------------------

run_once() {
  FAILED=(); WARNED=()
  local files n

  printf '\n%s### AI guard%s   mode=%s   %s\n' "$C_BLUE" "$C_OFF" "$MODE" "$(date '+%H:%M:%S')"

  files="$(collect_files)"
  n="$(printf '%s\n' "$files" | grep -c . || true)"

  if [ "$n" -eq 0 ]; then
    note "изменений нет, проверять нечего"
    return 0
  fi

  note "файлов в выборке: $n"
  [ "$VERBOSE" -eq 1 ] && printf '%s\n' "$files" | indent

  check_hygiene "$files"
  if [ "$MODE" != "hygiene" ]; then
    check_go "$files"
    check_web "$files"
    check_lib "$files"
  fi

  if [ "$UPDATE_BASELINE" -eq 1 ]; then
    section "Baseline"
    {
      printf '# AI guard baseline - известные проблемы, не считаются регрессией.\n'
      printf '# Обновление: make verify.baseline  (или scripts/verify-ai.sh --update-baseline)\n'
      if [ -f "$BASELINE_FILE" ]; then grep -v '^#' "$BASELINE_FILE" || true; fi
      if [ "${#NEW_PROBLEMS[@]}" -gt 0 ]; then printf '%s\n' "${NEW_PROBLEMS[@]}"; fi
    } | grep -v '^[[:space:]]*$' | sort -u > "$BASELINE_FILE.tmp"
    mv "$BASELINE_FILE.tmp" "$BASELINE_FILE"
    ok "baseline обновлён: $(grep -cv '^#' "$BASELINE_FILE" || true) записей"
    return 0
  fi

  section "Итог"
  if [ "${#FAILED[@]}" -eq 0 ]; then
    if [ "${#WARNED[@]}" -gt 0 ]; then
      ok "блокирующих проблем нет (предупреждений: ${#WARNED[@]})"
    else
      ok "всё чисто"
    fi
    return 0
  fi

  fail "блокирующих проблем: ${#FAILED[@]}"
  printf '%s\n' "${FAILED[@]}" | indent
  return 1
}

# --- watch / точка входа -----------------------------------------------------

watch_loop() {
  local last="" now
  echo "AI guard: watch-режим, интервал ${WATCH_INTERVAL}s, Ctrl+C для выхода"
  while true; do
    now="$(tree_fingerprint)"
    if [ "$now" != "$last" ]; then
      run_once || true
      last="$(tree_fingerprint)"
      printf '\n%sожидание изменений...%s\n' "$C_DIM" "$C_OFF"
    fi
    sleep "$WATCH_INTERVAL"
  done
}

if [ "$MODE" = "watch" ]; then
  watch_loop
else
  run_once
fi
