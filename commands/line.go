package commands

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/pkg/browser"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	charttypes "github.com/go-echarts/go-echarts/v2/types"

	"github.com/ainvaltin/nu-plugin"
	"github.com/ainvaltin/nu-plugin/syntaxshape"
	"github.com/ainvaltin/nu-plugin/types"
)

type LineData = []opts.LineData

type LineDataSeries = map[string]LineData

const DefaultSeries = "Items"

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
				nu.Flag{
					Long:     "title",
					Short:    "t",
					Shape:    syntaxshape.String(),
					Required: false,
					Desc:     "The chart title",
					VarId:    0,
					Default:  &nu.Value{Value: "Line chart"},
				},
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
	switch in := call.Input.(type) {
	case nil:
		log.Println("nuplotLineHandler:", "Input is nil")
		return nil
	case nu.Value:
		log.Println("nuplotLineHandler:", "Input is nu.Value")
		return plotLine(in.Value, call)
	case <-chan nu.Value:
		log.Println("nuplotLineHandler:", "Input is <-chan nu.Value")
		inValues := make([]nu.Value, 0)

		for v := range in {
			inValues = append(inValues, v)
		}

		return plotLine(inValues, call)
	case io.Reader:
		log.Println("nuplotLineHandler:", "Input is io.Reader")
		// decoder wants io.ReadSeeker so we need to read to buf.
		// could read just enough that the decoder can detect the
		// format and stream the rest?
		// buf, err := io.ReadAll(in)
		// if err != nil {
		// 	return fmt.Errorf("reding input: %w", err)
		// }
		// var v any
		// if _, err := plist.Unmarshal(buf, &v); err != nil {
		// 	return fmt.Errorf("decoding input as plist: %w", err)
		// }
		// rv, err := asValue(v)
		// if err != nil {
		// 	return fmt.Errorf("converting to Value: %w", err)
		// }
		// return call.ReturnValue(ctx, rv)
		return fmt.Errorf("1 unsupported input type: %T", call.Input)
	default:
		return fmt.Errorf("2 unsupported input type: %T", call.Input)
	}
}

func plotLine(input any, call *nu.ExecCommand) error {
	series := make(LineDataSeries)

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
					items := getSeries(series, k)
					series[k] = append(items, opts.LineData{Value: v.Value})
				}
			default:
				return fmt.Errorf("3 unsupported input value type: %T", inputValue)
			}
		}
	default:
		return fmt.Errorf("4 unsupported input value type: %T", inputValue)
	}

	// create a new line instance
	line := charts.NewLine()
	// set some global options like Title/Legend/ToolTip or anything else
	title, _ := call.FlagValue("title")
	log.Println("plotLine:", "title: ", title.Value.(string))
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: charttypes.ThemeWesteros}),
		charts.WithTitleOpts(opts.Title{
			Title:    title.Value.(string),
			Subtitle: "This chart was rendered by nuplot",
		}))

	// Put data into instance
	itemCount := 0
	for sName, sValues := range series {
		itemCount = len(sValues)
		log.Println("plotLine:", "Adding", itemCount, "items to series", sName, ":", sValues)
		line = line.AddSeries(sName, sValues)
	}

	xRange := make([]int, itemCount)
	for i := range itemCount {
		xRange[i] = i
	}

	line.SetXAxis(xRange).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)}))

	chartFile, _ := os.CreateTemp("", "chart-*.html")
	chartFileName := chartFile.Name()
	log.Println("plotLine:", "Rendering output to", chartFileName)
	line.Render(chartFile)
	chartFile.Close()

	browser.OpenFile(chartFileName)

	return nil
}
