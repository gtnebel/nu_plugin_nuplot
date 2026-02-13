module github.com/gtnebel/nu_plugin_nuplot

go 1.24.0

require (
	github.com/ainvaltin/nu-plugin v0.0.0-20250907111918-1d43779b9a0f
	github.com/go-echarts/go-echarts/v2 v2.6.2
	github.com/montanaflynn/stats v0.7.1
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c
	github.com/relvacode/iso8601 v1.7.0
)

require (
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
)

replace github.com/ainvaltin/nu-plugin => github.com/gtnebel/nu-plugin v0.0.0-20260213094314-1e10fe1ea48d
