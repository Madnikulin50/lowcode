# Compose

:leveloffset: +1

# Compose

[cols="1s,5a,5a"]
|===
| Operation| Description | Default


| [#rbac-compose-grant]#[rbac-compose-grant,grant](#rbac-compose-grant,grant)#
| Manage compose permissions
| Deny


| [#rbac-compose-settings.read]#[rbac-compose-settings.read,settings.read](#rbac-compose-settings.read,settings.read)#
| Read settings
| Deny


| [#rbac-compose-settings.manage]#[rbac-compose-settings.manage,settings.manage](#rbac-compose-settings.manage,settings.manage)#
| Manage settings
| Deny


| [#rbac-compose-namespace.create]#[rbac-compose-namespace.create,namespace.create](#rbac-compose-namespace.create,namespace.create)#
| Create namespace
| Deny


| [#rbac-compose-namespaces.search]#[rbac-compose-namespaces.search,namespaces.search](#rbac-compose-namespaces.search,namespaces.search)#
| List, search or filter namespaces
| Deny


| [#rbac-compose-resource-translations.manage]#[rbac-compose-resource-translations.manage,resource-translations.manage](#rbac-compose-resource-translations.manage,resource-translations.manage)#
| List, search, create, or update resource translations
| Deny


|===


<!-- include: include::./resource.attachment.gen.adoc[] -->

# chart

[cols="1s,5a,5a"]
|===
| Operation| Description | Default


| [#rbac-chart-read]#[rbac-chart-read,read](#rbac-chart-read,read)#
| read
| Deny


| [#rbac-chart-update]#[rbac-chart-update,update](#rbac-chart-update,update)#
| update
| Deny


| [#rbac-chart-delete]#[rbac-chart-delete,delete](#rbac-chart-delete,delete)#
| delete
| Deny


|===


# module

[cols="1s,5a,5a"]
|===
| Operation| Description | Default


| [#rbac-module-read]#[rbac-module-read,read](#rbac-module-read,read)#
| read
| Deny


| [#rbac-module-update]#[rbac-module-update,update](#rbac-module-update,update)#
| update
| Deny


| [#rbac-module-delete]#[rbac-module-delete,delete](#rbac-module-delete,delete)#
| delete
| Deny


| [#rbac-module-record.create]#[rbac-module-record.create,record.create](#rbac-module-record.create,record.create)#
| Create record
| Deny


| [#rbac-module-owned-record.create]#[rbac-module-owned-record.create,owned-record.create](#rbac-module-owned-record.create,owned-record.create)#
| Create record with custom owner
| Deny


| [#rbac-module-records.search]#[rbac-module-records.search,records.search](#rbac-module-records.search,records.search)#
| List, search or filter records
| Deny


|===


# module-field

[cols="1s,5a,5a"]
|===
| Operation| Description | Default


| [#rbac-module-field-record.value.read]#[rbac-module-field-record.value.read,record.value.read](#rbac-module-field-record.value.read,record.value.read)#
| Read field value on records
| Deny


| [#rbac-module-field-record.value.update]#[rbac-module-field-record.value.update,record.value.update](#rbac-module-field-record.value.update,record.value.update)#
| Update field value on records
| Deny


|===


# namespace

[cols="1s,5a,5a"]
|===
| Operation| Description | Default


| [#rbac-namespace-read]#[rbac-namespace-read,read](#rbac-namespace-read,read)#
| read
| Deny


| [#rbac-namespace-update]#[rbac-namespace-update,update](#rbac-namespace-update,update)#
| update
| Deny


| [#rbac-namespace-delete]#[rbac-namespace-delete,delete](#rbac-namespace-delete,delete)#
| delete
| Deny


| [#rbac-namespace-manage]#[rbac-namespace-manage,manage](#rbac-namespace-manage,manage)#
| Access to namespace admin panel
| Deny


| [#rbac-namespace-module.create]#[rbac-namespace-module.create,module.create](#rbac-namespace-module.create,module.create)#
| Create module on namespace
| Deny


| [#rbac-namespace-modules.search]#[rbac-namespace-modules.search,modules.search](#rbac-namespace-modules.search,modules.search)#
| List, search or filter module on namespace
| Deny


| [#rbac-namespace-chart.create]#[rbac-namespace-chart.create,chart.create](#rbac-namespace-chart.create,chart.create)#
| Create chart on namespace
| Deny


| [#rbac-namespace-charts.search]#[rbac-namespace-charts.search,charts.search](#rbac-namespace-charts.search,charts.search)#
| List, search or filter chart on namespace
| Deny


| [#rbac-namespace-page.create]#[rbac-namespace-page.create,page.create](#rbac-namespace-page.create,page.create)#
| Create page on namespace
| Deny


| [#rbac-namespace-pages.search]#[rbac-namespace-pages.search,pages.search](#rbac-namespace-pages.search,pages.search)#
| List, search or filter pages on namespace
| Deny


|===


# page

[cols="1s,5a,5a"]
|===
| Operation| Description | Default


| [#rbac-page-read]#[rbac-page-read,read](#rbac-page-read,read)#
| read
| Deny


| [#rbac-page-update]#[rbac-page-update,update](#rbac-page-update,update)#
| update
| Deny


| [#rbac-page-delete]#[rbac-page-delete,delete](#rbac-page-delete,delete)#
| delete
| Deny


| [#rbac-page-page-layout.create]#[rbac-page-page-layout.create,page-layout.create](#rbac-page-page-layout.create,page-layout.create)#
| Create page layout on namespace
| Deny


| [#rbac-page-page-layouts.search]#[rbac-page-page-layouts.search,page-layouts.search](#rbac-page-page-layouts.search,page-layouts.search)#
| List, search or filter page layouts on namespace
| Deny


|===


# page-layout

[cols="1s,5a,5a"]
|===
| Operation| Description | Default


| [#rbac-page-layout-read]#[rbac-page-layout-read,read](#rbac-page-layout-read,read)#
| read
| Deny


| [#rbac-page-layout-update]#[rbac-page-layout-update,update](#rbac-page-layout-update,update)#
| update
| Deny


| [#rbac-page-layout-delete]#[rbac-page-layout-delete,delete](#rbac-page-layout-delete,delete)#
| delete
| Deny


|===


# record

[cols="1s,5a,5a"]
|===
| Operation| Description | Default


| [#rbac-record-read]#[rbac-record-read,read](#rbac-record-read,read)#
| read
| Deny


| [#rbac-record-update]#[rbac-record-update,update](#rbac-record-update,update)#
| update
| Deny


| [#rbac-record-delete]#[rbac-record-delete,delete](#rbac-record-delete,delete)#
| delete
| Deny


| [#rbac-record-undelete]#[rbac-record-undelete,undelete](#rbac-record-undelete,undelete)#
| undelete
| Deny


| [#rbac-record-owner.manage]#[rbac-record-owner.manage,owner.manage](#rbac-record-owner.manage,owner.manage)#
| owner.manage
| Deny


| [#rbac-record-revisions.search]#[rbac-record-revisions.search,revisions.search](#rbac-record-revisions.search,revisions.search)#
| revisions.search
| Deny


|===


<!-- include: include::./resource.record-revision.gen.adoc[] -->


:leveloffset: -1
