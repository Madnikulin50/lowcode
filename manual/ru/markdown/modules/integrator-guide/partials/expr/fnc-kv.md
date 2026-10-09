# KV functions

!!! caution
    Результирующий тип функции KV зависит от первого аргумента.
    Нельзя передавать несколько различных типов KV (KV, KVV, Vars) в одну и ту же функцию.
    
    Например; `merge(KV, KVV, Vars)` не допускается.


## `set(kv, k, v)`

Функция `set` присваивает значение заданной переменной типа `KV`.

**Исходное значение остаётся неизменным.**

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": {}
}
|
```
```
out = set(in, "foo", "bar")
|
```json
```
{
  "in": {},
  "out": {"foo": "bar"}
}

|===

## `merge(kv, ...kv)`

|**merge**  |`merge(KV, arg1, ...argN)`|Функция объединяет все заданные типы KV в один тип KV.|`merge(&KVV{"foo": ["foo"]}, &KVV{"bar": ["bar"]})` результат: `&KVV{"foo": ["foo"], "bar": ["bar"]}`, То же самое для KV и Vars.

Функция `merge` объединяет все переменные типа `KV` в один `KV`.

**Исходное значение остаётся неизменным.**

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "a": {"foo": "foo"},
  "b": {"bar": "baz"}
}
|
```
```
out = merge(a, b)
|
```json
```
{
  "a": {"foo": "foo"},
  "b": {"bar": "baz"}
  "out": {"foo": "foo", "bar": "baz"}
}

|===

## `filter(kv, ...include)`

Функция `filter` возвращает новый `KV`, содержащий только указанные ключи.

**Исходное значение остаётся неизменным.**

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": {"foo": "foo", "bar": "bar"}
}
|
```
```
out = filter(in, "foo")
|
```json
```
{
  "in": {"foo": "foo", "bar": "bar"},
  "out": {"foo": "foo"}
}

|===

## `omit(kv, ...exclude)`

Функция `omit` возвращает новый `KV` без указанных ключей.

**Исходное значение остаётся неизменным.**

.Примеры:
[cols="1a,1a,1a"]
|===
|State |Expression |Result

|
```json
```
{
  "in": {"foo": "foo", "bar": "bar"}
}
|
```
```
out = filter(in, "foo")
|
```json
```
{
  "in": {"foo": "foo", "bar": "bar"},
  "out": {"bar": "bar"}
}

|===
