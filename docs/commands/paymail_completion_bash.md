## paymail completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(paymail completion bash)

To load completions for every new session, execute once:

#### Linux:

	paymail completion bash > /etc/bash_completion.d/paymail

#### macOS:

	paymail completion bash > /usr/local/etc/bash_completion.d/paymail

You will need to start a new shell for this setup to take effect.


```
paymail completion bash
```

### Options

```
  -h, --help              help for bash
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

