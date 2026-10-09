package dal

import (
	"context"
	"testing"
)

func TestAggregateAttr_IsDummyGroup(t *testing.T) {
	dummy := AggregateAttr{
		Type: &TypeNumber{HasDefault: true, DefaultValue: 0, Precision: -1, Scale: -1},
	}
	if !dummy.IsDummyGroup() {
		t.Fatal("empty group placeholder must be dummy")
	}
	if (AggregateAttr{Identifier: "count", RawExpr: "count(ID)"}).IsDummyGroup() {
		t.Fatal("count metric is not dummy")
	}
	if (AggregateAttr{Identifier: "dimension_0", RawExpr: "status"}).IsDummyGroup() {
		t.Fatal("dimension is not dummy")
	}
}

func TestMakeGroupKeyNilRunnerIsDummyGroup(t *testing.T) {
	gk, err := makeGroupKey(context.Background(), []*runnerGval{nil}, map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if len(gk) != 1 || gk[0] != nil {
		t.Fatalf("dummy group key = %#v", gk)
	}
}
