# Синхронизация узлов

**Синхронизация узлов** — это процесс определения **какими данными узел-источник хочет делиться**, и фактическая **синхронизация данных между двумя узлами**.

.Синхронизация узлов состоит из двух частей:
- [structure-sync,**синхронизация структур**](#structure-sync,**синхронизация структур**) определяет структуры, к которым узел-назначение может получить доступ,
- [data-sync,**синхронизация данных**](#data-sync,**синхронизация данных**) синхронизирует данные с узла-источника на узел-назначение.

!!! note
    Два узла должны пройти [сопряжение узлов](modules/developer-guide/pages/lowcode-server/federation/lowcode-server/federation/node-pair.md), прежде чем они смогут выполнять какую-либо синхронизацию.


<a id="structure-sync"></a>
## Синхронизация структур

**Синхронизация структур** — это процесс определения **к каким структурам** узлу-назначению **разрешено обращаться**.
В нашем случае мы указываем **модули и поля модулей**.

Узел-источник имеет **полный контроль над тем, к каким данным** узлу-назначению **разрешено обращаться**.
Это может быть так же просто, как разрешение доступа к конкретным модулям, и так же сложно, как разрешение доступа только к конкретным полям модулей (настраивается в [панели администрирования LowCoooode Low Code](modules/integrator-guide/pages/federation/index.md#expose-module) или через API).

Узел-источник определяет предоставленные структуры::
    Узел-источник указывает, к каким структурам узлу-назначению разрешено обращаться (модули и поля модулей).
    Узел-источник также может изменить или удалить эти структуры после первоначальной настройки.

Узел-назначение потребляет общие структуры::
    Узел-назначение получает доступные структуры (модули и поля) с узла-источника и определяет набор сопоставлений, чтобы определить, как (если вообще) данные хранятся.
    Узел-назначение также может изменить или удалить эти сопоставления после первоначальной настройки.

<a id="data-sync"></a>
## Синхронизация данных

**Синхронизация данных** — это процесс синхронизации данных **с узла-источника на узел-назначение**.
В нашем случае мы синхронизируем записи.

Узел-назначение имеет **полный контроль** над **тем, как хранятся данные** (если вообще хранятся) через сопоставления полей.

!!! important
    Синхронизация данных использует сервисы и слой хранилища Low Code, что упрощает общую архитектуру.
    
    Проектирование системы позволяет Federation быть развязанным и перемещённым в другое место.


Эта часть синхронизации узлов длинна и сложна, поэтому мы предоставляем две диаграммы, чтобы помочь вам визуализировать её.

.The diagram outlines the entire data syncing process from the API request to field mapping and storage from the destination node perspective.
[plantuml,data-sync-destination,svg,role=sequence]
@startuml
skinparam responseMessageBelowArrow true
actor "Origin Node" as NodeOrigin

box "Destination Node" #f7f7f7

participant "Federation Service" as FederationService
participant "Compose Record Service" as ComposeRS
database Store
participant "Federation Data Service" as FDataService
participant Decoder

activate FederationService

NodeOrigin <-> FederationService: Get federated records.

Store <-> FederationService: Get module M properties and field mapping definitions.

note over Decoder
The decoder converts federated records
from a specific format into the internal
format used by LowCoooode.
end note
Decoder <-> FederationService: Decode federated records into the internal structure.

FDataService <-> FederationService: Apply field mapping to obtain the final set of records to create or update.

FederationService -> ComposeRS: Create or update records
activate ComposeRS
ComposeRS -> Store: Write data to the store layer.

ComposeRS -> FederationService: status
deactivate ComposeRS

FederationService -> Store: Update data sync status
deactivate FederationService
end box
@enduml


.The diagram outlines the entire data syncing process from the API request to field mapping and storage from the origin node perspective.
[plantuml,data-sync-origin,svg,role=sequence]
@startuml
skinparam responseMessageBelowArrow true
actor "Destination Node" as NodeDestination

box "Origin Node" #f7f7f7

participant "Federation Rest Controller" as FRestController
participant "Federation Service" as FederationService
participant "Compose Record Service" as ComposeRS
database Store
participant "Federation Data Service" as FDataService
participant Encoder

activate NodeDestination
NodeDestination -> FRestController: Get federated records.
note over NodeDestination
This will only provide the data
that was updated after the last
successful data sync.
end note
activate FRestController

FRestController -> FederationService: Get federated records.
activate FederationService

FederationService <-> Store: Get module M properties and mapped fields.

FederationService -> ComposeRS: Get filtered Compose records.

activate ComposeRS
ComposeRS <-> Store: Get filtered Compose records.

ComposeRS -> FederationService: List of filtered Compose records.
deactivate ComposeRS

FederationService <-> FDataService: Prepare final list of records to share.
note over FDataService
This step removes fields
we do not wish to share.
end note


note over Encoder
The encoder converts records
from a specific format into the
requested format.
end note
FederationService <-> Encoder: Encode records into the requested format.


FederationService -> FRestController: Final list of federated records.
deactivate FederationService

FRestController -> NodeDestination: Final list of federated records.
deactivate FRestController
deactivate NodeDestination

end box
@enduml


Узел-назначение запрашивает изменённые данные::
    Узел-назначение запрашивает **все изменения данных узла-источника**, которые произошли **после последней успешной** синхронизации данных.
    Изменения получаются для каждой структуры (модуля) на фиксированном эндпоинте.

Узел-источник предоставляет изменённые данные::
    Узел-источник определяет, какие данные (записи) изменились после последней успешной синхронизации данных.
    Данные проходят через **контроль доступа**, где мы удаляем любые значения, которыми не хотим делиться.
    Наконец, **данные кодируются** в запрошенный формат и предоставляются узлу-назначению.

Узел-назначение сохраняет данные::
    Узел-назначение сначала **декодирует данные**, затем **преобразует данные на основе определения сопоставления полей** и, наконец, **создаёт или обновляет записи хранилища**.

Узел-назначение обновляет статус синхронизации::
    Узел-назначение обновляет статус синхронизации данных, **фиксируя временные метки**, чтобы обеспечить будущую синхронизацию данных.
