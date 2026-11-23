## paymail verify

Verify that a pubkey belongs to a paymail address

### Synopsis

```text
                   .__  _____
___  __ ___________|__|/ ____\__.__.
\  \/ // __ \_  __ \  \   __<   |  |
 \   /\  ___/|  | \/  ||  |  \___  |
  \_/  \___  >__|  |__||__|  / ____|
           \/                \/
```

Ask the provider of a paymail address whether a public key is a valid identity
key for it (when the provider supports the capability). The arguments may be
given in either order.

The key returned by pki for a handle can change over time, which makes data
signed with an old key hard to verify: holding the key is not proof that it
belonged to the user. This capability answers exactly that question.

The command exits non-zero when the provider reports no match.

Read more at: https://bitcoin-sv-specs.github.io/brfc-paymail/05-verify-public-key-owner.html

```
paymail verify <paymail> <pubkey> [flags]
```

### Examples

```
  paymail verify mrz@simply.cash 022d613a707aeb7b0e2ed73157d401d7157bff7b6c692733caa656e8e4ed5570ec
  paymail verify 022d613a707aeb7b0e2ed73157d401d7157bff7b6c692733caa656e8e4ed5570ec mrz@simply.cash
```

### Options

```
  -h, --help[=true|false]   help for verify
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

