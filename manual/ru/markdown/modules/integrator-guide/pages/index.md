# Руководство разработчика платформы Low-Code

Руководство разработчика low-code платформы описывает процесс расширения LowCoooode для покрытия потребностей вашего бизнеса.
Интеграция может быть как простой, как изменение экрана входа для включения вашей фирменной графики, так и сложной, как реализация полностью пользовательского приложения Low Code.

Если вы ещё не настроили свой инстанс LowCoooode, обратитесь к [руководству DevOps](modules/devops-guide/pages/index.md).

## Модель аутентификации и безопасности

Для **аутентификации** LowCoooode реализует протокол OAuth2, где LowCoooode может выступать как в роли клиента OAuth2, так и в роли сервера.
Обратитесь к разделу [аутентификация](modules/integrator-guide/pages/authentication/index.md) за подробностями.
Экран входа можно полностью персонализировать, чтобы он соответствовал вашему бренду и давал вашим пользователям спокойствие при вводе своих учётных данных.
Обратитесь к разделу [menu:personalization[authentication](modules/integrator-guide/pages/personalization/auth.md)] за подробностями.

!!! tip
    Может быть хорошей идеей включить вашу фирменную графику, когда к вашему инстансу LowCoooode обращаются внешние пользователи.


Для **контроля доступа** LowCoooode реализует [механизм RBAC](https://en.wikipedia.org/wiki/Role-based*access*control), который позволяет вам точно настроить разрешения доступа под ваши нужды.
Обратитесь к разделу [Security Model](modules/integrator-guide/pages/security-model/index.md) за подробностями.

## Интернационализация

LowCoooode позволяет полностью перевести большинство аспектов системы.

Обратитесь к странице [Static](modules/integrator-guide/pages/i18n/static.md), чтобы узнать, как перевести пользовательский интерфейс.

Обратитесь к странице [Resource](modules/integrator-guide/pages/i18n/resource.md), чтобы узнать, как перевести ваши приложения LowCoooode Low Code.

## Доступ к LowCoooode через API

LowCoooode является API-центричным, что означает, что всё можно сделать через эндпоинт API.
Обратитесь к разделам [доступ к LowCoooode](modules/integrator-guide/pages/accessing-lowcode/index.md) за подробностями о [аутентификация](modules/integrator-guide/pages/accessing-lowcode/index.md#authentication), [эндпоинты API](modules/integrator-guide/pages/accessing-lowcode/index.md#api-endpoints), [формат данных](modules/integrator-guide/pages/accessing-lowcode/index.md#response-format) и [язык запросов](modules/integrator-guide/pages/accessing-lowcode/ql.md).

Если вы хотите подключить внешнее приложение к LowCoooode для обмена данными, вы можете сделать это с помощью наших [Node.js API-клиентов](modules/integrator-guide/pages/accessing-lowcode/api-clients.md).

## Создание пользовательского приложения Low Code

Обратитесь к странице [конфигурация Low Code](modules/integrator-guide/pages/compose-configuration/index.md), чтобы узнать, как настроить собственное приложение Low Code.
За дополнительными подробностями обратитесь к [menu:Low Code configuration[справочнику типов полей](modules/integrator-guide/pages/compose-configuration/field-types.md)] для справки по доступным типам полей и [menu:Low Code configuration[справочнику блоков страниц](modules/integrator-guide/pages/compose-configuration/page-blocks.md)] для справки по доступным блокам страниц.

Обратитесь к странице [Api Gw](modules/integrator-guide/pages/api-gw/index.md), чтобы узнать, как определить пользовательские эндпоинты webhook, которые могут использовать внешние сервисы.

Обратитесь к странице [Reporting](modules/integrator-guide/pages/reporting/index.md), чтобы узнать, как создавать пользовательские отчёты для визуализации данных вашего Low Code.

## Автоматизация ваших процессов

LowCoooode предоставляет мощный механизм автоматизации, который позволяет реализовать почти всё!
Автоматизация определяется либо через [рабочие процессы (menu:automation[workflow](modules/integrator-guide/pages/automation/workflows/index.md))], либо через [скрипты автоматизации (menu:automation[automation script](modules/integrator-guide/pages/automation/automation-scripts/index.md))].

**Рабочий процесс** — это упрощённая BPMN-диаграмма, которая позволяет описать автоматизацию в удобной для пользователя форме.
Рабочий процесс — рекомендуемый способ реализации автоматизации.
Рабочие процессы проще читать и сопровождать.

**Скрипт автоматизации** — это фрагмент JavaScript-кода, выполняющий желаемую операцию.

!!! caution
    Когда автоматизация требует нескольких сложных операций, которые изначально не поддерживаются механизмом рабочих процессов, скрипт автоматизации — лучший выбор.


## LowCoooode Discovery

LowCoooode Discovery предоставляет мощную поисковую систему для взаимодействия с вашими данными.
LowCoooode Discovery определяет интуитивно понятный интерфейс для поиска и, в некоторых случаях, визуализации данных, таких как географические метаданные.

Обратитесь к странице [Discovery](modules/integrator-guide/pages/discovery/index.md), чтобы узнать, как настроить и использовать LowCoooode Discovery

## Создание пользовательских документов

!!! important
    Чтобы включить рендеринг PDF-документов, вам нужно запустить Docker-контейнер https://github.com/gotenberg/gotenberg[gotenberg].
    
    Обратитесь к [руководству DevOps](modules/devops-guide/pages/pdf-renderer.md) за подробностями о том, как его настроить.


LowCoooode определяет гибкий механизм шаблонов, который позволяет вам создавать пользовательские документы (такие как маркетинговые письма и PDF-цитаты), которые вы отправляете своим контактам.
Механизм шаблонов в настоящее время поддерживает форматы **PDF**, **HTML** и **обычный текст**.

Обратитесь к разделу [шаблоны](modules/integrator-guide/pages/templates/index.md), чтобы получить обзор процесса определения шаблона и рендеринга документа.

!!! tip
    Вы можете запрашивать рендеринг документов из внешних приложений через REST API.
    
    *DevNote* add a reference to the endpoint.


## Федерация LowCoooode

LowCoooode Federation позволяет разным инстансам LowCoooode устанавливать федеративную сеть для свободного и безопасного обмена информацией.

Обратитесь к странице [Federation](modules/integrator-guide/pages/federation/index.md), чтобы узнать, как настроить и использовать Федерацию LowCoooode.

## Устранение неполадок

Свяжитесь с нами на нашем [форуме](https://forum.lowcode.org).
Любые отзывы, вопросы или предложения всегда приветствуются!
