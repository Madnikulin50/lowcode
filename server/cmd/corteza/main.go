package main

import (
	"github.com/madnikulin50/lowcode/server/app"
	"github.com/madnikulin50/lowcode/server/pkg/cli"
	"github.com/madnikulin50/lowcode/server/pkg/logger"
)

func main() {
	// Initialize logger before any other action
	logger.Init()

	cli.HandleError(app.New().Execute())
}
