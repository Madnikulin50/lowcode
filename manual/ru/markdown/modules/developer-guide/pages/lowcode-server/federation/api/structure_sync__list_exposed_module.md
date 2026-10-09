.Used variables
```bash
```
# Base url for the federation api
$BASE_URL

# JWT of the user
$JWT

# Node id of the destination node
$NODE_ID

.Example request
```bash
```
curl -X GET "$BASE*URL/federation/nodes/$NODE*ID/modules/$MODULE_ID/exposed" \
  -H "Authorization: Bearer $JWT";

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
        "createdBy": "204158548916043781",
        "updatedAt": "2020-12-01T14:34:13Z",
        "updatedBy": "204158548916043781"

    }
}
