# Многофакторная аутентификация

Многофакторная аутентификация (MFA) обеспечивает дополнительный уровень безопасности для ваших пользователей.
LowCoooode предоставляет многофакторную аутентификацию через email или через мобильное приложение-аутентификатор.

## Настройка MFA

Многофакторную аутентификацию можно включить в веб-приложении LowCoooode Admin в разделе menu:system[authentication, multi-factor authentication].
Вы можете **разрешить** или **обязать** пользователей получать свой OTP (одноразовый пароль) либо через **email**, либо через **приложение-аутентификатор**.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/mfa/mfa-settings-admin-configure.png",
    "alias": "authentication-mfa-mfa-settings-admin-configure-email",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 572,
    "y": 330,
    "w": 1080,
    "h": 389
  },
  "annotations": [{
    "kind": "box",
    "x": 575,
    "y": 333,
    "w": 1029,
    "h": 46
  }, {
    "kind": "box",
    "x": 575,
    "y": 535,
    "w": 1029,
    "h": 70
  }]
}

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/mfa/mfa-settings-admin-configure.png",
    "alias": "authentication-mfa-mfa-settings-admin-configure-mobile",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 575,
    "y": 537,
    "w": 1076,
    "h": 340
  },
  "annotations": [{
    "kind": "box",
    "x": 575,
    "y": 535,
    "w": 1029,
    "h": 70
  }]
}

Далее, чтобы включить OTP для конкретного пользователя, перейдите в menu:system[users] и нажмите на значок «edit» рядом с пользователем, для которого вы хотите включить OTP.
Перейдите в раздел «многофакторная аутентификация» и нажмите кнопку btn:[enable].

.На скриншоте показан пользовательский интерфейс, используемый для включения MFA-аутентификации для конкретного пользователя.
[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/mfa/mfa-enable.png",
    "alias": "authentication-mfa-mfa-enable",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "focus": {
    "x": 575,
    "y": 755,
    "w": 1076,
    "h": 135
  },
  "annotations": [{
    "kind": "box",
    "x": 1570,
    "y": 811,
    "w": 82,
    "h": 32
  }]
}

С этого момента, когда пользователь пытается войти, он получает письмо с OTP, который нужно ввести до завершения входа.

## Использование MFA через email

Чтобы использовать MFA через email, перейдите в свой профиль и нажмите на вкладку «security».

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/mfa/mfa-settings-auth.png",
    "alias": "authentication-mfa-mfa-settings-auth-email",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": []
}

Нажмите кнопку btn:[configure] в разделе «дополнительная безопасность с одноразовым паролем через email».

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/mfa/mfa-settings-auth-email-enabled.png",
    "alias": "authentication-mfa-mfa-settings-auth-email-enabled",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": []
}

При следующем входе вам нужно будет подтвердить вход, введя OTP, отправленный на вашу почту.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/mfa/mfa-login-confirm.png",
    "alias": "authentication-mfa-mfa-login-confirm",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": []
}

## Использование MFA через приложение-аутентификатор

Чтобы использовать MFA через email, перейдите в свой профиль и нажмите на вкладку «security».

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/mfa/mfa-settings-auth.png",
    "alias": "authentication-mfa-mfa-settings-auth-app",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": []
}

Нажмите кнопку btn:[configure] в разделе «дополнительная безопасность с мобильным приложением (одноразовый пароль на основе времени)».

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/mfa/mfa-settings-auth-authenticator-add.png",
    "alias": "authentication-mfa-mfa-settings-auth-authenticator-add",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": []
}

Откройте ваше приложение-аутентификатор и настройте LowCoooode, отсканировав предоставленный QR-код или вручную указав параметры, отображаемые на экране аутентификации.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/mfa/mfa-settings-auth-authenticator-enabled.png",
    "alias": "authentication-mfa-mfa-settings-auth-authenticator-enabled",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": []
}

При следующем входе вам нужно будет подтвердить вход, введя OTP, сгенерированный вашим приложением-аутентификатором.

[annotation,role="data-zoomable"]
{
  "image": {
    "rel": "authentication/mfa/mfa-login-confirm.png",
    "alias": "authentication-mfa-mfa-login-confirm",
    "w": 1920,
    "h": 1080
  },
  "view": {},
  "annotations": []
}
