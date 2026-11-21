# Слой хранилища

Слой хранилища предоставляет гибкий способ определения того, как и где мы должны хранить данные.

## Управление схемой РСУБД

### Определение новой схемы

Вы можете определить новую схему в файле `store/rdbms/rdbms_schema.go`.
Система проходит по всем определеним и гарантирует, что все таблицы присутствуют в подключённой базе данных.

!!! important
    Этот шаг не учитывает никакие потенциальные манипуляции со схемой.
    См. <<store-update-schema>> за подробностями.


.Следующий пример определяет таблицу для модулей LowCoooode Low Code:
```go
```

func (s Schema) Tables() []*Table {
  return []*Table{

    s.ComposeModule(),

  }
}


func (Schema) ComposeModule() *Table {
  return TableDef("compose_module",
    ID,
    ColumnDef("rel_namespace", ColumnTypeIdentifier),
    ColumnDef("handle", ColumnTypeVarchar, ColumnTypeLength(handleLength)),
    ColumnDef("name", ColumnTypeText),
    ColumnDef("meta", ColumnTypeJson),
    CUDTimestamps,

    AddIndex("namespace", IColumn("rel_namespace")),
    AddIndex("unique*handle", IColumn("rel*namespace"), IExpr("LOWER(handle)"), IWhere("LENGTH(handle) > 0 AND deleted_at IS NULL")),
  )
}


<a id="store-update-schema"></a>
### Обновление существующих схем

!!! important
    При разработке новых функций, требующих новых определений хранилища, не засоряйте файл `generic_upgrades.go`.
    Либо удалите исходную таблицу и пересоздайте её, либо вручную примените обновления.
    
    Если вы работаете с другими людьми, обязательно скоординируйте это с ними.


Вы можете определить изменение схемы в файле `store/rdbms/generic_upgrades.go`.
Система пройдёт по всем определеним и убедится, что всё актуально.

!!! note
    Система применит только те изменения, которые ещё не были применены.
    Результат одинаков, независимо от того, сколько раз вы его запускаете.


.Следующий пример переименовывает таблицу actionlog:
```go
```

func (g genericUpgrades) Before(ctx context.Context) error {
  return g.all(ctx,
    g.RenameActionlog,

  )
}


func (g genericUpgrades) RenameActionlog(ctx context.Context) error {
  return g.RenameTable(ctx, "sys_actionlog", "actionlog")
}
