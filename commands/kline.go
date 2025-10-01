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

// A list of kline chart data points
type KlineDataList = []opts.KlineData

// Kline data series mapping
type KlineDataSeries = map[string]KlineDataList

// This function initializes the nuplot kline command.
func NuplotKline() *nu.Command {
	return &nu.Command{
		Signature: nu.PluginSignature{
			Name:        "nuplot kline",
			Category:    "Chart",
			Desc:        "Plots a kline chart",
			Description: "Title, size and color theme can be configured by flags. Each column that contains numbers will be plottet. The X axis can be set by means of the --xaxis flag.",
			SearchTerms: []string{"plot", "graph", "kline"},
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
				{In: types.List(types.Table(types.RecordDef{})), Out: types.Nothing()},
				{In: types.List(types.List(types.Number())), Out: types.Nothing()},
			},
			AllowMissingExamples: true,
		},
		Examples: []nu.Example{
			{
				Description: `Plot a kline graph of an array of array of numbers.`,
				Example:     `[[1 4 0 6] [5 3 1 4] [2 7 2 7]] | nuplot kline`,
			},
			// {
			// 	Description: `Plot a kline graph of a table with addidional date column.`,
			// 	Example:     `[[Date First Last Min Max]; ['2024-06-01' 1.53 1.64 1.50 1.66] ['2024-06-02' 1.63 1.73 1.61 1.75] ['2024-06-03' 1.72 1.57 1.52 1.77]] | nuplot kline --xaxis Date --fitted --title 'Fuel prices in June 24'`,
			// },
			{
				Description: `Load data from a file and display kline chart.`,
				Example:     `http get https://bulk.meteostat.net/v2/hourly/2023/10577.csv.gz | gunzip | from csv --noheaders | select column0 column2 | rename date temp | upsert date {|l| $l.date | into datetime | format date "%B"} | chunk-by {$in.date} | nuplot kline --xaxis date`,
			},
		},
		OnRun: nuplotKlineHandler,
	}
}

func nuplotKlineHandler(ctx context.Context, call *nu.ExecCommand) error {
	checkVerboseFlag(call)
	return handleCommandInput(call, plotKline)
}

func buildKlineDataValue(data []float64) opts.KlineData {
	if len(data) == 0 {
		return opts.KlineData{Value: [4]float64{0, 0, 0, 0}}
	}

	min, _ := stats.Min(data)
	max, _ := stats.Max(data)

	return opts.KlineData{Value: [4]float64{
		data[0],
		data[len(data)-1],
		min,
		max,
	}}
}

// func convertValueArray(arr []nu.Value) ([]float64, error) {
// 	res := make([]float64, 0)

// 	for _, item := range arr {
// 		v, err := ValueToFloat64(item)
// 		if err == nil {
// 			res = append(res, v)
// 		} else {
// 			return []float64{}, fmt.Errorf("Input array contains invalid values: %s", err.Error())
// 		}
// 	}

// 	return res, nil
// }

func klineReadInputListItem(listItem []nu.Value, klineSeries KlineDataSeries, xAxisName string) (xValue any, res error) {
	series := make(Float64Series)

	for itemIndex, item := range listItem {
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
			// Try to set xAxisName to one of the columns in the record.
			if itemIndex == 0 {
				xAxisName = autoSetXaxis(itemValue, xAxisName)
			}

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
			xv, err := klineReadInputListItem(itemValue, klineSeries, xAxisName)
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
			res = fmt.Errorf("klineReadInputListItem: unsupported input value type: %T", listItem)
			return
		}
	}

	for k, v := range series {
		items := getSeries(klineSeries, k)
		klineSeries[k] = append(items, buildKlineDataValue(v))
	}

	return
}

func plotKline(input any, call *nu.ExecCommand) error {
	series := make(KlineDataSeries)
	var xSeries []any = nil

	xAxisName := getCellPathFlag(call, "xaxis", XAxisSeries)
	slog.Debug("plotKline", "xAxisName", xAxisName)

	switch inputValue := input.(type) {
	case []nu.Value:
		xValue, err := klineReadInputListItem(inputValue, series, xAxisName)
		if err == nil {
			switch items := xValue.(type) {
			case []any:
				xSeries = items
			}
		} else {
			return err
		}
	default:
		return fmt.Errorf("plotKline: unsupported input value type: %T", inputValue)
	}

	// create a new kline instance
	kline := charts.NewKLine()

	kline.SetGlobalOptions(buildGlobalChartOptions(call)...)

	// Put data into instance
	itemCount := 0
	for sName, sValues := range series {
		if sName == xAxisName {
			continue
		}

		itemCount = len(sValues)
		slog.Debug("plotKline: Adding items to series", "series", sName, "items", itemCount)
		kline = kline.AddSeries(sName, sValues)
	}

	if xAxisName != XAxisSeries {
		kline = kline.SetXAxis(xSeries)
	} else {
		xRange := make([]int, itemCount)
		for i := range itemCount {
			xRange[i] = i
		}

		kline = kline.SetXAxis(xRange)
	}

	setPageTitle(call, &kline.BaseConfiguration)

	return renderChart(func(f *os.File) error { return kline.Render(f) })
}
