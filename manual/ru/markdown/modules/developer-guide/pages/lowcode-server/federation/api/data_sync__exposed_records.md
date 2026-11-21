!!! note
    $TOKEN_B is the token that was generated during the handshake and is used to authenticate the user on the Origin node (the one who shares the data) by the Destination node.


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

# Node B auth token
$TOKEN_B

.Example request
```bash
```
curl -X GET "$BASE*URL/federation/nodes/$NODE*ID/modules/$MODULE*ID/records?lastSync=$AFTER*TIMESTAMP" \
  -H "Authorization: Bearer $TOKEN_B";

.Example response
```bash
```
{
    "response": {
        "filter": {
            "moduleID": "132954639472525355",
            "query": "",
            "sort": "createdAt DESC",
            "page": 1,
            "perPage": 20,
            "count": 97,
            "deleted": 0
        },
        "set": [
            {
                "recordID": "$COMPOSE*RECORD*ID",
                "moduleID": "$FEDERATION*MODULE*ID",
                "values": [
                    {
                        "name": "name",
                        "value": "John"
                    },
                    {
                        "name": "surname",
                        "value": "Doe"
                    }
                ],
                "createdAt": "2020-09-08T19:56:14Z",
                "updatedAt": "2020-09-09T18:05:33Z"
            }
        ]
    }
}
