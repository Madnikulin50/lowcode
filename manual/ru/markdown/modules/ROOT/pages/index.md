# LowCoooode

LowCoooode — это Low-Code-платформа разработки с открытым исходным кодом, размещаемая самостоятельно, построенная на современных технологиях.
LowCoooode безопасна благодаря мощной системе контроля доступа, позволяющей задавать детальные разрешения.

LowCoooode стремится быть **безусловно заслуживающей доверия** в своих мотивациях и подходе к проектированию, разработке и сопровождению платформы.
Организации должны чувствовать, что их выбранная цифровая рабочая платформа всегда **под их контролем**, всегда **защищена** и постоянно **развивается в их интересах**.

!!! note
    Есть вопросы?
    Посетите https://lowcode.org[наш сайт] или https://forum.lowcode.org[свяжитесь с нами].


## Установка LowCoooode

!!! warning
    **Возможная потеря данных при использовании SQLite в памяти.**
    
    Мы настоятельно рекомендуем использовать другие движки БД (PostgreSQL или MySQL).
    Если вы всё же решили использовать SQLite, убедитесь, что используется постоянное хранилище.
    
    Используемый драйвер SQLite [mattn/go-sqlite3](https://pkg.go.dev/github.com/mattn/go-sqlite3) пересоздаёт всю базу данных при каждом новом подключении и удаляет всю базу данных при закрытии последнего подключения.
    
    Мы используем SQLite в основном для тестирования, поэтому для нас это не проблема.
    В будущих версиях мы поработаем над более надёжным решением.
    


Руководство DevOps [Devops Guide](modules/devops-guide/pages/index.md) проведёт вас через процесс установки для [демонстрационных/разработочных сред](modules/devops-guide/pages/index.md#deploy-offline) и [сред, приближённых к производственным](modules/devops-guide/pages/index.md#deploy-online).

Руководство DevOps также содержит разную дополнительную информацию: ссылки на конфигурацию [системы](modules/devops-guide/pages/references/configuration/server.md) и [Corredor](modules/devops-guide/pages/references/configuration/corredor.md), дополнительные [примеры автономного развёртывания](modules/devops-guide/pages/examples/deploy-offline/index.md), [примеры онлайн-развёртывания](modules/devops-guide/pages/examples/deploy-online/index.md), [резервные копии данных](modules/devops-guide/pages/maintenance/backups.md) и [устранение неполадок](modules/devops-guide/pages/troubleshooting/index.md).

!!! note
    Из коробки мы поддерживаем любую систему, на которой может работать Docker.
    Если вы хотите развернуть LowCoooode в другом месте (например, на «железе»), вам потребуется скомпилировать собственные бинарные файлы.
    
    *DevNote* добавить документацию по компиляции из исходного кода.


## Создание в LowCoooode

Руководство интегратора [Integrator Guide](modules/integrator-guide/pages/index.md) проведёт вас **через процесс интеграции**: от системы [Security Model](modules/integrator-guide/pages/security-model/index.md) до [Compose Configuration](modules/integrator-guide/pages/compose-configuration/index.md), [Automation](modules/integrator-guide/pages/automation/index.md), [Reporting](modules/integrator-guide/pages/reporting/index.md) и [Api Gw](modules/integrator-guide/pages/api-gw/index.md).

Руководство разработчика Low-Code-платформы также охватывает [Authentication](modules/integrator-guide/pages/authentication/index.md) и [Security Model](modules/integrator-guide/pages/security-model/index.md)

Есть также множество примеров «скопируй и вставь» и разная дополнительная информация: [Deploying](modules/integrator-guide/pages/automation/automation-scripts/deploying.md), отладка [скриптов автоматизации](modules/integrator-guide/pages/automation/automation-scripts/debugging.md) и советы по [Accessing LowCoooode](modules/integrator-guide/pages/accessing-lowcode/index.md).


## Обновление LowCoooode

!!! important
    При обновлении всегда сначала следует просмотреть [Changelog](modules/ROOT/pages/changelog/index.md) и [Upgrade Guide](upgrade-guide/index.md).
    
    Версии могут быть *несовместимы с предыдущими* и могут потребовать дополнительной работы для настройки.


Чтобы обновиться с `2022.9`, вы можете использовать [этот журнал изменений](modules/ROOT/pages/changelog/202303/index.md) и [это руководство по обновлению](upgrade-guide/index.md).

Все журналы изменений можно найти [здесь](modules/ROOT/pages/changelog/index.md), а все руководства по обновлению — [здесь](upgrade-guide/index.md).
