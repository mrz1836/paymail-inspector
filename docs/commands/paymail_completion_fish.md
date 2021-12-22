## paymail completion fish

Generate the autocompletion script for fish

### Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	paymail completion fish | source

To load completions for every new session, execute once:

	paymail completion fish > ~/.config/fish/completions/paymail.fish

You will need to start a new shell for this setup to take effect.


```
paymail completion fish [flags]
```

### Options

```
  -h, --help              help for fish
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --bsvalias string   The bsvalias version (default "1.0")
      --config string     Custom config file (default is $HOME/paymail/config.yaml)
      --docs              Generate docs from all commands (./docs/commands)
      --flush-cache       Flushes ALL cache, empties local database
      --no-cache          Turn off caching for this specific command
  -t, --skip-tracing      Turn off request tracing information
```

### SEE ALSO

* [paymail completion](paymail_completion.md)	 - Generate the autocompletion script for the specified shell

