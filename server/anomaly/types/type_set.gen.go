package types

// This file is auto-generated.
//
// Changes to this file may cause incorrect behavior and will be lost if
// the code is regenerated from anomaly/types/types.yaml

type (

	// BaselineSet slice of Baseline
	//
	// This type is auto-generated.
	BaselineSet []*Baseline

	// FindingSet slice of Finding
	//
	// This type is auto-generated.
	FindingSet []*Finding

	// RuleSet slice of Rule
	//
	// This type is auto-generated.
	RuleSet []*Rule
)

// Walk iterates through every slice item and calls w(Baseline) err
//
// This function is auto-generated.
func (set BaselineSet) Walk(w func(*Baseline) error) (err error) {
	for i := range set {
		if err = w(set[i]); err != nil {
			return
		}
	}

	return
}

// Filter iterates through every slice item, calls f(Baseline) (bool, err) and return filtered slice
//
// This function is auto-generated.
func (set BaselineSet) Filter(f func(*Baseline) (bool, error)) (out BaselineSet, err error) {
	var ok bool
	out = BaselineSet{}
	for i := range set {
		if ok, err = f(set[i]); err != nil {
			return
		} else if ok {
			out = append(out, set[i])
		}
	}

	return
}

// FindByID finds items from slice by its ID property
//
// This function is auto-generated.
func (set BaselineSet) FindByID(ID uint64) *Baseline {
	for i := range set {
		if set[i].ID == ID {
			return set[i]
		}
	}

	return nil
}

// IDs returns a slice of uint64s from all items in the set
//
// This function is auto-generated.
func (set BaselineSet) IDs() (IDs []uint64) {
	IDs = make([]uint64, len(set))

	for i := range set {
		IDs[i] = set[i].ID
	}

	return
}

// Walk iterates through every slice item and calls w(Finding) err
//
// This function is auto-generated.
func (set FindingSet) Walk(w func(*Finding) error) (err error) {
	for i := range set {
		if err = w(set[i]); err != nil {
			return
		}
	}

	return
}

// Filter iterates through every slice item, calls f(Finding) (bool, err) and return filtered slice
//
// This function is auto-generated.
func (set FindingSet) Filter(f func(*Finding) (bool, error)) (out FindingSet, err error) {
	var ok bool
	out = FindingSet{}
	for i := range set {
		if ok, err = f(set[i]); err != nil {
			return
		} else if ok {
			out = append(out, set[i])
		}
	}

	return
}

// FindByID finds items from slice by its ID property
//
// This function is auto-generated.
func (set FindingSet) FindByID(ID uint64) *Finding {
	for i := range set {
		if set[i].ID == ID {
			return set[i]
		}
	}

	return nil
}

// IDs returns a slice of uint64s from all items in the set
//
// This function is auto-generated.
func (set FindingSet) IDs() (IDs []uint64) {
	IDs = make([]uint64, len(set))

	for i := range set {
		IDs[i] = set[i].ID
	}

	return
}

// Walk iterates through every slice item and calls w(Rule) err
//
// This function is auto-generated.
func (set RuleSet) Walk(w func(*Rule) error) (err error) {
	for i := range set {
		if err = w(set[i]); err != nil {
			return
		}
	}

	return
}

// Filter iterates through every slice item, calls f(Rule) (bool, err) and return filtered slice
//
// This function is auto-generated.
func (set RuleSet) Filter(f func(*Rule) (bool, error)) (out RuleSet, err error) {
	var ok bool
	out = RuleSet{}
	for i := range set {
		if ok, err = f(set[i]); err != nil {
			return
		} else if ok {
			out = append(out, set[i])
		}
	}

	return
}

// FindByID finds items from slice by its ID property
//
// This function is auto-generated.
func (set RuleSet) FindByID(ID uint64) *Rule {
	for i := range set {
		if set[i].ID == ID {
			return set[i]
		}
	}

	return nil
}

// IDs returns a slice of uint64s from all items in the set
//
// This function is auto-generated.
func (set RuleSet) IDs() (IDs []uint64) {
	IDs = make([]uint64, len(set))

	for i := range set {
		IDs[i] = set[i].ID
	}

	return
}
