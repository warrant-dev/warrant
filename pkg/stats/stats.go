// Copyright 2023 Forerunner Labs, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stats

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type Stat struct {
	Store    string
	Tag      string
	Duration time.Duration
}

func (s Stat) MarshalZerologObject(e *zerolog.Event) {
	e.Str("store", s.Store).Str("tag", s.Tag).Dur("duration", s.Duration)
}

type RequestStats struct {
	mutex sync.Mutex
	stats []Stat
}

func (s *RequestStats) RecordStat(stat Stat) {
	s.mutex.Lock()
	s.stats = append(s.stats, stat)
	s.mutex.Unlock()
}

func (s *RequestStats) NumStats() int {
	s.mutex.Lock()
	numStats := len(s.stats)
	s.mutex.Unlock()
	return numStats
}

func (s *RequestStats) MarshalZerologObject(e *zerolog.Event) {
	arr := zerolog.Arr()
	s.mutex.Lock()
	for _, stat := range s.stats {
		arr.Object(stat)
	}
	s.mutex.Unlock()
	e.Array("stats", arr)
}

type requestStatsKey struct{}
type statTagKey struct{}

// Create & inject a 'per-request' stats object into request context.
func RequestStatsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqStats := RequestStats{
			stats: make([]Stat, 0),
		}
		ctxWithReqStats := context.WithValue(r.Context(), requestStatsKey{}, &reqStats)
		next.ServeHTTP(w, r.WithContext(ctxWithReqStats))
	})
}

// Get RequestStats from ctx, if present.
func GetRequestStatsFromContext(ctx context.Context) *RequestStats {
	if reqStats, ok := ctx.Value(requestStatsKey{}).(*RequestStats); ok {
		return reqStats
	}
	return nil
}

// Returns a blank context with only parent's existing *RequestStats (if present).
func BlankContextWithRequestStats(parent context.Context) context.Context {
	stats := GetRequestStatsFromContext(parent)
	if stats != nil {
		return context.WithValue(context.Background(), requestStatsKey{}, stats)
	}
	return context.Background()
}

// Append a new Stat to the RequestStats obj in provided context, if present.
func RecordStat(ctx context.Context, store string, tag string, start time.Time) {
	if reqStats, ok := ctx.Value(requestStatsKey{}).(*RequestStats); ok {
		if tagPrefix, ctxHasTag := ctx.Value(statTagKey{}).(string); ctxHasTag {
			tag = tagPrefix + "." + tag
		}
		reqStats.RecordStat(Stat{
			Store:    store,
			Tag:      tag,
			Duration: time.Since(start),
		})
	}
}

// Returns a new context with given crumb appended to existing tag, if present. Otherwise, tracks the new tag in returned context. Useful for adding breadcrumbs to a Stat prior to a recording it.
func ContextWithTagCrumb(ctx context.Context, crumb string) context.Context {
	if tag, ok := ctx.Value(statTagKey{}).(string); ok {
		return context.WithValue(ctx, statTagKey{}, tag+"."+crumb)
	}
	return context.WithValue(ctx, statTagKey{}, crumb)
}
