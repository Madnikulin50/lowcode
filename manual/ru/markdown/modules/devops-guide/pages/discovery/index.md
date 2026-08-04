# LowCoooode Discovery

LowCoooode Discovery предоставляет мощный поисковый движок для взаимодействия с вашими данными.
LowCoooode Discovery определяет интуитивно понятный интерфейс для поиска и, в некоторых случаях, визуализации данных, таких как географические метаданные.

!!! note
    Обратитесь к [Multi Discovery Pgsql](modules/devops-guide/pages/discovery/examples/deploy-online/multi-discovery-pgsql.md) за примером.


Обратитесь к [menu:Руководство разработчика low-code платформы[LowCoooode Discovery](modules/integrator-guide/pages/discovery/index.md)], чтобы узнать, как настроить и использовать LowCoooode Discovery.

## Настройка сервера LowCoooode

Чтобы включить LowCoooode Discovery, получите доступ к Docker-контейнеру и выполните следующие команды:

.Включите LowCoooode Discovery для пространств имён Low Code:
```bash
```
lowcode-server settings set discovery.compose-namespaces.enabled true

.Включите LowCoooode Discovery для модулей Low Code:
```bash
```
lowcode-server settings set discovery.compose-modules.enabled true

.Включите LowCoooode Discovery для записей Low Code:
```bash
```
lowcode-server settings set discovery.compose-records.enabled true

.Включите LowCoooode Discovery для пользователей:
```bash
```
lowcode-server settings set discovery.system-users.enabled true

.Далее задайте следующие переменные `.env` для вашего сервера LowCoooode:
```env
```
DISCOVERY_ENABLED=true
DISCOVERY*BASE*URL=your-discovery-server-base-url
# Optional variable for debugging
# DISCOVERY_DEBUG=true

!!! note
    За примером онлайн-развёртывания обратитесь к [Multi Discovery Pgsql](modules/devops-guide/pages/discovery/examples/deploy-online/multi-discovery-pgsql.md).


```bash
```
!!! note
    Чтобы получить доступ к Docker-контейнеру, выполните следующую команду:
    
    docker-compose exec server bash
    ----


## Настройка сервера LowCoooode Discovery

### Требования к поисковому движку
Для работы Discovery требуется запущенный инстанс ***Elasticsearch*** или ***OpenSearch***.

!!! note
    **Важно:** Обратите внимание, что переменные конфигурации ниже используют префикс `ES_` (например, `ES_ADDRESS`, `ES_USERNAME`). Этот префикс сохраняется для обратной совместимости. Вы должны использовать эти переменные `ES_` независимо от того, является ли ваш поисковый движок Elasticsearch или OpenSearch.


### Создание клиентов аутентификации
Перед настройкой переменных окружения вы должны создать клиентов аутентификации (Auth Clients) в LowCoooode, чтобы безопасно подключить Discovery Indexer и Searcher к основному серверу.

1. Войдите в ***Админ-зону*** LowCoooode.
1. Перейдите в ***Система*** > ***Клиенты аутентификации***.
1. Нажмите ***Новый***, чтобы создать клиента (например, назовите его «Discovery Indexer — Private»).
1. Убедитесь, что ***Тип предоставления*** настроен на разрешение `client_credentials`.
1. Выберите конкретную роль, у которой есть необходимые разрешения на чтение данных, которые вы хотите индексировать.
1. Нажмите ***Отправить***.
1. Безопасно скопируйте сгенерированные ***Client ID*** (используемый как Key) и ***Client Secret***.
1. Повторите этот процесс для Protected Indexer, Public Indexer и Searcher.

### Переменные окружения
Настройте следующие переменные `.env`, заменив значения-заполнители данными вашего поискового движка и учётными данными вновь созданных Auth Clients:

```env
```
ES_ADDRESS=your-open-search-url
ES_USERNAME=your-open-search-username
ES_PASSWORD=your-open-search-password
# Set to true if connecting over HTTPS
ES_SECURE=false
# Re-indexing interval in seconds
ES*INDEX*INTERVAL=60

LOWCODE*SERVER*BASE_URL=your-lowcode-server-url
LOWCODE*SERVER*AUTH_URL=your-lowcode-server-url/auth

DISCOVERY*INDEXER*ENABLED=true
DISCOVERY*SEARCHER*ENABLED=true

# Private indexer — indexes data visible only to the owning user
DISCOVERY*INDEXER*PRIVATE*INDEX*CLIENT_KEY=private-index-client-key
DISCOVERY*INDEXER*PRIVATE*INDEX*CLIENT_SECRET=private-index-client-secret
# Protected indexer — indexes data visible to authenticated users
DISCOVERY*INDEXER*PROTECTED*INDEX*CLIENT_KEY=protected-index-client-key
DISCOVERY*INDEXER*PROTECTED*INDEX*CLIENT_SECRET=protected-index-client-secret
# Public indexer — indexes data visible to everyone
DISCOVERY*INDEXER*PUBLIC*INDEX*CLIENT_KEY=public-index-client-key
DISCOVERY*INDEXER*PUBLIC*INDEX*CLIENT_SECRET=public-index-client-secret

# Searcher — used by the Discovery search interface
DISCOVERY*SEARCHER*CLIENT_KEY=your-searcher-client-key
DISCOVERY*SEARCHER*CLIENT_SECRET=your-searcher-client-secret
