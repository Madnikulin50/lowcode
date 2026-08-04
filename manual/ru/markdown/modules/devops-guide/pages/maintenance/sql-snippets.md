# SQL-фрагменты

!!! caution
    Перед выполнением разрушительных операций рекомендуется создать резервную копию базы данных.


!!! tip
    При удалении данных вы можете заменить `DELETE` на `SELECT`, чтобы просмотреть, что будет удалено.


## Удаление всех записей

```sql
```
-- Record values
delete from compose*record*value;

-- Records
delete from compose_record;

## Удаление всех записей пространства имён

```sql
```
-- Record values
delete from compose*record*value where record*id in (select id from compose*record where rel*namespace = {NAMESPACE*ID} AND rel*module = {MODULE*ID});

-- Records
delete from compose*record where rel*namespace = {NAMESPACE_ID};

## Удаление всех записей модуля

```sql
```
-- Record values
delete from compose*record*value where record*id in (select id from compose*record where rel*namespace = {NAMESPACE*ID} AND rel*module = {MODULE*ID});

-- Records
delete from compose*record where rel*namespace = {NAMESPACE*ID} AND rel*module = {MODULE_ID};


## Удаление конкретных значений записей

```sql
```
delete from compose*record*value where name = {FIELD*NAME} AND record*id in (select id from compose*record where rel*namespace = {NAMESPACE*ID} AND rel*module = {MODULE_ID});

## Удаление всех мягко удалённых записей

```sql
```
-- Record values
delete from compose*record*value where record*id in (select id from compose*record where rel*namespace = {NAMESPACE*ID} AND rel*module = {MODULE*ID} and deleted_at IS NOT NULL);

-- Records
delete from compose*record where rel*namespace = {NAMESPACE*ID} AND rel*module = {MODULE*ID} and deleted*at IS NOT NULL;
