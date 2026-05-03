package datasources

import (
	"context"
	"fmt"

	"github.com/cortezaproject/corteza/server/pkg/dal"
	"github.com/cortezaproject/corteza/server/system/types"
)

type Run struct {
	Pipeline dal.Pipeline
	Defs     FrameDefinitionSet
}

type dryRunner interface {
	ModelFinder
	Dryrun(context.Context, dal.Pipeline) error
}

type ModelFinder interface {
	FindModel(dal.ModelRef) *dal.Model
}

// Runs create a set of report runs based on step and frame definitions
func Runs(pr ModelFinder, steps types.ReportStepSet, defs FrameDefinitionSet) (out []Run, err error) {
	if len(steps) == 0 {
		return nil, fmt.Errorf("no steps found")
	}
	// Prepare runs based on the provided definitions
	//
	// - If consecutive definitions point to the same source with the same name
	//   consider them to fall under the same workload (the link step)
	// - else, one def per workload

	if len(defs) == 0 {
		lastStep := steps[len(steps)-1]
		def := FrameDefinition{Source: lastStep.Name()}
		defs = append(defs, &def)
	}
	auxDefs := make(FrameDefinitionSet, 0)
	var aux Run
	for i, def := range defs {
		if i == 0 {
			auxDefs = append(auxDefs, def)
			continue
		}

		// Definitions fall together
		if def.Name == defs[i-1].Name && def.Source == defs[i-1].Source {
			auxDefs = append(auxDefs, def)
			continue
		}

		// Make run for the previous definition (exclude current!!)
		aux, err = makeRun(pr, steps, auxDefs)
		if err != nil {
			return
		}
		out = append(out, aux)

		// Prepare next definition batch including the current one
		auxDefs = make(FrameDefinitionSet, 0)
		auxDefs = append(auxDefs, def)
	}

	// Handle the ones (potentially) not covered by the above loop
	if len(auxDefs) > 0 {
		aux, err = makeRun(pr, steps, auxDefs)
		if err != nil {
			return
		}
		out = append(out, aux)
	}

	return
}

// Frames returns a set of Frame for the given workload & iterator combo
func Frames(ctx context.Context, iter dal.Iterator, r Run) (ff []*Frame, err error) {
	// Preprocessing on the workload's frame definitions; assure all
	// columns/metdata are there to avoid nonesense later down the line
	updateDefinitionColumns(r)

	// @todo perhaps need to change; for now only this scenario introduces multiple
	//       frame defs per workload
	if len(r.Defs) > 1 {
		return stepLinkFrames(ctx, iter, r)
	}

	return stepFrames(ctx, iter, r)
}

func makeRun(pr ModelFinder, ss types.ReportStepSet, defs FrameDefinitionSet) (out Run, err error) {
	var pp dal.Pipeline
	pp, err = makePipeline(pr, ss, defs)
	if err != nil {
		return
	}

	out.Defs = defs
	out.Pipeline = pp.Slice(defs[0].Source)
	return
}

// makePipeline converts the given report steps into the DAL pipeline
func makePipeline(mf ModelFinder, ss types.ReportStepSet, defs FrameDefinitionSet) (pp dal.Pipeline, err error) {
	for _, step := range ss {
		switch {
		case step.Load != nil:
			aux, err := convStepLoad(mf, *step.Load, defs.FilterBySource(step.Load.Name))
			if err != nil {
				return nil, err
			}
			pp = append(pp, aux)

		case step.Aggregate != nil:
			aux, err := convStepAggregate(*step.Aggregate, defs.FilterBySource(step.Aggregate.Name))
			if err != nil {
				return nil, err
			}
			pp = append(pp, aux)

		case step.Join != nil:
			aux, err := convStepJoin(*step.Join, defs.FilterBySource(step.Join.Name))
			if err != nil {
				return nil, err
			}
			pp = append(pp, aux)

		case step.Link != nil:
			aux, err := convStepLink(*step.Link, defs.FilterBySource(step.Link.Name))
			if err != nil {
				return nil, err
			}
			pp = append(pp, aux)

		default:
			// this should never happen
			panic(fmt.Errorf("unknown step type: %v", step.Kind))
		}
	}

	return pp, pp.LinkSteps()
}
