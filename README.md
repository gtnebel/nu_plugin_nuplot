# nuplot

`nuplot` is a [nushell](https://www.nushell.sh) plugin for plotting charts. It
builds interactive charts from your data that are opened inside the web browser.

```shell
go install github.com/gtnebel/nu_plugin_nuplot@latest
```

## Features

- Supported chart types:
  - Line chart
  - Bar chart
  - Stacked bar chart
  - Pie chart
  - Boxplot chart
  - Kline chart
- Chart title, size and color theme can be adjusted
- Configure, which series is used for the x-axis

## Examples

#### A Simple Pie chart

```nushell
{'apples': 7 'oranges': 5 'bananas': 3} | nuplot pie --title "Fruits"
```

![Fruits](https://github.com/user-attachments/assets/848bdd94-364b-4c9e-b196-32e8d032bbd1)

#### Show weather forcast from wttr.in as bar chart

```nushell
http get http://wttr.in?format=j1
| get weather
| select date avgtempC
| each {|l| {date: ($l.date | into datetime) avgtempC: ($l.avgtempC | into int)} }
| nuplot bar --xaxis date --title "Weather forcast"
```

The data type conversion for `avgtempC` is needed, because nuplot only shows
series of numbers. The data type conversion of the `date` column can be omitted
but will lead to warnings in the moment because the date format is not
recognized correctly.

![Weather forcast (1)](https://github.com/user-attachments/assets/0674aa72-37e9-4868-a156-31cf990fbde9)

#### Show the average monthly temperatures as a boxplot chart

```nushell
http get https://bulk.meteostat.net/v2/hourly/2024/10389.csv.gz
| gunzip
| from csv --noheaders
| select column0 column2
| rename date temperature
| upsert date {|l| $l.date | format date "%B"}
| chunk-by {$in.date}
| nuplot boxplot --xaxis date --title "Average monthly temperatures for 2024 in Berlin"
```

![image](https://github.com/user-attachments/assets/760d626b-44c0-4979-88da-e20a4946a79c)

## Getting binaries

Binaries for a range of operating systems and architectures are provided with
each release on GitHub. Simply download the zip file for your os and
architecture.

## Build from source

**Prerequisits:** You will need the Go compiler to build the project.

Check out the project

```sh
git clone https://github.com/gtnebel/nu_plugin_nuplot.git
```

Build the project

```sh
go build
```

## Register the plugin an nushell

Use the `plugin add` and `plugin use` commands to register and use the plugin.

The `plugin use` command is only needed to activate the newly added plugin in
the currently running shell.

```nu
plugin add nu_plugin_nuplot
plugin use nuplot;
```

Now, `help nuplot line` should show the help for the line chart.

## Acknowledgments

This software is using other great open source libraries:

- [Nushell Plugin](https://github.com/ainvaltin/nu-plugin): A library for
  developing Nushell plugins in Golang
- [go-echarts](https://github.com/go-echarts/go-echarts): A charts library for
  Golang
- [Stats - Golang Statistics Package](https://github.com/montanaflynn/stats):
  Golang statistics library
- [Package browser](https://github.com/pkg/browser): Open generated chart in a
  browser window

[Additional dependencies](https://github.com/gtnebel/nu_plugin_nuplot/network/dependencies)

## BSD 2-Clause License

Copyright (c) 2025, Thomas Nebel

Redistribution and use in source and binary forms, with or without modification,
are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this
   list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
