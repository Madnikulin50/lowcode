package options

import (
	"github.com/madnikulin50/lowcode/server/codegen/schema"
)

webapp: schema.#optionsGroup & {
	handle: "webapp"

	options: {
		scss_dir_path: {
			description: "Path to custom SCSS source files directory"
		}
	}
}
