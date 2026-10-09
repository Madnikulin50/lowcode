# Type functions

## `coalesce(...Any)`

Функция `coalesce` возвращает первое не-`null` значение.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
}
|
```
```
out = coalesce(null, 0, 1, 2)
|
```json
```
{
  "out": 0
}

|===

## `isEmpty(Any)`

Функция `isEmpty` возвращает `true`, если значение пусто.

Что считается пустым, зависит от типа, см. таблицу ниже.

[cols="1,1,1"]
|===
|Type |Value |Result

.2+|***Array***|[]|true
|~undefined~|true
|***Array***|~undefined~|true
|***Any***|~undefined~|false
|***Any***|[]|true
|***Vars***|~undefined~|true
|***Boolean***|~undefined~|true
|***DateTime***|~undefined~|true
|***Float***|~undefined~|true
|***Integer***|~undefined~|true
|***UnsignedInteger***|~undefined~|true
|***String***|~undefined~|true
|***User***|~undefined~|false
|===

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": ""
}
|
```
```
out = isEmpty(in)
|
```json
```
{
  "in": "",
  "out": true
}


|
```json
```
{
  "in": 0
}
|
```
```
out = isEmpty(in)
|
```json
```
{
  "in": 0,
  "out": true
}

|===

## `isNil(Any)`

Функция `isNil` возвращает `true`, если заданное значение равно `null`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": null
}
|
```
```
out = isNil(in)
|
```json
```
{
  "in": null,
  "out": true
}


|
```json
```
{
  "in": 10
}
|
```
```
out = isNil(in)
|
```json
```
{
  "in": 10,
  "out": false
}

|===
