## paymail brfc

List, search, or generate BRFC specification IDs

### Synopsis

```text
___.           _____
\_ |__________/ ____\____
 | __ \_  __ \   __\/ ___\
 | \_\ \  | \/|  | \  \___
 |___  /__|   |__|  \___  >
     \/                 \/
```

BRFC (Bitcoin SV Request-For-Comments) specifications describe functionality
across the ecosystem. The bsvalias protocols behind paymail are defined across
a series of BRFC documents, and BRFC IDs are how a provider advertises the
extensions it supports during capability discovery.

  list      show every known BRFC specification
  search    find specifications by ID, title, or author
  generate  compute the BRFC ID for a new specification

Read more at: https://bitcoin-sv-specs.github.io/brfc-paymail/01-brfc-specifications.html

```
paymail brfc [flags]
```

### Examples

```
  paymail brfc list
  paymail brfc search nChain
  paymail brfc generate --title "BRFC Specifications" --author "andy (nChain)" --version 1
```

### Options

```
  -h, --help[=true|false]              help for brfc
      --skip-validation[=true|false]   Skip validating the IDs of existing BRFC specifications
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
* [paymail brfc generate](paymail_brfc_generate.md)	 - Generate the BRFC ID for a new specification
* [paymail brfc list](paymail_brfc_list.md)	 - List all known BRFC specifications
* [paymail brfc search](paymail_brfc_search.md)	 - Search BRFC specifications by ID, title, or author

