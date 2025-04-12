package flags

import (
	"github.com/ainvaltin/nu-plugin"
	"github.com/ainvaltin/nu-plugin/syntaxshape"
)

var (
	Title = nu.Flag{
		Long:     "title",
		Short:    "t",
		Shape:    syntaxshape.String(),
		Required: false,
		Desc:     "The chart title",
		VarId:    0,
		Default:  &nu.Value{Value: "Chart title"},
	}

	SubTitle = nu.Flag{
		Long:     "subtitle",
		Short:    "s",
		Shape:    syntaxshape.String(),
		Required: false,
		Desc:     "The chart subtitle",
		VarId:    0,
		Default:  &nu.Value{Value: "This chart was rendered by nuplot"},
	}

	XAxis = nu.Flag{
		Long:     "xaxis",
		Short:    "x",
		Shape:    syntaxshape.String(), // TODO: CellPath not supported...
		Required: false,
		Desc:     "Only if input is a table: the column name wich holds the values for the x-axis",
		VarId:    0,
		// Default:  nil,
	}

	ColorTheme = nu.Flag{
		Long:     "color-theme",
		Short:    "c",
		Shape:    syntaxshape.String(),
		Required: false,
		Desc:     "One of: chalk, essos, infographic, macarons, purple-passion, roma, romantic, shine, vintage, walden, westeros, wonderland,",
		VarId:    0,
		Default:  &nu.Value{Value: "westeros"},
	}

	Width = nu.Flag{
		Long:     "width",
		Short:    "W",
		Shape:    syntaxshape.Int(),
		Required: false,
		Desc:     "Width of chart in pixels",
		VarId:    0,
		Default:  &nu.Value{Value: 1200},
	}

	Height = nu.Flag{
		Long:     "height",
		Short:    "H",
		Shape:    syntaxshape.Int(),
		Required: false,
		Desc:     "Height of chart in pixels",
		VarId:    0,
		Default:  &nu.Value{Value: 600},
	}
)
