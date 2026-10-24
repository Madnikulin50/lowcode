package cmp

import (
	"strings"
	"testing"
)

func TestRolesMatrix(t *testing.T) {
	view := Actor{Name: "Наблюдатель", Known: true, Roles: RolesOf([]string{"stroykontrol_view", "authenticated"})}
	des := Actor{Name: "Петров П. П.", Known: true, Roles: RolesOf([]string{"stroykontrol_des"})}
	qc := Actor{Name: "Сидорова А. В.", Email: "sidorova@example.ru", Known: true, Roles: RolesOf([]string{"stroykontrol_qc"})}
	mgr := Actor{Name: "Руководитель", Known: true, Roles: RolesOf([]string{"STROYKONTROL_MGR"})}
	none := Actor{Name: "Гость", Known: true}
	cases := []struct {
		a    Actor
		act  string
		want bool
	}{
		{view, ActWork, false}, {view, ActUpload, false},
		{des, ActUpload, true}, {des, ActWork, true}, {des, ActConfirm, false}, {des, ActAssign, false}, {des, ActReview, false},
		{qc, ActConfirm, true}, {qc, ActReview, true}, {qc, ActAssign, false},
		{mgr, ActAssign, true}, {mgr, ActConfirm, true},
		{none, ActWork, false},
		{Actor{}, ActConfirm, true}, // roles not checked
	}
	for _, c := range cases {
		if got := c.a.Can(c.act); got != c.want {
			t.Errorf("%s %v can %s = %v", c.a.Name, c.a.Roles, c.act, got)
		}
	}
	if err := none.Refusal(ActWork); !strings.Contains(err.Error(), "нет ни одной роли") {
		t.Errorf("refusal %v", err)
	}
	if (Actor{}).Label() != DefaultActor || qc.Label() != "Сидорова А. В." {
		t.Errorf("labels")
	}
}

func TestCheckRemarkChange(t *testing.T) {
	des := Actor{Name: "Петров П. П.", Known: true, Roles: []string{SpaceDesigner}}
	qc := Actor{Name: "Сидорова А. В.", Email: "sidorova@example.ru", Known: true, Roles: []string{SpaceQC}}
	// the designer hands the remark over to the check, but cannot confirm it
	if err := des.CheckRemarkChange(RemarkChange{FromStatus: RemarkBeingFix, ToStatus: RemarkAnswerRev}); err != nil {
		t.Errorf("designer to review: %v", err)
	}
	if err := des.CheckRemarkChange(RemarkChange{FromStatus: RemarkAnswerRev, ToStatus: RemarkResolved}); err == nil {
		t.Error("designer confirmed")
	}
	if err := des.CheckRemarkChange(RemarkChange{AssignOrDue: true}); err == nil {
		t.Error("designer assigned")
	}
	// the specialist confirms — unless the remark is assigned to them
	if err := qc.CheckRemarkChange(RemarkChange{FromStatus: RemarkAnswerRev, ToStatus: RemarkResolved, Assignee: "Иванов И. И."}); err != nil {
		t.Errorf("qc confirm: %v", err)
	}
	for _, who := range []string{"Сидорова А. В.", "sidorova@example.ru", "sidorova"} {
		if err := qc.CheckRemarkChange(RemarkChange{OutcomeDecision: true, Assignee: who}); err == nil || !strings.Contains(err.Error(), "исполнитель") {
			t.Errorf("self-confirmation by %q allowed: %v", who, err)
		}
	}
	// the same status again is no confirmation
	if err := des.CheckRemarkChange(RemarkChange{FromStatus: RemarkResolved, ToStatus: RemarkResolved, WorkFields: true}); err != nil {
		t.Errorf("note on a resolved remark: %v", err)
	}
}
