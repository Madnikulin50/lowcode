package options

import (
	"github.com/madnikulin50/lowcode/server/codegen/schema"
)

monitor: schema.#optionsGroup & {
	handle: "monitor"

	imports: [
		"\"time\"",
	]

	options: {
		interval: {
			type:          "time.Duration"
			defaultGoExpr: "5 * time.Minute"
			defaultValue:  "5m"
			description:   "Output (log) interval for monitoring."
		}
	}
	title: "Monitoring"
}
