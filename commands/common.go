// This package holds all plugin subcommands along with common data types
// and functions used by all subcommands.
package commands

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gtnebel/nu_plugin_nuplot/commands/flags"
	"github.com/pkg/browser"

	"github.com/relvacode/iso8601"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	charttypes "github.com/go-echarts/go-echarts/v2/types"

	"github.com/ainvaltin/nu-plugin"
)

// Handler function that implements the output of a specific plot.
// This type is used in the [handleCommandInput] function.
type PlotHandlerFunc = func(any, *nu.ExecCommand) error

// Default name of a series if no other name is given.
const DefaultSeries = "Items"

// Internal name of the series representing the x axis.
const XAxisSeries = "__x_axis__"

// List of all availlable themes
var Themes = []string{
	"chalk", "essos", "infographic", "macarons", "purple-passion", "roma",
	"romantic", "shine", "vintage", "walden", "westeros", "wonderland",
}

// Abstract data type so that [getSeries] can be called for all plot types.
type ChartData interface {
	float64 | []float64 | opts.LineData | opts.BarData | opts.PieData | opts.BoxPlotData | opts.KlineData
}

// Retrieves a series with the given name from the series map. If the given
// series does not exist yet in the map, it is automatically  created and
// returned.
func getSeries[SeriesType ChartData](series map[string][]SeriesType, name string) []SeriesType {
	s, ok := series[name]

	if ok {
		return s
	} else {
		series[name] = make([]SeriesType, 0)
		return series[name]
	}
}

// Tries to find a column that is likely to be used as x-axis and returns it.
func autoSetXaxis(rec nu.Record, xAxisName string) string {
	// TODO: Find more column names for the x axis
	xAxisNames := []string{
		"timestamp", "ts", "date", "time", "datetime", "nr", "id",
	}

	if xAxisName != XAxisSeries {
		slog.Debug("autoSetXaxis: using user-provided x-axis name")
		return xAxisName
	} else {
		for k, _ := range rec {
			if slices.Contains(xAxisNames, strings.ToLower(k)) {
				slog.Debug("autoSetXaxis: using column of table as x-axis", "column", k)
				return k
			}
		}
	}

	slog.Debug("autoSetXaxis: no suitable column found")
	return xAxisName
}

// Retrieve the string value of a flag from the call. The name of the flag and
// a default value has to be provided.
func getStringFlag(call *nu.ExecCommand, name string, deflt string) string {
	value, _ := call.FlagValue(name)

	if value.Value != nil {
		return value.Value.(string)
	} else {
		return deflt
	}
}

// Retrieve the CellPath value of a flag from the call. The name of the flag and
// a default value has to be provided.
//
// This function returns the string representation of the cell path. Chained
// cell paths will be returned in "a.b" syntax.
//
// NOTE: Chained cell paths are not yet supported in the plotting commands.
func getCellPathFlag(call *nu.ExecCommand, name string, deflt string) string {
	value, _ := call.FlagValue(name)

	if value.Value != nil {
		path := value.Value.(nu.CellPath)
		return path.String()
	} else {
		return deflt
	}
}

// Retrieve the int64 value of a flag from the call. The name of the flag and
// a default value has to be provided.
func getIntFlag(call *nu.ExecCommand, name string, deflt int64) int64 {
	value, _ := call.FlagValue(name)

	if value.Value != nil {
		return value.Value.(int64)
	} else {
		return deflt
	}
}

// Retrieve the bool value of a flag from the call. The default value is false.
func getBoolFlag(call *nu.ExecCommand, name string) bool {
	value, _ := call.FlagValue(name)

	if value.Value != nil {
		return value.Value.(bool)
	} else {
		return false
	}
}

// Convert an int or float nushell value to float64.
func ValueToFloat64(value nu.Value) (float64, error) {
	switch v := value.Value.(type) {
	case int64:
		return float64(v), nil
	case float64:
		return v, nil
	default:
		return 0, fmt.Errorf("incompatible input type for ValueToFloat64(): %T", v)
	}
}

// Tries to parse the xaxis values into [time.Time] or [float]. This is useful
// when data is loaded from CSV or JSON files, where dates and numbers are
// sometimes represented as strings.
func matchXValue(nuValue nu.Value) any {
	const IsoDate = "2006-01-02 15:04:05 -07:00"
	const IsoDate_Local = "2006-01-02 15:04:05"
	const IsoDate_Date = "2006-01-02"

	switch value := nuValue.Value.(type) {
	case string:
		if date, err := iso8601.ParseString(value); err == nil {
			// slog.Debug("matchXValue: Value is ISO8601 date string")
			return date
		}
		if date, err := time.Parse(time.RFC3339, value); err == nil {
			// slog.Debug("matchXValue: Value is RFC3339 date string")
			return date
		}
		if date, err := time.Parse(IsoDate, value); err == nil {
			// slog.Debug("matchXValue: Value is ISO date string")
			return date
		}
		if date, err := time.ParseInLocation(IsoDate_Local, value, time.Local); err == nil {
			// slog.Debug("matchXValue: Value is ISO date (local time) string")
			return date
		}
		if date, err := time.ParseInLocation(IsoDate_Date, value, time.Local); err == nil {
			// slog.Debug("matchXValue: Value is ISO date (only date part) string")
			return date
		}
		if number, err := strconv.ParseFloat(value, 64); err == nil {
			// slog.Debug("matchXValue: Value is number string")
			return number
		}

		slog.Debug("matchXValue: Value is unknown string", "value", value)
		return value
	default:
		return value
	}
}

// Sets the debug level to [slog.LevelDebug], if the verbose flag is given on
// the command line.
func checkVerboseFlag(call *nu.ExecCommand) {
	if getBoolFlag(call, flags.Verbose.Long) {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
}

// This is the top level handler function that is called from the [nu.Command].
// The function analyzes, in which format the input values are given and than
// calls the provided plotFunc [PlotHandlerFunc] function.
func handleCommandInput(call *nu.ExecCommand, plotFunc PlotHandlerFunc) error {
	switch in := call.Input.(type) {
	case nil:
		slog.Debug("handleCommandInput: Input is nil")
		return nil
	case nu.Value:
		slog.Debug("handleCommandInput: Input is nu.Value")
		return plotFunc(in.Value, call)
	case <-chan nu.Value:
		slog.Debug("handleCommandInput: Input is <-chan nu.Value")
		inValues := make([]nu.Value, 0)

		for v := range in {
			inValues = append(inValues, v)
		}

		return plotFunc(inValues, call)
	case io.Reader:
		slog.Debug("handleCommandInput: Input is io.Reader")
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

// Builds the global chart options that determine the appearance of the chart.
func buildGlobalChartOptions(call *nu.ExecCommand) []charts.GlobalOpts {
	// set some global options like Title/Legend/ToolTip or anything else
	title := getStringFlag(call, "title", flags.Title.Default.Value.(string))
	subtitle := getStringFlag(call, "subtitle", "This chart was rendered by nuplot.")
	colorTheme := getStringFlag(call, "color-theme", charttypes.ThemeWesteros)
	width := getIntFlag(call, "width", 1200)
	height := getIntFlag(call, "height", 600)
	fitted := getBoolFlag(call, "fitted")
	slog.Debug("buildGlobalChartOptions", "title", title)
	slog.Debug("buildGlobalChartOptions", "subtitle", subtitle)
	slog.Debug("buildGlobalChartOptions", "color-theme", colorTheme)
	slog.Debug("buildGlobalChartOptions", "width", width)
	slog.Debug("buildGlobalChartOptions", "height", height)
	slog.Debug("buildGlobalChartOptions", "fitted", fitted)

	// If the given color theme is in the list of possible themes, we will
	// enable it.
	theme := charttypes.ThemeWesteros
	if slices.Contains(Themes, colorTheme) {
		theme = colorTheme
	}

	return []charts.GlobalOpts{
		charts.WithInitializationOpts(opts.Initialization{
			Theme:  theme,
			Width:  fmt.Sprintf("%dpx", width),
			Height: fmt.Sprintf("%dpx", height),
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    title,
			Subtitle: subtitle,
			// Right:    "40%",
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			// Trigger: "item",
			AxisPointer: &opts.AxisPointer{
				Type: "cross",
			},
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
		charts.WithDataZoomOpts(opts.DataZoom{
			Type: "inside",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Scale: opts.Bool(fitted),
		}),
	}
}

// Sets the page title for the given chart. We only have single charts and
// don't format them on a page, so we set the chart.PageTitle shortcut
// Field directly.
func setPageTitle(call *nu.ExecCommand, chart *charts.BaseConfiguration) {
	title := getStringFlag(call, "title", flags.Title.Default.Value.(string))
	chart.PageTitle = title
}

// Helper function that wraps the creation of the temporary file the chart is
// saved into and than opening the file in the web browser. The renderHandler
// function performs the actual plotting of the chart.
func renderChart(renderHandler func(f *os.File) error) error {
	chartFile, _ := os.CreateTemp("", "chart-*.html")
	chartFileName := chartFile.Name()
	slog.Debug("plotLine: Rendering output", "filename", chartFileName)
	err := renderHandler(chartFile)
	defer chartFile.Close()

	if err == nil {
		browser.OpenFile(chartFileName)
	}

	return err
}
