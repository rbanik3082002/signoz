package delta

import (
	"fmt"

	"go.signoz.io/signoz/pkg/query-service/app/metrics/v4/helpers"
	"go.signoz.io/signoz/pkg/query-service/constants"
	v3 "go.signoz.io/signoz/pkg/query-service/model/v3"
	"go.signoz.io/signoz/pkg/query-service/utils"
)

// prepareTimeAggregationSubQuery builds the sub-query to be used for temporal aggregation
func prepareTimeAggregationSubQuery(start, end, step int64, mq *v3.BuilderQuery) (string, error) {

	var subQuery string

	timeSeriesSubQuery, err := helpers.PrepareTimeseriesFilterQuery(mq)
	if err != nil {
		return "", err
	}

	samplesTableFilter := fmt.Sprintf("metric_name = %s AND timestamp_ms >= %d AND timestamp_ms <= %d", utils.ClickHouseFormattedValue(mq.AggregateAttribute.Key), start, end)

	// Select the aggregate value for interval
	queryTmpl :=
		"SELECT fingerprint, %s" +
			" toStartOfInterval(toDateTime(intDiv(timestamp_ms, 1000)), INTERVAL %d SECOND) as ts," +
			" %s as per_series_value" +
			" FROM " + constants.SIGNOZ_METRIC_DBNAME + "." + constants.SIGNOZ_SAMPLES_TABLENAME +
			" INNER JOIN" +
			" (%s) as filtered_time_series" +
			" USING fingerprint" +
			" WHERE " + samplesTableFilter +
			" GROUP BY fingerprint, ts" +
			" ORDER BY fingerprint, ts"

	selectLabelsAny := helpers.SelectLabelsAny(mq.GroupBy)

	switch mq.TimeAggregation {
	case v3.TimeAggregationAvg:
		op := "avg(value)"
		subQuery = fmt.Sprintf(queryTmpl, selectLabelsAny, step, op, timeSeriesSubQuery)
	case v3.TimeAggregationSum:
		op := "sum(value)"
		subQuery = fmt.Sprintf(queryTmpl, selectLabelsAny, step, op, timeSeriesSubQuery)
	case v3.TimeAggregationMin:
		op := "min(value)"
		subQuery = fmt.Sprintf(queryTmpl, selectLabelsAny, step, op, timeSeriesSubQuery)
	case v3.TimeAggregationMax:
		op := "max(value)"
		subQuery = fmt.Sprintf(queryTmpl, selectLabelsAny, step, op, timeSeriesSubQuery)
	case v3.TimeAggregationCount:
		op := "count(value)"
		subQuery = fmt.Sprintf(queryTmpl, selectLabelsAny, step, op, timeSeriesSubQuery)
	case v3.TimeAggregationCountDistinct:
		op := "count(distinct(value))"
		subQuery = fmt.Sprintf(queryTmpl, selectLabelsAny, step, op, timeSeriesSubQuery)
	case v3.TimeAggregationAnyLast:
		op := "anyLast(value)"
		subQuery = fmt.Sprintf(queryTmpl, selectLabelsAny, step, op, timeSeriesSubQuery)
	case v3.TimeAggregationRate:
		op := fmt.Sprintf("sum(value)/%d", step)
		subQuery = fmt.Sprintf(queryTmpl, selectLabelsAny, step, op, timeSeriesSubQuery)
	case v3.TimeAggregationIncrease:
		op := "sum(value)"
		subQuery = fmt.Sprintf(queryTmpl, selectLabelsAny, step, op, timeSeriesSubQuery)
	}
	return subQuery, nil
}

// See `canShortCircuit` below for details
func prepareQueryOptimized(start, end, step int64, mq *v3.BuilderQuery) (string, error) {

	groupBy := helpers.GroupingSetsByAttributeKeyTags(mq.GroupBy...)
	orderBy := helpers.OrderByAttributeKeyTags(mq.OrderBy, mq.GroupBy)
	selectLabels := helpers.SelectLabels(mq.GroupBy)

	var query string

	timeSeriesSubQuery, err := helpers.PrepareTimeseriesFilterQuery(mq)
	if err != nil {
		return "", err
	}

	samplesTableFilter := fmt.Sprintf("metric_name = %s AND timestamp_ms >= %d AND timestamp_ms <= %d", utils.ClickHouseFormattedValue(mq.AggregateAttribute.Key), start, end)

	// Select the aggregate value for interval
	queryTmpl :=
		"SELECT %s" +
			" toStartOfInterval(toDateTime(intDiv(timestamp_ms, 1000)), INTERVAL %d SECOND) as ts," +
			" %s as value" +
			" FROM " + constants.SIGNOZ_METRIC_DBNAME + "." + constants.SIGNOZ_SAMPLES_TABLENAME +
			" INNER JOIN" +
			" (%s) as filtered_time_series" +
			" USING fingerprint" +
			" WHERE " + samplesTableFilter +
			" GROUP BY %s" +
			" ORDER BY %s"

	switch mq.SpaceAggregation {
	case v3.SpaceAggregationSum:
		op := "sum(value)"
		if mq.TimeAggregation == v3.TimeAggregationRate {
			op = "sum(value)/" + fmt.Sprintf("%d", step)
		}
		query = fmt.Sprintf(queryTmpl, selectLabels, step, op, timeSeriesSubQuery, groupBy, orderBy)
	case v3.SpaceAggregationMin:
		op := "min(value)"
		query = fmt.Sprintf(queryTmpl, selectLabels, step, op, timeSeriesSubQuery, groupBy, orderBy)
	case v3.SpaceAggregationMax:
		op := "max(value)"
		query = fmt.Sprintf(queryTmpl, selectLabels, step, op, timeSeriesSubQuery, groupBy, orderBy)
	}
	return query, nil
}

// PrepareMetricQueryDeltaTimeSeries builds the query to be used for fetching metrics
func PrepareMetricQueryDeltaTimeSeries(start, end, step int64, mq *v3.BuilderQuery) (string, error) {

	if canShortCircuit(mq) {
		return prepareQueryOptimized(start, end, step, mq)
	}

	var query string

	temporalAggSubQuery, err := prepareTimeAggregationSubQuery(start, end, step, mq)
	if err != nil {
		return "", err
	}

	groupBy := helpers.GroupingSetsByAttributeKeyTags(mq.GroupBy...)
	orderBy := helpers.OrderByAttributeKeyTags(mq.OrderBy, mq.GroupBy)
	selectLabels := helpers.GroupByAttributeKeyTags(mq.GroupBy...)

	queryTmpl :=
		"SELECT %s," +
			" %s as value" +
			" FROM (%s)" +
			" WHERE isNaN(per_series_value) = 0" +
			" GROUP BY %s" +
			" ORDER BY %s"

	switch mq.SpaceAggregation {
	case v3.SpaceAggregationAvg:
		op := "avg(per_series_value)"
		query = fmt.Sprintf(queryTmpl, selectLabels, op, temporalAggSubQuery, groupBy, orderBy)
	case v3.SpaceAggregationSum:
		op := "sum(per_series_value)"
		query = fmt.Sprintf(queryTmpl, selectLabels, op, temporalAggSubQuery, groupBy, orderBy)
	case v3.SpaceAggregationMin:
		op := "min(per_series_value)"
		query = fmt.Sprintf(queryTmpl, selectLabels, op, temporalAggSubQuery, groupBy, orderBy)
	case v3.SpaceAggregationMax:
		op := "max(per_series_value)"
		query = fmt.Sprintf(queryTmpl, selectLabels, op, temporalAggSubQuery, groupBy, orderBy)
	case v3.SpaceAggregationCount:
		op := "count(per_series_value)"
		query = fmt.Sprintf(queryTmpl, selectLabels, op, temporalAggSubQuery, groupBy, orderBy)
	}

	return query, nil
}

// canShortCircuit returns true if we can use the optimized query
// for the given query
// This is used to avoid the group by fingerprint thus improving the performance
// for certain queries
// cases where we can short circuit:
// 1. time aggregation = (rate|increase) and space aggregation = sum
//   - rate = sum(value)/step, increase = sum(value) - sum of sums is same as sum of all values
//
// 2. time aggregation = sum and space aggregation = sum
//   - sum of sums is same as sum of all values
//
// 3. time aggregation = min and space aggregation = min
//   - min of mins is same as min of all values
//
// 4. time aggregation = max and space aggregation = max
//   - max of maxs is same as max of all values
//
// all of this is true only for delta metrics
func canShortCircuit(mq *v3.BuilderQuery) bool {
	if (mq.TimeAggregation == v3.TimeAggregationRate || mq.TimeAggregation == v3.TimeAggregationIncrease) && mq.SpaceAggregation == v3.SpaceAggregationSum {
		return true
	}
	if mq.TimeAggregation == v3.TimeAggregationSum && mq.SpaceAggregation == v3.SpaceAggregationSum {
		return true
	}
	if mq.TimeAggregation == v3.TimeAggregationMin && mq.SpaceAggregation == v3.SpaceAggregationMin {
		return true
	}
	if mq.TimeAggregation == v3.TimeAggregationMax && mq.SpaceAggregation == v3.SpaceAggregationMax {
		return true
	}
	return false
}
