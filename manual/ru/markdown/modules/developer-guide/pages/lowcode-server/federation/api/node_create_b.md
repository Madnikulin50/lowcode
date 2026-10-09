.Used variables
```bash
```
# Base URL of node B api
$API_B_BASE

# Main administrator JWT for node B
$MAIN*JWT*B

# Node B domain
$DOMAIN_B

# Node A domain
$DOMAIN_A

# Node name
$NODE_NAME

# Node B nodeID
$NODE*ID*B

# Node A nodeID
$NODE*ID*A

# Node URI
$NODE_URI

.Example request
```bash
```
curl -X POST "$API_B_BASE/federation/nodes" \
  -H "authorization: Bearer $MAIN*JWT*B" \
  --header "Content-Type: application/json" \
  --data "{
    \"baseURL\": \"$DOMAIN*B_BASE*URL\",
    \"name\": \"$DOMAIN_B_NAME\"
}";

.Example response
```bash
```
{
  "response": {
    "nodeID": "$NODE*ID*B",
    "name": "$DOMAIN_B_NAME",
    "status": "pending",
    "baseURL": "$DOMAIN*B_BASE*URL",
    "sharedNodeID": "$NODE*ID*A",
    "createdAt": "2020-12-01T14:24:47.246145938Z",
    "createdBy": "0"
  }
}
