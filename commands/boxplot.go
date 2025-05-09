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

	"github.com/montanaflynn/stats"

	"github.com/gtnebel/nu_plugin_nuplot/commands/flags"
)

type BoxPlotDataList = []opts.BoxPlotData

type BoxPlotDataSeries = map[string]BoxPlotDataList

func NuplotBoxPlot() *nu.Command {
	return &nu.Command{
		Signature: nu.PluginSignature{
			Name:        "nuplot boxplot",
			Category:    "Chart",
			Desc:        "Plots the data that is piped into the command as `echarts` graph.",
			Description: "Title, size and color theme can be configured by flags. Each column that contains numbers will be plottet. The X axis can be set by means of the --xaxis flag.",
			SearchTerms: []string{"plot", "graph", "boxplot"},
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
				{In: types.Table(types.RecordDef{}), Out: types.Nothing()},
				// {In: types.List(types.Table(types.RecordDef{})), Out: types.Nothing()},
				{In: types.List(types.Number()), Out: types.Nothing()},
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
		OnRun: nuplotBoxPlotHandler,
	}
}

func nuplotBoxPlotHandler(ctx context.Context, call *nu.ExecCommand) error {
	return handleCommandInput(call, plotBoxPlot)
}

func createBoxPlotDataValue(data []float64) []float64 {
	min, _ := stats.Min(data)
	max, _ := stats.Max(data)
	q, _ := stats.Quartile(data)

	return []float64{
		min,
		q.Q1,
		q.Q2,
		q.Q3,
		max,
	}
}

func plotBoxPlot(input any, call *nu.ExecCommand) error {
	series := make(BoxPlotDataSeries)

	xAxisName := getStringFlag(call, "xaxis", XAxisSeries)
	slog.Debug("plotBoxPlot", "xAxisName", xAxisName)

	switch inputValue := input.(type) {
	case []nu.Value:
		for _, item := range inputValue {
			switch itemValue := item.Value.(type) {
			case int64:
				items := getSeries(series, DefaultSeries)
				series[DefaultSeries] = append(items, opts.BoxPlotData{Value: itemValue})
			case float64:
				items := getSeries(series, DefaultSeries)
				series[DefaultSeries] = append(items, opts.BoxPlotData{Value: itemValue})
			case nu.Record:
				for k, v := range itemValue {
					if k == xAxisName {
						continue
					}

					_, ok1 := v.Value.(int64)
					_, ok2 := v.Value.(float64)
					if ok1 || ok2 {
						items := getSeries(series, k)
						series[k] = append(items, opts.BoxPlotData{Value: v.Value})
					}
				}

				// If a xaxis is defined, fill the series with the values.
				if xAxisName != XAxisSeries {
					if v, ok := itemValue[xAxisName]; ok {
						items := getSeries(series, xAxisName)
						series[xAxisName] = append(items, opts.BoxPlotData{Value: matchXValue(v)})
					} else {
						// If the column specified in --xaxis does not exist, we
						// set the `xAxisName` variable to XAxisSeries, so that a
						// simple int range is generated as x axis.
						xAxisName = XAxisSeries
					}
				}
			default:
				return fmt.Errorf("unsupported input value type: %T", inputValue)
			}
		}
	default:
		return fmt.Errorf("unsupported input value type: %T", inputValue)
	}

	// create a new boxplot instance
	boxplot := charts.NewBoxPlot()

	boxplot.SetGlobalOptions(buildGlobalChartOptions(call)...)

	// Reverse X/Y (only on bar charts)
	// line.XYReversal()

	// Put data into instance
	itemCount := 0
	for sName, sValues := range series {
		if sName == xAxisName {
			continue
		}

		itemCount = len(sValues)
		slog.Debug("plotBoxPlot: Adding items to series", "series", sName, "items", itemCount)
		boxplot = boxplot.AddSeries(sName, sValues)
	}

	if xAxisName != XAxisSeries {
		boxplot = boxplot.SetXAxis(series[xAxisName])
	} else {
		xRange := make([]int, itemCount)
		for i := range itemCount {
			xRange[i] = i
		}

		boxplot = boxplot.SetXAxis(xRange)
	}

	boxplot.SetSeriesOptions(
		charts.WithLineChartOpts(opts.LineChart{
			Smooth: opts.Bool(true),
		}),
		// For bar charts
		// charts.WithBarChartOpts(opts.BarChart{
		// 	Stack: "stackA",
		// }),
	)

	renderChart(func(f *os.File) error { return boxplot.Render(f) })

	return nil
}
