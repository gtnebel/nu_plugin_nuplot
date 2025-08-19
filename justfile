set shell := ['nu', '-c']

# Build the plugin
build:
	go build

# Update go.mod file
tidy:
	go mod tidy

# Remove the nuplot plugin from nushell
[private]
plugin_rm:
	if (plugin list | where name == 'nuplot' | length) > 0 { plugin rm nuplot }

# Add the nuplot plugin to nushell
init: build plugin_rm
	plugin add nu_plugin_nuplot
	@print $"\n✅ (ansi g)The plugin is now added to nushell. You can activate it with: (ansi yb)plugin use nuplot(ansi reset)\n"
