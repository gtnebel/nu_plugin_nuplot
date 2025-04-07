module github.com/gtnebel/nu_plugin_nuplot

go 1.23

toolchain go1.23.8

require (
	github.com/ainvaltin/nu-plugin v0.0.0-20250209110408-3e103b6f5c59
	// github.com/gtnebel/nu-plugin latest
	github.com/go-echarts/go-echarts/v2 v2.5.2
)

require (
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
)

replace github.com/ainvaltin/nu-plugin => github.com/gtnebel/nu-plugin v0.0.0-20250407192509-e6873fbd00e8
