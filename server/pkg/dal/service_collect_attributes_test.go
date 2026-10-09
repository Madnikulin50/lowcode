package dal

import (
	"context"
	"testing"
)

// TestCollectAttributesClobberedDatasource reproduces the "unknown attribute
// avg_plan_staffing for aggregate agg for staff_plan_agg" regression: a
// Datasource with an Aggregate clobbered into it (see
// optimization_clobber.go / def_datasource.go's clobbered field) must expose
// what the aggregate actually computes (group keys + output columns) to
// whatever step sources from it next — not the raw pre-aggregation column
// list, which is all collectAttributes() used to return for *Datasource.
func TestCollectAttributesClobberedDatasource(t *testing.T) {
	numType := &TypeNumber{Precision: -1, Scale: -1}

	// staff_plan_agg's own "Load" step, before its "Aggregate" step got
	// folded into it.
	ds := &Datasource{
		Ident: "staff_plan_agg",
		OutAttributes: []AttributeMapping{
			SimpleAttr{Ident: "store_id", Props: MapProperties{Type: numType}},
			SimpleAttr{Ident: "planned_staffing", Props: MapProperties{Type: numType}},
		},
	}

	// staff_plan_agg's own "Aggregate" step: AVG(planned_staffing) grouped
	// by store_id, producing avg_plan_staffing — this is what pipelines
	// with fewer than 3 steps already clobbered even before the len(in)<3
	// gate existed, so this shape alone was never the problem.
	innerAgg := &Aggregate{
		Ident:     "Aggregate",
		RelSource: "Load",
		rel:       ds,
		Group: []AggregateAttr{
			{Identifier: "store_id", Key: true, Type: numType},
		},
		OutAttributes: []AggregateAttr{
			{Identifier: "avg_plan_staffing", Type: numType},
		},
	}

	// Simulate what pipelineClobberSteps does when it folds an Aggregate
	// into its source Datasource.
	ds.clobbered = append(ds.clobbered, innerAgg)

	t.Run("collectAttributes exposes the clobbered aggregate's own output", func(t *testing.T) {
		aa := collectAttributes(ds)

		found := false
		for _, a := range aa {
			if a.Identifier() == "avg_plan_staffing" {
				found = true
			}
		}
		if !found {
			idents := make([]string, len(aa))
			for i, a := range aa {
				idents[i] = a.Identifier()
			}
			t.Fatalf("collectAttributes(clobbered datasource) = %v, want it to include avg_plan_staffing", idents)
		}
	})

	t.Run("a downstream aggregate referencing that output inits without error", func(t *testing.T) {
		// This mirrors the outer "agg" step from a report built on top of a
		// module (like staff_join) that in turn loads staff_plan_agg —
		// exactly the shape that produced the regression.
		outer := &Aggregate{
			Ident: "agg for staff_plan_agg",
			rel:   ds,
			Group: []AggregateAttr{
				{Identifier: "store_id", Key: true, RawExpr: "store_id", Type: numType},
			},
			OutAttributes: []AggregateAttr{
				{Identifier: "avg2", RawExpr: "avg_plan_staffing", Type: numType},
			},
		}

		if _, err := outer.init(context.Background(), nil); err != nil {
			t.Fatalf("outer aggregate referencing the clobbered aggregate's output failed to init: %v", err)
		}
	})

	t.Run("a downstream join sourcing from it picks up the computed column", func(t *testing.T) {
		// Mirrors store_perfomance_with_avg in test9: a plain Load joined
		// with a clobbered Load->Aggregate module (store_perfomance_avg_agg),
		// on the aggregate's own group key (store_id) — the join's
		// auto-derived OutAttributes must include the right side's computed
		// column (avg_plan_staffing), not just its raw source columns.
		left := &Datasource{
			Ident: "store_perfomance",
			OutAttributes: []AttributeMapping{
				SimpleAttr{Ident: "store_id", Props: MapProperties{Type: numType}},
				SimpleAttr{Ident: "revenue", Props: MapProperties{Type: numType}},
			},
		}

		j := &Join{
			Ident:    "Join",
			RelLeft:  "store_perfomance",
			RelRight: "staff_plan_agg",
			On:       JoinPredicate{Left: "store_id", Right: "store_id"},
			relLeft:  left,
			relRight: ds, // the clobbered Datasource from the outer test
		}

		exec, err := j.init(context.Background(), nil, nil)
		if err != nil {
			t.Fatalf("join sourcing from the clobbered aggregate failed to init: %v", err)
		}
		_ = exec

		found := false
		for _, a := range j.OutAttributes {
			if a.Source() == "avg_plan_staffing" {
				found = true
			}
		}
		if !found {
			idents := make([]string, len(j.OutAttributes))
			for i, a := range j.OutAttributes {
				idents[i] = a.Identifier()
			}
			t.Fatalf("join OutAttributes = %v, want one sourced from avg_plan_staffing", idents)
		}
	})
}

// TestCollectAttributesSelfClobberedAggregate reproduces the
// "unknown attribute ID for aggregate agg for receipt_positions" regression:
// a plain [Datasource -> Aggregate] report pipeline — the shape of almost
// every "sum/count/avg by dimension" report on a single module — where
// pipelineClobberSteps folds the Aggregate into its own upstream Datasource.
// That Aggregate's own init() must still validate its Group/OutAttributes
// expressions against the Datasource's raw, pre-aggregation columns, not
// against its own (not yet computed) output.
func TestCollectAttributesSelfClobberedAggregate(t *testing.T) {
	numType := &TypeNumber{Precision: -1, Scale: -1}

	// receipt_positions' raw Datasource, with real columns including the
	// sys "ID" field.
	ds := &Datasource{
		Ident: "receipt_positions",
		OutAttributes: []AttributeMapping{
			SimpleAttr{Ident: "ID", Props: MapProperties{Type: numType}},
			SimpleAttr{Ident: "position_sum", Props: MapProperties{Type: numType}},
		},
	}

	// The report's own wrapping aggregate ("agg for receipt_positions"),
	// referencing raw "position_sum" and the raw "ID" column in its own
	// output — same object pipelineClobberSteps folds into ds, becoming
	// ds.clobbered[0], since this two-step pipeline has nothing else
	// consuming ds.
	def := &Aggregate{
		Ident: "agg for receipt_positions",
		rel:   ds,
		OutAttributes: []AggregateAttr{
			{Identifier: "rp", RawExpr: "sum(position_sum)", Type: numType},
			{Identifier: "ID", RawExpr: "ID", Type: numType},
		},
	}
	ds.clobbered = append(ds.clobbered, def)

	if _, err := def.init(context.Background(), nil); err != nil {
		t.Fatalf("aggregate that clobbered into its own source failed to init: %v", err)
	}
}
