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
