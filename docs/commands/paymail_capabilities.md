## paymail capabilities

Get the capabilities of the paymail domain

### Synopsis

```
                          ___.   .__.__  .__  __  .__               
  ____ _____  ___________ \_ |__ |__|  | |__|/  |_|__| ____   ______
_/ ___\\__  \ \____ \__  \ | __ \|  |  | |  \   __\  |/ __ \ /  ___/
\  \___ / __ \|  |_> > __ \| \_\ \  |  |_|  ||  | |  \  ___/ \___ \ 
 \___  >____  /   __(____  /___  /__|____/__||__| |__|\___  >____  >
     \/     \/|__|       \/    \/                         \/     \/
```

This command will return the capabilities for a given paymail domain.

Capability Discovery is the process by which a paymail client learns the supported 
features of a paymail service and their respective endpoints and configurations.

Drawing inspiration from RFC 5785 and IANA's Well-Known URIs resource, the Capability Discovery protocol 
dictates that a machine-readable document is placed in a predictable location on a web server.

Read more at: http://bsvalias.org/02-02-capability-discovery.html

```
paymail capabilities [flags]
```

### Examples

```
paymail capabilities moneybutton.com
```

### Options

```
  -h, --help   help for capabilities
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

* [paymail](paymail.md)	 - Inspect, validate domains or resolve paymail addresses

