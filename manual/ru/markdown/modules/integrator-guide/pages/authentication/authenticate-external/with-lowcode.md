# Аутентификация с помощью LowCoooode

<a id="provider"></a>
## Настройка поставщика аутентификации

Создайте новый [клиент аутентификации](modules/integrator-guide/pages/authentication/authenticate-external/authentication/index.md#*auth*client) и убедитесь, что вы выбрали тип предоставления `authorization_code`.
Кроме того, отметьте возможность `allow client to use OIDC and verify user's identity`.

При желании отметьте флажок «trusted», чтобы пропустить финальный шаг проверки.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/with-lowcode/auth-client.png",
    "alias": "authentication-external-with-lowcode-auth-client",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "padding": "sm",
    "x": 854,
    "y": 409,
    "h": 16,
    "w": 527
  }, {
    "padding": "sm",
    "x": 854,
    "y": 660,
    "h": 16,
    "w": 404
  }]
}

## Настройка клиента аутентификации

Откройте LowCoooode Admin, перейдите в menu:system[settings] и прокрутите страницу вниз.

Убедитесь, что отмечено *enable external authentication*, и нажмите на btn:[add OIDC provider].

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/with-lowcode/provider-list.png",
    "alias": "authentication-external-with-lowcode-provider-list",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 563,
    "y": 505,
    "w": 1110,
    "h": 575
  },
  "annotations": [{
    "x": 584,
    "y": 582,
    "w": 249,
    "h": 17
  }, {
    "x": 1467,
    "y": 616,
    "w": 189,
    "h": 32
  }]
}

Заполните данные, полученные от клиента аутентификации, определённого [provider,выше](#provider,выше).

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/with-lowcode/provider-configure.png",
    "alias": "authentication-external-with-lowcode-provider-configure",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 560,
    "y": 30,
    "w": 800,
    "h": 768
  }
}

Нажмите кнопку btn:[submit], чтобы сохранить изменения.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/with-lowcode/provider-added.png",
    "alias": "authentication-external-with-lowcode-provider-added",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 563,
    "y": 505,
    "w": 1110,
    "h": 575
  },
  "annotations": [{
    "x": 1571,
    "y": 1032,
    "w": 85,
    "h": 37
  }]
}

## Аутентификация

Когда ваши пользователи пытаются пройти аутентификацию, им предлагается дополнительная опция внешней аутентификации.

!!! note
    Если OIDC-провайдер недоступен, проверьте журналы вашего сервера на предмет возможных ошибок.


[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/with-lowcode/login-oidc.png",
    "alias": "authentication-external-with-lowcode-login-oidc",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": [{
    "x": 672,
    "y": 645,
    "w": 577,
    "h": 44
  }]
}

Когда вы нажимаете на новую опцию OIDC-аутентификации, вы перенаправляетесь и вам предлагается пройти аутентификацию с вашими учётными данными на инстансе LowCoooode — поставщике.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/with-lowcode/login.png",
    "alias": "authentication-external-with-lowcode-login",
    "w": 1920,
    "h": 1080
  },
  "view": {}
}

После входа в инстанс LowCoooode — поставщик, аутентификация считается успешной.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/with-lowcode/logged-in.png",
    "alias": "authentication-external-with-lowcode-logged-in",
    "w": 1920,
    "h": 1080
  },
  "view": {}
}
