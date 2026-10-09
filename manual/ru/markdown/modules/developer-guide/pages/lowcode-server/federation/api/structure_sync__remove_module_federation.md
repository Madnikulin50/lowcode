.Used variables
```bash
```
# Base url for the federation api
$BASE_URL

# JWT of the user
$JWT

# Node id of the destination node (?exposed) or the origin node (?shared)
$NODE_ID

# Federation module id
$MODULE_ID

.Example request
```bash
```
curl -X DELETE "$BASE*URL/federation/nodes/$NODE*ID/modules/$MODULE_ID" \
  -H "authorization: Bearer $JWT";

.Example response
```bash
```
{}
