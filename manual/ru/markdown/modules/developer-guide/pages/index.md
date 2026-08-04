# Руководство разработчика платформы

Руководство разработчика платформы описывает процесс разработки основных компонентов платформы LowCoooode.

Руководство предназначено для ключевых разработчиков LowCoooode, а также для участников open source.

Если вы хотите узнать больше о разработке low-code приложений, обратитесь к [Integrator Guide](modules/integrator-guide/pages/index.md)

## Репозитории GitHub

Основная кодовая база LowCoooode находится в {GIT*REPO*LINK_PREFIX}[`lowcode` монорепозитории].
Монорепозиторий включает сервер, веб-приложения, NPM-пакеты и Discovery.
LowCoooode Corredor находится в собственном [репозитории](https://github.com/{GIT*REPO*GROUP}/{GIT*REPO*PREFIX}-server-corredor)

## Диаграмма компонентов

![LowCoooode component diagram](build-pipelines.png)

## Контрольный список внесения вклада

Все участники должны следовать контрольному списку внесения вклада, чтобы помочь сохранить согласованность и порядок по мере роста проекта.

.В приведённой ниже таблице перечислены пункты контрольного списка внесения вклада:
[cols="1s,5a"]
|===
| [#contribution-checklist-implement]#[contribution-checklist-implement,Реализация](#contribution-checklist-implement,Реализация)#
|
Реализуйте исправление ошибки, функцию или общее улучшение на основе трекера задач GitHub или собственных наблюдений.

.Реализация должна быть согласована с остальной частью компонента:
- [LowCoooode сервер](modules/developer-guide/pages/lowcode-server/index.md)
- [Веб-приложения](modules/developer-guide/pages/web-applications/index.md)
- [lowcode-js](modules/developer-guide/pages/lowcode-js/index.md)
- [lowcode-vue](modules/developer-guide/pages/lowcode-vue/index.md)
- [документация](modules/developer-guide/pages/documentation/index.md)

!!! note
    *DevNote* добавьте ссылку на индексатор, когда он станет доступен


**Внешние участники** должны отправлять свои изменения в виде pull request, который должен быть проверен ключевым контрибьютором.

**Ключевые контрибьюторы** могут вносить свои изменения непосредственно в версионные ветки, но для более крупных модификаций им всё же следует запрашивать рецензирование коллег.

| [#contribution-checklist-test]#[contribution-checklist-test,Тестирование](#contribution-checklist-test,Тестирование)#
|
Определите все требуемые тесты: модульные, интеграционные и ручное тестирование.
Любое дополнение к проекту должно определять как минимум базовые модульные и интеграционные тесты, чтобы помочь обеспечить стабильность системы и облегчить будущую разработку.

!!! note
    *DevNote* добавьте ссылку на руководство по тестированию


| [#contribution-checklist-document]#[contribution-checklist-document,Документация](#contribution-checklist-document,Документация)#
|
Документируйте свою работу, чтобы другие контрибьюторы, разработчики low-code платформы и конечные пользователи знали о вашем дополнении.

Обратитесь к [документации по документации](modules/developer-guide/pages/documentation/index.md) за подробностями о создании документации.

|===

## Куда дальше

Чтобы узнать больше о настройке и разработке для каждого компонента, обратитесь к [LowCoooode сервер](modules/developer-guide/pages/lowcode-server/index.md), [Веб-приложения](modules/developer-guide/pages/web-applications/index.md), [lowcode-js](modules/developer-guide/pages/lowcode-js/index.md) или [lowcode-vue](modules/developer-guide/pages/lowcode-vue/index.md).

Чтобы узнать больше о нашем цикле релизов и о том, как мы выпускаем продукт (как компилируется код и как собираются образы), обратитесь к документации [цикл релизов](modules/developer-guide/pages/release-cycle/index.md).

Чтобы узнать, как внести вклад в документацию, обратитесь к [документации по документации](modules/developer-guide/pages/documentation/index.md).
