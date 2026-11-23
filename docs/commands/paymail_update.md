## paymail update

Update paymail-inspector to the latest release

### Synopsis

Download the latest paymail-inspector release, verify its SHA-256 checksum against the
published checksums file, and atomically replace the running binary.

Nothing is written until the download has been verified. A binary that another
installer owns is refused rather than overwritten, with instructions for the
installer that owns it.

```
paymail update [flags]
```

### Examples

```
  paymail update
  paymail update --check
  paymail update --force
```

### Options

```
  -c, --check[=true|false]     Report whether an update is available without installing anything
  -f, --force[=true|false]     Install the latest release even when it is not newer than the running version
  -h, --help[=true|false]      help for update
  -v, --verbose[=true|false]   Narrate each step, and print the release notes with --check
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

