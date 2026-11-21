package rdbms

// Store implementation for composeResourceVersion (page, module and chart versions).
//
// Mirrors the automationSession section of rdbms.gen.go; the types, queries
// and filter it relies on come from the regular codegen (store/*.gen.go).

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	composeType "github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/filter"
	"github.com/madnikulin50/lowcode/server/store"
	"github.com/modern-go/reflect2"
)

var _ store.ComposeResourceVersions = &Store{}

// CreateComposeResourceVersion creates one or more rows in composeResourceVersion collection
//
// This function is auto-generated
func (s *Store) CreateComposeResourceVersion(ctx context.Context, rr ...*composeType.ResourceVersion) (err error) {
	for i := range rr {
		if err = s.checkComposeResourceVersionConstraints(ctx, rr[i]); err != nil {
			return
		}

		if err = s.Exec(ctx, composeResourceVersionInsertQuery(s.Dialect.GOQU(), rr[i])); err != nil {
			return
		}
	}

	return
}

// UpdateComposeResourceVersion updates one or more existing entries in composeResourceVersion collection
//
// This function is auto-generated
func (s *Store) UpdateComposeResourceVersion(ctx context.Context, rr ...*composeType.ResourceVersion) (err error) {
	for i := range rr {
		if err = s.checkComposeResourceVersionConstraints(ctx, rr[i]); err != nil {
			return
		}

		if err = s.Exec(ctx, composeResourceVersionUpdateQuery(s.Dialect.GOQU(), rr[i])); err != nil {
			return
		}
	}

	return
}

// UpsertComposeResourceVersion updates one or more existing entries in composeResourceVersion collection
//
// This function is auto-generated
func (s *Store) UpsertComposeResourceVersion(ctx context.Context, rr ...*composeType.ResourceVersion) (err error) {
	for i := range rr {
		if err = s.checkComposeResourceVersionConstraints(ctx, rr[i]); err != nil {
			return
		}

		// @todo this solution is ok for now but could be problematic when we start
		// batching together DB operations.
		if s.Dialect.Nuances().TwoStepUpsert {
			var rsp sql.Result
			rsp, err = s.ExecR(ctx, composeResourceVersionUpdateQuery(s.Dialect.GOQU(), rr[i]))
			if err != nil {
				return
			}
			if c, err := rsp.RowsAffected(); err != nil {
				return err
			} else if c > 0 {
				continue
			}

			err = s.Exec(ctx, composeResourceVersionInsertQuery(s.Dialect.GOQU(), rr[i]))
			if err != nil {
				return
			}
		} else {
			err = s.Exec(ctx, composeResourceVersionUpsertQuery(s.Dialect.GOQU(), rr[i]))
			if err != nil {
				return
			}
		}
	}

	return
}

// DeleteComposeResourceVersion Deletes one or more entries from composeResourceVersion collection
//
// This function is auto-generated
func (s *Store) DeleteComposeResourceVersion(ctx context.Context, rr ...*composeType.ResourceVersion) (err error) {
	for i := range rr {
		if err = s.Exec(ctx, composeResourceVersionDeleteQuery(s.Dialect.GOQU(), composeResourceVersionPrimaryKeys(rr[i]))); err != nil {
			return
		}
	}

	return nil
}

// DeleteComposeResourceVersionByID deletes single entry from composeResourceVersion collection
//
// This function is auto-generated
func (s *Store) DeleteComposeResourceVersionByID(ctx context.Context, id uint64) error {
	return s.Exec(ctx, composeResourceVersionDeleteQuery(s.Dialect.GOQU(), goqu.Ex{
		"id": id,
	}))
}

// TruncateComposeResourceVersions Deletes all rows from the composeResourceVersion collection
func (s *Store) TruncateComposeResourceVersions(ctx context.Context) error {
	return s.Exec(ctx, composeResourceVersionTruncateQuery(s.Dialect.GOQU()))
}

// SearchComposeResourceVersions returns (filtered) set of ComposeResourceVersions
//
// This function is auto-generated
func (s *Store) SearchComposeResourceVersions(ctx context.Context, f composeType.ResourceVersionFilter) (set composeType.ResourceVersionSet, _ composeType.ResourceVersionFilter, err error) {

	// Cleanup unwanted cursor values (only relevant is f.PageCursor, next&prev are reset and returned)
	f.PrevPage, f.NextPage = nil, nil

	if f.PageCursor != nil {
		if f.IncPageNavigation || f.IncTotal {
			return nil, f, fmt.Errorf("not allowed to fetch page navigation or total item count with page cursor")
		}

		// Page cursor exists; we need to validate it against used sort
		// To cover the case when paging cursor is set but sorting is empty, we collect the sorting instructions
		// from the cursor.
		// This (extracted sorting info) is then returned as part of response
		if f.Sort, err = f.PageCursor.Sort(f.Sort); err != nil {
			return
		}
	}

	// Make sure results are always sorted at least by primary keys
	if f.Sort.Get("id") == nil {
		f.Sort = append(f.Sort, &filter.SortExpr{
			Column:     "id",
			Descending: f.Sort.LastDescending(),
		})
	}

	// Cloned sorting instructions for the actual sorting
	// Original are passed to the etchFullPageOfComposeResourceVersions fn used for cursor creation;
	// direction information it MUST keep the initial
	sort := f.Sort.Clone()

	// When cursor for a previous page is used it's marked as reversed
	// This tells us to flip the descending flag on all used sort keys
	if f.PageCursor != nil && f.PageCursor.ROrder {
		sort.Reverse()
	}

	set, f.PrevPage, f.NextPage, err = s.fetchFullPageOfComposeResourceVersions(ctx, f, sort)

	f.PageCursor = nil
	if err != nil {
		return nil, f, err
	}

	if f.IncTotal {
		// Calc total from the number of items fetched
		// even if we do build the page navigation
		f.Total = len(set)

		if f.Limit > 0 && uint(len(set)) == f.Limit {
			// there are fewer items fetched then requested limit
			limit := f.Limit
			f.Limit = 0
			var navSet composeType.ResourceVersionSet
			if navSet, _, _, err = s.fetchFullPageOfComposeResourceVersions(ctx, f, sort); err != nil {
				return
			} else {
				f.Total = len(navSet)
				f.Limit = limit
			}
		}
	}

	return set, f, nil
}

// fetchFullPageOfComposeResourceVersions collects all requested results.
//
// Function applies:
//   - cursor conditions (where ...)
//   - limit
//
// Main responsibility of this function is to perform additional sequential queries in case when not enough results
// are collected due to failed check on a specific row (by check fn).
//
// # Function then moves cursor to the last item fetched
//
// This function is auto-generated
func (s *Store) fetchFullPageOfComposeResourceVersions(
	ctx context.Context,
	filter composeType.ResourceVersionFilter,
	sort filter.SortExprSet,
) (set []*composeType.ResourceVersion, prev, next *filter.PagingCursor, err error) {
	var (
		aux []*composeType.ResourceVersion

		// When cursor for a previous page is used it's marked as reversed
		// This tells us to flip the descending flag on all used sort keys
		reversedOrder = filter.PageCursor != nil && filter.PageCursor.ROrder

		// Copy no. of required items to limit
		// Limit will change when doing subsequent queries to fill
		// the set with all required items
		limit = filter.Limit

		reqItems = filter.Limit

		// cursor to prev. page is only calculated when cursor is used
		hasPrev = filter.PageCursor != nil

		// next cursor is calculated when there are more pages to come
		hasNext bool

		tryFilter composeType.ResourceVersionFilter
	)

	set = make([]*composeType.ResourceVersion, 0, DefaultSliceCapacity)

	for try := 0; try < MaxRefetches; try++ {
		// Copy filter & apply custom sorting that might be affected by cursor
		tryFilter = filter
		tryFilter.Sort = sort

		if limit > 0 {
			// fetching + 1 to peak ahead if there are more items
			// we can fetch (next-page cursor)
			tryFilter.Limit = limit + 1
		}

		if aux, hasNext, err = s.QueryComposeResourceVersions(ctx, tryFilter); err != nil {
			return nil, nil, nil, err
		}

		if len(aux) == 0 {
			// nothing fetched
			break
		}

		// append fetched items
		set = append(set, aux...)

		if reqItems == 0 || !hasNext {
			// no max requested items specified, break out
			break
		}

		collected := uint(len(set))

		if reqItems > collected {
			// not enough items fetched, try again with adjusted limit
			limit = reqItems - collected

			if limit < MinEnsureFetchLimit {
				// In case limit is set very low and we've missed records in the first fetch,
				// make sure next fetch limit is a bit higher
				limit = MinEnsureFetchLimit
			}

			// Update cursor so that it points to the last item fetched
			tryFilter.PageCursor = s.collectComposeResourceVersionCursorValues(set[collected-1], filter.Sort...)

			// Copy reverse flag from sorting
			tryFilter.PageCursor.LThen = filter.Sort.Reversed()
			continue
		}

		if reqItems < collected {
			set = set[:reqItems]
		}

		break
	}

	collected := len(set)

	if collected == 0 {
		return nil, nil, nil, nil
	}

	if reversedOrder {
		// Fetched set needs to be reversed because we've forced a descending order to get the previous page
		for i, j := 0, collected-1; i < j; i, j = i+1, j-1 {
			set[i], set[j] = set[j], set[i]
		}

		// when in reverse-order rules on what cursor to return change
		hasPrev, hasNext = hasNext, hasPrev
	}

	if hasPrev {
		prev = s.collectComposeResourceVersionCursorValues(set[0], filter.Sort...)
		prev.ROrder = true
		prev.LThen = !filter.Sort.Reversed()
	}

	if hasNext {
		next = s.collectComposeResourceVersionCursorValues(set[collected-1], filter.Sort...)
		next.LThen = filter.Sort.Reversed()
	}

	return set, prev, next, nil
}

// QueryComposeResourceVersions queries the database, converts and checks each row and returns collected set
//
// With generics, we can remove this per-resource-generated function
// and replace it with a single utility fetcher
//
// This function is auto-generated
func (s *Store) QueryComposeResourceVersions(
	ctx context.Context,
	f composeType.ResourceVersionFilter,
) (_ []*composeType.ResourceVersion, more bool, err error) {
	var (
		ok bool

		set         = make([]*composeType.ResourceVersion, 0, DefaultSliceCapacity)
		res         *composeType.ResourceVersion
		aux         *auxComposeResourceVersion
		rows        *sql.Rows
		count       uint
		expr, tExpr []goqu.Expression

		sortExpr []exp.OrderedExpression
	)

	if s.Filters.ComposeResourceVersion != nil {
		// extended filter set
		tExpr, f, err = s.Filters.ComposeResourceVersion(s, f)
	} else {
		// using generated filter
		tExpr, f, err = ComposeResourceVersionFilter(s.Dialect, f)
	}

	if err != nil {
		err = fmt.Errorf("could not generate filter expression for ComposeResourceVersion: %w", err)
		return
	}

	expr = append(expr, tExpr...)

	// paging feature is enabled
	if f.PageCursor != nil {
		if tExpr, err = cursorWithSorting(f.PageCursor, s.sortableComposeResourceVersionFields()); err != nil {
			return
		} else {
			expr = append(expr, tExpr...)
		}
	}

	query := composeResourceVersionSelectQuery(s.Dialect.GOQU()).Where(expr...)

	// sorting feature is enabled
	if sortExpr, err = order(f.Sort, s.sortableComposeResourceVersionFields()); err != nil {
		err = fmt.Errorf("could not generate order expression for ComposeResourceVersion: %w", err)
		return
	}

	if len(sortExpr) > 0 {
		query = query.Order(sortExpr...)
	}

	if f.Limit > 0 {
		query = query.Limit(f.Limit)
	}

	rows, err = s.Query(ctx, query)
	if err != nil {
		err = fmt.Errorf("could not query ComposeResourceVersion: %w", err)
		return
	}

	if err = rows.Err(); err != nil {
		err = fmt.Errorf("could not query ComposeResourceVersion: %w", err)
		return
	}

	defer func() {
		closeError := rows.Close()
		if err == nil {
			// return error from close
			err = closeError
		}
	}()

	for rows.Next() {
		if err = rows.Err(); err != nil {
			err = fmt.Errorf("could not query ComposeResourceVersion: %w", err)
			return
		}

		aux = new(auxComposeResourceVersion)
		if err = aux.scan(rows); err != nil {
			err = fmt.Errorf("could not scan rows for ComposeResourceVersion: %w", err)
			return
		}

		count++
		if res, err = aux.decode(); err != nil {
			err = fmt.Errorf("could not decode ComposeResourceVersion: %w", err)
			return
		}

		// check fn set, call it and see if it passed the test
		// if not, skip the item
		if f.Check != nil {
			if ok, err = f.Check(res); err != nil {
				return
			} else if !ok {
				continue
			}
		}

		set = append(set, res)
	}

	return set, f.Limit > 0 && count >= f.Limit, err

}

// LookupComposeResourceVersionByID searches for a compose resource version by ID
//
// This function is auto-generated
func (s *Store) LookupComposeResourceVersionByID(ctx context.Context, id uint64) (_ *composeType.ResourceVersion, err error) {
	var (
		rows   *sql.Rows
		aux    = new(auxComposeResourceVersion)
		lookup = composeResourceVersionSelectQuery(s.Dialect.GOQU()).Where(
			goqu.I("id").Eq(id),
		).Limit(1)
	)

	rows, err = s.Query(ctx, lookup)
	if err != nil {
		return
	}

	defer func() {
		closeError := rows.Close()
		if err == nil {
			// return error from close
			err = closeError
		}
	}()

	if err = rows.Err(); err != nil {
		return
	}

	if !rows.Next() {
		return nil, store.ErrNotFound.Stack(1)
	}

	if err = aux.scan(rows); err != nil {
		return
	}

	return aux.decode()
}

// sortableComposeResourceVersionFields returns all <no value> columns flagged as sortable
//
// # Notes
// With optional string arg, all columns are returned aliased
//
// This function is auto-generated
func (Store) sortableComposeResourceVersionFields() map[string]string {
	return map[string]string{
		"created_at":    "created_at",
		"createdat":     "created_at",
		"id":            "id",
		"resource_type": "resource_type",
		"resourcetype":  "resource_type",
		"version":       "version",
	}
}

// collectComposeResourceVersionCursorValues collects values from the given resource that and sets them to the cursor
// to be used for pagination
//
// Values that are collected must come from sortable, unique or primary columns/fields
// At least one of the collected columns must be flagged as unique, otherwise fn appends primary keys at the end
//
// # Known issues:
//
// When collecting cursor values for query that sorts by unique column with partial index (ie: unique handle on
// undeleted items)
//
// This function is auto-generated
func (s *Store) collectComposeResourceVersionCursorValues(res *composeType.ResourceVersion, cc ...*filter.SortExpr) *filter.PagingCursor {
	var (
		cur = &filter.PagingCursor{LThen: filter.SortExprSet(cc).Reversed()}

		hasUnique bool

		pkID bool

		collect = func(cc ...*filter.SortExpr) {
			getVal := func(col string) interface{} {
				switch col {
				case "id":
					pkID = true
					return res.ID
				case "resourceType":
					return res.ResourceType
				case "version":
					return res.Version
				case "createdAt":
					return res.CreatedAt
				}
				return nil
			}

			for _, c := range cc {
				switch c.Modifier() {
				case filter.COALESCE:
					var val interface{}
					for _, col := range c.Columns() {
						if reflect2.IsNil(val) {
							val = getVal(col)
						}
					}
					cur.SetModifier(c.Column, val, c.Descending, c.Modifier(), c.Columns()...)
				default:
					cur.Set(c.Column, getVal(c.Column), c.Descending)
				}
			}
		}
	)

	_ = hasUnique

	collect(cc...)
	if !hasUnique || !pkID {
		collect(&filter.SortExpr{Column: "id", Descending: false})
	}

	return cur

}

// checkComposeResourceVersionConstraints performs lookups (on valid) resource to check if any of the values on unique fields
// already exists in the store
//
// Using built-in constraint checking would be more performant, but unfortunately we cannot rely
// on the full support (MySQL does not support conditional indexes)
//
// This function is auto-generated
func (s *Store) checkComposeResourceVersionConstraints(ctx context.Context, res *composeType.ResourceVersion) (err error) {
	return nil
}
