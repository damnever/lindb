// Licensed to LinDB under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. LinDB licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package aggregation

import (
	"sync"

	"github.com/lindb/lindb/pkg/timeutil"
	"github.com/lindb/lindb/series"
	"github.com/lindb/lindb/series/field"
)

//go:generate mockgen -source=./series_agg.go -destination=./series_agg_mock.go -package=aggregation

// FieldAggregates represents aggregator which aggregates fields of a time series
type FieldAggregates []SeriesAggregator

// ResultSet returns the result set of aggregator
func (agg FieldAggregates) ResultSet(tags string) series.GroupedIterator {
	return newGroupedIterator(tags, agg)
}

// Reset resets the aggregator's context for reusing
func (agg FieldAggregates) Reset() {
	for _, aggregator := range agg {
		aggregator.Reset()
	}
}

// NewFieldAggregates creates the field aggregates based on aggregator specs and query time range.
// NOTE: if it does down sampling aggregator, aggregator specs must be in order by field id.
func NewFieldAggregates(
	queryInterval timeutil.Interval,
	intervalRatio int,
	queryTimeRange timeutil.TimeRange,
	aggSpecs AggregatorSpecs,
) FieldAggregates {
	aggregates := make(FieldAggregates, len(aggSpecs))
	for idx, aggSpec := range aggSpecs {
		aggregates[idx] = NewSeriesAggregator(queryInterval, intervalRatio, queryTimeRange, aggSpec)
	}
	return aggregates
}

// SeriesAggregator represents a series aggregator which aggregates one field of a time series
type SeriesAggregator interface {
	// FieldName returns field name
	FieldName() field.Name
	// GetFieldType returns field type
	GetFieldType() field.Type
	// GetAggregator gets field aggregator by segment start time, if not exist return (nil,false).
	GetAggregator(segmentStartTime int64) (agg FieldAggregator, ok bool)
	// GetAggregates returns all field aggregators.
	GetAggregates() []FieldAggregator
	// ResultSet returns the result set of series aggregator.
	ResultSet() series.Iterator
	// Reset resets the aggregator's context for reusing.
	Reset()
}

// seriesAggregator implements SeriesAggregator.
type seriesAggregator struct {
	fieldName field.Name
	fieldType field.Type

	queryInterval  timeutil.Interval
	queryTimeRange timeutil.TimeRange
	intervalRatio  int

	aggregates []FieldAggregator
	aggSpec    AggregatorSpec
	calc       timeutil.IntervalCalculator

	startTime int64

	mutex sync.Mutex
}

// NewSeriesAggregator creates a series aggregator.
func NewSeriesAggregator(
	queryInterval timeutil.Interval,
	intervalRatio int,
	queryTimeRange timeutil.TimeRange,
	aggSpec AggregatorSpec,
) SeriesAggregator {
	calc := queryInterval.Calculator()

	return &seriesAggregator{
		fieldName:      aggSpec.FieldName(),
		fieldType:      aggSpec.GetFieldType(),
		startTime:      calc.CalcFamilyTime(queryTimeRange.Start),
		calc:           calc,
		intervalRatio:  intervalRatio,
		queryInterval:  queryInterval,
		queryTimeRange: queryTimeRange,
		aggSpec:        aggSpec,
	}
}

// FieldName returns field name.
func (a *seriesAggregator) FieldName() field.Name {
	return a.fieldName
}

// GetFieldType returns the field type.
func (a *seriesAggregator) GetFieldType() field.Type {
	return a.fieldType
}

// ResultSet returns the result set of series aggregator.
func (a *seriesAggregator) ResultSet() series.Iterator {
	return newSeriesIterator(a)
}

// Reset resets the aggregator's context for reusing.
func (a *seriesAggregator) Reset() {
	for _, agg := range a.aggregates {
		if agg != nil {
			agg.reset()
		}
	}
}

// GetAggregates returns all field aggregators.
func (a *seriesAggregator) GetAggregates() []FieldAggregator {
	return a.aggregates
}

// GetAggregator gets field aggregator by segment start time, if not exist return (nil,false).
func (a *seriesAggregator) GetAggregator(segmentStartTime int64) (agg FieldAggregator, ok bool) {
	if segmentStartTime < a.startTime {
		return
	}
	familyTimeForQuery := a.calc.CalcFamilyTime(segmentStartTime)
	slotRange := a.queryInterval.CalcSlotRange(familyTimeForQuery, a.queryTimeRange)
	agg = NewFieldAggregator(a.aggSpec, familyTimeForQuery, int(slotRange.Start), int(slotRange.End))

	a.mutex.Lock()
	a.aggregates = append(a.aggregates, agg)
	a.mutex.Unlock()
	ok = true
	return
}
