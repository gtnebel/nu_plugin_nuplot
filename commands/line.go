package commands

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strconv"
	"time"

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
const XAxisSeries = "__x_axis__"

var Themes = []string{
	"chalk", "essos", "infographic", "macarons", "purple-passion", "roma",
	"romantic", "shine", "vintage", "walden", "westeros", "wonderland",
}

func getSeries(series LineDataSeries, name string) LineData {
	s, ok := series[name]

	if ok {
		return s
	} else {
		series[name] = make(LineData, 0)
		return series[name]
	}
}

func matchXValue(nuValue nu.Value) opts.LineData {
	const IsoDate = "2006-01-02 15:04:05 -07:00"

	switch value := nuValue.Value.(type) {
	case string:
		if date, err := time.Parse(time.RFC3339, value); err == nil {
			// log.Println("matchXValue:", "Value is RFC3339 date string")
			return opts.LineData{Value: date}
		}
		if date, err := time.Parse(IsoDate, value); err == nil {
			// log.Println("matchXValue:", "Value is ISO date string")
			return opts.LineData{Value: date}
		}
		if number, err := strconv.ParseFloat(value, 64); err == nil {
			// log.Println("matchXValue:", "Value is number string")
			return opts.LineData{Value: number}
		}

		log.Println("matchXValue:", "Value is unknown string:", value)
		return opts.LineData{Value: value}
	default:
		return opts.LineData{Value: value}
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
				nu.Flag{
					Long:     "subtitle",
					Short:    "s",
					Shape:    syntaxshape.String(),
					Required: false,
					Desc:     "The chart subtitle",
					VarId:    0,
					Default:  &nu.Value{Value: "This chart was rendered by nuplot"},
				},
				nu.Flag{
					Long:     "xaxis",
					Short:    "x",
					Shape:    syntaxshape.String(), // TODO: CellPath not supported...
					Required: false,
					Desc:     "Only if input is a table: the column name wich holds the values for the x-axis",
					VarId:    0,
					// Default:  nil,
				},
				nu.Flag{
					Long:     "color-theme",
					Short:    "c",
					Shape:    syntaxshape.String(),
					Required: false,
					Desc:     "One of: chalk, essos, infographic, macarons, purple-passion, roma, romantic, shine, vintage, walden, westeros, wonderland,",
					VarId:    0,
					Default:  &nu.Value{Value: "westeros"},
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
						series[XAxisSeries] = append(items, matchXValue(v))
					} else {
						// If the column specified in --xaxis does not exist, we
						// set the `xAxisOk` variable to false, so that a
						// simple int range is generated as x axis.
						xAxisOk = false
					}
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
	subtitle, _ := call.FlagValue("subtitle")
	colorTheme, _ := call.FlagValue("color-theme")
	log.Println("plotLine:", "title: ", title.Value.(string))
	log.Println("plotLine:", "subtitle: ", subtitle.Value.(string))
	log.Println("plotLine:", "color-theme: ", colorTheme.Value.(string))

	// If the given color theme is in the list of possible themes, we will
	// enable it.
	theme := charttypes.ThemeWesteros
	if slices.Contains(Themes, colorTheme.Value.(string)) {
		theme = colorTheme.Value.(string)
	}

	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme:  theme,
			Width:  "1200px",
			Height: "600px",
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    title.Value.(string),
			Subtitle: subtitle.Value.(string),
			// Right:    "40%",
		}),
		// charts.WithLegendOpts(opts.Legend{Right: "80%"}),
		charts.WithToolboxOpts(opts.Toolbox{
			// Right: "5%",
			Feature: &opts.ToolBoxFeature{
				SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{
					Type:  "jpg",
					Title: "Download as jpg",
				},
				// DataView: &opts.ToolBoxFeatureDataView{
				// 	Title: "DataView",
				// 	// set the language
				// 	// Chinese version: ["数据视图", "关闭", "刷新"]
				// 	Lang: []string{"data view", "turn off", "refresh"},
				// },
			}},
		),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type: "slider",
		}),
	)

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
