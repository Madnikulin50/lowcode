.Used variables
```bash
```
# Base URL of node B api
$API_B_BASE

# Main administrator JWT for node B
$MAIN*JWT*B

# Node B nodeID
$NODE*ID*B

.Example request
```bash
```
curl -X POST "$API*B_BASE/federation/nodes/$NODE*ID_B/pair" \
  -H "authorization: Bearer $MAIN*JWT*B" \
  --header "Content-Type: application/json";

.Example response
```bash
```
{}
