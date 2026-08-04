# Resource Operation Reference

:leveloffset: +1

# System

:leveloffset: +1

# System

[cols="1s,5a,5a"]
|===
| Operation| Description | Default


| [#rbac-system-grant]#[rbac-system-grant,grant](#rbac-system-grant,grant)#
| Manage system permissions
| Deny


| [#rbac-system-action-log.read]#[rbac-system-action-log.read,action-log.read](#rbac-system-action-log.read,action-log.read)#
| Access to action log
| Deny


| [#rbac-system-settings.read]#[rbac-system-settings.read,settings.read](#rbac-system-settings.read,settings.read)#
| Read system settings
| Deny


| [#rbac-system-settings.manage]#[rbac-system-settings.manage,settings.manage](#rbac-system-settings.manage,settings.manage)#
| Manage system settings
| Deny


| [#rbac-system-auth-client.create]#[rbac-system-auth-client.create,auth-client.create](#rbac-system-auth-client.create,auth-client.create)#
| Create auth clients
| Deny


| [#rbac-system-auth-clients.search]#[rbac-system-auth-clients.search,auth-clients.search](#rbac-system-auth-clients.search,auth-clients.search)#
| List, search or filter auth clients
| Deny


| [#rbac-system-role.create]#[rbac-system-role.create,role.create](#rbac-system-role.create,role.create)#
| Create roles
| Deny


| [#rbac-system-roles.search]#[rbac-system-roles.search,roles.search](#rbac-system-roles.search,roles.search)#
| List, search or filter roles
| Deny


| [#rbac-system-user.create]#[rbac-system-user.create,user.create](#rbac-system-user.create,user.create)#
| Create users
| Deny


| [#rbac-system-users.search]#[rbac-system-users.search,users.search](#rbac-system-users.search,users.search)#
| List, search or filter users
| Deny


| [#rbac-system-dal-connection.create]#[rbac-system-dal-connection.create,dal-connection.create](#rbac-system-dal-connection.create,dal-connection.create)#
| Create DAL connections
| Deny


| [#rbac-system-dal-connections.search]#[rbac-system-dal-connections.search,dal-connections.search](#rbac-system-dal-connections.search,dal-connections.search)#
| List, search or filter DAL connections
| Deny


| [#rbac-system-dal-sensitivity-level.manage]#[rbac-system-dal-sensitivity-level.manage,dal-sensitivity-level.manage](#rbac-system-dal-sensitivity-level.manage,dal-sensitivity-level.manage)#
| Can manage DAL sensitivity levels
| Deny


| [#rbac-system-application.create]#[rbac-system-application.create,application.create](#rbac-system-application.create,application.create)#
| Create applications
| Deny


| [#rbac-system-applications.search]#[rbac-system-applications.search,applications.search](#rbac-system-applications.search,applications.search)#
| List, search or filter auth clients
| Deny


| [#rbac-system-application.flag.self]#[rbac-system-application.flag.self,application.flag.self](#rbac-system-application.flag.self,application.flag.self)#
| Manage private flags for applications
| Deny


| [#rbac-system-application.flag.global]#[rbac-system-application.flag.global,application.flag.global](#rbac-system-application.flag.global,application.flag.global)#
| Manage global flags for applications
| Deny


| [#rbac-system-template.create]#[rbac-system-template.create,template.create](#rbac-system-template.create,template.create)#
| Create template
| Deny


| [#rbac-system-templates.search]#[rbac-system-templates.search,templates.search](#rbac-system-templates.search,templates.search)#
| List, search or filter templates
| Deny


| [#rbac-system-report.create]#[rbac-system-report.create,report.create](#rbac-system-report.create,report.create)#
| Create report
| Deny


| [#rbac-system-reports.search]#[rbac-system-reports.search,reports.search](#rbac-system-reports.search,reports.search)#
| List, search or filter reports
| Deny


| [#rbac-system-reminder.assign]#[rbac-system-reminder.assign,reminder.assign](#rbac-system-reminder.assign,reminder.assign)#
|  Assign reminders
| Deny


| [#rbac-system-queue.create]#[rbac-system-queue.create,queue.create](#rbac-system-queue.create,queue.create)#
| Create messagebus queues
| Deny


| [#rbac-system-queues.search]#[rbac-system-queues.search,queues.search](#rbac-system-queues.search,queues.search)#
| List, search or filter messagebus queues
| Deny


| [#rbac-system-apigw-route.create]#[rbac-system-apigw-route.create,apigw-route.create](#rbac-system-apigw-route.create,apigw-route.create)#
| Create API gateway route
| Deny


| [#rbac-system-apigw-routes.search]#[rbac-system-apigw-routes.search,apigw-routes.search](#rbac-system-apigw-routes.search,apigw-routes.search)#
| List search or filter API gateway routes
| Deny


| [#rbac-system-resource-translations.manage]#[rbac-system-resource-translations.manage,resource-translations.manage](#rbac-system-resource-translations.manage,resource-translations.manage)#
| List, search, create, or update resource translations
| Deny


| [#rbac-system-dal-schema-alterations.manage]#[rbac-system-dal-schema-alterations.manage,dal-schema-alterations.manage](#rbac-system-dal-schema-alterations.manage,dal-schema-alterations.manage)#
| List, search, apply, or dismiss DAL alterations
| Deny


| [#rbac-system-data-privacy-request.create]#[rbac-system-data-privacy-request.create,data-privacy-request.create](#rbac-system-data-privacy-request.create,data-privacy-request.create)#
| Create data privacy requests
| Deny


| [#rbac-system-data-privacy-requests.search]#[rbac-system-data-privacy-requests.search,data-privacy-requests.search](#rbac-system-data-privacy-requests.search,data-privacy-requests.search)#
| List, search or filter data privacy requests
| Deny


|===


<!-- include: include::./resource.attachment.gen.adoc[] -->

# application

[cols="1s,5a,5a"]
|===
| Operation| Description | Default


| [#rbac-application-read]#[rbac-application-read,read](#rbac-application-read,read)#
| Read application
| Deny


| [#rbac-application-update]#[rbac-application-update,update](#rbac-application-update,update)#
| Update application
| Deny


| [#rbac-application-delete]#[rbac-application-delete,delete](#rbac-application-delete,delete)#
| Delete application
| Deny


|===


# apigw-route

[cols="1s,5a,5a"]
|===
| Operation| Description | Default


| [#rbac-apigw-route-read]#[rbac-apigw-route-read,read](#rbac-apigw-route-read,read)#
| Read API Gateway route
| Deny


| [#rbac-apigw-route-update]#[rbac-apigw-route-update,update](#rbac-apigw-route-update,update)#
| Update API Gateway route
| Deny


| [#rbac-apigw-route-delete]#[rbac-apigw-route-delete,delete](#rbac-apigw-route-delete,delete)#
| Delete API Gateway route
| Deny


|===


<!-- include: include::./resource.apigw-filter.gen.adoc[] -->

# auth-client

[cols="1s,5a,5a"]
|===
| Operation| Description | Default


| [#rbac-auth-client-read]#[rbac-auth-client-read,read](#rbac-auth-client-read,read)#
| Read authorization client
| Deny


| [#rbac-auth-client-update]#[rbac-auth-client-update,update](#rbac-auth-client-update,update)#
| Update authorization client
| Deny


| [#rbac-auth-client-delete]#[rbac-auth-client-delete,delete](#rbac-auth-client-delete,delete)#
| Delete authorization client
| Deny


| [#rbac-auth-client-authorize]#[rbac-auth-client-authorize,authorize](#rbac-auth-client-authorize,authorize)#
| Authorize authorization client
| Deny


|===


<!-- include: include::./resource.auth-confirmed-client.gen.adoc[] -->

<!-- include: include::./resource.auth-session.gen.adoc[] -->

<!-- include: include::./resource.auth-oa2token.gen.adoc[] -->

<!-- include: include::./resource.credential.gen.adoc[] -->

# data-privacy-request

[cols="1s,5a,5a"]
|===
| Operation| Description | Default


| [#rbac-data-privacy-request-read]#[rbac-data-privacy-request-read,read](#rbac-data-privacy-request-read,read)#
| Read data privacy request
| Deny


| [#rbac-data-privacy-request-approve]#[rbac-data-privacy-request-approve,approve](#rbac-data-privacy-request-approve,approve)#
| Approve/Reject data privacy request
| Deny


|===


<!-- include: include::./resource.data-privacy-request-comment.gen.adoc[] -->

# queue

[cols="1s,5a,5a"]
|===
| Operation| Description | Default


| [#rbac-queue-read]#[rbac-queue-read,read](#rbac-queue-read,read)#
| Read queue
| Deny


| [#rbac-queue-update]#[rbac-queue-update,update](#rbac-queue-update,update)#
| Update queue
| Deny


| [#rbac-queue-delete]#[rbac-queue-delete,delete](#rbac-queue-delete,delete)#
| Delete queue
| Deny


| [#rbac-queue-queue.read]#[rbac-queue-queue.read,queue.read](#rbac-queue-queue.read,queue.read)#
| Read from queue
| Deny


| [#rbac-queue-queue.write]#[rbac-queue-queue.write,queue.write](#rbac-queue-queue.write,queue.write)#
| Write to queue
| Deny


|===


<!-- include: include::./resource.queue-message.gen.adoc[] -->

<!-- include: include::./resource.reminder.gen.adoc[] -->

# report

[cols="1s,5a,5a"]
|===
| Operation| Description | Default


| [#rbac-report-read]#[rbac-report-read,read](#rbac-report-read,read)#
| Read report
| Deny


| [#rbac-report-update]#[rbac-report-update,update](#rbac-report-update,update)#
| Update report
| Deny


| [#rbac-report-delete]#[rbac-report-delete,delete](#rbac-report-delete,delete)#
| Delete report
| Deny


| [#rbac-report-run]#[rbac-report-run,run](#rbac-report-run,run)#
| Run report
| Deny


|===


<!-- include: include::./resource.resource-translation.gen.adoc[] -->

# role

[cols="1s,5a,5a"]
|===
| Operation| Description | Default


| [#rbac-role-read]#[rbac-role-read,read](#rbac-role-read,read)#
| Read role
| Deny


| [#rbac-role-update]#[rbac-role-update,update](#rbac-role-update,update)#
| Update role
| Deny


| [#rbac-role-delete]#[rbac-role-delete,delete](#rbac-role-delete,delete)#
| Delete role
| Deny


| [#rbac-role-members.manage]#[rbac-role-members.manage,members.manage](#rbac-role-members.manage,members.manage)#
| Manage members
| Deny


|===


<!-- include: include::./resource.role-member.gen.adoc[] -->

<!-- include: include::./resource.settings.gen.adoc[] -->

# template

[cols="1s,5a,5a"]
|===
| Operation| Description | Default


| [#rbac-template-read]#[rbac-template-read,read](#rbac-template-read,read)#
| Read template
| Deny


| [#rbac-template-update]#[rbac-template-update,update](#rbac-template-update,update)#
| Update template
| Deny


| [#rbac-template-delete]#[rbac-template-delete,delete](#rbac-template-delete,delete)#
| Delete template
| Deny


| [#rbac-template-render]#[rbac-template-render,render](#rbac-template-render,render)#
| Render template
| Deny


|===


# user

[cols="1s,5a,5a"]
|===
| Operation| Description | Default


| [#rbac-user-read]#[rbac-user-read,read](#rbac-user-read,read)#
| Read user
| Deny


| [#rbac-user-update]#[rbac-user-update,update](#rbac-user-update,update)#
| Update user
| Deny


| [#rbac-user-delete]#[rbac-user-delete,delete](#rbac-user-delete,delete)#
| Delete user
| Deny


| [#rbac-user-suspend]#[rbac-user-suspend,suspend](#rbac-user-suspend,suspend)#
| Suspend user
| Deny


| [#rbac-user-unsuspend]#[rbac-user-unsuspend,unsuspend](#rbac-user-unsuspend,unsuspend)#
| Unsuspend user
| Deny


| [#rbac-user-email.unmask]#[rbac-user-email.unmask,email.unmask](#rbac-user-email.unmask,email.unmask)#
| Unmask email
| Deny


| [#rbac-user-name.unmask]#[rbac-user-name.unmask,name.unmask](#rbac-user-name.unmask,name.unmask)#
| Unmask name
| Deny


| [#rbac-user-impersonate]#[rbac-user-impersonate,impersonate](#rbac-user-impersonate,impersonate)#
| Impersonate user
| Deny


| [#rbac-user-credentials.manage]#[rbac-user-credentials.manage,credentials.manage](#rbac-user-credentials.manage,credentials.manage)#
| Manage user's credentials
| Deny


|===


# dal-connection

[cols="1s,5a,5a"]
|===
| Operation| Description | Default


| [#rbac-dal-connection-read]#[rbac-dal-connection-read,read](#rbac-dal-connection-read,read)#
| Read connection
| Deny


| [#rbac-dal-connection-update]#[rbac-dal-connection-update,update](#rbac-dal-connection-update,update)#
| Update connection
| Deny


| [#rbac-dal-connection-delete]#[rbac-dal-connection-delete,delete](#rbac-dal-connection-delete,delete)#
| Delete connection
| Deny


| [#rbac-dal-connection-dal-config.manage]#[rbac-dal-connection-dal-config.manage,dal-config.manage](#rbac-dal-connection-dal-config.manage,dal-config.manage)#
| Manage DAL configuration
| Deny


|===


<!-- include: include::./resource.dal-sensitivity-level.gen.adoc[] -->

<!-- include: include::./resource.dal-schema-alteration.gen.adoc[] -->


:leveloffset: -1


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


# Automation

:leveloffset: +1

# Automation

[cols="1s,5a,5a"]
|===
| Operation| Description | Default


| [#rbac-automation-grant]#[rbac-automation-grant,grant](#rbac-automation-grant,grant)#
| Manage automation permissions
| Deny


| [#rbac-automation-workflow.create]#[rbac-automation-workflow.create,workflow.create](#rbac-automation-workflow.create,workflow.create)#
| Create workflows
| Deny


| [#rbac-automation-triggers.search]#[rbac-automation-triggers.search,triggers.search](#rbac-automation-triggers.search,triggers.search)#
| List, search or filter triggers
| Deny


| [#rbac-automation-sessions.search]#[rbac-automation-sessions.search,sessions.search](#rbac-automation-sessions.search,sessions.search)#
| List, search or filter sessions
| Deny


| [#rbac-automation-workflows.search]#[rbac-automation-workflows.search,workflows.search](#rbac-automation-workflows.search,workflows.search)#
| List, search or filter workflows
| Deny


| [#rbac-automation-resource-translations.manage]#[rbac-automation-resource-translations.manage,resource-translations.manage](#rbac-automation-resource-translations.manage,resource-translations.manage)#
| List, search, create, or update resource translations
| Deny


|===


# workflow

[cols="1s,5a,5a"]
|===
| Operation| Description | Default


| [#rbac-workflow-read]#[rbac-workflow-read,read](#rbac-workflow-read,read)#
| Read workflow
| Deny


| [#rbac-workflow-update]#[rbac-workflow-update,update](#rbac-workflow-update,update)#
| Update workflow
| Deny


| [#rbac-workflow-delete]#[rbac-workflow-delete,delete](#rbac-workflow-delete,delete)#
| Delete workflow
| Deny


| [#rbac-workflow-undelete]#[rbac-workflow-undelete,undelete](#rbac-workflow-undelete,undelete)#
| Undelete workflow
| Deny


| [#rbac-workflow-execute]#[rbac-workflow-execute,execute](#rbac-workflow-execute,execute)#
| Execute workflow
| Deny


| [#rbac-workflow-triggers.manage]#[rbac-workflow-triggers.manage,triggers.manage](#rbac-workflow-triggers.manage,triggers.manage)#
| Manage workflow triggers
| Deny


| [#rbac-workflow-sessions.manage]#[rbac-workflow-sessions.manage,sessions.manage](#rbac-workflow-sessions.manage,sessions.manage)#
| Manage workflow sessions
| Deny


|===


<!-- include: include::./resource.session.gen.adoc[] -->

<!-- include: include::./resource.trigger.gen.adoc[] -->


:leveloffset: -1


:leveloffset: -1
