# Numeric functions

## `min(...number)`

Функция `min` возвращает число с наименьшим значением.

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
out = min(1, 2, 3, -1)
|
```json
```
{
  "out": -1
}

|===

## `max(...number)`

Функция `max` возвращает число с наибольшим значением.

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
out = min(1, 2, 3, -1)
|
```json
```
{
  "out": 3
}

|===

## `round(number, places)`

Функция `round` округляет `number` до указанного количества `places`.
Функция возвращает число с плавающей точкой.

!!! tip
    Чтобы избавиться от плавающей точки, просто приведите к `Integer`.


.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": 10.123
}
|
```
```
out = round(in, 2)
|
```json
```
{
  "in": 10.123,
  "out": 10.12
}


|
```json
```
{
  "in": 10.123
}
|
```
```
out = round(in, 0)
|
```json
```
{
  "in": 10.123,
  "out": 10
}
|===

## `floor(number)`

Функция `floor` округляет число вниз до ближайшего целого.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": 10.123
}
|
```
```
out = floor(in)
|
```json
```
{
  "in": 10.123,
  "out": 10
}

|===


## `ceil(number)`

Функция `ceil` округляет число вверх до ближайшего целого.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": 10.123
}
|
```
```
out = ceil(in)
|
```json
```
{
  "in": 10.123,
  "out": 11
}
|===


## `abs(number)`

Функция `abs` возвращает абсолютное значение переданного `number`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": -10
}
|
```
```
out = abs(in)
|
```json
```
{
  "in": -10
  "out": 10
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
out = abs(in)
|
```json
```
{
  "in": 10
  "out": 10
}
|===

## `log(number)`

Функция `log` возвращает десятичный логарифм (по основанию 10) заданного `number`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": 100
}
|
```
```
out = log(in)
|
```json
```
{
  "in": 100
  "out": 2
}
|===

## `pow(number, exp)`

Функция `pow` возвращает `number`, возведённое в степень `exp`.

|**pow**   |`pow(number, number)`|Функция возвращает x**y, x в степени y, см. [math.Pow](https://golang.org/pkg/math/#Pow)|`pow(2, 3)` результат: 8

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": 2,
  "exp": 3
}
|
```
```
out = pow(in, exp)
|
```json
```
{
  "in": 2,
  "exp": 3,
  "out": 8
}

|===


## `sqrt(number)`

Функция `sqrt` возвращает квадратный корень заданного `number`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": 4
}
|
```
```
out = sqrt(in)
|
```json
```
{
  "in": 4,
  "out": 2
}

|===


## `sum(...number)`

Функция `sum` возвращает сумму всех переданных аргументов.

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
out = sum(1, 2, 3, -1)
|
```json
```
{
  "out": 5
}
|===


## `average(...number)`

Функция `average` возвращает среднее значение из переданных аргументов.

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
out = average(1, 2)
|
```json
```
{
  "out": 1.5
}
|===


## `random(a, b?)`

Функция `random` возвращает случайное число.
При вызове с одним аргументом (`random(to)`) случайное число находится между `0` и `to`.
При вызове с двумя аргументами (`random(from, to)`) случайное число находится между `from` и `to`.

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "to": 10
}
|
```
```
out = random(to)
|
```json
```
{
  "to": 10,
  "out": 3.412
}


|
```json
```
{
  "from": 5,
  "to": 10
}
|
```
```
out = random(to)
|
```json
```
{
  "from": 5,
  "to": 10,
  "out": 5.9
}
|===

## `int(Any)`

Функция `int` приводит аргумент к `Integer`.
Если значение не может быть приведено, функция возвращает `0`.

!!! tip
    Когда вы присваиваете значение переменной, оно автоматически приводится к указанному типу.
    Явное приведение требуется только при передаче аргументов.


.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": "10"
}
|
```
```
out = int(in)
|
```json
```
{
  "in": "10",
  "out": 10
}


|
```json
```
{
  "in": "NO"
}
|
```
```
out = int(in)
|
```json
```
{
  "in": "NO",
  "out": 0
}
|===
