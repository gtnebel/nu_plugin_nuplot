module github.com/gtnebel/nu_plugin_nuplot

go 1.25.0

require (
	github.com/ainvaltin/nu-plugin v0.0.0-20260530183442-019dd4784d2e
	github.com/go-echarts/go-echarts/v2 v2.7.2
	github.com/montanaflynn/stats v0.9.0
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c
	github.com/relvacode/iso8601 v1.7.0
)

require (
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
)

// replace github.com/ainvaltin/nu-plugin => github.com/gtnebel/nu-plugin v0.0.0-20260220134143-4a2d0613d3c1
// replace github.com/ainvaltin/nu-plugin => ../nu-plugin
