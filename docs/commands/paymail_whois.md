## paymail whois

Find a handle across the public paymail providers

### Synopsis

```text
        .__           .__
__  _  _|  |__   ____ |__| ______
\ \/ \/ /  |  \ /  _ \|  |/  ___/
 \     /|   Y  (  <_> )  |\___ \
  \/\_/ |___|  /\____/|__/____  >
             \/               \/
```

Search 5 public paymail providers for a handle, in parallel, and show who
owns it on each one, or where it is still available.

The handle may be bare (mrz), a provider shortcut ($handle for HandCash,
1handle for RelayX), or a full paymail address, whose alias is searched.

```
paymail whois <handle> [flags]
```

### Examples

```
  paymail whois mrz
  paymail w 1mrz
  paymail w \$handcash
```

### Options

```
  -h, --help[=true|false]                  help for whois
      --skip-public-profile[=true|false]   Skip the public profile requests
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

