# Date and time functions

## `earliest(DateTime, ...DateTime)`

|**earliest**     |`earliest(arg1, arg2, ...argN)`|Функция возвращает самую раннюю дату и время.|`earliest(datefield1, datefield2)` результат: "1970-01-01T00:00:00"


Функция `earliest` возвращает самую раннюю `DateTime`.

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
out = earliest(parseISOTime("2019-01-01T00:00:00Z"), parseISOTime("2020-01-01T00:00:00Z"))
|
```json
```
{
  "out": "2019-01-01T00:00:00Z"
}

|===

## `latest`

Функция `latest` возвращает самую позднюю `DateTime`.

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
out = latest(parseISOTime("2019-01-01T00:00:00Z"), parseISOTime("2020-01-01T00:00:00Z"))
|
```json
```
{
  "out": "2020-01-01T00:00:00Z"
}

|===

## `parseISOTime`

Функция `parseISOTime` разбирает метку времени в формате ISO.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "2020-01-01T00:00:00Z"
}
|
```
```
out = parseISOTime("2020-01-01T00:00:00Z")
|
```json
```
{
  "in": "2020-01-01T00:00:00Z",
  "out": "2020-01-01T00:00:00Z"
}

|===

## `modTime`

Функция `modTime` возвращает новую `DateTime` с добавленной `duration`.
Функция `modTime` взаимодействует с временной частью `DateTime`.

Используйте `modDate`, `modWeek`, `modMonth` или `modYear`, если вы хотите изменить более крупные компоненты.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "2020-01-01T00:00:00Z",
  "d": "1h"
}
|
```
```
out = modTime(in, d)
|
```json
```
{
  "in": "2020-01-01T00:00:00Z",
  "d": "1h",
  "out": "2020-01-01T01:00:00Z",
}

|===

## `modDate(datetime, days)`

Функция `modDate` возвращает новую `DateTime` с добавленными `days`.
Функция `modDate` взаимодействует с частью даты (дни) `DateTime`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "2020-01-01T00:00:00Z",
  "d": "1"
}
|
```
```
out = modDate(in, d)
|
```json
```
{
  "in": "2020-01-01T00:00:00Z",
  "d": "1",
  "out": "2020-01-02T00:00:00Z",
}

|===

## `modWeek(datetime, weeks)`

Функция `modWeek` возвращает новую `DateTime` с добавленными `weeks`.
Функция `modWeek` взаимодействует с частью даты (дни) `DateTime`.


.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "2020-01-01T00:00:00Z",
  "d": "1"
}
|
```
```
out = modWeek(in, d)
|
```json
```
{
  "in": "2020-01-01T00:00:00Z",
  "d": "1",
  "out": "2020-01-08T00:00:00Z",
}

|===

## `modMonth(datetime, months)`

Функция `modMonth` возвращает новую `DateTime` с добавленными `months`.
Функция `modMonth` взаимодействует с частью месяца `DateTime`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "2020-01-01T00:00:00Z",
  "d": "1"
}
|
```
```
out = modMonth(in, d)
|
```json
```
{
  "in": "2020-01-01T00:00:00Z",
  "d": "1",
  "out": "2020-02-01T00:00:00Z",
}

|===

## `modYear(datetime, years)`

Функция `modYear` возвращает новую `DateTime` с добавленными `years`.
Функция `modYear` взаимодействует с частью года `DateTime`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "2020-01-01T00:00:00Z",
  "d": "1"
}
|
```
```
out = modYear(in, d)
|
```json
```
{
  "in": "2020-01-01T00:00:00Z",
  "d": "1",
  "out": "2021-00-01T00:00:00Z",
}

|===

## `parseDuration`

Функция `parseDuration` возвращает разобранную длительность из заданной строки.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "2h"
}
|
```
```
out = parseDuration(in)
|
```json
```
{
  "in": "2h",
  "out": "2h0m0s"
}

|===

## `strftime(datetime, pattern)`

Функция `strftime` возвращает отформатированную `DateTime` на основе заданного `pattern`.
Обратитесь к [Datetime Formatting](modules/integrator-guide/partials/expr/expr/datetime-formatting.md) за более подробной информацией.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "2020-01-01T00:00:00Z"
}
|
```
```
out = strftime(in, "%Y-%m-%d")
|
```json
```
{
  "in": "2020-01-01T00:00:00Z",
  "out": "2020-01-01"
}

|===

## `isLeapYear(datetime)`

Функция `isLeapYear` возвращает `true`, если заданный `DateTime` является високосным годом.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "2020-01-01T00:00:00Z"
}
|
```
```
out = isLeapYear(in)
|
```json
```
{
  "in": "2020-01-01T00:00:00Z",
  "out": true
}


|
```json
```
{
  "in": "2019-01-01T00:00:00Z"
}
|
```
```
out = isLeapYear(in)
|
```json
```
{
  "in": "2019-01-01T00:00:00Z",
  "out": false
}

|===

## `now`

Функция `now` возвращает текущую `DateTime`.

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
out = now()
|
```json
```
{
  "out": "2022-02-24T18:00:00Z"
}

|===

## `isWeekDay`

|**isWeekDay**    |`isWeekDay(datetime)`|Функция возвращает true, если указанный день является будним днём.|`isWeekDay(datefield)` результат: `true`


Функция `isWeekDay` возвращает `true`, если заданный `DateTime` является будним днём.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "2022-02-24T00:00:00Z"
}
|
```
```
out = isWeekDay(in)
|
```json
```
{
  "in": "2022-02-24T00:00:00Z",
  "out": true
}


|
```json
```
{
  "in": "2022-02-26T00:00:00Z"
}
|
```
```
out = isWeekDay(in)
|
```json
```
{
  "in": "2022-02-26T00:00:00Z",
  "out": false
}

|===

## `sub(from, to)`

Функция `sub` возвращает разницу между двумя `DateTime` в миллисекундах.

`from` должен быть больше `to`; в противном случае функция вызовет ошибку.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "from": "2022-02-02T00:00:00Z",
  "to": "2022-02-01T00:00:00Z"
}
|
```
```
out = sub(from, to)
|
```json
```
{
  "from": "2022-02-02T00:00:00Z",
  "to": "2022-02-01T00:00:00Z"
  "out": 86400000
}

|===
