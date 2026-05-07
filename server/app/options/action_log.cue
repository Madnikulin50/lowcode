package options

import (
	"github.com/madnikulin50/lowcode/server/codegen/schema"
)

actionLog: schema.#optionsGroup & {
	handle: "action-log"
	env:    "ACTIONLOG"
	options: {
		enabled: {
			type:          "bool"
			defaultGoExpr: "true"
		}
		debug: {
			type: "bool"
		}
		workflow_functions_enabled: {
			type: "bool"
		}
	}
	title: "Actionlog"
}
