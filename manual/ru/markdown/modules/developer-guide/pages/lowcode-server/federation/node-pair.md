# Сопряжение узлов

**Сопряжение узлов** — это процесс **установления федеративной сети** между двумя инстансами LowCoooode (узлами).

Сопряжение узлов состоит из [node-identification,**идентификации узлов**](#node-identification,**идентификации узлов**), которая **идентифицирует два узла** — обеспечивает связь; и [node-handshake,**рукопожатия узлов**](#node-handshake,**рукопожатия узлов**), которое **обменивается учётными данными аутентификации** для безопасной связи.

.Диаграмма описывает весь процесс сопряжения узлов. Далее он разбивается на идентификацию узлов и рукопожатие узлов и подробно описывается.
[plantuml,node-pair,svg,role=sequence]
@startuml
skinparam responseMessageBelowArrow true

actor "Administrator A" as AdministratorA
participant "Node A" as NodeA
participant "Node B" as NodeB
actor "Administrator B" as AdministratorB

## Node identification ==
AdministratorA->NodeA: Register federated node B.
AdministratorA-->AdministratorB: Send node URI.
note right of AdministratorA
Administrator A sends the node URI
to administrator B via a secure channel.
end note
...
AdministratorB->NodeB: Register federated node A.
note left of AdministratorB
Administrator B registers node A
using the node URI.
end note


## Node handshake ==
AdministratorB->NodeB: Initialize the handshake.
note over NodeB #FFAAAA
Initialize node B for the federated network.
Create a system user, generate authentication
token, and configure the system.
end note

NodeB->NodeA: Request the handshake.
NodeA-->AdministratorA: Notify administrator A.
note right of AdministratorA #ffffff
Administrator A must manually
approve the handshake request.
end note
...
AdministratorA->NodeA: Approve the handshake request.
note over NodeA #FFAAAA
Initialize node A for the federated network.
Create a system user, generate authentication
token, and configure the system.
end note
NodeA->NodeB: Complete the handshake.
@enduml

<a id="node-identification"></a>
## Идентификация узлов

Шаг идентификации узлов **обменивается информацией** об узлах, необходимой для **установления подключения** (**URL-адрес**, **имя узла** и некоторые другие **метаданные**).

!!! note
    Шаг идентификации узлов *не обменивается никакими токенами аутентификации*, кроме токена OTT.


.The diagram outlines the node identification step which exchanges the information that are required to establish a connection.
[plantuml,node-pair,svg,role=sequence]
@startuml
skinparam responseMessageBelowArrow true

actor "Administrator A" as AdministratorA
participant "Node A" as NodeA
participant "Node B" as NodeB
actor "Administrator B" as AdministratorB

AdministratorA->NodeA: Register federated node B.
AdministratorA-->AdministratorB: Send node URI.
note right of AdministratorA
Administrator A sends the node URI
to administrator B via a secure channel.
end note
...
AdministratorB->NodeB: Register federated node A.
note left of AdministratorB
Administrator B registers node A
using the node URI.
end note
@enduml


Администратор узла A регистрирует узел B и генерирует URI узла::
    Шаг регистрации узла **позволяет узлу A узнать об узле B**.
    Сгенерированный URI узла **идентифицирует узел A** и выглядит так: `lowcode://$NODE*ID*A:$OTT@$DOMAIN_A?name=$NAME`.

!!! note
    `$OTT` позволяет выполнить первоначальную аутентификацию при выполнении <<node-handshake,рукопожатия ниже>>.


Администратор узла A отправляет URI узла администратору узла B::
    Переданный **URI узла** позволяет администратору узла B быстро **зарегистрировать узел A**.

!!! note
    Этот шаг *выполняется вручную* администраторами узлов.
    Два администратора *должны использовать защищённый канал* для обмена этой информацией.


Администратор узла B регистрирует узел A с помощью URI узла::
    Шаг регистрации узла **позволяет узлу B узнать об узле A**.
    Оба узла идентифицированы и готовы выполнить [node-handshake,**рукопожатие узлов**](#node-handshake,**рукопожатие узлов**).

<a id="node-handshake"></a>
## Рукопожатие узлов

Шаг рукопожатия узлов **настраивает узлы** и **обменивается токенами аутентификации**, которые узлы используют для доступа к защищённым ресурсам.

LowCoooode Federation использует уже установленный механизм аутентификации LowCoooode, применяя **системных пользователей и JWT-токены** (далее называемые токеном).

Это позволяет нам **уменьшить потенциальные дыры в безопасности** и **использовать наш механизм контроля доступа RBAC**.

!!! note
    Все *токены аутентификации уникальны*, даже в пределах одной пары узлов.


.The diagram outlines the node handshake step which exchanges the authentication tokens used to access protected resources.
[plantuml,node-pair,svg,role=sequence]
@startuml
skinparam responseMessageBelowArrow true

actor "Administrator A" as AdministratorA
participant "Node A" as NodeA
participant "Node B" as NodeB
actor "Administrator B" as AdministratorB

AdministratorB->NodeB: Initialize the handshake.
note over NodeB #FFAAAA
Initialize node B for the federated network.
Create a system user, generate authentication
token, and configure the system.
end note

NodeB->NodeA: Request the handshake.
NodeA-->AdministratorA: Notify administrator A.
note right of AdministratorA #ffffff
Administrator A must manually
approve the handshake request.
end note
...
AdministratorA->NodeA: Approve the handshake request.
note over NodeA #FFAAAA
Initialize node A for the federated network.
Create a system user, generate authentication
token, and configure the system.
end note
NodeA->NodeB: Complete the handshake.
@enduml


Администратор узла B инициализирует рукопожатие с узлом A::
    Узел B инициализирует состояние и генерирует `$TOKEN_B`, который может использоваться узлом A при доступе к защищённым ресурсам.

Узел B отправляет запрос рукопожатия узлу A::
    Администратор узла A уведомляется (по электронной почте), что узел B хочет установить федеративную сеть.
    Запрос рукопожатия **должен быть вручную подтверждён** администратором узла A.

!!! important
    Этот запрос аутентифицируется токеном `$OTT` (сгенерированным на шаге <<node-identification,идентификации узлов>>), *вне* стандартного *механизма аутентификации*.
    
    Фактические токены аутентификации аутентифицируют все последующие запросы.


Администратор узла A подтверждает запрос рукопожатия::
    Узел A инициализирует состояние и генерирует `$TOKEN_A`, который может использоваться узлом B при доступе к защищённым ресурсам.

Узел A завершает шаг рукопожатия::
    Узел A отправляет сгенерированный `$TOKEN_A` узлу B с подтверждением того, что рукопожатие прошло успешно.
