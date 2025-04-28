# nuplot

`nuplot` is a [nushell](https://www.nushell.sh) plugin for plotting charts. It builds interactive charts from your data that are opened inside the web browser. 

## Features

- The plugin shows line, (stacked) bar and pie charts.
- Chart title, size and color theme can be adjusted
- Input types:
  - **line chart** and **bar chart**
    - List of numbers (a single series)
    - Table (one series per column)
  - **pie chart**
    - List of numbers (values without labels)
    - Record (record key is then label)

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

The data type conversion for `avgtempC` is needed, because nuplot only shows series of numbers. The data type conversion of the `date` column can be omitted but will lead to warnings in the moment because the date format is not recognized correctly.

![Weather forcast (1)](https://github.com/user-attachments/assets/0674aa72-37e9-4868-a156-31cf990fbde9)

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

- [ ] Implement chart types
  - [x] Line chart
  - [x] Bar chart
  - [x] Stacked bar chart
  - [x] Pie chart
  - [ ] ...
- [x] Define a default set of flags for all chart types
- [x] Define reasonable default features for all charts
- [ ] Documentation
- [ ] Packaging
