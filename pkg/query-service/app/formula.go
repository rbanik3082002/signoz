package app

import (
	"math"
	"sort"

	"github.com/SigNoz/govaluate"
	v3 "go.signoz.io/signoz/pkg/query-service/model/v3"
)

// Define the ExpressionEvalFunc type
type ExpressionEvalFunc func(*govaluate.EvaluableExpression, map[string]float64) float64

// Helper function to check if one label set is a subset of another
func isSubset(super, sub map[string]string) bool {
	for k, v := range sub {
		if val, ok := super[k]; !ok || val != v {
			return false
		}
	}
	return true
}

// Function to find unique label sets
func findUniqueLabelSets(results []*v3.Result) []map[string]string {
	uniqueSets := make([]map[string]string, 0)

	for _, result := range results {
		for _, series := range result.Series {
			isUnique := true
			for _, uSet := range uniqueSets {
				if isSubset(series.Labels, uSet) || isSubset(uSet, series.Labels) {
					isUnique = false
					break
				}
			}
			if isUnique {
				uniqueSets = append(uniqueSets, series.Labels)
			}
		}
	}
	return uniqueSets
}

// Function to join series on timestamp and calculate new values
func joinAndCalculate(results []*v3.Result, uniqueLabelSet map[string]string, expression *govaluate.EvaluableExpression) (*v3.Series, error) {

	uniqueTimestamps := make(map[int64]struct{})
	// map[queryNmae]map[timestamp]value
	seriesMap := make(map[string]map[int64]float64)
	for _, result := range results {
		var matchingSeries *v3.Series
		for _, series := range result.Series {
			if isSubset(series.Labels, uniqueLabelSet) {
				matchingSeries = series
				break
			}
		}
		if matchingSeries != nil {
			for _, point := range matchingSeries.Points {
				if _, ok := seriesMap[result.QueryName]; !ok {
					seriesMap[result.QueryName] = make(map[int64]float64)
				}
				seriesMap[result.QueryName][point.Timestamp] = point.Value
				uniqueTimestamps[point.Timestamp] = struct{}{}
			}
		}
	}

	vars := expression.Vars()
	var doesNotHaveAllVars bool
	for _, v := range vars {
		if _, ok := seriesMap[v]; !ok {
			doesNotHaveAllVars = true
			break
		}
	}
	if doesNotHaveAllVars {
		return nil, nil
	}

	resultSeries := &v3.Series{
		Labels: uniqueLabelSet,
	}
	timestamps := make([]int64, 0)
	for timestamp := range uniqueTimestamps {
		timestamps = append(timestamps, timestamp)
	}
	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i] < timestamps[j]
	})

	for timestamp := range uniqueTimestamps {
		values := make(map[string]interface{})
		for queryName, series := range seriesMap {
			values[queryName] = series[timestamp]
		}
		newValue, err := expression.Evaluate(values)
		if err != nil {
			return nil, err
		}

		resultSeries.Points = append(resultSeries.Points, v3.Point{
			Timestamp: timestamp,
			Value:     newValue.(float64),
		})
	}
	return resultSeries, nil
}

// Main function to process the Results
func processResults(results []*v3.Result, expression *govaluate.EvaluableExpression) (*v3.Result, error) {
	uniqueLabelSets := findUniqueLabelSets(results)
	newSeries := make([]*v3.Series, 0)

	for _, labelSet := range uniqueLabelSets {
		series, err := joinAndCalculate(results, labelSet, expression)
		if err != nil {
			return nil, err
		}
		if series != nil {
			newSeries = append(newSeries, series)
		}
	}

	return &v3.Result{
		Series: newSeries,
	}, nil
}

var SupportedFunctions = []string{"exp", "log", "ln", "exp2", "log2", "exp10", "log10", "sqrt", "cbrt", "erf", "erfc", "lgamma", "tgamma", "sin", "cos", "tan", "asin", "acos", "atan", "degrees", "radians"}

func evalFuncs() map[string]govaluate.ExpressionFunction {
	GoValuateFuncs := make(map[string]govaluate.ExpressionFunction)
	// Returns e to the power of the given argument.
	GoValuateFuncs["exp"] = func(args ...interface{}) (interface{}, error) {
		return math.Exp(args[0].(float64)), nil
	}
	// Returns the natural logarithm of the given argument.
	GoValuateFuncs["log"] = func(args ...interface{}) (interface{}, error) {
		return math.Log(args[0].(float64)), nil
	}
	// Returns the natural logarithm of the given argument.
	GoValuateFuncs["ln"] = func(args ...interface{}) (interface{}, error) {
		return math.Log(args[0].(float64)), nil
	}
	// Returns the base 2 exponential of the given argument.
	GoValuateFuncs["exp2"] = func(args ...interface{}) (interface{}, error) {
		return math.Exp2(args[0].(float64)), nil
	}
	// Returns the base 2 logarithm of the given argument.
	GoValuateFuncs["log2"] = func(args ...interface{}) (interface{}, error) {
		return math.Log2(args[0].(float64)), nil
	}
	// Returns the base 10 exponential of the given argument.
	GoValuateFuncs["exp10"] = func(args ...interface{}) (interface{}, error) {
		return math.Pow10(int(args[0].(float64))), nil
	}
	// Returns the base 10 logarithm of the given argument.
	GoValuateFuncs["log10"] = func(args ...interface{}) (interface{}, error) {
		return math.Log10(args[0].(float64)), nil
	}
	// Returns the square root of the given argument.
	GoValuateFuncs["sqrt"] = func(args ...interface{}) (interface{}, error) {
		return math.Sqrt(args[0].(float64)), nil
	}
	// Returns the cube root of the given argument.
	GoValuateFuncs["cbrt"] = func(args ...interface{}) (interface{}, error) {
		return math.Cbrt(args[0].(float64)), nil
	}
	// Returns the error function of the given argument.
	GoValuateFuncs["erf"] = func(args ...interface{}) (interface{}, error) {
		return math.Erf(args[0].(float64)), nil
	}
	// Returns the complementary error function of the given argument.
	GoValuateFuncs["erfc"] = func(args ...interface{}) (interface{}, error) {
		return math.Erfc(args[0].(float64)), nil
	}
	// Returns the natural logarithm of the absolute value of the gamma function of the given argument.
	GoValuateFuncs["lgamma"] = func(args ...interface{}) (interface{}, error) {
		v, _ := math.Lgamma(args[0].(float64))
		return v, nil
	}
	// Returns the gamma function of the given argument.
	GoValuateFuncs["tgamma"] = func(args ...interface{}) (interface{}, error) {
		return math.Gamma(args[0].(float64)), nil
	}
	// Returns the sine of the given argument.
	GoValuateFuncs["sin"] = func(args ...interface{}) (interface{}, error) {
		return math.Sin(args[0].(float64)), nil
	}
	// Returns the cosine of the given argument.
	GoValuateFuncs["cos"] = func(args ...interface{}) (interface{}, error) {
		return math.Cos(args[0].(float64)), nil
	}
	// Returns the tangent of the given argument.
	GoValuateFuncs["tan"] = func(args ...interface{}) (interface{}, error) {
		return math.Tan(args[0].(float64)), nil
	}
	// Returns the arcsine of the given argument.
	GoValuateFuncs["asin"] = func(args ...interface{}) (interface{}, error) {
		return math.Asin(args[0].(float64)), nil
	}
	// Returns the arccosine of the given argument.
	GoValuateFuncs["acos"] = func(args ...interface{}) (interface{}, error) {
		return math.Acos(args[0].(float64)), nil
	}
	// Returns the arctangent of the given argument.
	GoValuateFuncs["atan"] = func(args ...interface{}) (interface{}, error) {
		return math.Atan(args[0].(float64)), nil
	}
	// Returns the argument converted from radians to degrees.
	GoValuateFuncs["degrees"] = func(args ...interface{}) (interface{}, error) {
		return args[0].(float64) * 180 / math.Pi, nil
	}
	// Returns the argument converted from degrees to radians.
	GoValuateFuncs["radians"] = func(args ...interface{}) (interface{}, error) {
		return args[0].(float64) * math.Pi / 180, nil
	}
	return GoValuateFuncs
}