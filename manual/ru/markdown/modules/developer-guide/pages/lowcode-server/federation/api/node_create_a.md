.Used variables
```bash
```
# Base URL of node A api
$API_A_BASE

# Main administrator JWT for node A
$MAIN*JWT*A

# Node A domain
$DOMAIN_A

# Node B domain
$DOMAIN_B

# Node name
$NODE_NAME

# Node A nodeID
$NODE*ID*A

# Node B nodeID
$NODE*ID*B

# Node URI
$NODE_URI

.Example request
```bash
```
curl -X POST "$API_A_BASE/federation/nodes/" \
  -H "authorization: Bearer $MAIN*JWT*A" \
  --header "Content-Type: application/json" \
  --data "{
    \"baseURL\": \"$DOMAIN*A_BASE*URL\",
    \"name\": \"$DOMAIN_A_NAME\"
}";

.Example response
```bash
```
{
  "response": {
    "nodeID": "$NODE*ID*A",
    "name": "$DOMAIN_A_NAME",
    "status": "pending",
    "baseURL": "$DOMAIN*A_BASE*URL",
    "sharedNodeID": "$NODE*ID*A",
    "createdAt": "2020-12-01T14:24:47.246145938Z",
    "createdBy": "0"
  }
}
