<a id="saml-okta"></a>
# Интеграция SAML LowCoooode с Okta

Чтобы включить SAML SSO в LowCoooode через Okta, вам необходимо создать новое приложение в IdP Okta, а также настроить его в веб-приложении LowCoooode Admin.

!!! caution
    Значения, использованные на скриншоте и в таблице, приведены только в информационных целях и должны быть изменены в соответствии с настройками вашего инстанса.


## Добавление нового приложения в Okta

Перейдите на https://www.okta.com/ и войдите в свою учётную запись Okta.
Если у вас нет учётной записи, создайте её перед продолжением.

!!! note
    В целях документирования была создана пробная учётная запись Okta.
    Пробная версия может не поддерживать все функции, которые есть в подписной версии, но базовые функции для SAML SSO включены.
    
    За дополнительной помощью по Okta обратитесь на их https://support.okta.com/help[официальный сайт поддержки].


Во-первых, создайте новое приложение, перейдя в боковое меню menu:Applications[Applications].

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/okta/okta-application.png",
    "alias": "authentication-saml-okta",
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
    "x": 22,
    "y": 228,
    "w": 236,
    "h": 77
  }]
}

Нажмите кнопку btn:[Create App Integration] и выберите **SAML 2.0 integration**.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/okta/okta-application-screen2.png",
    "alias": "authentication-saml-okta-screen2",
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
    "x": 874,
    "y": 463,
    "h": 65,
    "w": 530
  },
  {
    "kind": "box",
    "x": 1352,
    "y": 754,
    "w": 60,
    "h": 40
  }]
}

Мастер приложений приведёт вас к следующему экрану, где вы указываете имя и логотип вашего приложения.
После завершения нажмите кнопку btn:[Next].

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/okta/okta-application-screen3.png",
    "alias": "authentication-saml-okta-screen3",
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
    "x": 1239,
    "y": 580,
    "w": 58,
    "h": 38
  }]
}

На следующем экране вам нужно настроить Okta с параметрами SAML вашего поставщика услуг, в данном случае — вашего инстанса LowCoooode, где настроен SAML.
После настройки нажмите кнопку btn:[Next], чтобы перейти к необязательному шагу обратной связи.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/okta/okta-application-screen4.png",
    "alias": "authentication-saml-okta-screen4",
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
    "x": 615,
    "y": 376,
    "h": 60,
    "w": 671
  },
  {
    "kind": "box",
    "x": 615,
    "y": 498,
    "h": 120,
    "w": 671
  }]
}

На необязательном шаге обратной связи нажмите кнопку btn:[finish], чтобы завершить настройку Okta.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/okta/okta-application-screen5.png",
    "alias": "authentication-saml-okta-screen5",
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
    "x": 1230,
    "y": 560,
    "h": 38,
    "w": 67
  }]
}

После создания приложения вы можете просмотреть его в боковом меню menu:Applications[Applications].
Здесь вы выполняете управление пользователями, так как созданное приложение нужно подключить к пользователям, которые будут его использовать.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/okta/okta-application-preview.png",
    "alias": "authentication-external-saml-okta-okta-application-preview.png",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": []
}

### Параметры приложения Okta SAML

[Attributes]
|===
|**Параметр** |**Значение**

|**Single Sign On URL**
|https://your-lowcode-instance.tld/auth/external/saml/callback

|**Recipient URL**
|https://your-lowcode-instance.tld/auth/external/saml/callback

|**Destination URL**
|https://your-lowcode-instance.tld/auth/external/saml/callback

|**Audience Restriction**
|https://your-lowcode-instance.tld/auth/external/saml/metadata

|**Name ID Format**
|EmailAddress
|===

## Добавление настроек Okta в LowCoooode

После настройки приложения в Okta вам нужно будет настроить LowCoooode с параметрами, предоставленными Okta.

Перейдите в LowCoooode Admin, откройте menu:system[settings] и нажмите на значок гаечного ключа рядом с внешним поставщиком аутентификации SAML.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/okta/lowcode-select-saml.png",
    "alias": "authentication-saml-lowcode",
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
    "padding": "lg",
    "x": 1610,
    "y": 933,
    "w": 18,
    "h": 16
  }]
}

Настройки, связанные с Okta, указываются в разделе **identity provider**, где необходимо указать URL метаданных.

!!! note
    Как сгенерировать URL метаданных для Okta, описано https://support.okta.com/help/s/question/0D50Z00008G7VVzSAN/how-to-use-okta-idp-metadata-in-service-provider-application?language=en_US[здесь].


[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/okta/lowcode-saml-details2.png",
    "alias": "authentication-saml-lowcode-details-identity-provider",
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
    "x": 581,
    "y": 340,
    "w": 770,
    "h": 700
  }
}

Вы можете найти больше информации о создании сертификата на странице [Saml](modules/integrator-guide/pages/authentication/external/saml/authentication/external/saml/index.md).

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/okta/lowcode-saml-details.png",
    "alias": "authentication-saml-lowcode-details-certificate",
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
    "x": 580,
    "y": 210,
    "w": 770,
    "h": 220
  }
}

Опубликуйте настройки, нажав кнопку btn:[OK].

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/external/saml/okta/lowcode-saml-details2.png",
    "alias": "authentication-saml-lowcode-details2",
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
    "x": 1298,
    "y": 1004,
    "h": 30,
    "w": 50
  }]
}

### Параметры приложения Okta SAML для LowCoooode

[Attributes]
|===
|**Параметр** |**Значение**

|**URL**
|https://okta-example-instance.okta.com/app/exk1fgv7123/sso/saml/metadata

|**Name field**
|http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname

|**Handle field**
|http://schemas.microsoft.com/identity/claims/objectidentifier

|**Identifier field**
|http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name
|===
