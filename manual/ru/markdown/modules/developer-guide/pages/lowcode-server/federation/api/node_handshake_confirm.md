.Used variables
```bash
```
# Base URL of node A api
$API_A_BASE

# Node A nodeID
$NODE*ID*A

# Main administrator JWT for node A
$MAIN*JWT*A

.Example request
```bash
```
curl -X POST "$API*A_BASE/federation/nodes/$NODE*ID_A/handshake-confirm" \
  -H "authorization: Bearer $MAIN*JWT*A" \
  --header "Content-Type: application/json";

.Example response
```bash
```
{}
