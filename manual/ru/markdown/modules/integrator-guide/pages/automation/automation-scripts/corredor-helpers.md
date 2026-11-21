# Хелперы Corredor

Хелперы Corredor реализуют наиболее распространённые операции, такие как создание записей и отправка электронных писем.

!!! tip
    Хелперы Corredor поставляются предварительно настроенными в [контексте выполнения](modules/integrator-guide/pages/automation/automation-scripts/index.md#execution-context).


Хелперы Corredor учитывают контекст, то есть могут автоматически определять базовые параметры, необходимые для операции.
Например, при создании записи хелперы Corredor будут знать, какое пространство имён, модуль и запись вы используете.

!!! tip
    Вы можете использовать хелперы Corredor за пределами скриптов автоматизации.


Исходный код можно найти на [GitHub/lowcode/lowcode-js](https://github.com/lowcode/lowcode-js).
Пакеты NPM можно найти на [NPM/@lowcode/lowcode-js](https://www.npmjs.com/package/@lowcode/lowcode-js).

!!! important
    Убедитесь, что вы используете ту же версию пакета `@lowcode/lowcode-js`, что и ваш инстанс LowCoooode.


При использовании нашего пакета `lowcode-js` клиенты API можно импортировать с помощью `import { corredor } from '@lowcode/lowcode-js'`.

.Доступные хелперы Corredor LowCoooode:
- [`corredor.SystemHelper`](https://github.com/lowcode/lowcode-js/blob/2021.3.x/src/corredor/helpers/system.ts)
- [`corredor.ComposeHelper`](https://github.com/lowcode/lowcode-js/blob/2021.3.x/src/corredor/helpers/compose.ts)

!!! note
    *DevNote*: обратите внимание на хелпер compose-ui.
