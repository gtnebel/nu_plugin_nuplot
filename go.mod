module github.com/gtnebel/nu_plugin_nuplot

go 1.26.0

require (
	github.com/ainvaltin/nu-plugin v0.0.0-20260711090100-edde0f88c6d9
	github.com/go-echarts/go-echarts/v2 v2.7.3
	github.com/montanaflynn/stats v0.13.0
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c
	github.com/relvacode/iso8601 v1.8.2
)

require (
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
)

replace github.com/ainvaltin/nu-plugin => github.com/gtnebel/nu-plugin v0.0.0-20261008103812-2ea362700e5a

// replace github.com/ainvaltin/nu-plugin => ../nu-plugin
