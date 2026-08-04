# Вычисление разрешений

Функция вычисления разрешений позволяет администраторам проверять, какие действия или операции может выполнять пользователь или комбинация ролей.
Вычисление предполагает сравнение различных ролей, назначенных пользователю, чтобы определить его способность выполнять различные операции в рамках ресурса.

!!! note
    RBAC LowCoooode использует роли данного пользователя, чтобы определить, есть ли у него доступ к данному ресурсу.


## Как вычислить применённые разрешения

!!! note
    Для большинства ресурсов LowCoooode можно определить, каким [ролям](modules/integrator-guide/pages/security-model/security-model/roles.md) разрешён доступ к ресурсу.
    Чтобы управлять ресурсом, найдите кнопку btn:[permissions] в пользовательском интерфейсе.
    
    [annotation,role="data-zoomable"]
    ----
    {
      "image": {
        "rel": "security-model/access-control/general-permissions-button.png",
        "alias": "general-permissions-button",
        "w": 1920,
        "h": 1080
      },
      "view": {
        "x": 320,
        "y": 0,
        "h": 550,
        "w": 1600
      },
      "focus": {
        "x": 1466,
        "y": 80,
        "h": 31,
        "w": 206
      },
      "annotations": [{
        "kind": "box",
        "x": 1539,
        "y": 80,
        "w": 130,
        "h": 30
      }]
    }
    ----


Перейдите к нужному ресурсу и нажмите btn:[permissions].

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "security-model/access-control/permission-button-page.png",
    "alias": "new-user",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "kind": "box",
    "x": 1539,
    "y": 80,
    "w": 131,
    "h": 31
  }]
}

Нажмите кнопку btn:[add +].
Появится всплывающее окно с двумя раскрывающимися списками: btn:[select role(s)] и btn:[search or select user].

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "security-model/access-control/add-role.png",
    "alias": "add-role",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "kind": "box",
    "x": 762,
    "y": 145,
    "w": 758,
    "h": 97
}]
}

Нажмите на раскрывающийся список btn:[select roles] или btn:[select user] и выберите роль или пользователя для вычисления.
Можно применить только один из двух вариантов.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "security-model/access-control/eval-permission.png",
    "alias": "eval-permission.png",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "padding": "xs",
    "x": 585,
    "y": 254,
    "w": 783,
    "h": 410
  },
  "annotations": [{
    "kind": "box",
    "x": 603,
    "y": 353,
    "w": 751,
    "h": 217
  }]
}

Нажмите кнопку btn:[save].

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "security-model/access-control/save-evaluation-permission.png",
    "alias": "save-evaluation-permission",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "padding": "xs",
    "x": 585,
    "y": 254,
    "w": 783,
    "h": 410
  },
  "annotations": [{
    "kind": "box",
    "x": 1148,
    "y": 615,
    "w": 200,
    "h": 43
  }]
}

Если вы выбрали вычисление по роли, повторите шаг, выбрав другую роль, чтобы вычислить разрешения, применённые к разным ролям.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "security-model/access-control/save-evaluation.png",
    "alias": "save-evaluation",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": []
}

## Вычисление разрешений по компонентам

LowCoooode имеет три ключевых компонента ресурсов, а именно:
**System**, **Compose** и **Automation**.
В каждом компоненте ресурсов выполняются свои операции.

!!! important
    Разрешения на различные операции внутри ресурса компонента определяются ролями, назначенными пользователю.


Разрешения можно вычислить, выбрав комбинацию ролей или конкретного пользователя.

Выполните следующие шаги, чтобы получить доступ к разрешениям системы, Compose и автоматизации и вычислить их.

Перейдите к своему инстансу LowCoooode (например, local.lowcode.org) и нажмите на приложение btn:[Admin area].

Появится новый экран со всеми доступными компонентами.
У каждого компонента есть кнопка btn:[permissions].

Нажмите кнопку btn:[permissions] у компонента system, compose или automation.
Появится новый экран со списком всех доступных операций в компоненте.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "security-model/access-control/system-permissions-button.png",
    "alias": "system-permissions-button.png",
    "w": 1920,
    "h": 1080
  },
  "view": {},

  "annotations": [{
    "kind": "box",
    "padding": "xs",
    "x": 14,
    "y": 480,
    "w": 170,
    "h": 32
}]
}

Нажмите кнопку btn:[add +].
Появится всплывающее окно с двумя раскрывающимися списками.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "security-model/access-control/system-permissions-add-button.png",
    "alias": "system-permissions-add-button.png",
    "w": 1920,
    "h": 1080
  },
  "view": {},

  "annotations": [{
    "kind": "box",
    "x": 633,
    "y": 152,
    "w": 1265,
    "h": 129
  }]
}

Нажмите кнопку btn:[evaluate], затем нажмите на раскрывающийся список btn:[select roles] или btn:[select user] и выберите роль или пользователя для вычисления.
Можно применить только один из двух вариантов.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "security-model/access-control/evaluate-permissions.png",
    "alias": "user-create",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "kind": "box",
    "x": 772,
    "y": 510,
    "w": 376,
    "h": 121
  }]
}

Нажмите кнопку btn:[save].

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "security-model/access-control/save-eval-permission.png",
    "alias": "save-eval-permission-button",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "padding": "xs",
    "x": 585,
    "y": 254,
    "w": 783,
    "h": 430
  },
  "annotations": [{
    "kind": "box",
    "x": 1148,
    "y": 615,
    "w": 200,
    "h": 42
  }]
}

Если вы выбрали вычисление по роли, повторите шаг, выбрав другую роль, чтобы вычислить разрешения, применённые к разным ролям.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "security-model/access-control/system-evaluated-permissions.png",
    "alias": "system-evaluated-permissions.png",
    "w": 1920,
    "h": 1080
  },
  "view": {},

  "annotations": []
}
