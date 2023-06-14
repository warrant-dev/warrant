package stats

import (
	"time"

	"github.com/rs/zerolog"
)

type RequestStatsKey struct{}

type QueryStat struct {
	Store     string
	QueryType string
	Duration  time.Duration
}

func (q QueryStat) MarshalZerologObject(e *zerolog.Event) {
	e.Str("store", q.Store).Str("type", q.QueryType).Dur("duration", q.Duration)
}

type RequestStats struct {
	NumQueries int
	Queries    []QueryStat
}

func (s *RequestStats) RecordQuery(stat QueryStat) {
	s.Queries = append(s.Queries, stat)
	s.NumQueries++
}

func (s *RequestStats) MarshalZerologObject(e *zerolog.Event) {
	e.Int("numQueries", s.NumQueries)
	arr := zerolog.Arr()
	for _, query := range s.Queries {
		arr.Object(query)
	}
	e.Array("queries", arr)
}
