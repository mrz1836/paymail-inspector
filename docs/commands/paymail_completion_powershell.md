## paymail completion powershell

Generate the autocompletion script for powershell

### Synopsis

Generate the autocompletion script for powershell.

To load completions in your current shell session:

	paymail completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.


```
paymail completion powershell [flags]
```

### Options

```
  -h, --help              help for powershell
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

