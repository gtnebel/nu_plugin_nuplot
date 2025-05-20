package commands

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"

	"github.com/ainvaltin/nu-plugin"
	"github.com/ainvaltin/nu-plugin/types"

	"github.com/gtnebel/nu_plugin_nuplot/commands/flags"
)

type KlineDataList = []opts.KlineData

type KlineDataSeries = map[string]KlineDataList

func NuplotKline() *nu.Command {
	return &nu.Command{
		Signature: nu.PluginSignature{
			Name:        "nuplot kline",
			Category:    "Chart",
			Desc:        "Plots a kline chart",
			Description: "Title, size and color theme can be configured by flags. Each column that contains numbers will be plottet. The X axis can be set by means of the --xaxis flag.",
			SearchTerms: []string{"plot", "graph", "kline"},
			// OptionalPositional: nu.PositionalArgs{},
			Named: nu.Flags{
				flags.XAxis,
				flags.Title,
				flags.SubTitle,
				flags.Width,
				flags.Height,
				flags.ColorTheme,
				flags.Fitted,
				flags.Verbose,
			},
			InputOutputTypes: []nu.InOutTypes{
				{
					In: types.Table(types.RecordDef{
						"First": types.Number(),
						"Last":  types.Number(),
						"Min":   types.Number(),
						"Max":   types.Number(),
					}),
					Out: types.Nothing(),
				},
				// Array List of 4-element Lists with [first, last, min, max] values.
				{In: types.List(types.List(types.Number())), Out: types.Nothing()},
			},
			AllowMissingExamples: true,
		},
		Examples: nu.Examples{
			{
				Description: `Plot a kline graph of an array of array of numbers.`,
				Example:     `[[1 4 0 6] [5 3 1 5] [2 7 2 7]] | nuplot kline`,
			},
			{
				Description: `Plot a kline graph of a table with addidional date column.`,
				Example:     `[[Date First Last Min Max]; ['2024-06-01' 1.53 1.64 1.50 1.66] ['2024-06-02' 1.63 1.73 1.61 1.75] ['2024-06-03' 1.72 1.57 1.52 1.77]] | nuplot kline --xaxis Date --fitted --title 'Fuel prices in June 24'`,
			},
			{
				Description: `Load data from a file and prepare it for the kline chart.`,
				Example:     `open temperatures-2023.csv | chunk-by {|l| $l.date | into datetime | format date '%Y-%m-%d' } | each {|d| {Date: $d.0.date First: $d.0.value Last: ($d | last | get value) Min: ($d | get value | math min) Max: ($d | get value | math max)}} | nuplot kline --title "Temperatures 2023" --xaxis Date`,
			},
		},
		OnRun: nuplotKlineHandler,
	}
}

func nuplotKlineHandler(ctx context.Context, call *nu.ExecCommand) error {
	checkVerboseFlag(call)
	return handleCommandInput(call, plotKline)
}

func ValueToFloat64(value nu.Value) (float64, error) {
	switch v := value.Value.(type) {
	case int64:
		return float64(v), nil
	case float64:
		return v, nil
	default:
		return 0, fmt.Errorf("incompatible input type for ValueToFloat64(): %T", v)
	}
}

func convertValueArray(arr []nu.Value) ([]float64, error) {
	res := make([]float64, 0)

	for _, item := range arr {
		v, err := ValueToFloat64(item)
		if err == nil {
			res = append(res, v)
		} else {
			return []float64{}, fmt.Errorf("Input array contains invalid values: %s", err.Error())
		}
	}

	return res, nil
}

func plotKline(input any, call *nu.ExecCommand) error {
	series := make(KlineDataSeries)

	xAxisName := getStringFlag(call, "xaxis", XAxisSeries)
	slog.Debug("plotKline", "xAxisName", xAxisName)

	switch inputValue := input.(type) {
	case []nu.Value:
		for _, item := range inputValue {
			switch itemValue := item.Value.(type) {
			case nu.Record:
				var first float64 = 0
				var last float64 = 0
				var min float64 = 0
				var max float64 = 0

				var ok1 error = nil
				var ok2 error = nil
				var ok3 error = nil
				var ok4 error = nil

				for k, v := range itemValue {
					if k == xAxisName {
						continue
					}

					if k == "First" {
						first, ok1 = ValueToFloat64(v)
					}
					if k == "Last" {
						last, ok2 = ValueToFloat64(v)
					}
					if k == "Min" {
						min, ok3 = ValueToFloat64(v)
					}
					if k == "Max" {
						max, ok4 = ValueToFloat64(v)
					}
				}

				if ok1 == nil && ok2 == nil && ok3 == nil && ok4 == nil {
					items := getSeries(series, DefaultSeries)
					series[DefaultSeries] = append(
						items,
						opts.KlineData{Value: [4]float64{first, last, min, max}},
					)
				} else {
					return fmt.Errorf("plotKline: Some values in the First, Last Min, Max columns contain invalid values.")
				}

				// If a xaxis is defined, fill the series with the values.
				if xAxisName != XAxisSeries {
					if v, ok := itemValue[xAxisName]; ok {
						items := getSeries(series, xAxisName)
						series[xAxisName] = append(items, opts.KlineData{Value: matchXValue(v)})
					} else {
						// If the column specified in --xaxis does not exist, we
						// set the `xAxisName` variable to XAxisSeries, so that a
						// simple int range is generated as x axis.
						xAxisName = XAxisSeries
					}
				}
			case []nu.Value:
				if len(itemValue) == 4 {
					lst, err := convertValueArray(itemValue)
					if err == nil {
						items := getSeries(series, DefaultSeries)
						series[DefaultSeries] = append(items, opts.KlineData{Value: lst})
					} else {
						return err
					}
				} else {
					return fmt.Errorf(
						"plotKline: Sub lists in a <list<list<number>>> input have to have length of 4 elements.",
					)
				}
			default:
				return fmt.Errorf("plotKline: unsupported input value type: %T", inputValue)
			}
		}
	default:
		return fmt.Errorf("plotKline: unsupported input value type: %T", inputValue)
	}

	// create a new line instance
	line := charts.NewKLine()

	line.SetGlobalOptions(buildGlobalChartOptions(call)...)

	// Reverse X/Y (only on bar charts)
	// line.XYReversal()

	// Put data into instance
	itemCount := 0
	for sName, sValues := range series {
		if sName == xAxisName {
			continue
		}

		itemCount = len(sValues)
		slog.Debug("plotKline: Adding items to series", "series", sName, "items", itemCount)
		line = line.AddSeries(sName, sValues)
	}

	if xAxisName != XAxisSeries {
		line = line.SetXAxis(series[xAxisName])
	} else {
		xRange := make([]int, itemCount)
		for i := range itemCount {
			xRange[i] = i
		}

		line = line.SetXAxis(xRange)
	}

	renderChart(func(f *os.File) error { return line.Render(f) })

	return nil
}
