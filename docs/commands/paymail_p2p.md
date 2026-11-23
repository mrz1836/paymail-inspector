## paymail p2p

Start a P2P payment request and list its outputs

### Synopsis

```text
       ________
______ \_____  \______
\____ \ /  ____/\____ \
|  |_> >       \|  |_> >
|   __/\_______ \   __/
|__|           \/__|
```

Start a P2P payment destination request with the receiver for an amount in
satoshis, and list the outputs (script, satoshis, address) and the reference
it returns.

P2P payment destination is the successor to basic address resolution: rather
than one address it returns a set of outputs for a P2P transaction.

Read more at: https://tsc.bsvblockchain.org/standards/paymail/ (BRFC 2a40af698840)

```
paymail p2p <paymail> [flags]
```

### Examples

```
  paymail p2p \$handcash
  paymail p2p \$handcash --satoshis 5000
  paymail p2p handcash@handcash.io
```

### Options

```
  -h, --help[=true|false]                  help for p2p
      --satoshis uint                      Amount in satoshis to request (default 1000)
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

