package dal

import "testing"

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
