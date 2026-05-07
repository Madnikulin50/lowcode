package app

import (
	"github.com/madnikulin50/lowcode/server/codegen/schema"
	"github.com/madnikulin50/lowcode/server/app/options"
	"github.com/madnikulin50/lowcode/server/system"
	"github.com/madnikulin50/lowcode/server/compose"
	"github.com/madnikulin50/lowcode/server/automation"
	"github.com/madnikulin50/lowcode/server/federation"
)

corteza: schema.#platform & {
	"ident": "corteza"

	"options": [
		options.DB,
		options.HTTPClient,
		options.HTTPServer,
		options.RBAC,
		options.SCIM,
		options.SMTP,
		options.actionLog,
		options.apigw,
		options.auth,
		options.corredor,
		options.environment,
		options.eventbus,
		options.federation,
		options.limit,
		options.locale,
		options.log,
		options.messagebus,
		options.monitor,
		options.objectStore,
		options.provision,
		options.sentry,
		options.template,
		options.upgrade,
		options.waitFor,
		options.websocket,
		options.workflow,
		options.discovery,
		options.attachment,
		options.webapp,
	]

	// platform resources
	"resources": resources

	"components": [
		system.component,
		compose.component,
		automation.component,
		federation.component,
	]
}
