.Used variables
```bash
```
# Base url for the federation api
$BASE_URL

# JWT of the user
$JWT

# Node id of the origin node
$NODE_ID

.Example request
```bash
```
curl -X GET "$BASE*URL/federation/nodes/$NODE*ID/modules/$MODULE_ID/shared" \
  -H "Authorization: Bearer $JWT";

.Example response
```bash
```
{
    "response": {
        "moduleID": "122709113267335170",
        "handle": "Account",
        "name": "Account",
        "createdAt": "2019-12-18T17:45:15Z",
        "updatedAt": "2020-05-26T13:29:36Z",
        "fields": [
            {
                "kind": "Url",
                "name": "LinkedIn",
                "label": "LinkedIn",
                "isMulti": false,
            },
            {
                "kind": "String",
                "name": "Phone",
                "label": "Phone",
                "isMulti": false,
            }
        ]
    }
}
