package options

import (
	"github.com/madnikulin50/lowcode/server/codegen/schema"
)

environment: schema.#optionsGroup & {
	handle: "environment"
	options: {
		environment: {
			defaultValue: "production"
			env:          "ENVIRONMENT"
		}
	}
	title: "Environment"
}
