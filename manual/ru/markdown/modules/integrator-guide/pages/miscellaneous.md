# Разное

## Системные настройки

Большая часть системной конфигурации, которая влияет на поведение системы, выполняется на уровне развёртывания.

Обратитесь к [Devops Guide](modules/devops-guide/pages/index.md) за подробностями.

### Пользовательские интерфейсы

Интерфейс управления ролями находится в веб-приложении [LowCoooode Admin](modules/integrator-guide/pages/index.md#webapp-admin), в разделах:

- menu:system[settings]
- menu:system[email settings]
- menu:compose[settings]
- menu:user interface[settings]

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "miscellaneous/webapp-admin-dashboard.png",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "w": 320,
    "h": 1080,
    "x": 0,
    "y": 0
  },
  "annotations": [{
    "kind": "box",
    "padding": "xs",
    "x": 16,
    "y": 295,
    "w": 288,
    "h": 30
  }, {
    "kind": "box",
    "padding": "xs",
    "x": 16,
    "y": 330,
    "w": 288,
    "h": 30
  }, {
    "kind": "box",
    "padding": "xs",
    "x": 16,
    "y": 583,
    "w": 288,
    "h": 37
  }, {
    "kind": "box",
    "padding": "xs",
    "x": 16,
    "y": 868,
    "w": 288,
    "h": 37
  }]
}

### Системные настройки

**Системные настройки** находятся в разделе menu:system[settings] в левом навигационном меню.

Системные настройки позволяют вам настроить внутреннюю и внешнюю аутентификацию, а также [многофакторную аутентификацию](modules/integrator-guide/pages/authentication/mfa.md) и различные потоки аутентификации.

!!! tip
    Вы можете отключить внутреннюю регистрацию и вручную добавлять пользователей, которым разрешён доступ к вашей системе.


Обратитесь к [External](modules/integrator-guide/pages/authentication/external/index.md) за подробностями о настройке внешних поставщиков.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "miscellaneous/system-settings.png",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "kind": "box",
    "padding": "xs",
    "x": 16,
    "y": 292,
    "w": 288,
    "h": 30
  }]
}

### Настройки электронной почты

**Настройки электронной почты** находятся в разделе menu:system[email settings] в левом навигационном меню.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "miscellaneous/email-settings.png",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "kind": "box",
    "padding": "xs",
    "x": 16,
    "y": 330,
    "w": 288,
    "h": 30
  }]
}

Интерфейс настроек электронной почты позволяет вам настроить SMTP-серверы для отправки писем из LowCoooode.
Пожалуйста, обратитесь к вашему поставщику почтовых услуг за информацией о SMTP-сервере (host, port, user, password).

!!! caution
    Диагностика и дополнительная информация в случае неправильной конфигурации доступны только в журналах сервера.


В случае каких-либо проблем с TLS/сертификатами вы можете изменить имя сервера для TLS-проверки или отключить проверку и разрешить использование недействительных сертификатов.

!!! important
    Первоначальные настройки копируются из [переменных окружения `SMTP_*`](modules/devops-guide/pages/references/configuration/server.md#\_email_sending).
    LowCoooode записывает предупреждение в журнал, если вы сохраняете переменные окружения и вносите изменения либо в переменные, либо в настройки.


Откройте панель администрирования и перейдите в menu:System[Email settings].

.Скриншот показывает настройки SMTP-сервера в панели администрирования.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "email-settings/form.png",
    "alias": "email-settings-form.png",
    "w": 1484,
    "h": 1210
  },
  "view": {},
  "annotations": []
}

- Укажите требуемые имя сервера и порт и необязательные имя пользователя и пароль.
Изменения применяются немедленно и не требуют перезапуска сервера.


### Настройки Low Code

**Настройки Low Code** находятся в разделе menu:compose[settings] в левом навигационном меню.
Настройки Low Code позволяют вам настроить общие параметры вложений, такие как максимальный размер и списки разрешённых типов.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "miscellaneous/compose-settings.png",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "kind": "box",
    "padding": "xs",
    "x": 16,
    "y": 577,
    "w": 288,
    "h": 37
  }]
}

### Настройки пользовательского интерфейса

**Настройки пользовательского интерфейса** находятся в разделе menu:user interface[settings] в левом навигационном меню.
Настройки пользовательского интерфейса позволяют вам настроить основной логотип и логотип-иконку, которые будут использовать приложения LowCoooode.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "miscellaneous/ui-settings.png",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "kind": "box",
    "padding": "xs",
    "x": 16,
    "y": 855,
    "w": 288,
    "h": 37
  }]
}
