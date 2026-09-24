package def

// RbacOperation mirrors codegen/schema/rbac.cue's #rbacOperation (the subset
// needed by the rbac/types_rbac template, which does not use checkFuncName).
type RbacOperation struct {
	Description string // default = the operation's handle
}

// Rbac mirrors codegen/schema/rbac.cue's #rbacResource / #rbacComponent.
// A nil *Rbac on a Resource means the resource has no rbac block at all
// (mirrors CUE's optional `rbac?:` being entirely absent), which excludes it
// from rbac/types_rbac generation.
type Rbac struct {
	Operations map[string]RbacOperation
}
