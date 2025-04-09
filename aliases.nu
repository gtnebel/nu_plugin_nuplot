
alias gmt = go mod tidy
alias build = go build
alias gb = go build 

alias reimport = if (plugin list | where name == nuplot | length) > 0 {
  plugin rm nuplot; plugin add nu_plugin_nuplot; plugin use nuplot
} else {
  plugin add nu_plugin_nuplot; plugin use nuplot
}
