package dalutils

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/madnikulin50/lowcode/server/compose/service/values"
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/dal"
	"github.com/madnikulin50/lowcode/server/pkg/filter"
	"github.com/spf13/cast"
)

type (
	creator interface {
		Create(ctx context.Context, m dal.ModelRef, operations dal.OperationSet, vv ...dal.ValueGetter) error
	}

	updater interface {
		Update(ctx context.Context, m dal.ModelRef, operations dal.OperationSet, rr ...dal.ValueGetter) (err error)
	}

	searcher interface {
		Search(ctx context.Context, m dal.ModelRef, operations dal.OperationSet, f filter.Filter) (dal.Iterator, error)
	}

	lookuper interface {
		Lookup(ctx context.Context, m dal.ModelRef, operations dal.OperationSet, lookup dal.ValueGetter, dst dal.ValueSetter) (err error)
	}

	deleter interface {
		Delete(ctx context.Context, m dal.ModelRef, operations dal.OperationSet, pkv ...dal.ValueGetter) (err error)
	}

	counter interface {
		Count(ctx context.Context, m dal.ModelRef, operations dal.OperationSet, f filter.Filter) (uint, error)
	}
)

// ComposeRecordsList iterates over results and collects all available records
func ComposeRecordsList(ctx context.Context, s searcher, mod *types.Module, filter types.RecordFilter) (set types.RecordSet, outFilter types.RecordFilter, err error) {
	iter, err := prepIterator(ctx, s, mod, filter)
	if err != nil {
		return
	}

	set, _, outFilter, err = drainIterator(ctx, iter, mod, filter)
	return
}

func ComposeRecordsListN(ctx context.Context, s searcher, mod *types.Module, filter types.RecordFilter) (set types.RecordSet, summaries map[string]types.RecordSummary, outFilter types.RecordFilter, err error) {
	iter, err := prepIterator(ctx, s, mod, filter)
	if err != nil {
		return
	}

	set, summaries, outFilter, err = drainIterator(ctx, iter, mod, filter)
	return
}

func ComposeRecordsIterator(ctx context.Context, s searcher, mod *types.Module, filter types.RecordFilter) (iter dal.Iterator, outFilter types.RecordFilter, err error) {
	iter, err = prepIterator(ctx, s, mod, filter)
	if err != nil {
		return
	}

	outFilter = filter
	outFilter.Paging = *filter.Paging.Clone()

	return
}

func ComposeRecordsFind(ctx context.Context, l lookuper, mod *types.Module, recordID uint64) (out *types.Record, err error) {
	out = prepareRecordTarget(mod)

	err = l.Lookup(ctx, mod.ModelRef(), recLookupOperations(mod), dal.PKValues{"id": recordID}, out)
	if err != nil {
		return
	}

	return
}

func ComposeRecordsCount(ctx context.Context, c counter, mod *types.Module, filter types.RecordFilter) (cnt uint, err error) {
	// Same constraints as list: module DAL config + model.Constraints
	// (namespaceID/moduleID on compose_record). Do not force namespaceID here —
	// omitted system fields have no matching attribute.
	dalFilter := prepFilter(filter, mod)

	return c.Count(ctx, mod.ModelRef(), recLookupOperations(mod), dalFilter)
}

// ComposeRecordsCountWithTimeout is the Metric/Progress report COUNT path.
// Uses recordReportCountTimeout; on timeout return TotalUnknown (-1). List
// incTotal does not share this wait — it skips JSON Record/User filters or
// uses recordListCountTimeout (1s).
func ComposeRecordsCountWithTimeout(ctx context.Context, c counter, mod *types.Module, flt types.RecordFilter) (cnt int, err error) {
	if queryFiltersJSONRecordField(mod, flt.Query) {
		return filter.TotalUnknown, nil
	}
	return countWithTimeout(ctx, recordReportCountTimeout, func(countCtx context.Context) (uint, error) {
		return ComposeRecordsCount(countCtx, c, mod, flt)
	})
}

func ComposeRecordCreate(ctx context.Context, c creator, mod *types.Module, records ...*types.Record) (err error) {
	return c.Create(ctx, mod.ModelRef(), recCreateOperations(mod), recToGetters(records...)...)
}

func ComposeRecordUpdate(ctx context.Context, u updater, mod *types.Module, records ...*types.Record) (err error) {
	return u.Update(ctx, mod.ModelRef(), recUpdateOperations(mod), recToGetters(records...)...)
}

func ComposeRecordSoftDelete(ctx context.Context, u updater, mod *types.Module, records ...*types.Record) (err error) {
	return u.Update(ctx, mod.ModelRef(), recUpdateOperations(mod), recToGetters(records...)...)
}

func ComposeRecordUndelete(ctx context.Context, u updater, mod *types.Module, records ...*types.Record) (err error) {
	return u.Update(ctx, mod.ModelRef(), recUpdateOperations(mod), recToGetters(records...)...)
}

func ComposeRecordDelete(ctx context.Context, d deleter, mod *types.Module, records ...*types.Record) (err error) {
	return d.Delete(ctx, mod.ModelRef(), recDeleteOperations(mod), recToGetters(records...)...)
}

func WalkIterator(ctx context.Context, iter dal.Iterator, mod *types.Module, f func(r *types.Record) error) (err error) {
	for iter.Next(ctx) {
		r := prepareRecordTarget(mod)
		if err = iter.Scan(r); err != nil {
			return
		}

		if len(r.Values) < len(mod.Fields) {
			values.Expression(ctx, mod, r, nil, nil)
		}

		if err = f(r); err != nil {
			return
		}
	}

	return iter.Err()
}

// // // // // // // // // // // // // // // // // // // // // // // // //
// Utils

func prepFilter(filter types.RecordFilter, mod *types.Module) filter.Filter {
	return filter.ToConstraintedFilter(mod.Config.DAL.Constraints)
}

func prepIterator(ctx context.Context, dal searcher, mod *types.Module, filter types.RecordFilter) (iter dal.Iterator, err error) {
	dalFilter := prepFilter(filter, mod)

	iter, err = dal.Search(ctx, mod.ModelRef(), recSearchOperations(mod, filter), dalFilter)
	return
}

// drains iterator and collects all records
//
// Collection of records is done with respect to check function and limit constraint on record filter
// For any other filter constraint we assume that underlying DAL took care of it
func drainIterator(ctx context.Context, iter dal.Iterator, mod *types.Module, f types.RecordFilter) (set types.RecordSet, summaries map[string]types.RecordSummary, outFilter types.RecordFilter, err error) {
	// close iterator after we've drained it
	defer iter.Close()

	if f.PageCursor != nil {
		if f.IncPageNavigation || f.IncTotal {
			err = fmt.Errorf("not allowed to fetch page navigation or total item count with page cursor")
			return
		}
	}

	var (
		ok         bool
		fetched    uint
		filtered   uint
		lastRecord *types.Record
		fetchLimit = f.Limit
	)

	// Get the requested number of record
	if f.Limit > 0 {
		set = make(types.RecordSet, 0, f.Limit)
	} else {
		set = make(types.RecordSet, 0, 1000)
	}

	for f.Limit == 0 || uint(len(set)) < f.Limit {
		var firstRecord *types.Record

		// reset counters every drain
		fetched = 0
		filtered = 0

		add := make(types.RecordSet, 0, 12)
		err = WalkIterator(ctx, iter, mod, func(r *types.Record) error {
			lastRecord = r
			if firstRecord == nil {
				firstRecord = r
			}

			// check fetched record
			if f.Check != nil {
				if ok, err = f.Check(r); err != nil {
					return err
				} else if !ok {
					filtered++
					return nil
				}
			}

			fetched++
			add = append(add, r)
			return err
		})

		// if an error occurred inside Next()/WalkIterator,
		// we need to stop draining
		if err != nil {
			if isCountTimeout(err) {
				// Page SELECT hit statement_timeout / ctx deadline. Keep
				// whatever rows were already scanned so HTTP can return.
				err = nil
				if f.PageCursor != nil && f.PageCursor.ROrder {
					set = append(add, set...)
				} else {
					set = append(set, add...)
				}
				outFilter = f
				outFilter.IncPageNavigation = false
				if f.IncTotal {
					outFilter.Total = filter.TotalUnknown
				}
			}
			return
		}

		// If it's reverse, we need to add extra fetches to the start
		if f.PageCursor != nil && f.PageCursor.ROrder {
			set = append(add, set...)
		} else {
			set = append(set, add...)
		}

		total := fetched + filtered
		if total == 0 || f.Limit == 0 {
			// iterator empty, or no limit (everything was fetched)
			break
		}
		if fetchLimit > 0 && total < fetchLimit {
			// SQL returned a short page — no more rows. Do not probe with
			// More()/cursor: that extra SELECT is what broke bulk delete
			// (recordID='a' OR recordID='b' … with fewer matches than Limit).
			break
		}

		// Fetch more records (AC/check filtered some rows out of a full SQL page)
		setLen := uint(len(set))
		if total > 0 && setLen < f.Limit {
			fetchMore := f.Limit - setLen
			var crsrRec *types.Record

			// request more items
			if f.PageCursor == nil || !f.PageCursor.ROrder {
				crsrRec = lastRecord
			} else {
				crsrRec = firstRecord
			}

			if err = iter.More(fetchMore, crsrRec); err != nil {
				return
			}
			fetchLimit = fetchMore
		}
	}

	// Get the page nav/total/next-prev cursors
	nav, auxSm, err := generatePageNavigation(ctx, iter, mod, f, set)
	if err != nil {
		return
	}

	summaries = auxSm

	// Make out filter
	outFilter = f
	outFilter.Paging = nav.Paging
	outFilter.Sorting = nav.Sorting

	return
}

// generatePageNavigation generates page navigation for a given record set using an iterator and filter limit.
// If the limit is not defined, the page navigation will consist of only one page without a cursor.
// If the limit is defined and is greater than the total number of records in the set,
// the page navigation will consist of only one page without a cursor.
// If the limit is defined and is less than the total number of records in the set,
// the page navigation will have multiple pages with cursor(s) based on the total number of records and the provided limit.
// @todo revisit and clean up this function properly
func generatePageNavigation(ctx context.Context, iter dal.Iterator, mod *types.Module, p types.RecordFilter, set types.RecordSet) (out types.RecordFilter, summaries map[string]types.RecordSummary, err error) {
	const (
		howMuchMore = 1000
	)
	var (
		ok      bool
		total   = uint(len(set))
		setLen  = len(set)
		counter = 0

		first *types.Record
		last  *types.Record
		page  filter.Page

		pageNavigation = []*filter.Page{
			{
				Page:   1,
				Count:  0,
				Cursor: nil,
			},
		}

		// generatePage generates pageNavigation for given record set
		generatePage = func(last *types.Record) (err error) {
			if !p.IncPageNavigation || p.Limit == 0 || len(pageNavigation) == 0 {
				return
			}

			lastNavPageNo := len(pageNavigation) - 1
			nextPage, err := iter.ForwardCursor(last)
			if err != nil {
				return
			}

			if total < p.Limit {
				pageNavigation[lastNavPageNo].Count = total
			}

			// prepare page
			if total != 0 && (total%p.Limit) == 0 {
				pageNavigation[lastNavPageNo].Count = p.Limit
				page = filter.Page{
					Page:   uint(len(pageNavigation) + 1),
					Count:  p.Limit,
					Cursor: nextPage,
				}
			}

			expectedItemCountUpToPage := uint(lastNavPageNo+1) * p.Limit
			if p.Limit == 1 {
				expectedItemCountUpToPage = uint(lastNavPageNo) * p.Limit
			}

			if expectedItemCountUpToPage < total {
				// push page when limit is matched with the previous page item size
				if pageNavigation[lastNavPageNo].Count == p.Limit {
					pageNavigation = append(pageNavigation, &filter.Page{
						Page:   page.Page,
						Count:  total % p.Limit,
						Cursor: page.Cursor, // prev cursor
					})
				}
			}

			return
		}

		recordChecker = func(i dal.Iterator) (ok bool, err error) {
			if p.Check == nil {
				return true, err
			}

			rc := prepareRecordTarget(mod)
			err = i.Scan(rc)

			if err != nil {
				return
			}

			return p.Check(rc)
		}
	)

	if len(p.Summaries) > 0 {
		summaries = make(map[string]types.RecordSummary, len(p.Summaries))
	}

	if setLen == 0 {
		return
	}

	existing := make(map[any]struct{}, 24)
	looped := false

	procSummary := func(r *types.Record) (err error) {
		for _, smDef := range p.Summaries {
			// Get record value
			var vv []any
			vv, err = r.GetValues(smDef.Field)
			if err != nil {
				return
			}

			bit := summaries[fmt.Sprintf("%s %s", smDef.Name, smDef.Field)]

			// This will be constant so we're good
			bit.Name = smDef.Name

			// Skip empty
			if len(vv) == 0 {
				bit.EmptyCount++
				summaries[fmt.Sprintf("%s %s", smDef.Name, smDef.Field)] = bit
				continue
			}

			bit.NotEmptyCount++

			for _, v := range vv {
				bit.Count++

				if _, ok := existing[v]; !ok {
					existing[v] = struct{}{}
					bit.UniqueCount++
				}
			}

			switch smDef.Name {
			case "min":
				for _, v := range vv {
					if !looped {
						bit.Min = cast.ToFloat64(v)
					} else {
						bit.Min = math.Min(bit.Min, cast.ToFloat64(v))
					}
				}

			case "max":
				for _, v := range vv {
					if !looped {
						bit.Max = cast.ToFloat64(v)
					} else {
						bit.Max = math.Max(bit.Max, cast.ToFloat64(v))
					}
				}

			case "avg":
				for _, v := range vv {
					bit.Sum += cast.ToFloat64(v)
				}

			case "sum":
				for _, v := range vv {
					bit.Sum += cast.ToFloat64(v)
				}

			case "earliest":
				for _, v := range vv {
					if bit.Earliest.IsZero() {
						bit.Earliest = cast.ToTime(v)
					} else {
						aux := cast.ToTime(v)
						if !aux.IsZero() && aux.Before(bit.Earliest) {
							bit.Earliest = aux
						}
					}
				}

			case "latest":
				for _, v := range vv {
					if bit.Latest.IsZero() {
						bit.Latest = cast.ToTime(v)
					} else {
						aux := cast.ToTime(v)
						if !aux.IsZero() && aux.After(bit.Latest) {
							bit.Latest = aux
						}
					}
				}
			}

			looped = true
			summaries[fmt.Sprintf("%s %s", smDef.Name, smDef.Field)] = bit
		}
		return
	}

	// Firstly sort out the current things
	if len(p.Summaries) > 0 {
		for _, r := range set {
			err = procSummary(r)
			if err != nil {
				return
			}
		}
	}

	first = set[0]
	last = set[setLen-1]

	// Limit
	out.Limit = p.Limit

	// Sorting
	out.Sort = dal.IteratorSorting(iter)

	// First page is already in `set`. Next/prev cursor probes and page-nav
	// walks issue extra SELECTs; a SQL error there must not drop the page
	// (admin lists always set IncPageNavigation and would otherwise render empty).
	keepCollected := func() {
		err = nil
		if !p.IncTotal {
			return
		}
		if p.PageCursor == nil && (p.Limit == 0 || uint(setLen) < p.Limit) {
			out.Total = setLen
			return
		}
		out.Total = filter.TotalUnknown
	}

	// Probe next/prev only when the page is full. A short page means the
	// iterator is exhausted; PreLoadCursor would still run an extra SELECT.
	if p.Limit > 0 && uint(len(set)) >= p.Limit {
		// PrevPage
		if p.PageCursor != nil {
			out.PrevPage, err = dal.PreLoadCursor(ctx, iter, 100, true, first, recordChecker)
			if err != nil {
				keepCollected()
				return
			}
		}

		// NextPage
		out.NextPage, err = dal.PreLoadCursor(ctx, iter, 100, false, last, recordChecker)
		if err != nil {
			keepCollected()
			return
		}
	}

	if iter.Err() != nil {
		keepCollected()
		return
	}

	// Fast path: total only (no page-nav, no summaries).
	// Never block the HTTP response on a long COUNT — the first page (and
	// next/prev cursors) is already in `out`. JSON-stored Record/User filters
	// (e.g. `device = ID`) make COUNT a seq-scan of compose_record.values;
	// skip it and return TotalUnknown (-1). Other filters get a 1s budget.
	if p.IncTotal && !p.IncPageNavigation && len(p.Summaries) == 0 {
		if p.PageCursor == nil && (p.Limit == 0 || uint(setLen) < p.Limit) {
			out.Total = setLen
			return
		}
		if queryFiltersJSONRecordField(mod, p.Query) {
			out.Total = filter.TotalUnknown
			return
		}
		if _, ok := iter.(interface {
			Count(context.Context) (uint, error)
		}); ok {
			out.Total, err = countRecordsWithTimeout(ctx, iter)
			if err != nil {
				keepCollected()
			}
			return
		}
		out.Total = filter.TotalUnknown
		return
	}

	if p.IncTotal || p.IncPageNavigation || len(p.Summaries) > 0 {
		// For the first page nav
		err = generatePage(last)
		if err != nil {
			keepCollected()
			return
		}

		for counter == 0 || counter < howMuchMore {
			counter++

			interLoop := 0

			if err = iter.More(howMuchMore, last); err != nil {
				keepCollected()
				return
			}

			err = WalkIterator(ctx, iter, mod, func(rec *types.Record) error {
				// check fetched record
				if p.Check != nil {
					if ok, err = p.Check(rec); err != nil {
						return err
					} else if !ok {
						return nil
					}
				}

				err = procSummary(rec)
				if err != nil {
					return err
				}

				interLoop++
				total++
				last = rec
				return generatePage(rec)
			})
			if err != nil {
				keepCollected()
				return
			}

			if interLoop < howMuchMore {
				break
			}
		}
	}

	// Total
	if p.IncTotal {
		out.Total = int(total)
	}

	// Page navigation
	if p.IncPageNavigation {
		// Ensure that the last page count is correct if it's not equal to the limit.
		lastPageCount := pageNavigation[len(pageNavigation)-1].Count
		if lastPageCount > 0 && lastPageCount != p.Limit && lastPageCount != total%p.Limit {
			pageNavigation[len(pageNavigation)-1].Count = total % p.Limit
		}

		if p.Limit == 1 {
			pageNavigation = pageNavigation[:len(pageNavigation)-1]
		}

		out.PageNavigation = pageNavigation
	}

	// Do averages
	if len(p.Summaries) > 0 {
		for n, s := range summaries {
			if s.Name != "avg" {
				continue
			}

			s.Avg = s.Sum / float64(s.Count)
			summaries[n] = s
		}
	}

	return
}

var (
	// List incTotal must not stall the HTTP response. 1s is enough for an
	// indexed COUNT; JSON-extract COUNT is skipped entirely (TotalUnknown).
	recordListCountTimeout = time.Second
	// Metric/Progress report COUNT for non-JSON filters. JSON Record/User
	// predicates skip COUNT (TotalUnknown) instead of waiting.
	recordReportCountTimeout = 8 * time.Second
	// RecordSearchBudget caps list/report DAL work per HTTP request so a
	// JSON TypeRef seq-scan cannot hold the handler forever.
	RecordSearchBudget = 8 * time.Second
)

// BoundRecordSearchContext puts a hard deadline on list/report DAL calls.
func BoundRecordSearchContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if dl, ok := ctx.Deadline(); ok {
		if time.Until(dl) <= RecordSearchBudget {
			return ctx, func() {}
		}
	}
	return context.WithTimeout(ctx, RecordSearchBudget)
}

// IsSearchTimeout reports a canceled/deadline/statement_timeout from list or COUNT.
func IsSearchTimeout(err error) bool {
	return isCountTimeout(err)
}

// QueryFiltersJSONRecordField is true when the QL query mentions a module
// field stored in compose_record.values as a Record/User TypeRef. COUNT of
// those predicates seq-scans JSON and must not block list or report HTTP.
func QueryFiltersJSONRecordField(mod *types.Module, query string) bool {
	return queryFiltersJSONRecordField(mod, query)
}

func queryFiltersJSONRecordField(mod *types.Module, query string) bool {
	if mod == nil || strings.TrimSpace(query) == "" {
		return false
	}
	q := strings.ToLower(query)
	for _, f := range mod.Fields {
		if f == nil {
			continue
		}
		if f.Kind != "Record" && f.Kind != "User" {
			continue
		}
		name := strings.ToLower(f.Name)
		if name == "" {
			continue
		}
		if strings.Contains(q, name) {
			return true
		}
	}
	return false
}

// countWithTimeout runs COUNT with the given budget. If the budget expires
// while the parent ctx is still valid, TotalUnknown (-1) is returned.
// The wait is in a goroutine so HTTP returns even when the DB driver
// does not abort QueryRow immediately on context cancel.
func countWithTimeout(ctx context.Context, timeout time.Duration, fn func(context.Context) (uint, error)) (int, error) {
	countCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	type outcome struct {
		n   uint
		err error
	}
	ch := make(chan outcome, 1)
	go func() {
		n, err := fn(countCtx)
		ch <- outcome{n, err}
	}()

	select {
	case o := <-ch:
		if o.err != nil {
			if ctx.Err() != nil {
				return 0, ctx.Err()
			}
			if countCtx.Err() != nil || isCountTimeout(o.err) {
				return filter.TotalUnknown, nil
			}
			return 0, o.err
		}
		return int(o.n), nil
	case <-countCtx.Done():
		if ctx.Err() != nil {
			return 0, ctx.Err()
		}
		return filter.TotalUnknown, nil
	}
}

func isCountTimeout(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "deadline exceeded") ||
		strings.Contains(s, "statement timeout") ||
		strings.Contains(s, "query canceled") ||
		strings.Contains(s, "canceling statement")
}

// countRecordsWithTimeout runs iterator COUNT after the first page is already
// fetched. If COUNT exceeds 1s the query is canceled and TotalUnknown (-1)
// is returned so the HTTP handler can send the page without hanging.
func countRecordsWithTimeout(ctx context.Context, iter dal.Iterator) (int, error) {
	c, ok := iter.(interface {
		Count(context.Context) (uint, error)
	})
	if !ok {
		return 0, fmt.Errorf("iterator does not support count")
	}

	return countWithTimeout(ctx, recordListCountTimeout, c.Count)
}

func prepareRecordTarget(module *types.Module) *types.Record {
	// so we can avoid some code later involving (non)partitioned modules :seenoevil:
	r := &types.Record{
		ModuleID:    module.ID,
		NamespaceID: module.NamespaceID,
		Values:      make(types.RecordValueSet, 0, len(module.Fields)),
	}
	r.SetModule(module)

	return r
}

func recToGetters(rr ...*types.Record) (out []dal.ValueGetter) {
	out = make([]dal.ValueGetter, len(rr))

	for i := range rr {
		out[i] = rr[i]
	}

	return
}

func recCreateOperations(m *types.Module) (out dal.OperationSet) {
	return dal.CreateOperations()
}

func recUpdateOperations(m *types.Module) (out dal.OperationSet) {
	return dal.UpdateOperations()
}

func recDeleteOperations(m *types.Module) (out dal.OperationSet) {
	return dal.DeleteOperations()
}

func recFilterOperations(f types.RecordFilter) (out dal.OperationSet) {
	if f.PageCursor != nil {
		out = append(out, dal.Paging)
	}

	if f.IncPageNavigation {
		out = append(out, dal.Paging)
	}

	if f.Sort != nil {
		out = append(out, dal.Sorting)
	}

	return
}

func recSearchOperations(m *types.Module, f types.RecordFilter) (out dal.OperationSet) {
	return dal.SearchOperations().
		Union(recFilterOperations(f))
}

func recLookupOperations(m *types.Module) (out dal.OperationSet) {
	return dal.LookupOperations()
}
