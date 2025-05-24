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

type BoxPlotSeriesHelper = map[string][][]float64

type Float64Series = map[string][]float64

func NuplotBoxPlot() *nu.Command {
	return &nu.Command{
		Signature: nu.PluginSignature{
			Name:        "nuplot boxplot",
			Category:    "Chart",
			Desc:        "Plots a boxplot chart",
			Description: "Title, size and color theme can be configured by flags. Each column that contains numbers will be plottet. The X axis can be set by means of the --xaxis flag.",
			SearchTerms: []string{"plot", "graph", "boxplot"},
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
				{In: types.List(types.Number()), Out: types.Nothing()},
				{In: types.List(types.Table(types.RecordDef{})), Out: types.Nothing()},
				{In: types.List(types.List(types.Number())), Out: types.Nothing()},
			},
			AllowMissingExamples: true,
		},
		Examples: []nu.Example{
			{
				Description: `Simple boxplot from an array of numbers.`,
				Example:     `[5, 4, 3, 2, 5, 7, 8] | nuplot boxplot`,
			},
			{
				Description: `Boxplot from a table with two columns.`,
				Example:     `[[value1 value2]; [2 3] [3 5] [6 8] [1 8]] | nuplot boxplot`,
			},
			{
				Description: `Make a boxplot of monthly values.`,
				Example:     `open Temperatures.csv | upsert date {|l| $l.date | format date "%B"} | chunk-by {$in.date} | nuplot boxplot --xaxis date`,
			},
		},
		OnRun: nuplotBoxPlotHandler,
	}
}

func nuplotBoxPlotHandler(ctx context.Context, call *nu.ExecCommand) error {
	checkVerboseFlag(call)
	return handleCommandInput(call, plotBoxPlot)
}

func createBoxPlotDataValue(data []float64) ([]float64, error) {
	if len(data) == 0 {
		return []float64{0, 0, 0, 0, 0},
			fmt.Errorf("createBoxPlotDataValue: zero input data")
	}

	min, _ := stats.Min(data)
	max, _ := stats.Max(data)
	q, _ := stats.Quartile(data)

	return []float64{min, q.Q1, q.Q2, q.Q3, max}, nil
}

func boxplotReadInputListItem(listItem []nu.Value, seriesHelper BoxPlotSeriesHelper, xAxisName string) (xValue any, res error) {
	series := make(Float64Series)

	for _, item := range listItem {
		switch itemValue := item.Value.(type) {
		case int64:
			// items := getSeries(seriesHelper, DefaultSeries)
			items := getSeries(series, DefaultSeries)
			series[DefaultSeries] = append(items, float64(itemValue))
		case float64:
			// items := getSeries(seriesHelper, DefaultSeries)
			items := getSeries(series, DefaultSeries)
			series[DefaultSeries] = append(items, itemValue)
		case nu.Record:
			for k, v := range itemValue {
				if k == xAxisName {
					continue
				}

				items := getSeries(series, k)
				if float64Val, err := ValueToFloat64(v); err == nil {
					series[k] = append(items, float64Val)
				}
			}
			// If a xaxis is defined, fill the series with the values.
			if xAxisName != XAxisSeries {
				if v, ok := itemValue[xAxisName]; ok {
					// Only the first Item in the sub-table is as x value for
					// this sub-table
					if xValue == nil {
						xValue = matchXValue(v)
					}
				} else {
					// If the column specified in --xaxis does not exist, we
					// set the `xAxisName` variable to XAxisSeries, so that a
					// simple int range is generated as x axis.
					xAxisName = XAxisSeries
				}
			}
		case []nu.Value:
			// The input value is a list of tables or list of lists
			xv, err := boxplotReadInputListItem(itemValue, seriesHelper, xAxisName)
			if err == nil {
				if xValue == nil {
					xValue = make([]any, 0)
				}

				switch items := xValue.(type) {
				case []any:
					xValue = append(items, xv)
				}
			} else {
				res = err
				return
			}
		default:
			res = fmt.Errorf("boxplotReadInputListItem: unsupported input value type: %T", listItem)
			return
		}
	}

	for k, v := range series {
		items := getSeries(seriesHelper, k)
		seriesHelper[k] = append(items, v)
	}

	return
}

func plotBoxPlot(input any, call *nu.ExecCommand) error {
	seriesHelper := make(BoxPlotSeriesHelper)
	var xSeries []any = nil

	xAxisName := getCellPathFlag(call, "xaxis", XAxisSeries)
	slog.Debug("plotBoxPlot", "xAxisName", xAxisName)

	switch inputValue := input.(type) {
	case []nu.Value:
		xValue, err := boxplotReadInputListItem(inputValue, seriesHelper, xAxisName)
		if err == nil {
			switch items := xValue.(type) {
			case []any:
				xSeries = items
			}
		} else {
			return err
		}
	default:
		return fmt.Errorf("plotBoxPlot: unsupported input value type: %T", inputValue)
	}

	// create a new boxplot instance
	boxplot := charts.NewBoxPlot()

	boxplot.SetGlobalOptions(buildGlobalChartOptions(call)...)

	// Put data into instance
	itemCount := 0
	for sName, sValues := range seriesHelper {
		if sName == xAxisName {
			continue
		}

		// Check, if series is completely empty. In this case, no box plot
		// data is created for it.
		empty := true
		for _, sVal := range sValues {
			if len(sVal) > 0 {
				empty = false
				break
			}
		}
		if empty {
			slog.Debug("plotBoxPlot: Skipping empty series", "series", sName)
			continue
		}

		itemCount = len(sValues)
		slog.Debug("plotBoxPlot: Adding items to series", "series", sName, "items", itemCount)

		data := make(BoxPlotDataList, 0)
		for _, sVal := range sValues {
			bpValues, err := createBoxPlotDataValue(sVal)
			if err == nil {
				data = append(data, opts.BoxPlotData{Value: bpValues})
			} else {
				// slog.Debug(err.Error())
				return err
			}
		}
		boxplot = boxplot.AddSeries(sName, data)
	}

	if xAxisName != XAxisSeries {
		slog.Debug("Setting x axis to", "series", xAxisName)
		boxplot = boxplot.SetXAxis(xSeries)
	} else {
		xRange := make([]int, itemCount)
		for i := range itemCount {
			xRange[i] = i
		}

		boxplot = boxplot.SetXAxis(xRange)
	}

	return renderChart(func(f *os.File) error { return boxplot.Render(f) })
}
