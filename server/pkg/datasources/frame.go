package datasources

import (
	"context"
	"fmt"
	"strings"

	"github.com/madnikulin50/lowcode/server/pkg/dal"
	"github.com/madnikulin50/lowcode/server/pkg/filter"
	"github.com/madnikulin50/lowcode/server/system/types"
	"github.com/spf13/cast"
)

type FrameColumnSet []FrameColumn
type FrameColumn struct {
	Name    string `json:"name"`
	Label   string `json:"label"`
	Kind    string `json:"kind"`
	Primary bool   `json:"primary"`
	Unique  bool   `json:"unique"`
	System  bool   `json:"system"`

	Multivalue          bool   `json:"multivalue"`
	MultivalueDelimiter string `json:"multivalueDelimiter"`
}

type (
	Frame struct {
		Name   string `json:"name"`
		Source string `json:"source"`
		Ref    string `json:"ref,omitempty"`

		// RefValue is the common value between the two datasources
		RefValue string `json:"refValue,omitempty"`
		// RelColumn is what column in the local ds was used
		RelColumn string `json:"relColumn,omitempty"`
		// RelSource is the ds that this frame is related to
		RelSource string `json:"relSource,omitempty"`

		Columns FrameColumnSet `json:"columns"`
		Rows    []FrameRow     `json:"rows"`

		Paging *filter.Paging          `json:"paging"`
		Sort   filter.SortExprSet      `json:"sort"`
		Filter *types.ReportFilterExpr `json:"filter"`
	}

	FrameDescription struct {
		Source  string         `json:"source"`
		Ref     string         `json:"ref,omitempty"`
		Columns FrameColumnSet `json:"columns"`
	}

	FrameRow []string

	FrameDefinition struct {
		Name    string         `json:"name"`
		Source  string         `json:"source"`
		Ref     string         `json:"ref"`
		Columns FrameColumnSet `json:"columns"`

		Filter *types.ReportFilterExpr `json:"filter"`
		Paging *filter.Paging          `json:"paging"`
		Sort   filter.SortExprSet      `json:"sort"`
	}
	FrameDefinitionSet []*FrameDefinition
)

// @todo nicer formatting and alignment
func (f Frame) String() string {
	out := fmt.Sprintf("n: %10s; src: %10s\n", f.Name, f.Source)

	if f.Ref != "" {
		out += fmt.Sprintf("ref: %10s; col: %10s\n; key: %10s\n", f.Ref, f.RelColumn, f.RefValue)
	}

	if f.RelSource != "" {
		out += fmt.Sprintf("Rel source: %10s;\n", f.RelSource)
	}

	for _, c := range f.Columns {
		out += fmt.Sprintf("%s<%s>, ", c.Name, c.Kind)
	}
	out = strings.TrimRight(out, " ,")
	out += "\n"

	for i, r := range f.Rows {
		out += fmt.Sprintf("%d| %s", i+1, strings.Join(r, ", "))
		out += "\n"
	}

	if f.Paging != nil {
		out += "\n"
		out += fmt.Sprintf("< %s; =%s; > %s", f.Paging.PrevPage.String(), f.Paging.PageCursor.String(), f.Paging.NextPage.String())
	}

	if len(f.Sort) > 0 {
		out += "\n"
		out += f.Sort.String()
	}

	out += "\n"
	out += fmt.Sprintf("ix %d; len %d", 0, len(f.Rows))

	return out
}

func (cc FrameColumnSet) String() string {
	out := ""
	for _, c := range cc {
		out += fmt.Sprintf("%s<%s>, ", c.Name, c.Kind)
	}
	return strings.TrimRight(out, " ,")
}

// OmitSys returns the columns that are not system defined
func (cc FrameColumnSet) OmitSys() FrameColumnSet {
	out := make(FrameColumnSet, 0, len(cc))
	for _, c := range cc {
		if !c.System {
			out = append(out, c)
		}
	}

	return out
}

func (r FrameRow) String() string {
	return strings.Join(r, ", ")
}

// FilterBySource returns a set of definitions for the requested identifier
func (dd FrameDefinitionSet) FilterBySource(ident string) FrameDefinitionSet {
	out := make(FrameDefinitionSet, 0, len(dd))
	for _, d := range dd {
		if d.Source == ident {
			out = append(out, d)
		}
	}

	return out
}

// Describe returns a set of frame descriptions based on the given pipeline
func Describe(ctx context.Context, rr dryRunner, ss types.ReportStepSet, sources []string) (out []*FrameDescription, err error) {
	// Make a run for the whole thing
	pp, err := makePipeline(rr, ss, nil)
	if err != nil {
		return
	}

	var aux []*FrameDescription
	for _, src := range sources {
		// Use the requested source as root
		sub := pp.Slice(src)
		err = rr.Dryrun(ctx, sub)
		if err != nil {
			return
		}

		s := sub[0]
		aux, err = describePipeline(s, src)
		if err != nil {
			return
		}
		out = append(out, aux...)
	}

	return
}

// stepLinkFrames is dedicated for the link step due to it's unique output
func stepLinkFrames(ctx context.Context, iter dal.Iterator, r Run) (ff []*Frame, err error) {
	defs := r.Defs
	// @note this will only be called for the link step so it can freely panic if violated
	defLink := r.Pipeline[0].(*dal.Link)

	// Unpack frame definitions for the link
	defLeft, defRight := unpackLinkDefs(defs, r.Pipeline)

	// Init vars to keep track of the progress
	// @note true is left, false is right
	counters := make(map[bool]uint)

	builders := make(map[bool]*reportFrameBuilder)
	builders[true] = newReportFrameBuilder(defLeft)
	builders[false] = newReportFrameBuilder(defRight)
	builders[false].linked(defLink.On.Right, defLink.On.Left, defLink.RelLeft)

	limits := make(map[bool]uint)
	if defLeft.Paging != nil {
		limits[true] = defLeft.Paging.Limit
	}
	if defRight.Paging != nil {
		limits[false] = defRight.Paging.Limit
	}

	// Helper to determine if we need a next cursor
	nextCursor := false

	// Helpers for reading iterators
	var (
		ref    string
		row    = &dal.Row{}
		doingF = false
	)

	for iter.Next(ctx) {
		if limits[true] > 0 && counters[true] >= limits[true] {
			nextCursor = true
			break
		}
		row.Reset()

		_ = iter.Scan(row)

		// Determine ref and which vars to use
		aux, _ := row.GetValue(dal.LinkRefIdent, 0)
		ref = cast.ToString(aux)
		if ref == "" {
			ref = defLeft.Ref
		}
		left := ref == defLeft.Ref

		// When needed, flush the finished frames to the output
		if left && doingF {
			ff = append(ff, builders[false].done())
			doingF = false
		} else if !left {
			doingF = true
		}

		builders[left].addRow(row)
		counters[left]++
	}
	if err = iter.Err(); err != nil {
		return
	}

	// If the loop ended before the limit cut it off, we need to finish the
	// last right frame as it wasn't yet in the above loop
	if !nextCursor {
		if doingF {
			ff = append(ff, builders[false].done())
		}
	}

	// Apply paging cursor to the left frame
	// @todo consider applying them to the right as well, for now, no
	if nextCursor {
		if builders[true].frame.Paging == nil {
			builders[true].frame.Paging = &filter.Paging{}
		}
		builders[true].frame.Paging.NextPage, err = iter.ForwardCursor(row)
		if err != nil {
			return
		}
	}

	// Complete the output with the left frame
	// @note the left frame goes to the start and the right frames are in the same order
	//       as the related rows from the left frame.
	return append([]*Frame{builders[true].done()}, ff...), nil
}

// unpackLinkDefs returns the left and the right frame definition disregarding
// the order of the definitions in the input
func unpackLinkDefs(defs FrameDefinitionSet, pp dal.Pipeline) (left, right *FrameDefinition) {
	r := pp[0]

	l := r.(*dal.Link)

	find := func(defs FrameDefinitionSet, ref string) *FrameDefinition {
		for _, def := range defs {
			if def.Ref == ref {
				return def
			}
		}
		return nil
	}

	return find(defs, l.RelLeft), find(defs, l.RelRight)
}

// stepFrames is a generic iter to Frame converter
func stepFrames(ctx context.Context, iter dal.Iterator, r Run) (ff []*Frame, err error) {
	defs := r.Defs

	// @note only the link step takes multiple defs and that one is not covered
	//       by this function
	if len(defs) != 1 {
		panic(fmt.Sprintf("impossible state: expecting one frame definition, got %d", len(defs)))
	}
	def := defs[0]

	// Init vars to keep track of the progress
	limit := uint(0)
	counter := uint(0)
	builder := newReportFrameBuilder(def)
	if def.Paging != nil {
		limit = def.Paging.Limit
	}

	// Helper to determine if we need a next cursor
	nextCursor := false

	// Helpers for reading iterators
	row := &dal.Row{}

	for iter.Next(ctx) {
		if limit > 0 && counter >= limit {
			nextCursor = true
			break
		}
		row.Reset()

		_ = iter.Scan(row)
		builder.addRow(row)
		counter++
	}
	if err = iter.Err(); err != nil {
		return
	}

	// Apply paging cursor to the frame
	if nextCursor {
		if builder.frame.Paging == nil {
			builder.frame.Paging = &filter.Paging{}
		}
		builder.frame.Paging.NextPage, err = iter.ForwardCursor(row)
		if err != nil {
			return
		}
	}

	return append(ff, builder.done()), nil
}

func describePipeline(s dal.PipelineStep, src string) (out []*FrameDescription, err error) {
	aa := s.Attributes()

	out = make([]*FrameDescription, len(aa))
	for i, a := range aa {
		out[i] = &FrameDescription{
			Source: src, Columns: mappingToFrameCols(a),
		}
	}

	// @note this case is only possible for the link step; expand when/if needed
	if len(out) == 2 {
		l := s.(*dal.Link)
		out[0].Ref = l.RelLeft
		out[1].Ref = l.RelRight
	}

	return
}

// mappingToFrameCols converts pipeline AttributeMapping to FrameColumnSet
func mappingToFrameCols(mm []dal.AttributeMapping) FrameColumnSet {
	out := make(FrameColumnSet, 0, len(mm))

	for _, m := range mm {
		out = append(out, mappingToFrameCol(m))
	}

	return out
}

// @note current implementation a bit _rushed_ since I'll probably rethink
//
//	how the pipeline handles attributes -- will revisit then.
func mappingToFrameCol(m dal.AttributeMapping) FrameColumn {
	p := m.Properties()

	const (
		// Coppied around to reduce imports
		emailLength = 254
		urlLength   = 2048

		attachmentResType = "corteza::system:attachment"
		userResType       = "corteza::system:user"
		moduleResType     = "corteza::compose:module"
	)

	l := m.Properties().Label
	if l == "" {
		l = m.Identifier()
	}
	out := FrameColumn{
		Name:  m.Identifier(),
		Label: l,
		Kind:  "String",

		Primary:             p.IsPrimary,
		System:              p.IsSystem,
		Multivalue:          p.IsMultivalue,
		MultivalueDelimiter: p.MultivalueDelimiter,
	}
	if out.MultivalueDelimiter == "" {
		out.MultivalueDelimiter = "\n"
	}

	switch t := p.Type.(type) {
	case *dal.TypeBoolean:
		out.Kind = "Boolean"
	case *dal.TypeDate, *dal.TypeTime, *dal.TypeTimestamp:
		out.Kind = "DateTime"

	case *dal.TypeNumber:
		out.Kind = "Number"

	case *dal.TypeEnum:
		out.Kind = "Select"

	case *dal.TypeText:
		// @note temporary solution; we should push some meta along with it
		if t.Length == emailLength {
			out.Kind = "Email"
		} else if t.Length == urlLength {
			out.Kind = "URL"
		} else {
			out.Kind = "String"
		}

	case *dal.TypeRef:
		switch t.RefModel.ResourceType {
		case moduleResType:
			out.Kind = "Record"
		case userResType:
			out.Kind = "User"
		case attachmentResType:
			out.Kind = "File"
		}
	}

	return out
}

// Report step -> DAL step conversion

// convStepLoad converts ReportStepLoad to dal.Datasource
func convStepLoad(pr ModelFinder, step types.ReportStepLoad, defs FrameDefinitionSet) (out *dal.Datasource, err error) {
	// Validation
	if len(defs) > 1 {
		err = fmt.Errorf("cannot convert load step: expecting at most one definition, got %d", len(defs))
		return
	}

	// Get additional filtering
	var extf filter.Filter
	if len(defs) == 1 {
		extf = filterFromDef(defs[0])
	}

	// Prepare model ref
	mfr, err := MakeModelRef(step)
	if err != nil {
		return
	}

	// @todo refactor this out after we support other resources with potentially missing soft delete fields
	f, err := dal.FilterFromExpr(step.Filter.Node()).
		MergeFilters(filter.Generic(filter.WithStateConstraint("deletedAt", filter.StateExcluded)))
	if err != nil {
		return
	}

	f, err = f.MergeFilters(extf)
	if err != nil {
		return
	}

	// Make pipeline step
	return &dal.Datasource{
		Ident:         step.Name,
		Filter:        f,
		ModelRef:      mfr,
		OutAttributes: ModelAttributes(pr, step, mfr),
	}, nil
}

// convStepAggregate converts ReportStepAggregate to dal.Aggregate
func convStepAggregate(step types.ReportStepAggregate, defs FrameDefinitionSet) (out *dal.Aggregate, err error) {
	// Validation
	if len(defs) > 1 {
		err = fmt.Errorf("cannot convert aggregate step: expecting at most one definition, got %d", len(defs))
		return
	}

	// Get additional filtering
	var extf filter.Filter
	if len(defs) == 1 {
		extf = filterFromDef(defs[0])
	}

	f, err := dal.FilterFromExpr(step.Filter.Node()).
		MergeFilters(extf)
	if err != nil {
		return
	}

	ggs := make([]dal.AggregateAttr, 0, len(step.Keys))
	for _, c := range step.Keys {
		ggs = append(ggs, dal.AggregateAttr{
			Key:        true,
			Identifier: c.Name,
			Label:      c.Label,
			Expression: c.Def.Node(),
		})
	}

	vvs := make([]dal.AggregateAttr, 0, len(step.Columns))
	for _, c := range step.Columns {
		vvs = append(vvs, dal.AggregateAttr{
			Identifier: c.Name,
			Label:      c.Label,
			Expression: c.Def.Node(),
		})
	}

	// Make pipeline step
	out = &dal.Aggregate{
		Ident:     step.Name,
		RelSource: step.Source,
		Filter:    f,

		Group:         ggs,
		OutAttributes: vvs,
	}
	return
}

// updateDefinitionColumns assures run's frame column completeness
func updateDefinitionColumns(r Run) {
	ppAttrs := r.Pipeline[0].Attributes()
	for i, def := range r.Defs {
		if len(def.Columns) > 0 {
			continue
		}

		def.Columns = mappingToFrameCols(ppAttrs[i])
	}
}

// makeModelRef returns the model ref based on the step load definition
// @todo should be expanded when we support models that are not compose modules

func filterFromDef(def *FrameDefinition) (out filter.Filter) {
	aux := filter.Generic(
		filter.WithExpressionParsed(def.Filter.Node()),
		filter.WithOrderBy(def.Sort),
	)

	if def.Paging != nil {
		aux = aux.With(
			filter.WithCursor(def.Paging.PageCursor),
			filter.WithLimit(def.Paging.Limit),
		)
	}

	return aux
}
