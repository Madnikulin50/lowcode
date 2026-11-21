.Used variables
```bash
```
# Base url for the federation api
$BASE_URL

# JWT of the user
$JWT

# Node id of the destination node
$NODE_ID

# Federation module id
$MODULE_ID

.Example request
```bash
```
curl -X PUT "$BASE*URL/federation/nodes/$NODE*ID/modules/" \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  --data "{
    \"composeModuleID\": \"$COMPOSE*MODULE*ID\",
    \"composeNamespaceID\": \"$COMPOSE*NAMESPACE*ID\",
    \"name\": \"Account\",
    \"handle\": \"Account\",
    \"fields\": [
        {
            \"kind\": \"String\",
            \"name\": \"AccountName\",
            \"label\": \"Account Name\",
            \"isMulti\": false,
            \"value\": true,
            \"map\": null
        },
        {
            \"kind\": \"User\",
            \"name\": \"OwnerId\",
            \"label\": \"Account Owner\",
            \"isMulti\": false,
            \"value\": true,
            \"map\": null
        }
    ]
}";

.Example response
```bash
```
{
    "response": {
        "moduleID": "$MODULE_ID",
        "nodeID": "$NODE_ID",
        "composeModuleID": "$COMPOSE*MODULE*ID",
        "composeNamespaceID": "$COMPOSE*NAMESPACE*ID",
        "handle": "Account",
        "name": "Account",
        "fields": [
            {
                "kind": "String",
                "name": "AccountName",
                "label": "Account Name",
                "isMulti": false
            },
            {
                "kind": "User",
                "name": "OwnerId",
                "label": "Account Owner",
                "isMulti": false
            }
        ],
        "createdAt": "2020-12-01T14:33:14.010034106Z",
        "createdBy": "204158548916043781"
    }
}
