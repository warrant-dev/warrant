package stats

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type QueryStat struct {
	Store    string
	Query    string
	Duration time.Duration
}

func (q QueryStat) MarshalZerologObject(e *zerolog.Event) {
	e.Str("store", q.Store).Str("query", q.Query).Dur("duration", q.Duration)
}

type RequestStats struct {
	Queries []QueryStat
}

func (s *RequestStats) MarshalZerologObject(e *zerolog.Event) {
	arr := zerolog.Arr()
	for _, query := range s.Queries {
		arr.Object(query)
	}
	e.Array("queries", arr)
}

type requestStatsKey struct{}
type queryCrumbsKey struct{}

// Create & inject a 'per-request' stats object into request context
func RequestStatsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqStats := RequestStats{
			Queries: make([]QueryStat, 0),
		}
		ctxWithReqStats := context.WithValue(r.Context(), requestStatsKey{}, &reqStats)
		next.ServeHTTP(w, r.WithContext(ctxWithReqStats))
	})
}

// Get RequestStats from ctx, if present
func GetRequestStatsFromContext(ctx context.Context) *RequestStats {
	if reqStats, ok := ctx.Value(requestStatsKey{}).(*RequestStats); ok {
		return reqStats
	}
	return nil
}

// Append a new QueryStat to the RequestStats obj in provided context, if present
func RecordQueryStat(ctx context.Context, store string, query string, duration time.Duration) {
	if reqStats, ok := ctx.Value(requestStatsKey{}).(*RequestStats); ok {
		if queryPrefix, ctxHasQuery := ctx.Value(queryCrumbsKey{}).(string); ctxHasQuery {
			query = queryPrefix + "." + query
		}
		reqStats.Queries = append(reqStats.Queries, QueryStat{
			Store:    store,
			Query:    query,
			Duration: duration,
		})
	}
}

// Returns a new context with given crumb appended to existing query, if present. Otherwise, tracks the new query in returned context. Useful for adding breadcrumbs to QueryStats prior to a RecordQuery.
func ContextWithQueryCrumb(ctx context.Context, crumb string) context.Context {
	if query, ok := ctx.Value(queryCrumbsKey{}).(string); ok {
		return context.WithValue(ctx, queryCrumbsKey{}, query+"."+crumb)
	}
	return context.WithValue(ctx, queryCrumbsKey{}, crumb)
}
