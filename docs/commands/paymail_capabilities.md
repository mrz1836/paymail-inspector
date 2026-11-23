## paymail capabilities

List the capabilities a paymail domain publishes

### Synopsis

```text
                          ___.   .__.__  .__  __  .__
  ____ _____  ___________ \_ |__ |__|  | |__|/  |_|__| ____   ______
_/ ___\\__  \ \____ \__  \ | __ \|  |  | |  \   __\  |/ __ \ /  ___/
\  \___ / __ \|  |_> > __ \| \_\ \  |  |_|  ||  | |  \  ___/ \___ \
 \___  >____  /   __(____  /___  /__|____/__||__| |__|\___  >____  >
     \/     \/|__|       \/    \/                         \/     \/
```

Capability discovery is how a paymail client learns the features a paymail
service supports, and the endpoint of each. Drawing on RFC 5785, the service
publishes a machine-readable document at /.well-known/bsvalias on the host
named by its SRV record.

Each capability is listed with the BRFC specification it implements, when the
specification is known.

Read more at: https://bitcoin-sv-specs.github.io/brfc-paymail/02-02-capability-discovery.html

```
paymail capabilities <domain|paymail> [flags]
```

### Examples

```
  paymail capabilities relayx.io
  paymail c mrz@relayx.io
```

### Options

```
  -h, --help[=true|false]   help for capabilities
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

* [paymail](paymail.md)	 - Inspect, validate and resolve paymail addresses and domains

