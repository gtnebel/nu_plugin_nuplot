set shell := ['nu', '-c']

# Build the plugin
build:
    go build

# Update go.mod file
gmt:
    go mod tidy

# Update all project dependencies
update: && gmt
    go get -u ./...

# Remove the nuplot plugin from nushell
[private]
plugin_rm:
    if (plugin list | where name == 'nuplot' | length) > 0 { plugin rm nuplot }

# Add the nuplot plugin to nushell
add: build plugin_rm
    plugin add nu_plugin_nuplot
    @print $"\n✅ (ansi g)The plugin is now added to nushell. You can activate it with: (ansi yb)plugin use nuplot(ansi reset)\n"

# Rename build artifacts to match release file names
rename-artifacts path nu_version:
    ls -f {{ path }} \
    | where name like 'nu_plugin_nuplot-' \
    | each {|f| \
        let p = $f.name | path parse; \
        let new_name = ($p.stem | str replace 'darwin' 'macos') + "-nushell_{{ nu_version }}"; \
        mv $f.name ({ parent: $p.parent stem: $new_name extension: $p.extension} | path join) \
    }
