.Used variables
```bash
```
# Base URL of node B api
$API_B_BASE

# Node B nodeID
$NODE*ID*B

# Node B auth token
$TOKEN_B

# Node A auth token
$TOKEN_A

.Example request
```bash
```
curl -X POST "$API*B_BASE/federation/nodes/$NODE*ID_B/handshake-complete" \
  -H "authorization: Bearer $TOKEN_B" \
  --header "Content-Type: application/json" \
  --data "{
    \"token\": \"$TOKEN_A\"
  }";

.Example response
```bash
```
{}
