# API

## Сопряжение узлов

### Создание федеративного узла из набора параметров

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


### Создание федеративного узла из URI узла

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


### Инициализация рукопожатия

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


### Запрос рукопожатия с узлом A

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


### Подтверждение запрошенного рукопожатия

.Used variables
```bash
```
# Base URL of node A api
$API_A_BASE

# Node A nodeID
$NODE*ID*A

# Main administrator JWT for node A
$MAIN*JWT*A

.Example request
```bash
```
curl -X POST "$API*A_BASE/federation/nodes/$NODE*ID_A/handshake-confirm" \
  -H "authorization: Bearer $MAIN*JWT*A" \
  --header "Content-Type: application/json";

.Example response
```bash
```
{}


### Завершение рукопожатия

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


## Структуры источника

### Добавление модуля в федерацию

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


### Изменение полей общего доступа

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
curl -X PUT "$BASE*URL/federation/nodes/$NODE*ID/modules/$MODULE_ID/exposed" \
  -H "Authorization: Bearer $JWT"
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
    ]}";

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
        "updatedAt": "2020-12-01T14:34:12.972692192Z",
        "updatedBy": "204158548916043781"
    }
}


### Информация о предоставленном федеративном модуле

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


### Удаление модуля из федерации

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


## Структуры назначения

### Информация о общем федеративном модуле

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


### Список общих модулей

.Used variables
```bash
```
# Base url for the federation api
$BASE_URL

# JWT of the user
$JWT

# Node id of the destination node (?exposed) or the origin node (?shared)
$NODE_ID

.Example request
```bash
```
curl -X GET "$BASE*URL/federation/nodes/$NODE*ID/modules" \
  -H "Authorization: Bearer $JWT";

.Example response
```bash
```
{
    "response": {
        "filter": {
            "query": "",
            "handle": "",
            "name": "",
            "sort": "name ASC",
            "count": 1
        },
        "set": [
            {
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
        ]
    }
}


### Задание сопоставления модулей для модуля

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
curl -X PUT "$BASE*URL/federation/nodes/$NODE*ID/modules/$MODULE_ID/mapped" \
  -H "Authorization: Bearer $JWT"
  -H "Content-Type: application/json" \
  --data "[{
      \"origin\":{
          \"name\":\"LinkedIn\",
          \"kind\":\"Url\",
          \"is_multi\":0
        },
        \"destination\":{
            \"name\":\"Social\",
            \"kind\":\"String\",
            \"is_multi\":0
        }
    }]";

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
        "mapping": [
            {
                "origin": {
                    "name": "LinkedIn",
                    "kind": "Url",
                    "is_multi": 0
                },
                "destination": {
                    "name": "Social",
                    "kind": "String",
                    "is_multi": 0
                }
            }
        ]
    }
}


### Получение сопоставления модулей для модуля

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
curl -X GET "$BASE*URL/federation/nodes/$NODE*ID/modules/$MODULE_ID/mapped" \
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
        "mapping": [
            {
                "origin": {
                    "name": "LinkedIn",
                    "kind": "Url",
                    "is_multi": 0
                },
                "destination": {
                    "name": "Social",
                    "kind": "String",
                    "is_multi": 0
                }
            }
        ]
    }
}


### Удаление сопоставления модулей из федерации

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


## Синхронизация структур

### Получение изменений источника

```bash
```
```bash
```
```bash
```

### Синхронизация структуры общего модуля

```bash
```
```bash
```
```bash
```

## Получение предоставленных данных

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
