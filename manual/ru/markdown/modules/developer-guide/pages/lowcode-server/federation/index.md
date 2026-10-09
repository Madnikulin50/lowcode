<a id="federation"></a>
# Federation

!!! caution
    LowCoooode Federation в настоящее время находится в экспериментальном режиме и по умолчанию отключён,
    Задайте переменную `.env` `FEDERATION_ENABLED` равной `true`, если хотите его включить.


LowCoooode Federation позволяет разным **инстансам LowCoooode** устанавливать **федеративную сеть** для **свободного и безопасного** обмена информацией.

Федеративная сеть состоит из узлов-**источников** и узлов-**назначений**, где каждый узел может выполнять обе роли.
Ознакомьтесь с [глоссарием](modules/developer-guide/pages/lowcode-server/federation/lowcode-server/federation/glossary.md) для справки по нашей терминологии.

.Диаграмма предоставляет абстрактный обзор всей системы LowCoooode Federation. Она состоит из трёх основных частей — сопряжения узлов (узлы устанавливают федеративную сеть), синхронизации структур (узлы договариваются о том, чем будут делиться) и синхронизации данных (узел-назначение обращается к данным на узле-источнике).
[plantuml,federation-overview,svg,role=sequence]
@startuml
skinparam ParticipantPadding 50
participant "Node A\n(Origin Node)" as NodeA
participant "Node B\n(Destination Node)" as NodeB

## Node pairing ==
NodeA <-> NodeB: Pair nodes.
note over NodeA, NodeB: The two nodes identify and exchange authentication credentials via a node handshake.
...
## Structure syncing ==
NodeA -> NodeA: Expose structures.
note over NodeA
The origin node determines what structures
it wishes to expose to the destination node.
end note
NodeA -> NodeB: Structure sync.

NodeB -> NodeB: Module mapping.
note over NodeB
The destination node determines what fields
from the origin structure it wishes to consume.
The destination node also specifies where
the data should be stored (field mapping).
end note
...
## Data syncing ==
NodeA -> NodeB: Data sync.

@enduml

## Глоссарий

Мы рекомендуем вам сначала ознакомиться с [**глоссарием**](modules/developer-guide/pages/lowcode-server/federation/lowcode-server/federation/glossary.md), чтобы вы поняли нашу терминологию.
Это поможет вам следить за ходом изложения и понимать, о чём мы говорим.

## Безопасность и логирование

В [**«Безопасность и логирование»**](modules/developer-guide/pages/lowcode-server/federation/lowcode-server/federation/security-logging.md) описано, как LowCoooode Federation обрабатывает **аутентификацию узлов и доступ к защищённым ресурсам**, интеграцию механизма **контроля доступа RBAC** и интеграцию логирования в **журнал действий**.

## Сопряжение узлов

[**Сопряжение узлов**](modules/developer-guide/pages/lowcode-server/federation/lowcode-server/federation/node-pair.md) — это процесс **установления федеративной сети** между двумя инстансами LowCoooode (узлами).

LowCoooode Federation определяет **рукопожатие при сопряжении узлов**, которое позволяет двум узлам безопасно обмениваться учётными данными аутентификации.

## Синхронизация узлов

[**Синхронизация узлов**](modules/developer-guide/pages/lowcode-server/federation/lowcode-server/federation/node-sync.md) — это процесс определения **какими данными узел-источник хочет делиться**, и фактическая **синхронизация данных между двумя узлами**.

.Синхронизация узлов состоит из двух частей:
- [**синхронизация структур**](modules/developer-guide/pages/lowcode-server/federation/lowcode-server/federation/node-sync.md#structure-sync) определяет структуры, к которым узел-назначение может получить доступ,
- [**синхронизация данных**](modules/developer-guide/pages/lowcode-server/federation/lowcode-server/federation/node-sync.md#data-sync) синхронизирует данные с узла-источника на узел-назначение.

## Справочник API

[**Справочник API**](modules/developer-guide/pages/lowcode-server/federation/lowcode-server/federation/api/index.md) предоставляет полный список **доступных эндпоинтов API** с рабочими **примерами cURL**.

## Локальная разработка

[**Заметки по разработке**](modules/developer-guide/pages/lowcode-server/federation/lowcode-server/federation/dev-notes.md) проведут вас через настройку инстанса LowCoooode федеративного источника и назначения.
