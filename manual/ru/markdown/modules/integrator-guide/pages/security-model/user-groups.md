# Группы пользователей

Группы пользователей позволяют определить иерархию между пользователями, обеспечивая контроль доступа на основе иерархии.

!!! caution
    Убедитесь, что вы знакомы с [вычислением доступа](modules/integrator-guide/pages/security-model/security-model/index.md).


!!! important
    Каждый несистемный пользователь должен принадлежать к группе пользователей.
    Если вы обновились со старой версии, все существующие пользователи будут назначены в группу пользователей по умолчанию.
    
    Когда новый пользователь регистрируется, он назначается в группу пользователей, определённую клиентом аутентификации.
    
    При создании пользователей через административное веб-приложение, автоматизацию или через API необходимо указывать группу пользователей вручную.


## Настройка групп пользователей

Группы пользователей настраиваются и управляются в веб-приложении LowCoooode Admin.

В веб-приложении Admin перейдите в menu:System[User Groups], чтобы увидеть список текущих групп пользователей.
Нажмите кнопку btn:[New User Group] в левом верхнем углу, чтобы открыть редактор.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "security-model/user-groups/user-groups-list.png",
    "alias": "security-model-user-groups-user-groups-list",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "kind": "box",
    "y": 256,
    "x": 21,
    "w": 113,
    "h": 16
  }, {
    "kind": "box",
    "y": 95,
    "x": 581,
    "w": 164,
    "h": 44
  }]
}

Заполните обязательные поля и выберите группы пользователей, которым отчитывается эта группа.
Нажмите кнопку btn:[Submit], чтобы создать группу пользователей.

!!! note
    Каждая группа пользователей может отчитываться нескольким группам пользователей.
    Каждая группа пользователей может отчитываться одной и той же группе пользователей по путям с разными названиями.
    
    Название пути можно указать в функции [`isDescendantOf`](modules/integrator-guide/pages/security-model/expr/fnc-reference.md#isdescendantof)


[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "security-model/user-groups/user-group-basic-filled.png",
    "alias": "security-model-user-groups-user-group-basic-filled",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 614,
    "y": 81,
    "w": 979,
    "h": 548
  },
  "annotations": [{
    "kind": "box",
    "x": 1514,
    "y": 597,
    "w": 74,
    "h": 28
  }]
}

После создания группы пользователей в нижней части страницы появляются два новых раздела.
В разделе «Члены группы пользователей» можно назначать пользователей в группу.

!!! note
    Группу пользователей конкретного пользователя также можно изменить на экране редактирования пользователя.
    Найдите раскрывающийся список «Группа пользователей», чтобы изменить группу.
    
    Не забудьте сохранить изменения, нажав кнопку btn:[Submit].
    
    [annotation,role="data-zoomable"]
    ----
    {
      "image": {
        "rel": "security-model/user-groups/user-edit.png",
        "alias": "security-model-user-groups-user-edit",
        "w": 1920,
        "h": 1080
      },
      "view": {},
      "focus": {
        "x": 575,
        "y": 127,
        "h": 543,
        "w": 1078
      },
      "annotations": [{
        "kind": "box",
        "x": 1129,
        "y": 273,
        "h": 63,
        "w": 522
      }, {
        "kind": "box",
        "x": 1568,
        "y": 634,
        "w": 83,
        "h": 32
      }]
    }
    ----


[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "security-model/user-groups/user-group-bottom.png",
    "alias": "security-model-user-groups-user-group-bottom-users",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 608,
    "y": 647,
    "w": 980,
    "h": 186
  },
  "annotations": []
}

В разделе «Членство в ролях» можно назначать роли группе пользователей.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "security-model/user-groups/user-group-bottom.png",
    "alias": "security-model-user-groups-user-group-bottom-roles",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 608,
    "y": 867,
    "w": 980,
    "h": 186
  },
  "annotations": []
}

## Дополнения

### Дополнения клиента аутентификации

Теперь клиенты аутентификации определяют группу пользователей по умолчанию, в которую назначаются новые пользователи.
Группа пользователей по умолчанию (создаваемая LowCoooode) назначена всем существующим клиентам аутентификации.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "security-model/user-groups/auth-client-default-user-group.png",
    "alias": "security-model-user-groups-auth-client-default-user-group",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 567,
    "y": 120,
    "h": 948,
    "w": 1093
  },
  "annotations": [{
    "kind": "box",
    "x": 575,
    "y": 918,
    "h": 64,
    "w": 521
  }]
}


### Выражения контекстных ролей

Теперь контекстные роли предоставляют набор выражений, которые можно использовать для вычисления иерархии.

1. [`isDescendantOf`](modules/integrator-guide/pages/security-model/expr/fnc-reference.md#isdescendantof)
1. [`isDescendantOfC`](modules/integrator-guide/pages/security-model/expr/fnc-reference.md#isdescendantofc)
1. [`isDescendantOfR`](modules/integrator-guide/pages/security-model/expr/fnc-reference.md#isdescendantofr)
1. [`isDescendantOfU`](modules/integrator-guide/pages/security-model/expr/fnc-reference.md#isdescendantofu)
1. [`isDescendantOfD`](modules/integrator-guide/pages/security-model/expr/fnc-reference.md#isdescendantofd)

Мы рекомендуем создать новую контекстную роль, которая позволит пользователям получать доступ к ресурсам в нижестоящих группах пользователей.
