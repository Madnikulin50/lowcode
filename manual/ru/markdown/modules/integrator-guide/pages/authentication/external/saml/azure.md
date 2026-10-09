<a id="saml-azure"></a>
# Интеграция SAML LowCoooode с Azure

Чтобы включить SAML SSO в LowCoooode через Microsoft Azure, вам необходимо создать новое приложение в Azure, а также настроить его в веб-приложении LowCoooode Admin.

!!! caution
    Значения, использованные на скриншоте и в таблице, приведены только в информационных целях и должны быть изменены в соответствии с настройками вашего инстанса.


## Предварительные требования

<a id="prereq-certificates"></a>
### Сертификаты

LowCoooode требует предоставить пару сертификата и закрытого ключа.
Вам нужно либо предоставить существующую пару, либо сгенерировать новую.

.Для генерации требуемых параметров вы можете использовать следующие команды:
```shell
```
# This generates a private key
openssl genpkey -algorithm RSA -out private.key -pkeyopt rsa*keygen*bits:2048

# This generates a self-signed certificate using the private key
openssl req -new -x509 -key private.key -out certificate.crt -days 365

# This packs everything into a .pfx file
openssl pkcs12 -export -out certificate.pfx -inkey private.key -in certificate.crt

!!! important
    Файл `.pfx` требует указать пароль.
    Убедитесь, что вы используете что-то надёжное, и запишите его, так как он понадобится позже.


## Настройка Microsoft Azure

### Создание приложения Azure

Перейдите на https://portal.azure.com/ и войдите в свою учётную запись Azure (если у вас нет учётной записи, создайте её перед продолжением).
На главной странице нажмите кнопку btn:[Enterprise applications].

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/azure/homepage.png",
    "alias": "authentication-saml-azure-homepage",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 0,
    "y": 0,
    "w": 1920,
    "h": 1080
  },
  "annotations": [{
    "kind": "box",
    "x": 550,
    "y": 113,
    "h": 74,
    "w": 70
  }]
}

В списке приложений (если вашего приложения ещё нет) нажмите кнопку btn:[New application].
Если вы уже создали приложение, вы можете пропустить шаг создания приложения

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/azure/enterprise-applications.png",
    "alias": "authentication-saml-azure-enterprise-applications",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 0,
    "y": 0,
    "w": 1920,
    "h": 1080
  },
  "annotations": [{
    "kind": "box",
    "x": 293,
    "y": 142,
    "w": 116,
    "h": 17
  }]
}

Нажмите на btn:[create your own application] и заполните имя приложения.
Выберите опцию btn:[Integrate any other application you don't find in the gallery (non-gallery)] и нажмите кнопку btn:[create].

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/azure/create-new-app.png",
    "alias": "authentication-saml-azure-create-new-app",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 0,
    "y": 0,
    "w": 1920,
    "h": 1080
  },
  "annotations": [{
    "kind": "box",
    "x": 28,
    "y": 143,
    "w": 185,
    "h": 17
  }, {
    "kind": "box",
    "x": 1355,
    "y": 242,
    "w": 445,
    "h": 24
  }, {
    "kind": "box",
    "x": 1356,
    "y": 365,
    "w": 444,
    "h": 18
  }, {
    "kind": "box",
    "x": 1355,
    "y": 1036,
    "w": 80,
    "h": 24
  }]
}

### Настройка SAML

На странице обзора нажмите на btn:[set up single sign on]

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/azure/app-overview.png",
    "alias": "authentication-saml-azure-app-overview",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 0,
    "y": 0,
    "w": 1920,
    "h": 1080
  },
  "annotations": [{
    "kind": "box",
    "x": 665,
    "y": 439,
    "w": 332,
    "h": 132
  }]
}

Затем нажмите на опцию btn:[SAML], которая приведёт вас на экран настройки входа на основе SAML.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/azure/sso-type-select.png",
    "alias": "authentication-saml-azure-sso-type-select",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 0,
    "y": 0,
    "w": 1920,
    "h": 1080
  },
  "annotations": [{
    "kind": "box",
    "x": 645,
    "y": 333,
    "w": 332,
    "h": 161
  }]
}

В разделе «basic SAML configuration» нажмите кнопку btn:[edit] и вставьте следующее:

- **Identifier (Entry ID)**: `https://api.{API_DOMAIN}/auth/external/saml/metadata`
- **Reply URL**: `https://api.{API_DOMAIN}/auth/external/saml/callback`

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/azure/sso-config-basic-saml-configuration.png",
    "alias": "authentication-saml-azure-sso-config-basic-saml-configuration",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 0,
    "y": 0,
    "w": 1920,
    "h": 1080
  },
  "annotations": [{
    "kind": "box",
    "x": 956,
    "y": 362,
    "h": 17,
    "w": 47
  }, {
    "kind": "box",
    "x": 956,
    "y": 362,
    "h": 17,
    "w": 47
  }, {
    "kind": "box",
    "x": 1093,
    "y": 278,
    "h": 24,
    "w": 581
  }, {
    "kind": "box",
    "x": 1093,
    "y": 480,
    "h": 24,
    "w": 581
  }, {
    "kind": "box",
    "x": 1093,
    "y": 114,
    "h": 16,
    "w": 50
  }]
}

На экране настройки входа на основе SAML в разделе «SAML Certificates» нажмите кнопку btn:[edit], а затем кнопку btn:[Import Certificate].
Предоставьте файл `.pfx`, [prereq-certificates,который мы сгенерировали в начале](#prereq-certificates,который мы сгенерировали в начале).

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/azure/sso-config-certificate-import.png",
    "alias": "authentication-saml-azure-sso-config-certificate-import.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 0,
    "y": 0,
    "w": 1920,
    "h": 1080
  },
  "annotations": [{
    "kind": "box",
    "x": 957,
    "y": 761,
    "h": 16,
    "w": 46
  }, {
    "kind": "box",
    "x": 1290,
    "y": 114,
    "h": 17,
    "w": 121
  }, {
    "kind": "box",
    "x": 1204,
    "y": 243,
    "h": 23,
    "w": 294
  }, {
    "kind": "box",
    "x": 1204,
    "y": 282,
    "h": 24,
    "w": 269
  }, {
    "kind": "box",
    "x": 1086,
    "y": 322,
    "h": 24,
    "w": 80
  }]
}

После предоставления сертификата нажмите на три вертикальные точки, чтобы развернуть меню, затем отметьте сертификат как активный.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/azure/activate-certificate.png",
    "alias": "authentication-saml-azure-activate-certificate.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 0,
    "y": 0,
    "w": 1920,
    "h": 1080
  },
  "annotations": [{
    "kind": "box",
    "x": 1870,
    "y": 222,
    "w": 26,
    "h": 26
  }, {
    "kind": "box",
    "x": 1696,
    "y": 234,
    "w": 194,
    "h": 33
  }]
}

Нажмите btn:[yes] во всплывающем окне предупреждения, чтобы завершить активацию сертификата.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/azure/activate-certificate-warning.png",
    "alias": "authentication-saml-azure-activate-certificate-warning.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 0,
    "y": 0,
    "w": 1920,
    "h": 1080
  },
  "annotations": [{
    "kind": "box",
    "x": 1086,
    "y": 239,
    "h": 24,
    "w": 80
  }]
}

## Настройка LowCoooode

Перейдите на your-lowcode-instance.tld и войдите в свой инстанс LowCoooode.
На главной странице нажмите на приложение btn:[Admin Area].

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/azure/lowcode-homepage.png",
    "alias": "authentication-saml-azure-lowcode-homepage.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 0,
    "y": 0,
    "w": 1920,
    "h": 1080
  },
  "annotations": [{
    "kind": "box",
    "x": 529,
    "y": 388,
    "w": 207,
    "h": 169
  }]
}

В навигационном меню нажмите на menu:System[Auth Settings] и перейдите в раздел «External Authentication Providers».
Найдите поставщика SAML и нажмите на значок гаечного ключа, чтобы открыть модальное окно конфигурации.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/azure/auth-settings.png",
    "alias": "authentication-saml-azure-auth-settings.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 0,
    "y": 0,
    "w": 1920,
    "h": 1080
  },
  "annotations": [{
    "kind": "box",
    "y": 326,
    "x": 15,
    "w": 275,
    "h": 33
  }, {
    "kind": "box",
    "y": 549,
    "x": 1609,
    "w": 48,
    "h": 56
  }]
}

Укажите следующие параметры:

- **Name**: это подпись, которая отображается на экране аутентификации («Login with ...your name here...»).
- **Certificate/public key**: скопируйте и вставьте содержимое файла `certificate.crt`, сгенерированного в [#prereq-certificates,разделе предварительных требований](##prereq-certificates,разделе предварительных требований).
- **Certificate/private key**: скопируйте и вставьте содержимое файла `private.key`, сгенерированного в [#prereq-certificates,разделе предварительных требований](##prereq-certificates,разделе предварительных требований)
- **Requests/sign requests**: отметьте этот флажок
- **signature method**: `SHA256`
- **Binding**: `HTTP POST`
- **Identity provider/URL**: скопируйте и вставьте значение, указанное в «App Federation Metadata Url»
- **Name Field**: `http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname`
- **Handle Field**: `http://schemas.microsoft.com/identity/claims/objectidentifier`
- **Identifier Field**: `http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name`

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/azure/saml-params-top.png",
    "alias": "authentication-saml-azure-saml-params-top.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 0,
    "y": 0,
    "w": 1920,
    "h": 1080
  },
  "focus": {
    "x": 571,
    "y": 38,
    "w": 778,
    "h": 1004
  },
  "annotations": []
}

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/azure/saml-params-bottom.png",
    "alias": "authentication-saml-azure-saml-params-bottom.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 0,
    "y": 0,
    "w": 1920,
    "h": 1080
  },
  "focus": {
    "x": 571,
    "y": 38,
    "w": 778,
    "h": 1004
  },
  "annotations": []
}

Нажмите кнопку btn:[Ok] и отправьте изменения поставщика внешней аутентификации.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/azure/confirm-modal.png",
    "alias": "authentication-saml-azure-confirm-modal.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 0,
    "y": 0,
    "w": 1920,
    "h": 1080
  },
  "annotations": [{
    "kind": "box",
    "x": 1568,
    "y": 935,
    "w": 83,
    "h": 31
  }]
}


## Конечный результат

Откройте приватное окно или другой браузер, затем перейдите на свой инстанс LowCoooode.
Если всё настроено правильно, вы должны увидеть новую опцию входа в центре экрана.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/azure/login-screen.png",
    "alias": "authentication-saml-azure-login-screen.png",
    "w": 1920,
    "h": 1080
  },
  "view": {
    "x": 0,
    "y": 0,
    "w": 1920,
    "h": 1080
  },
  "annotations": [{
    "kind": "box",
    "x": 670,
    "y": 663,
    "w": 579,
    "h": 42
  }]
}
