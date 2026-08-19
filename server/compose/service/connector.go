package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/errors"
	"github.com/madnikulin50/lowcode/server/store"
)

type connectorSvc struct {
	store store.Storer
}

func connector(s store.Storer) *connectorSvc {
	return &connectorSvc{store: s}
}

func Connector() *connectorSvc {
	return &connectorSvc{store: DefaultStore}
}

func (svc *connectorSvc) Fetch(ctx context.Context, mod *types.Module, filter types.RecordFilter) (set types.RecordSet, outFilter types.RecordFilter, err error) {
	switch mod.Config.Connector.Type {
	case "rest":
		return svc.fetchREST(ctx, mod, filter)
	case "graphql":
		return svc.fetchGraphQL(ctx, mod, filter)
	case "elasticsearch":
		return svc.fetchElasticSearch(ctx, mod, filter)
	case "db":
		return svc.fetchDB(ctx, mod, filter)
	case "mongodb":
		return nil, filter, errors.Internal("mongodb connector not implemented")
	case "kafka":
		return nil, filter, errors.Internal("kafka connector not implemented")
	case "redis":
		return nil, filter, errors.Internal("redis connector not implemented")
	case "grpc":
		return nil, filter, errors.Internal("grpc connector not implemented")
	default:
		return nil, filter, errors.Internal("unknown connector type: %s", mod.Config.Connector.Type)
	}
}

func (svc *connectorSvc) Test(ctx context.Context, cfg types.ModuleConfigConnector) error {
	switch cfg.Type {
	case "rest", "graphql":
		return svc.testHTTP(ctx, cfg)
	case "elasticsearch":
		return svc.testElasticSearch(ctx, cfg)
	case "db":
		return svc.testDB(ctx, cfg)
	case "mongodb", "kafka", "redis", "grpc":
		return errors.Internal("connector test not implemented for " + cfg.Type)
	default:
		return errors.Internal("unknown connector type: %s", cfg.Type)
	}
}

func (svc *connectorSvc) testHTTP(ctx context.Context, cfg types.ModuleConfigConnector) error {
	if cfg.RestURL == "" {
		return errors.Internal("URL is empty")
	}
	req, err := http.NewRequestWithContext(ctx, "HEAD", cfg.RestURL, nil)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	for k, v := range cfg.RestHeaders {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 500 {
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}
	return nil
}

func (svc *connectorSvc) testElasticSearch(ctx context.Context, cfg types.ModuleConfigConnector) error {
	if cfg.RestURL == "" {
		return errors.Internal("URL is empty")
	}
	url := strings.TrimRight(cfg.RestURL, "/") + "/_cluster/health"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	for k, v := range cfg.RestHeaders {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 500 {
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}
	return nil
}

func (svc *connectorSvc) testDB(ctx context.Context, cfg types.ModuleConfigConnector) error {
	if cfg.DBConnectionString != "" {
		driver := cfg.DBDriver
		if driver == "" {
			driver = "postgres"
		}
		dbConn, err := sql.Open(driver, cfg.DBConnectionString)
		if err != nil {
			return fmt.Errorf("open failed: %w", err)
		}
		defer dbConn.Close()
		if err := dbConn.PingContext(ctx); err != nil {
			return fmt.Errorf("ping failed: %w", err)
		}
		return nil
	}

	rs := rdbmsStore(svc.store)
	if rs == nil {
		return errors.Internal("store type not supported")
	}
	sqlxDB, ok := rs.DB.(*sqlx.DB)
	if !ok {
		return errors.Internal("store DB type assertion failed")
	}
	if err := sqlxDB.DB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	return nil
}

func (svc *connectorSvc) fetchREST(ctx context.Context, mod *types.Module, filter types.RecordFilter) (set types.RecordSet, outFilter types.RecordFilter, err error) {
	cfg := mod.Config.Connector

	if cfg.RestURL == "" {
		return nil, filter, errors.Internal("connector REST URL is empty")
	}

	method := cfg.RestMethod
	if method == "" {
		method = "GET"
	}

	raw, err := svc.doHTTP(ctx, method, cfg.RestURL, cfg.RestHeaders, cfg.RestBody, cfg, filter)
	if err != nil {
		return nil, filter, err
	}

	items, err := parseJSONItems(raw, cfg.RestDataPath)
	if err != nil {
		return nil, filter, fmt.Errorf("connector rest parse: %w", err)
	}

	set, err = mapToRecordSet(items, mod)
	if err != nil {
		return nil, filter, err
	}

	outFilter = filter
	outFilter.Total = len(set)

	return set, outFilter, nil
}

func (svc *connectorSvc) fetchGraphQL(ctx context.Context, mod *types.Module, filter types.RecordFilter) (set types.RecordSet, outFilter types.RecordFilter, err error) {
	cfg := mod.Config.Connector

	if cfg.RestURL == "" {
		return nil, filter, errors.Internal("connector GraphQL URL is empty")
	}

	body := cfg.RestBody
	if body == "" {
		return nil, filter, errors.Internal("connector GraphQL query is empty")
	}

	gqlPayload := map[string]any{
		"query": body,
	}
	gqlBytes, err := json.Marshal(gqlPayload)
	if err != nil {
		return nil, filter, fmt.Errorf("connector graphql marshal: %w", err)
	}

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	for k, v := range cfg.RestHeaders {
		headers[k] = v
	}

	raw, err := svc.doHTTP(ctx, "POST", cfg.RestURL, headers, string(gqlBytes), cfg, filter)
	if err != nil {
		return nil, filter, err
	}

	dataPath := cfg.RestDataPath
	if dataPath == "" {
		dataPath = "data"
	}

	items, err := parseJSONItems(raw, dataPath)
	if err != nil {
		return nil, filter, fmt.Errorf("connector graphql parse: %w", err)
	}

	set, err = mapToRecordSet(items, mod)
	if err != nil {
		return nil, filter, err
	}

	outFilter = filter
	outFilter.Total = len(set)

	return set, outFilter, nil
}

func (svc *connectorSvc) fetchElasticSearch(ctx context.Context, mod *types.Module, filter types.RecordFilter) (set types.RecordSet, outFilter types.RecordFilter, err error) {
	cfg := mod.Config.Connector

	if cfg.RestURL == "" {
		return nil, filter, errors.Internal("connector ElasticSearch URL is empty")
	}

	url := strings.TrimRight(cfg.RestURL, "/")
	if cfg.EsIndex != "" {
		url = url + "/" + strings.TrimLeft(cfg.EsIndex, "/")
	}
	url = url + "/_search"

	method := cfg.RestMethod
	if method == "" {
		method = "POST"
	}

	body := cfg.RestBody
	if body == "" {
		body = `{"query": {"match_all": {}}}`
	}

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	for k, v := range cfg.RestHeaders {
		headers[k] = v
	}

	if cfg.RestLimitParam != "" && filter.Limit > 0 {
		url = fmt.Sprintf("%s?%s=%d", url, cfg.RestLimitParam, filter.Limit)
	}

	raw, err := svc.doHTTP(ctx, method, url, headers, body, cfg, filter)
	if err != nil {
		return nil, filter, err
	}

	items, err := parseJSONItems(raw, "hits.hits")
	if err != nil {
		return nil, filter, fmt.Errorf("connector elasticsearch parse: %w", err)
	}

	esItems := make([]map[string]any, len(items))
	for i, hit := range items {
		if source, ok := hit["_source"]; ok {
			if m, ok := source.(map[string]any); ok {
				esItems[i] = m
			}
		}
	}

	set, err = mapToRecordSet(esItems, mod)
	if err != nil {
		return nil, filter, err
	}

	outFilter = filter
	outFilter.Total = len(set)

	return set, outFilter, nil
}

func (svc *connectorSvc) doHTTP(ctx context.Context, method, url string, headers map[string]string, body string, cfg types.ModuleConfigConnector, filter types.RecordFilter) ([]byte, error) {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = bytes.NewBufferString(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("connector http request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if method == "GET" || method == "" {
		q := req.URL.Query()
		if cfg.RestLimitParam != "" && filter.Limit > 0 {
			q.Set(cfg.RestLimitParam, strconv.Itoa(int(filter.Limit)))
		}
		if cfg.RestOffsetParam != "" && filter.Limit > 0 && filter.PageCursor != nil {
			offset := filter.PageCursor.Values()
			if len(offset) > 0 {
				q.Set(cfg.RestOffsetParam, fmt.Sprintf("%v", offset[0]))
			}
		}
		if len(q) > 0 {
			req.URL.RawQuery = q.Encode()
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connector http do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("connector http bad status: %d", resp.StatusCode)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("connector http read: %w", err)
	}

	return raw, nil
}

func parseJSONItems(raw []byte, dataPath string) ([]map[string]any, error) {
	var items []map[string]any

	if dataPath != "" {
		var root map[string]any
		if err := json.Unmarshal(raw, &root); err != nil {
			return nil, fmt.Errorf("json decode root: %w", err)
		}

		parts := splitPath(dataPath)
		current := any(root)
		for _, part := range parts {
			m, ok := current.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("json path %q: expected object at %q", dataPath, part)
			}
			current = m[part]
			if current == nil {
				return nil, fmt.Errorf("json path %q: key %q not found", dataPath, part)
			}
		}

		arr, ok := current.([]any)
		if !ok {
			return nil, fmt.Errorf("json path %q: expected array, got %T", dataPath, current)
		}
		items = make([]map[string]any, len(arr))
		for i, v := range arr {
			m, ok := v.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("json path %q: item %d is not object", dataPath, i)
			}
			items[i] = m
		}
	} else {
		if err := json.Unmarshal(raw, &items); err != nil {
			var single map[string]any
			if err2 := json.Unmarshal(raw, &single); err2 != nil {
				return nil, fmt.Errorf("json decode: %w", err)
			}
			items = []map[string]any{single}
		}
	}

	return items, nil
}

func splitPath(path string) []string {
	var parts []string
	for _, part := range bytes.Split([]byte(path), []byte{'.'}) {
		parts = append(parts, string(part))
	}
	return parts
}

func mapToRecordSet(items []map[string]any, mod *types.Module) (types.RecordSet, error) {
	set := make(types.RecordSet, len(items))

	mapping := make(map[string]string)
	for _, fm := range mod.Config.Connector.FieldMapping {
		mapping[fm.Field] = fm.Source
	}

	for i, item := range items {
		values := make(types.RecordValueSet, 0, len(mod.Fields))
		for _, f := range mod.Fields {
			sourceKey := f.Name
			if mapped, ok := mapping[f.Name]; ok {
				sourceKey = mapped
			}

			v, exists := item[sourceKey]
			if !exists {
				continue
			}

			if v == nil {
				continue
			}

			values = append(values, &types.RecordValue{
				Name:  f.Name,
				Value: fmt.Sprintf("%v", v),
			})
		}
		set[i] = &types.Record{
			NamespaceID: mod.NamespaceID,
			ModuleID:    mod.ID,
			Values:      values,
		}
	}

	return set, nil
}

func (svc *connectorSvc) fetchDB(ctx context.Context, mod *types.Module, filter types.RecordFilter) (set types.RecordSet, outFilter types.RecordFilter, err error) {
	cfg := mod.Config.Connector

	if cfg.DBQuery == "" {
		return nil, filter, errors.Internal("connector DB query is empty")
	}

	var rows *sql.Rows

	if cfg.DBConnectionString != "" {
		driver := cfg.DBDriver
		if driver == "" {
			driver = "postgres"
		}
		dbConn, openErr := sql.Open(driver, cfg.DBConnectionString)
		if openErr != nil {
			return nil, filter, fmt.Errorf("connector db open: %w", openErr)
		}
		defer dbConn.Close()

		if pingErr := dbConn.PingContext(ctx); pingErr != nil {
			return nil, filter, fmt.Errorf("connector db ping: %w", pingErr)
		}

		rows, err = dbConn.QueryContext(ctx, cfg.DBQuery)
	} else {
		rs := rdbmsStore(svc.store)
		if rs == nil {
			return nil, filter, errors.Internal("store type not supported for DB connector")
		}
		rows, err = rs.DB.QueryContext(ctx, cfg.DBQuery)
	}
	if err != nil {
		return nil, filter, fmt.Errorf("connector db query: %w", err)
	}
	defer rows.Close()

	result, scanErr := scanRows(rows)
	if scanErr != nil {
		return nil, filter, scanErr
	}

	set, err = mapToRecordSet(result, mod)
	if err != nil {
		return nil, filter, err
	}

	outFilter = filter
	outFilter.Total = len(set)

	return set, outFilter, nil
}

func scanRows(rows *sql.Rows) ([]map[string]any, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("connector db columns: %w", err)
	}

	var result []map[string]any
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("connector db scan: %w", err)
		}

		item := make(map[string]any, len(columns))
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				item[col] = string(b)
			} else {
				item[col] = val
			}
		}
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("connector db rows: %w", err)
	}

	return result, nil
}
