package dal

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/madnikulin50/lowcode/server/pkg/ql"
	"github.com/modern-go/reflect2"
	"github.com/spf13/cast"
)

type (
	// aggregator performs value aggregation primarily used for the aggregate step
	// @todo consider moving into a separate package
	//
	// The aggregator performs the requested value expressions as well as
	// aggregation operations over the evaluated values.
	//
	// The aggregator computes the values on the fly (for the ones it can).
	aggregator struct {
		// aggregates holds output aggregation values
		// @note we'll use float64 for all values, but it might make more sense to split it up
		//       in the future. For now, it'll be ok.
		aggregates []float64
		last       ValueGetter
		unique     []*map[float64]bool
		list       []*[]float64

		// counts holds the number of values for each aggregate including multi value fields.
		// Counts are currently only used for average.
		counts []int

		// def provides a list of aggregates.
		//
		// Each index corresponds to the aggregates and counts slices
		def []aggregateDef
		// scanned indicates whether the aggregator has been scanned since
		// we need to block writes after the first scan.
		scanned bool
	}

	// aggregateDef is a wrapper for the provided value aggregation
	aggregateDef struct {
		outIdent string

		aggOp string

		inIdent string
		eval    evaluator
	}
)

var (
	// aggregateFunctionIndex specifies all of the registered aggregate functions
	//
	// @todo consider making this expandable via some registry/plugin/...
	aggregateFunctionIndex = map[string]bool{
		"count":       true,
		"sum":         true,
		"min":         true,
		"max":         true,
		"avg":         true,
		"":            true,
		"uniquecount": true,
		"stddev":      true,
		"std":         true,
	}
)

// Aggregator initializes a new aggregator for the given set of mappings
//
// The aggregator is not routine safe; consider defining multiple aggregators
// and then combining them together.
func Aggregator() *aggregator {
	return &aggregator{
		aggregates: make([]float64, 0, 16),
		last:       nil,
		unique:     make([]*map[float64]bool, 0, 16),
		list:       make([]*[]float64, 0, 16),
		counts:     make([]int, 0, 16),
	}
}

// AddAggregateE adds a new aggregate from a raw expression
func (a *aggregator) AddAggregateE(ident, expr string) (err error) {
	// @note the converter is reused so we can safely do this
	n, err := newConverterGval().Parse(expr)
	if err != nil {
		return
	}

	return a.AddAggregate(ident, n)
}

// AddAggregate adds a new aggregate from an already parsed expression
func (a *aggregator) AddAggregate(ident string, expr *ql.ASTNode) (err error) {
	def := aggregateDef{
		outIdent: ident,
	}

	inIdent, expr, err := unpackMappingSource(expr)
	if err != nil {
		return
	}

	// Take it from the source
	if inIdent != "" {
		def.inIdent = inIdent
	}

	// Take it from the expression
	// - agg. op.
	def.aggOp, expr, err = unpackExpressionNode(expr)
	if err != nil {
		return
	}
	// Prepare a runner in case we're not simply copying values
	if inIdent == "" {
		// - make evaluator
		if expr != nil {
			def.eval, err = newRunnerGvalParsed(expr)
			if err != nil {
				return
			}
		}
	}

	a.aggregates = append(a.aggregates, 0)
	a.last = nil
	a.unique = append(a.unique, &map[float64]bool{})
	a.list = append(a.list, &[]float64{})
	a.counts = append(a.counts, 0)
	a.def = append(a.def, def)
	return
}

// Aggregate aggregates the given value
func (a *aggregator) Aggregate(ctx context.Context, v ValueGetter) (err error) {
	if a.scanned {
		// If we attempt to add data to the aggregator, the previous data might
		// no longer be valid.
		return fmt.Errorf("cannot call Aggregate on an already scanned aggregator")
	}

	// @todo consider throwing all of the aggregates into a routine.
	//       I think (for now) performance gains will be negligible.
	for i, attr := range a.def {
		err = a.aggregate(ctx, attr, i, v)
		if err != nil {
			return
		}
	}

	return nil
}

// Scan scans the aggregated values into the setter
func (a *aggregator) Scan(s ValueSetter) (err error) {
	// On first scan, complete partial aggregates
	if !a.scanned {
		a.completePartials(context.TODO())
	}

	a.scanned = true

	// Set the values

	if a.last != nil {
		m := a.last.CountValues()
		for k, v := range m {
			var i uint
			for i = 0; i < v; i++ {
				val, err := a.last.GetValue(k, i)
				if err != nil {
					continue
				}
				err = s.SetValue(k, i, val)
			}
		}
	}

	for i, attr := range a.def {
		// @note each aggregated value can be at most one so no need for multi-value
		//       suport here.
		err = s.SetValue(attr.outIdent, 0, a.aggregates[i])
		if err != nil {
			return
		}
	}
	return
}

// aggregate applies the provided value into the requested aggregate
func (a *aggregator) aggregate(ctx context.Context, attr aggregateDef, i int, v ValueGetter) (err error) {
	a.last = v
	switch attr.aggOp {
	case "count":
		return a.count(ctx, attr, i, v)

	case "sum":
		return a.sum(ctx, attr, i, v)

	case "min":
		return a.min(ctx, attr, i, v)

	case "max":
		return a.max(ctx, attr, i, v)
	case "stddev", "std":
		return a.stddev(ctx, attr, i, v)
	case "":
		return a.fix(ctx, attr, i, v)
	case "avg":
		return a.avg(ctx, attr, i, v)
	case "uniquecount":
		return a.uniqueCount(ctx, attr, i, v)
	}

	return fmt.Errorf("unsupported aggregate function: %s", attr.aggOp)
}

// walkValues traverses the available values for the specified attribute
func (a *aggregator) walkValues(ctx context.Context, r ValueGetter, cc map[string]uint, attr aggregateDef, run func(v any, isNil bool)) (err error) {
	var out any
	if attr.inIdent == "" {
		if attr.eval != nil {
			out, err = attr.eval.Eval(ctx, r)
			if err != nil {
				return nil
			}
		} else {
			out = r
		}

		run(out, reflect2.IsNil(out))
		return nil
	}

	for i := uint(0); i < cc[attr.inIdent]; i++ {
		v, err := r.GetValue(attr.inIdent, i)
		if err != nil {
			return err
		}
		run(v, reflect2.IsNil(v))
	}

	return nil
}

// Aggregate methods

func (a *aggregator) count(ctx context.Context, attr aggregateDef, i int, v ValueGetter) (err error) {
	err = a.walkValues(ctx, v, v.CountValues(), attr, func(v any, isNil bool) {
		if isNil {
			return
		}

		a.aggregates[i]++
		a.counts[i]++
	})
	if err != nil {
		return
	}

	return
}

func (a *aggregator) sum(ctx context.Context, attr aggregateDef, i int, v ValueGetter) (err error) {
	err = a.walkValues(ctx, v, v.CountValues(), attr, func(v any, isNil bool) {
		if isNil {
			return
		}
		a.aggregates[i] += cast.ToFloat64(v)
		a.counts[i]++
	})
	if err != nil {
		return
	}

	return
}

func (a *aggregator) min(ctx context.Context, attr aggregateDef, i int, v ValueGetter) (err error) {
	err = a.walkValues(ctx, v, v.CountValues(), attr, func(v any, isNil bool) {
		if isNil {
			return
		}

		if a.counts[i] == 0 {
			a.aggregates[i] = cast.ToFloat64(v)
		} else {
			a.aggregates[i] = math.Min(a.aggregates[i], cast.ToFloat64(v))
		}
		a.counts[i]++
	})
	if err != nil {
		return
	}

	return
}

func (a *aggregator) fix(ctx context.Context, attr aggregateDef, i int, v ValueGetter) (err error) {
	err = a.walkValues(ctx, v, v.CountValues(), attr, func(value any, isNil bool) {
		if isNil {
			return
		}
		if attr.eval != nil {
			out, err := attr.eval.Eval(ctx, v)
			if err == nil {
				value = out
			} else {
				value = -1
			}
		}

		if a.counts[i] == 0 {
			a.aggregates[i] = cast.ToFloat64(value)
		} else {
			a.aggregates[i] = math.Max(a.aggregates[i], cast.ToFloat64(value))
		}
		a.counts[i]++
	})
	if err != nil {
		return
	}

	return
}

func (a *aggregator) max(ctx context.Context, attr aggregateDef, i int, v ValueGetter) (err error) {
	err = a.walkValues(ctx, v, v.CountValues(), attr, func(v any, isNil bool) {
		if isNil {
			return
		}

		if a.counts[i] == 0 {
			a.aggregates[i] = cast.ToFloat64(v)
		} else {
			a.aggregates[i] = math.Max(a.aggregates[i], cast.ToFloat64(v))
		}
		a.counts[i]++
	})
	if err != nil {
		return
	}

	return
}

func (a *aggregator) avg(ctx context.Context, attr aggregateDef, i int, v ValueGetter) (err error) {
	err = a.walkValues(ctx, v, v.CountValues(), attr, func(v any, isNil bool) {
		if isNil {
			return
		}

		a.aggregates[i] += cast.ToFloat64(v)

		a.counts[i]++
	})
	if err != nil {
		return
	}

	return
}

func (a *aggregator) uniqueCount(ctx context.Context, attr aggregateDef, i int, v ValueGetter) (err error) {
	err = a.walkValues(ctx, v, v.CountValues(), attr, func(v any, isNil bool) {
		if isNil {
			return
		}

		(*a.unique[i])[cast.ToFloat64(v)] = true

		a.counts[i]++
	})
	if err != nil {
		return
	}

	return
}

func (a *aggregator) stddev(ctx context.Context, attr aggregateDef, i int, v ValueGetter) (err error) {
	err = a.walkValues(ctx, v, v.CountValues(), attr, func(v any, isNil bool) {
		if isNil {
			return
		}

		(*a.list[i]) = append((*a.list[i]), cast.ToFloat64(v))

		a.counts[i]++
	})
	if err != nil {
		return
	}

	return
}

func calcStdDev(data []float64) float64 {
	if len(data) <= 1 {
		return 0.0
	}

	// 1. Calculate the mean
	var sum float64
	for _, v := range data {
		sum += v
	}
	mean := sum / float64(len(data))

	// 2. Sum the squared differences from the mean
	var squaredDiffSum float64
	for _, v := range data {
		diff := v - mean
		squaredDiffSum += diff * diff
	}

	// 3. Divide by (n - 1) for sample stddev, then take square root
	variance := squaredDiffSum / float64(len(data)-1)
	stddev := math.Sqrt(variance)
	return stddev
}

func (a *aggregator) completePartials(ctx context.Context) {
	hasExprs := false
	for i, attr := range a.def {
		if attr.aggOp == "avg" {
			if a.counts[i] == 0 {
				return
			}
			a.aggregates[i] = a.aggregates[i] / float64(a.counts[i])
		}
		if attr.aggOp == "uniquecount" {
			if a.counts[i] == 0 {
				return
			}
			a.aggregates[i] = float64(len(*a.unique[i]))
		}
		if attr.aggOp == "stddev" || attr.aggOp == "std" {
			if a.counts[i] == 0 {
				return
			}
			a.aggregates[i] = calcStdDev(*a.list[i])
		}
		if attr.inIdent == "" && attr.eval != nil {
			hasExprs = true
		}

	}
	if hasExprs {
		record := map[string]any{}
		for i, attr := range a.def {
			record[attr.outIdent] = a.aggregates[i]
		}
		for i, attr := range a.def {
			if attr.inIdent != "" {
				continue
			}
			out, e := attr.eval.Eval(ctx, record)
			if e == nil {
				fl, ok := out.(float64)
				if ok {
					a.aggregates[i] = fl
					record[attr.outIdent] = a.aggregates[i]
				}

			}

		}
	}
}

// Utilities

func unpackMappingSource(n *ql.ASTNode) (ident string, expr *ql.ASTNode, err error) {
	// Check if first arg of agg. fnc. is an attr.
	if len(n.Args) == 1 && n.Args[0].Symbol != "" {
		return n.Args[0].Symbol, n, nil
	}

	expr = n
	return
}

func unpackExpressionNode(n *ql.ASTNode) (aggOp string, expr *ql.ASTNode, err error) {
	if n.Ref != "" {
		aggOp = strings.ToLower(n.Ref)
	}
	if !aggregateFunctionIndex[aggOp] {
		//err = fmt.Errorf("root expression must be an aggregate function")
		return "", n, nil
	}

	if len(n.Args) > 0 {
		expr = n.Args[0]
	} else {
		return "", n, nil
	}
	return
}

// reset is a benchmarking utility to reset the aggregator
//
// Don't use it in production code.
func (a *aggregator) reset() {
	for i := 0; i < len(a.aggregates); i++ {
		a.aggregates[i] = 0
		a.unique[i] = &map[float64]bool{}
		a.last = nil
		a.counts[i] = 0
	}
	a.scanned = false
}
