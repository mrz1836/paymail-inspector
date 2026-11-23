## paymail resolve

Resolve a paymail address to its pubkey, output script, and profile

### Synopsis

```text
                            .__
_______   ____   __________ |  |___  __ ____
\_  __ \_/ __ \ /  ___/  _ \|  |\  \/ // __ \
 |  | \/\  ___/ \___ (  <_> )  |_\   /\  ___/
 |__|    \___  >____  >____/|____/\_/  \___  >
             \/     \/                     \/
```

Resolve a paymail address into its public key, a hex-encoded Bitcoin output
script and address, and its public profile (when published).

Given a sender and a receiver, where the sender knows the receiver's paymail
handle <alias>@<domain>.<tld>, the sender performs service discovery against
the receiver and requests a payment destination from the receiver's paymail
service.

Read more at: https://bitcoin-sv-specs.github.io/brfc-paymail/04-01-basic-address-resolution.html

```
paymail resolve <paymail> [flags]
```

### Examples

```
  paymail resolve mrz@relayx.io
  paymail r mrz@relayx.io --sender-handle you@relayx.io
  paymail r 1mrz
```

### Options

```
  -a, --amount uint                        Amount in satoshis for the payment request
  -h, --help[=true|false]                  help for resolve
  -p, --purpose string                     Purpose of the payment
      --sender-handle string               Sender's paymail address (default: the receiver's address)
      --sender-name string                 Sender's name
  -s, --signature string                   Signature of the sender request (when the provider enforces sender validation)
      --skip-public-profile[=true|false]   Skip the public profile request
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

