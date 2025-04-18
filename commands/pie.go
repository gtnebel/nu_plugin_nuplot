package commands

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"

	"github.com/ainvaltin/nu-plugin"
	"github.com/ainvaltin/nu-plugin/types"

	"github.com/gtnebel/nu_plugin_nuplot/commands/flags"
)

type PieDataList = []opts.PieData

type PieDataSeries = map[string]PieDataList

func NuplotPie() *nu.Command {
	return &nu.Command{
		Signature: nu.PluginSignature{
			Name:        "nuplot pie",
			Category:    "Chart",
			Desc:        "Plots the data that is piped into the command as pi chart.",
			Description: "Title, size and color theme can be configured by flags. Each column that contains numbers will be plottet. The X axis can be set by means of the --xaxis flag.",
			SearchTerms: []string{"plot", "graph", "pie"},
			// OptionalPositional: nu.PositionalArgs{},
			Named: nu.Flags{
				// flags.XAxis,
				flags.Title,
				flags.SubTitle,
				flags.Width,
				flags.Height,
				flags.ColorTheme,
			},
			InputOutputTypes: []nu.InOutTypes{
				{In: types.Record(types.RecordDef{}), Out: types.Nothing()},
				{In: types.Table(types.RecordDef{}), Out: types.Nothing()},
				{In: types.List(types.Table(types.RecordDef{})), Out: types.Nothing()},
				{In: types.List(types.Number()), Out: types.Nothing()},
			},
			AllowMissingExamples: true,
		},
		Examples: nu.Examples{
			{
				Description: `Plot a pie graph of an array of numbers.`,
				Example:     `{'apples': 7 'oranges': 5 'bananas': 3} | nuplot pie --title "Fruits"`,
			},
		},
		OnRun: nuplotPieHandler,
	}
}

func nuplotPieHandler(ctx context.Context, call *nu.ExecCommand) error {
	return handleCommandInput(call, plotPie)
}

func plotPie(input any, call *nu.ExecCommand) error {
	series := make(PieDataSeries)

	xAxisName := getStringFlag(call, "xaxis", XAxisSeries)
	log.Println("plotPie:", "xAxisName:", xAxisName)

	switch inputValue := input.(type) {
	case []nu.Value:
		for _, item := range inputValue {
			switch itemValue := item.Value.(type) {
			case int64:
				items := getSeries(series, DefaultSeries)
				series[DefaultSeries] = append(items, opts.PieData{Value: itemValue})
			case float64:
				items := getSeries(series, DefaultSeries)
				series[DefaultSeries] = append(items, opts.PieData{Value: itemValue})
			case nu.Record:
				for k, v := range itemValue {
					if k == xAxisName {
						continue
					}

					_, ok1 := v.Value.(int64)
					_, ok2 := v.Value.(float64)
					if ok1 || ok2 {
						items := getSeries(series, k)
						series[k] = append(items, opts.PieData{Value: v.Value})
					}
				}

				// If a xaxis is defined, fill the series with the values.
				if xAxisName != XAxisSeries {
					if v, ok := itemValue[xAxisName]; ok {
						items := getSeries(series, xAxisName)
						series[xAxisName] = append(items, opts.PieData{Value: matchXValue(v)})
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
	case nu.Record:
		for k, v := range inputValue {
			if k == xAxisName {
				continue
			}

			_, ok1 := v.Value.(int64)
			_, ok2 := v.Value.(float64)
			if ok1 || ok2 {
				items := getSeries(series, "Items")
				series["Items"] = append(items, opts.PieData{Name: k, Value: v.Value})
			}
		}
	default:
		return fmt.Errorf("unsupported input value type: %T", inputValue)
	}

	// create a new pie instance
	pie := charts.NewPie()

	pie.SetGlobalOptions(buildGlobalChartOptions(call)...)

	// Put data into instance
	itemCount := 0
	for sName, sValues := range series {
		if sName == xAxisName {
			continue
		}

		itemCount = len(sValues)
		log.Println("plotPie:", "Adding", itemCount, "items to series", sName)
		pie = pie.AddSeries(sName, sValues)
	}

	renderChart(func(f *os.File) error { return pie.Render(f) })

	return nil
}
