# nuplot

`nuplot` is a [nushell](https://www.nushell.sh) plugin for plotting charts. It
builds interactive charts from your data that are opened inside the web browser.

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

## Build and install

**Prerequisits:** You will need the Go compiler to build the project.

Check out the project

```sh
git clone https://github.com/gtnebel/nu_plugin_nuplot.git
```

Build the project

```sh
go build
```

Install and use the plugin

```nu
plugin add nu_plugin_nuplot; plugin use nuplot;
```

Now, `help nuplot line` should show the help for the line chart.

## TODO

- [x] Implement chart types
  - [x] Line chart
  - [x] Bar chart
  - [x] Stacked bar chart
  - [x] Pie chart
  - [x] Boxplot chart
  - [x] Kline chart
- [x] Implement main command `nuplot`
- [x] Define a default set of flags for all chart types
- [x] Define reasonable default features for all charts
- [ ] Documentation
- [ ] Packaging
