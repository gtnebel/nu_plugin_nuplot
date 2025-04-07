package commands

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	charttypes "github.com/go-echarts/go-echarts/v2/types"

	"github.com/ainvaltin/nu-plugin"
	"github.com/ainvaltin/nu-plugin/types"
)

func NuplotLine() *nu.Command {
	return &nu.Command{
		Signature: nu.PluginSignature{
			Name:        "nuplot line",
			Category:    "Chart",
			Desc:        `Plots the data that is piped into the command as 'echarts' graph.`,
			SearchTerms: []string{"plot", "graph", "line", "bar"},
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
		return nil
	case nu.Value:
		items := make([]opts.LineData, 0)

		switch data := in.Value.(type) {
		case []int64:
			for _, val := range data {
				items = append(items, opts.LineData{Value: val})
			}
		case []float64:
			for _, val := range data {
				items = append(items, opts.LineData{Value: val})
			}
		case []nu.Value:
			for _, val := range data {
				items = append(items, opts.LineData{Value: val.Value})
				// switch v := val.Value.(type) {
				// case int64:
				// 	items = append(items, opts.LineData{Value: v})
				// case float64:
				// 	items = append(items, opts.LineData{Value: v})
				// }
			}
		default:
			return fmt.Errorf("unsupported input value type %T", data)
		}

		// create a new line instance
		line := charts.NewLine()
		// set some global options like Title/Legend/ToolTip or anything else
		line.SetGlobalOptions(
			charts.WithInitializationOpts(opts.Initialization{Theme: charttypes.ThemeWesteros}),
			charts.WithTitleOpts(opts.Title{
				Title:    "Line example in Westeros theme",
				Subtitle: "Line chart rendered by the http server this time",
			}))

		xRange := make([]int, len(items))
		for i := range len(items) {
			xRange[i] = i
		}

		// Put data into instance
		line.SetXAxis(xRange).
			AddSeries("Items", items).
			SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)}))

		f1, _ := os.Create("line.html")
		line.Render(f1)

		// If the call returns nothing, we can return nil here.
		return nil

		// var v any
		// if _, err := plist.Unmarshal(buf, &v); err != nil {
		// 	return fmt.Errorf("decoding input as plist: %w", err)
		// }
		// rv, err := asValue(v)
		// if err != nil {
		// 	return fmt.Errorf("converting to Value: %w", err)
		// }
		// return call.ReturnValue(ctx, rv)
	case <-chan nu.Value:
		return fmt.Errorf("unsupported input type: <-chan %T", call.Input)

	case io.Reader:
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
		return fmt.Errorf("unsupported input type: io.Reader %T", call.Input)
	default:
		return fmt.Errorf("unsupported input type %T", call.Input)
	}
}

func asValue(v any) (_ nu.Value, err error) {
	switch in := v.(type) {
	case uint64, float64, bool, string, []byte:
		return nu.Value{Value: in}, nil
	case []any:
		lst := make([]nu.Value, len(in))
		for i := 0; i < len(in); i++ {
			if lst[i], err = asValue(in[i]); err != nil {
				return nu.Value{}, err
			}
		}
		return nu.Value{Value: lst}, nil
	case map[string]any:
		rec := nu.Record{}
		for k, v := range in {
			if rec[k], err = asValue(v); err != nil {
				return nu.Value{}, err
			}
		}
		return nu.Value{Value: rec}, nil
	default:
		return nu.Value{}, fmt.Errorf("unsupported value type %T", in)
	}
}
