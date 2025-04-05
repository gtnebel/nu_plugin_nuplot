package commands

import (
	"context"
	"fmt"
	"io"

	"github.com/ainvaltin/nu-plugin"
	"github.com/ainvaltin/nu-plugin/types"
)

func NuplotLine() *nu.Command {
	return &nu.Command{
		Signature: nu.PluginSignature{
			Name:                 "nuplot line",
			Category:             "Formats",
			Desc:                 `Plots the data that is piped into the command as 'echarts' graph.`,
			SearchTerms:          []string{"plot", "graph", "line", "bar"},
			InputOutputTypes:     []nu.InOutTypes{{In: types.Binary(), Out: types.Any()}, {In: types.String(), Out: types.Any()}},
			AllowMissingExamples: true,
		},
		Examples: nu.Examples{
			{
				Description: `Convert an Open Step array to list of Nu values`,
				Example:     `'(10,foo,)' | from plist`,
				Result:      &nu.Value{Value: []nu.Value{{Value: 10}, {Value: "foo"}}},
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
		// var buf []byte
		switch data := in.Value.(type) {
		// case []byte:
		// 	buf = data
		// case string:
		// 	buf = []byte(data)
		default:
			return fmt.Errorf("unsupported input value type %T", data)
		}
		var v any
		// if _, err := plist.Unmarshal(buf, &v); err != nil {
		// 	return fmt.Errorf("decoding input as plist: %w", err)
		// }
		rv, err := asValue(v)
		if err != nil {
			return fmt.Errorf("converting to Value: %w", err)
		}
		return call.ReturnValue(ctx, rv)
	case io.Reader:
		// decoder wants io.ReadSeeker so we need to read to buf.
		// could read just enough that the decoder can detect the
		// format and stream the rest?
		// buf, err := io.ReadAll(in)
		// if err != nil {
		// 	return fmt.Errorf("reding input: %w", err)
		// }
		var v any
		// if _, err := plist.Unmarshal(buf, &v); err != nil {
		// 	return fmt.Errorf("decoding input as plist: %w", err)
		// }
		rv, err := asValue(v)
		if err != nil {
			return fmt.Errorf("converting to Value: %w", err)
		}
		return call.ReturnValue(ctx, rv)
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
