package commands

import (
	"context"
	"fmt"
	// "io"
	"log"
	"os"
	// "slices"
	// "strconv"
	// "time"

	"github.com/pkg/browser"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	// charttypes "github.com/go-echarts/go-echarts/v2/types"

	"github.com/ainvaltin/nu-plugin"
	// "github.com/ainvaltin/nu-plugin/syntaxshape"
	"github.com/ainvaltin/nu-plugin/types"

	"github.com/gtnebel/nu_plugin_nuplot/commands/flags"
)

type LineData = []opts.LineData

type LineDataSeries = map[string]LineData

func getSeries(series LineDataSeries, name string) LineData {
	s, ok := series[name]

	if ok {
		return s
	} else {
		series[name] = make(LineData, 0)
		return series[name]
	}
}

func NuplotLine() *nu.Command {
	return &nu.Command{
		Signature: nu.PluginSignature{
			Name:        "nuplot line",
			Category:    "Chart",
			Desc:        `Plots the data that is piped into the command as 'echarts' graph.`,
			Description: "TODO: Long description....",
			SearchTerms: []string{"plot", "graph", "line", "bar"},
			// OptionalPositional: nu.PositionalArgs{},
			Named: nu.Flags{
				flags.Title,
				flags.SubTitle,
				flags.XAxis,
				flags.ColorTheme,
				flags.Width,
				flags.Height,
			},
			InputOutputTypes: []nu.InOutTypes{
				{In: types.List(types.Table(types.RecordDef{})), Out: types.Nothing()},
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
		OnRun: nuplotLineHandler,
	}
}

func nuplotLineHandler(ctx context.Context, call *nu.ExecCommand) error {
	return handleCommandInput(call, plotLine)
}

func plotLine(input any, call *nu.ExecCommand) error {
	series := make(LineDataSeries)

	xAxis, xAxisOk := call.FlagValue("xaxis")
	xAxisName := xAxis.Value.(string)
	log.Println("plotLine:", "xAxis:", xAxisOk, xAxis)

	switch inputValue := input.(type) {
	case []nu.Value:
		for _, item := range inputValue {
			switch itemValue := item.Value.(type) {
			case int64:
				items := getSeries(series, DefaultSeries)
				series[DefaultSeries] = append(items, opts.LineData{Value: itemValue})
			case float64:
				items := getSeries(series, DefaultSeries)
				series[DefaultSeries] = append(items, opts.LineData{Value: itemValue})
			case nu.Record:
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
				if xAxisOk {
					if v, ok := itemValue[xAxisName]; ok {
						items := getSeries(series, XAxisSeries)
						series[XAxisSeries] = append(items, opts.LineData{Value: matchXValue(v)})
					} else {
						// If the column specified in --xaxis does not exist, we
						// set the `xAxisOk` variable to false, so that a
						// simple int range is generated as x axis.
						xAxisOk = false
					}
				}
			default:
				return fmt.Errorf("unsupported input value type: %T", inputValue)
			}
		}
	default:
		return fmt.Errorf("unsupported input value type: %T", inputValue)
	}

	// create a new line instance
	line := charts.NewLine()

	line.SetGlobalOptions(buildGlobalChartOptions(call)...)

	// Reverse X/Y (only on bar charts)
	// line.XYReversal()

	// Put data into instance
	itemCount := 0
	for sName, sValues := range series {
		if sName == XAxisSeries {
			continue
		}

		itemCount = len(sValues)
		log.Println("plotLine:", "Adding", itemCount, "items to series", sName)
		line = line.AddSeries(sName, sValues)
	}

	if xAxisOk {
		line = line.SetXAxis(series[XAxisSeries])
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

	chartFile, _ := os.CreateTemp("", "chart-*.html")
	chartFileName := chartFile.Name()
	log.Println("plotLine:", "Rendering output to", chartFileName)
	line.Render(chartFile)
	chartFile.Close()

	browser.OpenFile(chartFileName)

	return nil
}
