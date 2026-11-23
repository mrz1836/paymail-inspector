## paymail completion zsh

Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(paymail completion zsh)

To load completions for every new session, execute once:

#### Linux:

	paymail completion zsh > "${fpath[1]}/_paymail"

#### macOS:

	paymail completion zsh > $(brew --prefix)/share/zsh/site-functions/_paymail

You will need to start a new shell for this setup to take effect.


```
paymail completion zsh [flags]
```

### Options

```
  -h, --help[=true|false]              help for zsh
      --no-descriptions[=true|false]   disable completion descriptions
```

### Options inherited from parent commands

```
      --bsvalias string             The bsvalias version to expect from providers (default "1.0")
      --config string               Config file (default: <user config dir>/paymail-inspector/config.yaml)
      --flush-cache[=true|false]    Empty the local response cache before running
      --no-cache[=true|false]       Skip cached responses for this command (fresh results are still cached)
      --no-color[=true|false]       Disable colored output (NO_COLOR is honored too)
  -t, --skip-tracing[=true|false]   Hide per-request network timing
```

### SEE ALSO

* [paymail completion](paymail_completion.md)	 - Generate the autocompletion script for the specified shell

