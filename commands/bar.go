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

// A list of bar chart data points
type BarDataList = []opts.BarData

// This type maps series names to their values
type BarDataSeries = map[string]BarDataList

// This function initializes the nuplot bar command.
func NuplotBar() *nu.Command {
	return &nu.Command{
		Signature: nu.PluginSignature{
			Name:        "nuplot bar",
			Category:    "Chart",
			Desc:        "Plots a bar chart or stacked bar chart.",
			Description: "Title, size and color theme can be configured by flags. Each column that contains numbers will be plottet. The X axis can be set by means of the --xaxis flag.",
			SearchTerms: []string{"plot", "graph", "bar"},
			// OptionalPositional: nu.PositionalArgs{},
			Named: []nu.Flag{
				flags.XAxis,
				flags.XYReverse,
				flags.Stacked,
				flags.Title,
				flags.SubTitle,
				flags.Width,
				flags.Height,
				flags.ColorTheme,
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
				Description: `Plot a bar graph of an array of numbers.`,
				Example:     `[5, 4, 3, 2, 5, 7, 8] | nuplot bar`,
				// Result:      &nu.Value{Value: []nu.Value{{Value: 10}, {Value: "foo"}}},
			},
		},
		OnRun: nuplotBarHandler,
	}
}

func nuplotBarHandler(ctx context.Context, call *nu.ExecCommand) error {
	checkVerboseFlag(call)
	return handleCommandInput(call, plotBar)
}

func plotBar(input any, call *nu.ExecCommand) error {
	series := make(BarDataSeries)

	xAxisName := getCellPathFlag(call, "xaxis", XAxisSeries)
	slog.Debug("plotBar", "xAxisName", xAxisName)

	switch inputValue := input.(type) {
	case []nu.Value:
		for _, item := range inputValue {
			switch itemValue := item.Value.(type) {
			case int64:
				items := getSeries(series, DefaultSeries)
				series[DefaultSeries] = append(items, opts.BarData{Value: itemValue})
			case float64:
				items := getSeries(series, DefaultSeries)
				series[DefaultSeries] = append(items, opts.BarData{Value: itemValue})
			case nu.Record:
				for k, v := range itemValue {
					if k == xAxisName {
						continue
					}

					_, ok1 := v.Value.(int64)
					_, ok2 := v.Value.(float64)
					if ok1 || ok2 {
						items := getSeries(series, k)
						series[k] = append(items, opts.BarData{Value: v.Value})
					}
				}

				// If a xaxis is defined, fill the series with the values.
				if xAxisName != XAxisSeries {
					if v, ok := itemValue[xAxisName]; ok {
						items := getSeries(series, xAxisName)
						series[xAxisName] = append(items, opts.BarData{Value: matchXValue(v)})
					} else {
						// If the column specified in --xaxis does not exist, we
						// set the `xAxisName` variable to XAxisSeries, so that a
						// simple int range is generated as x axis.
						xAxisName = XAxisSeries
					}
				}
			default:
				return fmt.Errorf("plotBar: unsupported input value type: %T", inputValue)
			}
		}
	default:
		return fmt.Errorf("plotBar: unsupported input value type: %T", inputValue)
	}

	// create a new bar instance
	bar := charts.NewBar()

	bar.SetGlobalOptions(buildGlobalChartOptions(call)...)

	// Reverse X/Y (only on bar charts)
	if getBoolFlag(call, flags.XYReverse.Long) {
		bar.XYReversal()
	}

	// Put data into instance
	itemCount := 0
	for sName, sValues := range series {
		if sName == xAxisName {
			continue
		}

		itemCount = len(sValues)
		slog.Debug("plotBar: Adding items to series", "series", sName, "items", itemCount)
		bar = bar.AddSeries(sName, sValues)
	}

	if xAxisName != XAxisSeries {
		bar = bar.SetXAxis(series[xAxisName])
	} else {
		xRange := make([]int, itemCount)
		for i := range itemCount {
			xRange[i] = i
		}

		bar = bar.SetXAxis(xRange)
	}

	if getBoolFlag(call, flags.Stacked.Long) {
		bar.SetSeriesOptions(
			charts.WithBarChartOpts(opts.BarChart{
				Stack: "stackA",
			}),
		)
	}

	return renderChart(func(f *os.File) error { return bar.Render(f) })
}
