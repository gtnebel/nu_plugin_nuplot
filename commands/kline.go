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
			Desc:        "Plots the data that is piped into the command as `echarts` graph.",
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
				Description: `Plot a line graph of an array of numbers.`,
				Example:     `[5, 4, 3, 2, 5, 7, 8] | nuplot line`,
				// Result:      &nu.Value{Value: []nu.Value{{Value: 10}, {Value: "foo"}}},
			},
		},
		OnRun: nuplotKlineHandler,
	}
}

func nuplotKlineHandler(ctx context.Context, call *nu.ExecCommand) error {
	return handleCommandInput(call, plotKline)
}

func ValueToFloat64(value nu.Value) float64 error {
	
}

func plotKline(input any, call *nu.ExecCommand) error {
	series := make(KlineDataSeries)

	xAxisName := getStringFlag(call, "xaxis", XAxisSeries)
	slog.Debug("plotKline", "xAxisName", xAxisName)

	switch inputValue := input.(type) {
	case []nu.Value:
		for _, item := range inputValue {
			switch itemValue := item.Value.(type) {
			case int64:
				items := getSeries(series, DefaultSeries)
				series[DefaultSeries] = append(items, opts.KlineData{Value: itemValue})
			case float64:
				items := getSeries(series, DefaultSeries)
				series[DefaultSeries] = append(items, opts.KlineData{Value: itemValue})
			case nu.Record:
				for k, v := range itemValue {
					if k == xAxisName {
						continue
					}

					_, ok1 := v.Value.(int64)
					_, ok2 := v.Value.(float64)
					if ok1 || ok2 {
						items := getSeries(series, k)
						series[k] = append(items, opts.KlineData{Value: v.Value})
					}
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
					items := getSeries(series, DefaultSeries)
					series[DefaultSeries] = append(items, opts.KlineData{Value: [4]float64{
						
					}})
				} else {
					return fmt.Errorf(
						"Sub lists in a <list<list<number>>> input have to have length of 4 elements.",
					)
				}
			default:
				return fmt.Errorf("unsupported input value type: %T", inputValue)
			}
		}
	default:
		return fmt.Errorf("unsupported input value type: %T", inputValue)
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

	line.SetSeriesOptions(
		charts.WithLineChartOpts(opts.LineChart{
			Smooth: opts.Bool(true),
		}),
		// For bar charts
		// charts.WithBarChartOpts(opts.BarChart{
		// 	Stack: "stackA",
		// }),
	)

	renderChart(func(f *os.File) error { return line.Render(f) })

	return nil
}
