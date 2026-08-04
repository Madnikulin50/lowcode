.Used variables
```bash
```
# Base URL of node A api
$API_A_BASE

# Node A nodeID
$NODE*ID*A

# Node URI
$NODE_URI

# Node B auth token
$TOKEN_B

# Node B nodeID
$NODE*ID*B

.Example request
```bash
```
curl -X POST "$API*A_BASE/federation/nodes/$NODE*ID_A/handshake" \
  --header "Content-Type: application/json" \
  --data "{
    \"nodeURI\": \"$NODE_URI\",
    \"token\": \"$TOKEN_B\",
    \"nodeIDB\": \"$NODE*ID*B\"
  }";

.Example response
```bash
```
{}
