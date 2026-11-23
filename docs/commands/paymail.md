## paymail

Inspect, validate and resolve paymail addresses and domains

### Synopsis

```text
__________                             .__.__    .___                                     __
\______   \_____  ___.__. _____ _____  |__|  |   |   | ____   ____________   ____   _____/  |_  ___________
 |     ___/\__  \<   |  |/     \\__  \ |  |  |   |   |/    \ /  ___/\____ \_/ __ \_/ ___\   __\/  _ \_  __ \
 |    |     / __ \\___  |  Y Y  \/ __ \|  |  |__ |   |   |  \\___ \ |  |_> >  ___/\  \___|  | (  <_> )  | \/
 |____|    (____  / ____|__|_|  (____  /__|____/ |___|___|  /____  >|   __/ \___  >\___  >__|  \____/|__|
                \/\/          \/     \/                   \/     \/ |__|        \/     \/
```

Inspect, validate and resolve paymail addresses, discover a provider's
capabilities, and explore the BRFC specifications behind them.

Docs: https://github.com/mrz1836/paymail-inspector

```
paymail [flags]
```

### Examples

```
  paymail resolve mrz@relayx.io
  paymail capabilities relayx.io
  paymail validate relayx.io
  paymail whois mrz
```

### Options

```
      --bsvalias string             The bsvalias version to expect from providers (default "1.0")
      --config string               Config file (default: <user config dir>/paymail-inspector/config.yaml)
      --flush-cache[=true|false]    Empty the local response cache before running
  -h, --help[=true|false]           help for paymail
      --no-cache[=true|false]       Skip cached responses for this command (fresh results are still cached)
      --no-color[=true|false]       Disable colored output (NO_COLOR is honored too)
  -t, --skip-tracing[=true|false]   Hide per-request network timing
```

### SEE ALSO

* [paymail brfc](paymail_brfc.md)	 - List, search, or generate BRFC specification IDs
* [paymail capabilities](paymail_capabilities.md)	 - List the capabilities a paymail domain publishes
* [paymail completion](paymail_completion.md)	 - Generate the autocompletion script for the specified shell
* [paymail p2p](paymail_p2p.md)	 - Start a P2P payment request and list its outputs
* [paymail resolve](paymail_resolve.md)	 - Resolve a paymail address to its pubkey, output script, and profile
* [paymail update](paymail_update.md)	 - Update paymail-inspector to the latest release
* [paymail validate](paymail_validate.md)	 - Validate a paymail domain or address (SRV, DNSSEC, SSL, capabilities)
* [paymail verify](paymail_verify.md)	 - Verify that a pubkey belongs to a paymail address
* [paymail version](paymail_version.md)	 - Show version, commit, and build information
* [paymail whois](paymail_whois.md)	 - Find a handle across the public paymail providers

