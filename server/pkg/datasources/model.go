package datasources

import (
	"fmt"

	"github.com/cortezaproject/corteza/server/pkg/dal"
	"github.com/cortezaproject/corteza/server/system/types"
	"github.com/spf13/cast"
)

func MakeModelRef(step types.ReportStepLoad) (out dal.ModelRef, err error) {
	var (
		connectionID          uint64
		moduleID, namespaceID uint64
		module, namespace     string

		aux any
		ok  bool
	)

	if aux, ok = step.Definition["moduleID"]; ok {
		moduleID = cast.ToUint64(aux)
	} else if aux, ok = step.Definition["module"]; ok {
		module = cast.ToString(aux)
	} else {
		err = fmt.Errorf("step definition is missing moduleID or module")
		return
	}

	if aux, ok = step.Definition["namespaceID"]; ok {
		namespaceID = cast.ToUint64(aux)
	} else if aux, ok = step.Definition["namespace"]; ok {
		namespace = cast.ToString(aux)
	} else {
		err = fmt.Errorf("step definition is missing namespaceID or namespace")
		return
	}

	// Connection is optional, default is primary connection
	if aux, ok = step.Definition["connectionID"]; ok {
		connectionID = cast.ToUint64(aux)
	}

	out.ConnectionID = connectionID
	out.Refs = make(map[string]any)

	// Use only one of the two identifier variations with priority to ID
	if moduleID > 0 {
		out.Refs["moduleID"] = moduleID
	} else {
		out.Refs["module"] = module
	}

	if namespaceID > 0 {
		out.Refs["namespaceID"] = namespaceID
	} else {
		out.Refs["namespace"] = namespace
	}

	return
}

// ModelAttributes returns the attributes defined on the referenced model
func ModelAttributes(pr ModelFinder, step types.ReportStepLoad, mfr dal.ModelRef) []dal.AttributeMapping {
	// All of the attributes
	attrs, err := getModelAttrs(pr, mfr)
	if err != nil {
		return nil
	}

	return attrToMapping(attrs...)
}

func getModelAttrs(pr ModelFinder, mfr dal.ModelRef) (attrs dal.AttributeSet, err error) {
	m := pr.FindModel(mfr)
	if m == nil {
		return nil, fmt.Errorf("model not found: %v", mfr)
	}

	return m.Attributes, nil
}

func attrToMapping(aa ...*dal.Attribute) (out []dal.AttributeMapping) {
	for _, a := range aa {
		out = append(out, dal.SimpleAttr{
			Ident: a.Ident,
			Src:   a.Ident,
			Props: dal.MapProperties{
				Label:     a.Label,
				Type:      a.Type,
				IsSystem:  a.System,
				Nullable:  a.Type.IsNullable(),
				IsPrimary: a.PrimaryKey,
				// @todo add multi value delimiter; it's currently not set on the
				//       attribute so we might need to rethink this a bit
				IsMultivalue: a.MultiValue,
			},
		})
	}
	return
}
