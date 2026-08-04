# Federation

:leveloffset: +1

# Federation

[cols="1s,5a,5a"]
|===
| Operation| Description | Default


| [#rbac-federation-grant]#[rbac-federation-grant,grant](#rbac-federation-grant,grant)#
| Manage federation permissions
| Deny


| [#rbac-federation-pair]#[rbac-federation-pair,pair](#rbac-federation-pair,pair)#
| Pair federation nodes
| Deny


| [#rbac-federation-settings.read]#[rbac-federation-settings.read,settings.read](#rbac-federation-settings.read,settings.read)#
| Read settings
| Deny


| [#rbac-federation-settings.manage]#[rbac-federation-settings.manage,settings.manage](#rbac-federation-settings.manage,settings.manage)#
| Manage settings
| Deny


| [#rbac-federation-node.create]#[rbac-federation-node.create,node.create](#rbac-federation-node.create,node.create)#
| Create new federation node
| Deny


| [#rbac-federation-nodes.search]#[rbac-federation-nodes.search,nodes.search](#rbac-federation-nodes.search,nodes.search)#
| List, search or filter federation nodes
| Deny


|===


# node

[cols="1s,5a,5a"]
|===
| Operation| Description | Default


| [#rbac-node-manage]#[rbac-node-manage,manage](#rbac-node-manage,manage)#
| Manage federation node
| Deny


| [#rbac-node-module.create]#[rbac-node-module.create,module.create](#rbac-node-module.create,module.create)#
| Create shared module
| Deny


|===


<!-- include: include::./resource.node-sync.gen.adoc[] -->

# exposed-module

[cols="1s,5a,5a"]
|===
| Operation| Description | Default


| [#rbac-exposed-module-manage]#[rbac-exposed-module-manage,manage](#rbac-exposed-module-manage,manage)#
| Manage exposed module module
| Deny


|===


# shared-module

[cols="1s,5a,5a"]
|===
| Operation| Description | Default


| [#rbac-shared-module-map]#[rbac-shared-module-map,map](#rbac-shared-module-map,map)#
| Map shared module
| Deny


|===


<!-- include: include::./resource.module-mapping.gen.adoc[] -->


:leveloffset: -1
