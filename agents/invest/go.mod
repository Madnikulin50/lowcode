module github.com/madnikulin50/lowcode/agents/invest

go 1.23

require (
	github.com/go-chi/chi/v5 v5.3.1
	github.com/madnikulin50/lowcode/agents/sdk v0.0.0
	github.com/madnikulin50/lowcode/agents/services/calc-evm v0.0.0
)

require github.com/google/uuid v1.6.0 // indirect

replace github.com/madnikulin50/lowcode/agents/services/calc-evm => ../services/calc-evm

replace github.com/madnikulin50/lowcode/agents/sdk => ../sdk
