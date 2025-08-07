// This package holds the flags used in all subcommands.
package flags

import (
	"github.com/ainvaltin/nu-plugin"
	"github.com/ainvaltin/nu-plugin/syntaxshape"
)

var (
	Verbose = nu.Flag{
		Long:     "verbose",
		Short:    'v',
		Shape:    nil,
		Required: false,
		Desc:     "Prints debug messages to the terminal.",
		VarId:    0,
		Default:  nil,
	}

	Title = nu.Flag{
		Long:     "title",
		Short:    'T',
		Shape:    syntaxshape.String(),
		Required: false,
		Desc:     "The chart title",
		VarId:    0,
		Default:  &nu.Value{Value: "Nuplot Chart"},
	}

	SubTitle = nu.Flag{
		Long:     "subtitle",
		Short:    'S',
		Shape:    syntaxshape.String(),
		Required: false,
		Desc:     "The chart subtitle",
		VarId:    0,
		Default:  &nu.Value{Value: "This chart was rendered by nuplot"},
	}

	Width = nu.Flag{
		Long:     "width",
		Short:    'W',
		Shape:    syntaxshape.Int(),
		Required: false,
		Desc:     "Width of chart in pixels",
		VarId:    0,
		Default:  &nu.Value{Value: 1200},
	}

	Height = nu.Flag{
		Long:     "height",
		Short:    'H',
		Shape:    syntaxshape.Int(),
		Required: false,
		Desc:     "Height of chart in pixels",
		VarId:    0,
		Default:  &nu.Value{Value: 600},
	}

	ColorTheme = nu.Flag{
		Long:     "color-theme",
		Short:    'C',
		Shape:    syntaxshape.String(),
		Required: false,
		Desc:     "One of: chalk, essos, infographic, macarons, purple-passion, roma, romantic, shine, vintage, walden, westeros, wonderland,",
		VarId:    0,
		Default:  &nu.Value{Value: "westeros"},
	}

	XAxis = nu.Flag{
		Long:     "xaxis",
		Short:    'x',
		Shape:    syntaxshape.CellPath(),
		Required: false,
		Desc:     "Only if input is a table: the column name wich holds the values for the x-axis",
		VarId:    0,
		// Default:  nil,
	}

	Fitted = nu.Flag{
		Long:     "fitted",
		Short:    'f',
		Shape:    nil,
		Required: false,
		Desc:     "Removes zero offset from y-axis to fit values into chart area.",
		VarId:    0,
		Default:  nil,
	}

	Stacked = nu.Flag{
		Long:     "stacked",
		Short:    's',
		Shape:    nil,
		Required: false,
		Desc:     "Plots data series as stacked bar chart.",
		VarId:    0,
		Default:  nil,
	}

	XYReverse = nu.Flag{
		Long:     "xyreverse",
		Short:    'r',
		Shape:    nil,
		Required: false,
		Desc:     "Reverse the x and y axes",
		VarId:    0,
		Default:  nil,
	}
)
