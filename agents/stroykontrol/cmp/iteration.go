package cmp

// State of a finding relative to the previous iteration of the same
// comparison (the same ПД/РД pair after the designer's corrections).
const (
	IterNew        = "new"        // not there before
	IterPersisting = "persisting" // still there — not fixed
	IterResolved   = "resolved"   // was there, gone now — fixed
)

// Prior is a finding of the previous iteration as it is stored.
type Prior struct {
	Fingerprint string
	Status      string // the reviewer's decision: new | confirmed | rejected
	Values      Values // the stored record, to re-create a resolved finding
}

// Carry matches the findings of a new iteration against the previous one by
// fingerprint (which is built from the compared fragments, not from their
// position, so it survives re-pagination):
//   - a finding that is still there is "persisting" and keeps the reviewer's
//     decision;
//   - one that is not in the previous iteration is "new";
//   - a previous finding that is gone is "resolved" — except those the
//     reviewer rejected as false alarms, which were never real.
func Carry(prior []Prior, cur []Finding) (state, status map[string]string, resolved []Prior) {
	state, status = map[string]string{}, map[string]string{}
	prev := map[string]Prior{}
	for _, p := range prior {
		if p.Fingerprint != "" {
			prev[p.Fingerprint] = p
		}
	}
	now := map[string]bool{}
	for _, f := range cur {
		now[f.Fingerprint] = true
		if p, ok := prev[f.Fingerprint]; ok {
			state[f.Fingerprint] = IterPersisting
			if p.Status != "" && p.Status != StatusNew {
				status[f.Fingerprint] = p.Status
			}
		} else {
			state[f.Fingerprint] = IterNew
		}
	}
	for _, p := range prior {
		if p.Fingerprint == "" || now[p.Fingerprint] || p.Status == StatusRejected {
			continue
		}
		now[p.Fingerprint] = true // each fingerprint once
		resolved = append(resolved, p)
	}
	return
}
