# nuplot

`nuplot` is a Nushell plugin for plotting charts.

## Project status

The plugin is still in early development. Currently, only the line chart is
supported and the plugin still writes lots of debugging infos to `stderr`.

## Build and install

**Prerequisits:** You will need the Go compiler to compile the project.

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
  - [ ] Bar chart
  - [ ] Stacked bar chart
  - [ ] ...
- [ ] Define a default set of flags for all chart types
- [ ] Define reasonable default features for all charts
- [ ] Documentation
- [ ] Packaging
