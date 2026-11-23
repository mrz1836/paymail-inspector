## paymail brfc generate

Generate the BRFC ID for a new specification

```
paymail brfc generate [flags]
```

### Examples

```
  paymail brfc generate --title "BRFC Specifications" --author "andy (nChain)" --version 1
```

### Options

```
      --author string       Author(s) of the new BRFC specification
  -h, --help[=true|false]   help for generate
      --title string        Title of the new BRFC specification (required)
      --version string      Version of the new BRFC specification
```

### Options inherited from parent commands

```
      --bsvalias string                The bsvalias version to expect from providers (default "1.0")
      --config string                  Config file (default: <user config dir>/paymail-inspector/config.yaml)
      --flush-cache[=true|false]       Empty the local response cache before running
      --no-cache[=true|false]          Skip cached responses for this command (fresh results are still cached)
      --no-color[=true|false]          Disable colored output (NO_COLOR is honored too)
  -t, --skip-tracing[=true|false]      Hide per-request network timing
      --skip-validation[=true|false]   Skip validating the IDs of existing BRFC specifications
```

### SEE ALSO

* [paymail brfc](paymail_brfc.md)	 - List, search, or generate BRFC specification IDs

