package dal

type (
	modelDiffType string
	// ModelDiff defines one identified missmatch between two models
	ModelDiff struct {
		Type modelDiffType
		// Original will be nil when a new attribute is being added
		Original *Attribute
		// Asserted will be nil wen an existing attribute is being removed
		Asserted *Attribute

		// OriginalIndex/AssertedIndex mirror Original/Asserted above, but for
		// IndexMissing diffs. Kept as separate fields (rather than reusing
		// Original/Asserted) since an Index isn't an Attribute.
		OriginalIndex *Index
		AssertedIndex *Index
	}

	ModelDiffSet []*ModelDiff
)

const (
	AttributeMissing             modelDiffType = "attributeMissing"
	AttributeTypeMissmatch       modelDiffType = "typeMissmatch"
	AttributeSensitivityMismatch modelDiffType = "sensitivityMismatch"
	AttributeCodecMismatch       modelDiffType = "codecMismatch"

	// IndexMissing covers both directions (see ModelDiff.OriginalIndex vs
	// AssertedIndex): the index doesn't exist yet, or an index under the
	// same Ident changed shape (fields/uniqueness) — the latter surfaces as
	// a matched pair of "missing on each side" diffs (drop then recreate),
	// same as how attribute type changes are handled via AttributeReType
	// rather than here, except indexes don't support an in-place ALTER.
	IndexMissing modelDiffType = "indexMissing"
)

// Diff calculates the diff between models a and b where a is used as base
func (a *Model) Diff(b *Model) (out ModelDiffSet) {
	if a == nil {
		a = &Model{}
	}
	if b == nil {
		b = &Model{}
	}

	bIndex := make(map[string]struct {
		found bool
		attr  *Attribute
	})
	for _, _attr := range b.Attributes {
		attr := _attr
		bIndex[attr.Ident] = struct {
			found bool
			attr  *Attribute
		}{
			attr: attr,
		}
	}

	aIndex := make(map[string]struct {
		found bool
		attr  *Attribute
	})
	for _, _attr := range a.Attributes {
		attr := _attr
		aIndex[attr.Ident] = struct {
			found bool
			attr  *Attribute
		}{
			attr: attr,
		}
	}

	// Deleted and update ones
	for _, _attrA := range a.Attributes {
		attrA := _attrA
		// store is an interface to something that could be a pointer.
		// we need to copy it to make sure we don't get a nil pointer
		// make sure not to modify this since it would modify the original
		attrA.Store = _attrA.Store

		// Missmatches
		attrBAux, ok := bIndex[attrA.Ident]
		if !ok {
			out = append(out, &ModelDiff{
				Type:     AttributeMissing,
				Original: attrA,
			})
			continue
		}

		// Typecheck
		if attrA.Type.Type() != attrBAux.attr.Type.Type() {
			out = append(out, &ModelDiff{
				Type:     AttributeTypeMissmatch,
				Original: attrA,
				Asserted: attrBAux.attr,
			})
		}

		// Other stuff
		// @todo improve; for now it'll do
		if attrA.SensitivityLevelID != attrBAux.attr.SensitivityLevelID {
			out = append(out, &ModelDiff{
				Type:     AttributeSensitivityMismatch,
				Original: attrA,
				Asserted: attrBAux.attr,
			})
		}
		if attrA.Store.Type() != attrBAux.attr.Store.Type() {
			out = append(out, &ModelDiff{
				Type:     AttributeCodecMismatch,
				Original: attrA,
				Asserted: attrBAux.attr,
			})
		}
	}

	// New
	for _, _attrB := range b.Attributes {
		attrB := _attrB

		// Missmatches
		_, ok := aIndex[attrB.Ident]
		if !ok {
			out = append(out, &ModelDiff{
				Type:     AttributeMissing,
				Original: nil,
				Asserted: attrB,
			})
			continue
		}
	}

	out = append(out, diffIndexes(a.Indexes, b.Indexes)...)

	return
}

// diffIndexes compares two Index sets by Ident, additionally treating an
// Ident present on both sides but with a different shape (fields order,
// attributes, or uniqueness) as dropped-then-recreated, since there's no
// in-place ALTER INDEX equivalent across the supported drivers.
func diffIndexes(a, b IndexSet) (out ModelDiffSet) {
	bByIdent := make(map[string]*Index, len(b))
	for _, idx := range b {
		bByIdent[idx.Ident] = idx
	}
	aByIdent := make(map[string]*Index, len(a))
	for _, idx := range a {
		aByIdent[idx.Ident] = idx
	}

	for _, idxA := range a {
		idxB, ok := bByIdent[idxA.Ident]
		if !ok || !indexesEqualShape(idxA, idxB) {
			out = append(out, &ModelDiff{Type: IndexMissing, OriginalIndex: idxA})
		}
	}
	for _, idxB := range b {
		idxA, ok := aByIdent[idxB.Ident]
		if !ok || !indexesEqualShape(idxA, idxB) {
			out = append(out, &ModelDiff{Type: IndexMissing, AssertedIndex: idxB})
		}
	}

	return
}

// indexesEqualShape reports whether two indexes would produce the same DDL:
// same uniqueness and the same attributes, in the same order. Sort/Nulls/
// Modifiers per field are intentionally not compared yet (v1 only exposes
// plain ascending indexes to admins; nothing can produce a difference there).
func indexesEqualShape(a, b *Index) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Unique != b.Unique || len(a.Fields) != len(b.Fields) {
		return false
	}
	for i, f := range a.Fields {
		if f.AttributeIdent != b.Fields[i].AttributeIdent {
			return false
		}
	}
	return true
}

func (dd ModelDiffSet) Alterations() (out []*Alteration) {
	add := func(a *Alteration) {
		out = append(out, a)
	}

	for _, d := range dd {
		switch d.Type {
		case AttributeMissing:
			if d.Asserted == nil {
				// @todo if this was the last attribute we can consider dropping this column
				if d.Original.Store.Type() == AttributeCodecRecordValueSetJSON {
					break
				}

				add(&Alteration{
					AttributeDelete: &AttributeDelete{
						Attr: d.Original,
					},
				})
			} else {
				if d.Asserted.Store.Type() == AttributeCodecRecordValueSetJSON {
					add(&Alteration{
						AttributeAdd: &AttributeAdd{
							Attr: &Attribute{
								Ident: d.Asserted.StoreIdent(),
								Type:  &TypeJSON{Nullable: false},
								Store: &CodecPlain{},
							},
						},
					})
				} else {
					add(&Alteration{
						AttributeAdd: &AttributeAdd{
							Attr: d.Asserted,
						},
					})
				}

			}

		case AttributeTypeMissmatch:
			// @todo we might have to do some validation earlier on
			if d.Original.Store.Type() == AttributeCodecRecordValueSetJSON {
				break
			}

			add(&Alteration{
				AttributeReType: &AttributeReType{
					Attr: d.Asserted,
					To:   d.Asserted.Type,
				},
			})

		case AttributeCodecMismatch:
			add(&Alteration{
				AttributeReEncode: &AttributeReEncode{
					Attr: d.Asserted,
					To:   d.Asserted.Store,
				},
			})

		case IndexMissing:
			if d.AssertedIndex == nil {
				add(&Alteration{
					IndexDelete: &IndexDelete{
						Ident: d.OriginalIndex.Ident,
					},
				})
			} else {
				add(&Alteration{
					IndexAdd: &IndexAdd{
						Index: d.AssertedIndex,
					},
				})
			}
		}
	}

	return
}
