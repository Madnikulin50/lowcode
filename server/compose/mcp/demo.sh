#!/bin/bash
# End-to-end demo: lowcode AI platform
# Requires: server running on localhost with API proxy

BASE="${BASE:-http://localhost:8080/api/compose}"
AUTH="${AUTH:-}"

echo "=== Lowcode AI Platform — End-to-End Demo ==="
echo ""

# 1. Health check
echo ">>> 1. Health check"
curl -s "$BASE/health" | python3 -m json.tool 2>/dev/null || curl -s "$BASE/health"
echo ""

# 2. List available demo chains
echo ">>> 2. List rule chains"
curl -s "$BASE/rulechain/" | python3 -m json.tool 2>/dev/null || curl -s "$BASE/rulechain/"
echo ""

# 3. Test demo_welcome_email chain
echo ">>> 3. Test demo_welcome_email"
curl -s -X POST "$BASE/rulechain/demo_welcome_email/run" \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}' | python3 -m json.tool 2>/dev/null
echo ""

# 4. Test demo_lead_scoring chain (hot lead)
echo ">>> 4. Test demo_lead_scoring (hot lead: budget=15000)"
curl -s -X POST "$BASE/rulechain/demo_lead_scoring/run" \
  -H "Content-Type: application/json" \
  -d '{"name":"Acme Corp","company":"Acme","budget":"15000","email":"lead@acme.com"}' | python3 -m json.tool 2>/dev/null
echo ""

# 5. Run AI script
echo ">>> 5. Run a JS script: normalize data"
curl -s -X POST "$BASE/ai/script/run" \
  -H "Content-Type: application/json" \
  -d '{"script":"var name = (context.name || \"\").trim(); var email = (context.email || \"\").toLowerCase(); runtime.log.info(\"Normalized: \" + name + \" / \" + email); ({name: name, email: email})","input":{"name":"  John Doe  ","email":"JOHN@EXAMPLE.COM"}}' | python3 -m json.tool 2>/dev/null
echo ""

# 6. Admin API — list node types
echo ">>> 6. Available node types"
curl -s "$BASE/admin/rulechain/nodes" | python3 -c "import json,sys; data=json.load(sys.stdin); [print(f'  {n[\"type\"]:12s} — {n[\"label\"]}') for n in data.get('nodes',[])]" 2>/dev/null
echo ""

# 7. Admin API — stats
echo ">>> 7. System stats"
curl -s "$BASE/admin/rulechain/stats" | python3 -m json.tool 2>/dev/null
echo ""

# 8. Pageblock trigger
echo ">>> 8. PageBlock trigger with record context"
curl -s -X POST "$BASE/pageblock/trigger" \
  -H "Content-Type: application/json" \
  -d '{"chainID":"demo_data_cleanup","pageID":1,"blockID":"button_1","recordID":"123","record":{"name":"  Dirty Data  ","email":"DIRTY@example.com"}}' | python3 -m json.tool 2>/dev/null
echo ""

# 9. Full health info
echo ">>> 9. System info"
curl -s "$BASE/health/info" | python3 -m json.tool 2>/dev/null
echo ""

echo "=== Demo complete ==="
