# nuplot

`nuplot` is a Nushell plugin for plotting charts.

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
