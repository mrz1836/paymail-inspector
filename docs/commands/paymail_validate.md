## paymail validate

Validate a paymail domain or address (SRV, DNSSEC, SSL, capabilities)

### Synopsis

```text
              .__  .__    .___       __
___  _______  |  | |__| __| _/____ _/  |_  ____
\  \/ /\__  \ |  | |  |/ __ |\__  \\   __\/ __ \
 \   /  / __ \|  |_|  / /_/ | / __ \|  | \  ___/
  \_/  (____  /____/__\____ |(____  /__|  \___  >
            \/             \/     \/          \/
```

Validate a paymail domain, or a specific paymail address (alias@domain.tld),
against the bsvalias specification:

  1. SRV record   _bsvalias._tcp.<domain> (falls back to <domain>:443)
  2. DNSSEC       on the SRV target (recommended, reported as a warning)
  3. SSL          a valid certificate on the SRV target
  4. Capabilities the required pki and paymentDestination capabilities
  5. PKI          for an address: its public key resolves

The command exits non-zero when any required check fails, so it can gate CI.

Read more at: https://tsc.bsvblockchain.org/standards/paymail/

```
paymail validate <domain|paymail> [flags]
```

### Examples

```
  paymail validate relayx.io
  paymail v mrz@relayx.io
  paymail v relayx.io --skip-dnssec --nameserver 1.1.1.1
```

### Options

```
  -h, --help[=true|false]          help for validate
  -n, --nameserver string          DNS name server used for the lookups (default "8.8.8.8")
  -p, --port uint16                Expected port in the SRV record (default 443)
      --priority uint16            Expected priority in the SRV record (default 10)
      --protocol string            Expected protocol in the SRV record (default "tcp")
  -s, --service string             Expected service name in the SRV record (default "bsvalias")
  -d, --skip-dnssec[=true|false]   Skip the DNSSEC check
      --skip-srv[=true|false]      Skip the SRV record check
      --skip-ssl[=true|false]      Skip the SSL check
  -w, --weight uint16              Expected weight in the SRV record (default 10)
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

