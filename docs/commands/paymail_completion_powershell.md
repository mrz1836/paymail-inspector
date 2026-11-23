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
  -h, --help[=true|false]              help for powershell
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

