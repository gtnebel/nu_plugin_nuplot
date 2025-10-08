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

// A List of line chart data points
type LineDataList = []opts.LineData

// Line data series mapping
type LineDataSeries = map[string]LineDataList

// This function initializes the nuplot line command.
func NuplotLine() *nu.Command {
	return &nu.Command{
		Signature: nu.PluginSignature{
			Name:        "nuplot line",
			Category:    "Chart",
			Desc:        "Plots a line chart",
			Description: "Title, size and color theme can be configured by flags. Each column that contains numbers will be plottet. The X axis can be set by means of the --xaxis flag.",
			SearchTerms: []string{"plot", "graph", "line"},
			// OptionalPositional: nu.PositionalArgs{},
			Named: []nu.Flag{
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
				{In: types.Table(types.RecordDef{}), Out: types.Nothing()},
				// {In: types.List(types.Table(types.RecordDef{})), Out: types.Nothing()},
				{In: types.List(types.Number()), Out: types.Nothing()},
			},
			AllowMissingExamples: true,
		},
		Examples: []nu.Example{
			{
				Description: `Plot a line graph of an array of numbers.`,
				Example:     `[5, 4, 3, 2, 5, 7, 8] | nuplot line`,
				// Result:      &nu.Value{Value: []nu.Value{{Value: 10}, {Value: "foo"}}},
			},
		},
		OnRun: nuplotLineHandler,
	}
}

func nuplotLineHandler(ctx context.Context, call *nu.ExecCommand) error {
	checkVerboseFlag(call)
	return handleCommandInput(call, plotLine)
}

func plotLine(input any, call *nu.ExecCommand) error {
	series := make(LineDataSeries)

	xAxisName := getCellPathFlag(call, "xaxis", XAxisSeries)
	slog.Debug("plotLine", "xAxisName", xAxisName)

	switch inputValue := input.(type) {
	case []nu.Value:
		for itemIndex, item := range inputValue {
			switch itemValue := item.Value.(type) {
			case int64:
				items := getSeries(series, DefaultSeries)
				series[DefaultSeries] = append(items, opts.LineData{Value: itemValue})
			case float64:
				items := getSeries(series, DefaultSeries)
				series[DefaultSeries] = append(items, opts.LineData{Value: itemValue})
			case nu.Record:
				// Try to set xAxisName to one of the columns in the record.
				if itemIndex == 0 {
					xAxisName = autoSetXaxis(itemValue, xAxisName)
				}

				for k, v := range itemValue {
					if k == xAxisName {
						continue
					}

					_, ok1 := v.Value.(int64)
					_, ok2 := v.Value.(float64)
					if ok1 || ok2 {
						items := getSeries(series, k)
						series[k] = append(items, opts.LineData{Value: v.Value})
					}
				}

				// If a xaxis is defined, fill the series with the values.
				if xAxisName != XAxisSeries {
					if v, ok := itemValue[xAxisName]; ok {
						items := getSeries(series, xAxisName)
						series[xAxisName] = append(items, opts.LineData{Value: matchXValue(v)})
					} else {
						slog.Warn("Specified x-axis is not continuous. Reseting x-axis to default value.")
						// If the column specified in --xaxis does not exist, we
						// set the `xAxisName` variable to XAxisSeries, so that a
						// simple int range is generated as x axis.
						xAxisName = XAxisSeries
					}
				}
			default:
				return fmt.Errorf("plotLine: unsupported input value type: %T", inputValue)
			}
		}
	default:
		return fmt.Errorf("plotLine: unsupported input value type: %T", inputValue)
	}

	// create a new line instance
	line := charts.NewLine()

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
		slog.Debug("plotLine: Adding items to series", "series", sName, "items", itemCount)
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

	setPageTitle(call, &line.BaseConfiguration)

	return renderChart(func(f *os.File) error { return line.Render(f) })
}
