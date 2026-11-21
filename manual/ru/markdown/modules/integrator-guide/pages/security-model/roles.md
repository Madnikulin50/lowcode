# Управление ролями


Роль — это группа безопасности, которая может иметь или не иметь доступ к определённым пользовательским интерфейсам или системным ресурсам в зависимости от вашей [модели безопасности](modules/integrator-guide/pages/security-model/security-model/index.md).

Роль может включать ряд [пользователей](modules/integrator-guide/pages/security-model/security-model/users.md), которым предоставляется модель безопасности, определённая данной ролью.

## Пользовательский интерфейс

Пользовательский интерфейс управления ролями находится в веб-приложении [LowCoooode Admin](modules/integrator-guide/pages/security-model/index.md#webapp-admin) в разделе menu:system[roles].

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "security-model/roles/list.png",
    "alias": "security-model-roles-list-index",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "kind": "box",
    "x": 22,
    "y": 238,
    "w": 190,
    "h": 14
  }]
}

## Советы и рекомендации

### Модульное проектирование

!!! note
    *DevNote* @todo


## Список ролей

.Чтобы вывести список текущих системных ролей:
1. Перейдите в menu:system[roles],
1. при необходимости задайте параметры фильтрации.
1. Список обновится автоматически.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "security-model/roles/list.png",
    "alias": "security-model-roles-list",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 321,
    "y": 0,
    "w": 1599,
    "h": 960
  },
  "annotations": []
}

## Создание ролей

!!! important
    Разрешения можно определять только для существующих ролей.


.Чтобы создать новую роль:
1. Перейдите в menu:system[roles],
1. нажмите кнопку btn:[new],
1. укажите параметры новой роли,
1. нажмите кнопку btn:[submit].

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "security-model/roles/create.png",
    "alias": "security-model-roles-create",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 321,
    "y": 0,
    "w": 1599,
    "h": 460
  },
  "annotations": []
}

## Редактирование ролей

.Чтобы отредактировать существующую роль:
1. Перейдите в menu:system[roles],
1. нажмите кнопку btn:[edit] рядом с ролью, которую хотите отредактировать,
1. обновите параметры роли,
1. нажмите кнопку btn:[submit].

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "security-model/roles/edit.png",
    "alias": "security-model-roles-edit-index",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 354,
    "y": 0,
    "w": 1561,
    "h": 707
  },
  "annotations": []
}

<a id="archiving-roles"></a>
## Архивация ролей

.Чтобы заархивировать существующую роль:
1. Перейдите в menu:system[roles],
1. нажмите кнопку btn:[edit] рядом с ролью, которую хотите заархивировать,
1. нажмите кнопку btn:[archive] и подтвердите действие.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "security-model/roles/edit.png",
    "alias": "security-model-roles-edit-archive",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 354,
    "y": 0,
    "w": 1561,
    "h": 707
  },
  "annotations": [{
    "kind": "box",
    "padding": "sm",
    "y": 646,
    "x": 635,
    "w": 91,
    "h": 34
  }]
}

## Удаление ролей

!!! tip
    Вместо удаления ролей вы можете <<archiving-roles,заархивировать роль>>.


.Чтобы удалить существующую роль:
1. Перейдите в menu:system[roles],
1. нажмите кнопку btn:[edit] рядом с ролью, которую хотите удалить,
1. нажмите кнопку btn:[delete] и подтвердите действие.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "security-model/roles/edit.png",
    "alias": "security-model-roles-edit-delete",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 354,
    "y": 1,
    "w": 1565,
    "h": 984
  },
  "annotations": [{
    "kind": "box",
    "padding": "sm",
    "x": 544,
    "y": 646,
    "w": 82,
    "h": 34
  }]
}

## Клонирование разрешений роли

.Чтобы скопировать разрешения роли на другую роль:
1. Перейдите в menu:system[roles],
1. нажмите кнопку btn:[clone permissions],
1. выберите роли, на которые нужно скопировать разрешения,
1. нажмите кнопку btn:[clone].

.На скриншоте показана кнопка клонирования разрешений.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "security-model/roles/clone-permissions-button.png",
    "alias": "security-model-roles-clone-permissions-button",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "kind": "box",
    "x": 1572,
    "y": 88,
    "w": 148,
    "h": 18
  }]
}

.На скриншоте показан диалог клонирования разрешений для выбора ролей, на которые нужно скопировать разрешения.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "security-model/roles/clone-permissions.png",
    "alias": "security-model-roles-clone-permissions",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "kind": "box",
    "x": 685,
    "y": 412,
    "w": 550,
    "h": 256
  }]
}
