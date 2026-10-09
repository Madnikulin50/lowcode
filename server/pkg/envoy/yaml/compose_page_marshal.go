package yaml

import (
	"context"
	"strconv"

	automationTypes "github.com/madnikulin50/lowcode/server/automation/types"
	composeTypes "github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/envoy"
	"github.com/madnikulin50/lowcode/server/pkg/envoy/resource"
)

func composePageFromResource(r *resource.ComposePage, cfg *EncoderConfig) *composePage {
	return &composePage{
		res:           r.Res,
		encoderConfig: cfg,
	}
}

func (n *composePage) Prepare(ctx context.Context, state *envoy.ResourceState) (err error) {
	pg, ok := state.Res.(*resource.ComposePage)
	if !ok {
		return encoderErrInvalidResource(composeTypes.PageResourceType, state.Res.ResourceType())
	}

	// Get the related namespace
	n.relNs = resource.FindComposeNamespace(state.ParentResources, pg.RefNs.Identifiers)
	if n.relNs == nil {
		return resource.ComposeNamespaceErrUnresolved(pg.RefNs.Identifiers)
	}

	n.refNamespace = relNsToRef(n.relNs)

	// Get the related parent; if any
	if pg.RefParent != nil {
		n.relParent = resource.FindComposePage(state.ParentResources, pg.RefParent.Identifiers)
		if n.relParent == nil {
			return resource.ComposePageErrUnresolved(pg.RefParent.Identifiers)
		}
	}

	// Get the related module; if any
	if pg.RefMod != nil {
		n.relMod = resource.FindComposeModule(state.ParentResources, pg.RefMod.Identifiers)
		if n.relMod == nil {
			return resource.ComposePageErrUnresolved(pg.RefMod.Identifiers)
		}
	}

	n.blocks = make(composePageBlockSet, 0, len(pg.Res.Blocks))
	for i, b := range pg.Res.Blocks {
		auxb := b
		cpb := &composePageBlock{
			res: &auxb,
		}

		// Check for refs
		if refs, has := pg.BlockRefs[i]; has {
			for _, ref := range refs {
				switch ref.ResourceType {
				case composeTypes.ModuleResourceType:
					relMod := resource.FindComposeModule(state.ParentResources, ref.Identifiers)
					if relMod == nil {
						return resource.ComposeModuleErrUnresolved(ref.Identifiers)
					}

					cpb.relMod = append(cpb.relMod, relMod)
					cpb.refMod = append(cpb.refMod, relModToRef(relMod))

				case automationTypes.WorkflowResourceType:
					relWf := resource.FindAutomationWorkflow(state.ParentResources, ref.Identifiers)
					if relWf == nil {
						return resource.AutomationWorkflowErrUnresolved(ref.Identifiers)
					}

					cpb.relWf = append(cpb.relWf, relWf)
					cpb.refWf = append(cpb.refWf, relWfToRef(relWf))

				case composeTypes.ChartResourceType:
					relChart := resource.FindComposeChart(state.ParentResources, ref.Identifiers)
					if relChart == nil {
						return resource.ComposeChartErrUnresolved(ref.Identifiers)
					}

					cpb.relChart = append(cpb.relChart, relChart)
					cpb.refChart = append(cpb.refChart, relChartToRef(relChart))

				default:
					return ErrUnknownResource
				}
			}
		}

		n.blocks = append(n.blocks, cpb)
	}

	return nil
}

func (n *composePage) Encode(ctx context.Context, doc *Document, state *envoy.ResourceState) (err error) {
	if n.res.ID <= 0 {
		n.res.ID = nextID()
	}

	if state.Conflicting {
		return nil
	}

	// Timestaps
	n.ts, err = resource.MakeTimestampsCUDA(&n.res.CreatedAt, n.res.UpdatedAt, n.res.DeletedAt, nil).
		Model(n.encoderConfig.TimeLayout, n.encoderConfig.Timezone)
	if err != nil {
		return err
	}

	// @todo skip eval?

	// if n.encoderConfig.CompactOutput {
	// 	if n.refParent != "" {
	// 		err = doc.nestComposePageChild(n.refParent, n)
	// 	} else {
	// 		err = doc.nestComposePage(n.refNamespace, n)
	// 	}
	// } else {
	// 	doc.addComposePage(n)
	// }
	doc.addComposePage(n)

	return err
}

func (p *composePage) MarshalYAML() (interface{}, error) {
	var err error

	if p.blocks != nil && len(p.blocks) > 0 {
		p.blocks.configureEncoder(p.encoderConfig)
	}

	nn, _ := makeMap()

	if p.relMod != nil {
		nn, err = addMap(nn, "module", firstOkString(p.relMod.Handle, strconv.FormatUint(p.relMod.ID, 10)))
		if err != nil {
			return nil, err
		}
	}

	if p.relParent != nil {
		nn, err = addMap(nn, "parent", firstOkString(p.relParent.Handle, strconv.FormatUint(p.relParent.ID, 10)))
		if err != nil {
			return nil, err
		}
	}

	nn, err = addMap(nn,
		"pageID", p.res.ID,
		"handle", p.res.Handle,
		"title", p.res.Title,
		"description", p.res.Description,

		"blocks", p.blocks,

		"labels", p.res.Labels,
		"visible", p.res.Visible,
		"weight", p.res.Weight,
	)
	if err != nil {
		return nil, err
	}
	if p.children != nil && len(p.children) > 0 {
		p.children.configureEncoder(p.encoderConfig)

		nn, err = encodeResource(nn, "pages", p.children, p.encoderConfig.MappedOutput, "handle")
		if err != nil {
			return nil, err
		}
	}

	nn, err = encodeTimestamps(nn, p.ts)
	if err != nil {
		return nil, err
	}

	return nn, nil
}

func (c *composePageBlock) MarshalYAML() (interface{}, error) {
	if c.res.Kind == "RecordList" {
		c.cleanupModuleFields(c.res.Options)
	}

	opt := c.res.Options
	switch c.res.Kind {
	case "RecordList":
		delete(opt, "moduleID")
		if len(c.refMod) > 0 {
			opt["moduleID"] = c.refMod[0]
		}
		delete(opt, "module")
		break

	case "RecordOrganizer":
		delete(opt, "moduleID")
		if len(c.refMod) > 0 {
			opt["moduleID"] = c.refMod[0]
		}
		delete(opt, "module")
		break

	case "Chart":
		delete(opt, "chartID")
		if len(c.refChart) > 0 {
			opt["chartID"] = c.refChart[0]
		}
		delete(opt, "chart")
		break

	case "Calendar":
		ff, _ := opt["feeds"].([]interface{})
		n := 0
		for _, f := range ff {
			feed, _ := f.(map[string]interface{})
			fOpts, _ := (feed["options"]).(map[string]interface{})
			// the references were collected for the feeds that name a module
			if !namesModule(fOpts) || n >= len(c.refMod) {
				continue
			}
			fOpts["module"] = c.refMod[n]
			delete(fOpts, "moduleID")
			n++
		}
		break

	case "Automation":
		bb, _ := opt["buttons"].([]interface{})
		i := 0
		for _, b := range bb {
			button, _ := b.(map[string]interface{})
			if _, has := button["workflowID"]; !has {
				continue
			}

			button["workflow"] = c.refWf[i]
			delete(button, "workflowID")
			i++
		}
		break

	case "RecordGraph":
		// the references were collected for the modules that are named, in order
		mm, _ := opt["excludeModules"].([]interface{})
		n := 0
		for i, m := range mm {
			id, _ := m.(string)
			if id == "" || id == "0" || n >= len(c.refMod) {
				continue
			}
			mm[i] = c.refMod[n]
			n++
		}

		// then the name templates, in the same order the references were collected
		ll, _ := opt["labels"].([]interface{})
		for _, l := range ll {
			label, _ := l.(map[string]interface{})
			if !namesModule(label) || n >= len(c.refMod) {
				continue
			}
			label["module"] = c.refMod[n]
			delete(label, "moduleID")
			n++
		}
		break

	case "RelatedRecords":
		// the references were collected for the relations that name a module, in order
		rr, _ := opt["relations"].([]interface{})
		n := 0
		for _, r := range rr {
			rel, _ := r.(map[string]interface{})
			if !namesModule(rel) || n >= len(c.refMod) {
				continue
			}
			rel["module"] = c.refMod[n]
			delete(rel, "moduleID")
			n++
		}
		break

	case "Metric":
		mm, _ := opt["metrics"].([]interface{})
		n := 0
		for _, m := range mm {
			mops, _ := m.(map[string]interface{})
			// the references were collected for the metrics that name a module
			if !namesModule(mops) || n >= len(c.refMod) {
				continue
			}
			mops["module"] = c.refMod[n]
			delete(mops, "moduleID")
			n++
		}
		break

	case "Comment":
		delete(opt, "moduleID")
		if len(c.refMod) > 0 {
			opt["moduleID"] = c.refMod[0]
		}
		delete(opt, "module")
		break

	}

	return makeMap(
		"blockID", c.res.BlockID,
		"title", c.res.Title,
		"description", c.res.Description,
		"options", c.res.Options,
		"style", c.res.Style,
		"kind", c.res.Kind,
		// @todo figure out a way to have [x, y, w, h]; stringifying doesn't work
		"xywh", c.res.XYWH,
	)
}

// cleanupModuleFields handles cases where module fields are provided as full blown fields
// instead of just a name reference.
func (c *composePageBlock) cleanupModuleFields(opt map[string]interface{}) {
	rawFF, has := opt["fields"]
	if !has {
		return
	}

	ff, ok := rawFF.([]interface{})
	if !ok {
		return
	}

	retFF := make([]interface{}, 0, len(ff))
	for _, rawF := range ff {
		switch c := rawF.(type) {
		case string:
			retFF = append(retFF, map[string]interface{}{"name": c})
		case map[string]interface{}, map[string]string:
			retFF = append(retFF, c)
		default:
			retFF = append(retFF, rawF)
		}
	}

	opt["fields"] = retFF
}

// namesModule: the options of a block's item (a feed, a metric, a relation)
// refer to a module - the same test that decides whether a reference to one was
// collected for it (see resource.ComposePage), so the references can be matched
// to the items in order.
func namesModule(opt map[string]interface{}) bool {
	// the first key that is present decides, as when the references are collected
	for _, k := range []string{"module", "moduleID"} {
		if v, has := opt[k]; has {
			id, _ := v.(string)
			return id != "" && id != "0"
		}
	}
	return false
}
