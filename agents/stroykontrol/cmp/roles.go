package cmp

import (
	"fmt"
	"strings"
)

// Роли (ТЗ 3.8): права на объекты, модули и записи даёт платформа LowCoode; агент, который пишет в
// пространство служебным токеном, сам проверяет, что вызвавший его пользователь вправе сделать то,
// что просит, — по ролям платформы, в которых он состоит. Роль узнаётся по её handle
// (compose/roles.mjs создаёт четыре роли пространства). Наблюдатель только смотрит; проектировщик
// загружает документацию, работает с замечаниями и передаёт их на проверку; специалист по контролю
// качества разбирает выводы, подтверждает устранение и возвращает на доработку; руководитель проекта
// вдобавок назначает исполнителей и сроки. Подтвердить устранение замечания не может его исполнитель.

// Roles of the space.
const (
	SpaceAdmin    = "admin"
	SpaceManager  = "manager"
	SpaceDesigner = "designer"
	SpaceQC       = "qc"
	SpaceObserver = "observer"
)

// RoleHandles maps the handles of the platform roles to the roles of the space.
var RoleHandles = map[string]string{
	"admins":            SpaceAdmin, // the platform's administrators
	"stroykontrol_mgr":  SpaceManager,
	"stroykontrol_des":  SpaceDesigner,
	"stroykontrol_qc":   SpaceQC,
	"stroykontrol_view": SpaceObserver,
}

// RoleLabels are the names of the roles.
var RoleLabels = map[string]string{
	SpaceAdmin: "Администратор", SpaceManager: "Руководитель проекта", SpaceDesigner: "Проектировщик",
	SpaceQC: "Специалист по контролю качества", SpaceObserver: "Наблюдатель",
}

// Actions the agent checks.
const (
	ActUpload  = "upload"  // load documents, run comparisons and checks, import exports, the volume list
	ActReview  = "review"  // decisions on findings, pairing, excluded regions, remark cards from findings
	ActWork    = "work"    // a remark: note, estimate, requirements, AI, a status other than the two below
	ActAssign  = "assign"  // a remark: the assignee and the due date
	ActConfirm = "confirm" // a remark: «Устранено», «Возвращено на доработку», the decision on the result
	ActIDCheck = "idcheck" // the check of the executive documentation
)

// ActionLabels say what an action is, for the refusal.
var ActionLabels = map[string]string{
	ActUpload: "загрузка документации и запуск проверок", ActReview: "решения по находкам сравнения",
	ActWork: "работа с замечанием", ActAssign: "назначение исполнителя и срока",
	ActConfirm: "подтверждение устранения и возврат на доработку", ActIDCheck: "проверка ИД",
}

var roleActs = map[string][]string{
	SpaceAdmin:    {ActUpload, ActReview, ActWork, ActAssign, ActConfirm, ActIDCheck},
	SpaceManager:  {ActUpload, ActReview, ActWork, ActAssign, ActConfirm, ActIDCheck},
	SpaceQC:       {ActUpload, ActReview, ActWork, ActConfirm, ActIDCheck},
	SpaceDesigner: {ActUpload, ActWork},
	SpaceObserver: nil,
}

// Actor is the person behind a request. Known is false when the agent runs without the check of
// roles (--roles=off, fixtures): everything is allowed and the actions are signed DefaultActor.
type Actor struct {
	ID    string   `json:"id,omitempty"`
	Name  string   `json:"name"`
	Email string   `json:"email,omitempty"`
	Roles []string `json:"roles"`
	Known bool     `json:"known"`
}

// DefaultActor signs the actions when the person is not known.
const DefaultActor = "специалист ПТО"

// Label is how the actor signs the history.
func (a Actor) Label() string {
	if !a.Known {
		return DefaultActor
	}
	return firstNonEmpty(a.Name, a.Email, a.ID)
}

// RolesOf maps the handles of the person's platform roles to the roles of the space.
func RolesOf(handles []string) []string {
	var out []string
	seen := map[string]bool{}
	for _, h := range handles {
		if r, ok := RoleHandles[strings.ToLower(strings.TrimSpace(h))]; ok && !seen[r] {
			seen[r] = true
			out = append(out, r)
		}
	}
	return out
}

// Can says whether the actor may do the action.
func (a Actor) Can(act string) bool {
	if !a.Known {
		return true
	}
	for _, r := range a.Roles {
		for _, x := range roleActs[r] {
			if x == act {
				return true
			}
		}
	}
	return false
}

// Allowed lists the actions of the actor (for the interface: what to offer).
func (a Actor) Allowed() []string {
	var out []string
	for _, act := range []string{ActUpload, ActReview, ActWork, ActAssign, ActConfirm, ActIDCheck} {
		if a.Can(act) {
			out = append(out, act)
		}
	}
	return out
}

// Refusal is the error of an action the role does not allow.
func (a Actor) Refusal(act string) error {
	if len(a.Roles) == 0 {
		return fmt.Errorf("недостаточно прав: «%s» — у пользователя %s нет ни одной роли пространства «Стройконтроль»", ActionLabels[act], a.Label())
	}
	var names []string
	for _, r := range a.Roles {
		names = append(names, RoleLabels[r])
	}
	return fmt.Errorf("недостаточно прав: «%s» не входит в роль %s", ActionLabels[act], strings.Join(names, ", "))
}

// RemarkChange is what a hand-made change of a remark touches.
type RemarkChange struct {
	Assignee        string // the current assignee of the remark
	ToStatus        string // "" — the status is not changed
	FromStatus      string
	AssignOrDue     bool // the assignee or the due date changes
	OutcomeDecision bool // the decision on the result is made or cleared
	WorkFields      bool // note, estimate
}

// CheckRemarkChange checks a change of a remark against the actor's roles and forbids the assignee to
// confirm the remark resolved himself.
func (a Actor) CheckRemarkChange(c RemarkChange) error {
	confirm := c.OutcomeDecision || (c.ToStatus != "" && c.ToStatus != c.FromStatus && (c.ToStatus == RemarkResolved || c.ToStatus == RemarkRework))
	switch {
	case confirm && !a.Can(ActConfirm):
		return a.Refusal(ActConfirm)
	case c.AssignOrDue && !a.Can(ActAssign):
		return a.Refusal(ActAssign)
	case (c.WorkFields || (c.ToStatus != "" && c.ToStatus != c.FromStatus)) && !a.Can(ActWork):
		return a.Refusal(ActWork)
	}
	if confirm && a.Known && c.Assignee != "" && a.isNamed(c.Assignee) {
		return fmt.Errorf("исполнитель замечания не может сам подтвердить устранение или принять решение по результату: это делает другой специалист")
	}
	return nil
}

// isNamed: the assignee field names the actor (by the name, the e-mail or the e-mail's login).
func (a Actor) isNamed(who string) bool {
	w := NormName(who)
	if w == "" {
		return false
	}
	for _, s := range []string{a.Name, a.Email} {
		if s != "" && NormName(s) == w {
			return true
		}
	}
	if at := strings.Index(a.Email, "@"); at > 0 && NormName(a.Email[:at]) == w {
		return true
	}
	return false
}
