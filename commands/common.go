package commands

import (
	// "context"
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
	// "github.com/ainvaltin/nu-plugin/syntaxshape"
	// "github.com/ainvaltin/nu-plugin/types"
	// "github.com/gtnebel/nu_plugin_nuplot/commands/flags"
)

type PlotHandlerFunc = func(any, *nu.ExecCommand) error

const DefaultSeries = "Items"
const XAxisSeries = "__x_axis__"

var Themes = []string{
	"chalk", "essos", "infographic", "macarons", "purple-passion", "roma",
	"romantic", "shine", "vintage", "walden", "westeros", "wonderland",
}

func matchXValue(nuValue nu.Value) any {
	const IsoDate = "2006-01-02 15:04:05 -07:00"

	switch value := nuValue.Value.(type) {
	case string:
		if date, err := time.Parse(time.RFC3339, value); err == nil {
			// log.Println("matchXValue:", "Value is RFC3339 date string")
			return date
		}
		if date, err := time.Parse(IsoDate, value); err == nil {
			// log.Println("matchXValue:", "Value is ISO date string")
			return date
		}
		if number, err := strconv.ParseFloat(value, 64); err == nil {
			// log.Println("matchXValue:", "Value is number string")
			return number
		}

		log.Println("matchXValue:", "Value is unknown string:", value)
		return value
	default:
		return value
	}
}

func handleCommandInput(call *nu.ExecCommand, plotFunc PlotHandlerFunc) error {
	switch in := call.Input.(type) {
	case nil:
		log.Println("handleCommandInput:", "Input is nil")
		return nil
	case nu.Value:
		log.Println("handleCommandInput:", "Input is nu.Value")
		return plotFunc(in.Value, call)
	case <-chan nu.Value:
		log.Println("handleCommandInput:", "Input is <-chan nu.Value")
		inValues := make([]nu.Value, 0)

		for v := range in {
			inValues = append(inValues, v)
		}

		return plotFunc(inValues, call)
	case io.Reader:
		log.Println("handleCommandInput:", "Input is io.Reader")
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

func buildGlobalChartOptions(call *nu.ExecCommand) []charts.GlobalOpts {
	// set some global options like Title/Legend/ToolTip or anything else
	title, _ := call.FlagValue("title")
	subtitle, _ := call.FlagValue("subtitle")
	colorTheme, _ := call.FlagValue("color-theme")
	width, _ := call.FlagValue("width")
	height, _ := call.FlagValue("height")
	log.Println("plotLine:", "title: ", title.Value.(string))
	log.Println("plotLine:", "subtitle: ", subtitle.Value.(string))
	log.Println("plotLine:", "color-theme: ", colorTheme.Value.(string))
	log.Println("plotLine:", "width: ", width.Value.(int64))
	log.Println("plotLine:", "height: ", height.Value.(int64))

	// If the given color theme is in the list of possible themes, we will
	// enable it.
	theme := charttypes.ThemeWesteros
	if slices.Contains(Themes, colorTheme.Value.(string)) {
		theme = colorTheme.Value.(string)
	}

	return []charts.GlobalOpts{
		charts.WithInitializationOpts(opts.Initialization{
			Theme:  theme,
			Width:  fmt.Sprintf("%dpx", width.Value),
			Height: fmt.Sprintf("%dpx", height.Value),
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
	}
}

func renderChart(renderHandler func(f *os.File) error) {

	chartFile, _ := os.CreateTemp("", "chart-*.html")
	chartFileName := chartFile.Name()
	log.Println("plotLine:", "Rendering output to", chartFileName)
	renderHandler(chartFile) // TODO: handle errors
	chartFile.Close()

	browser.OpenFile(chartFileName)

}
