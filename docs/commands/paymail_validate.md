## paymail validate

Validate a paymail address or domain

### Synopsis

```
              .__  .__    .___       __          
___  _______  |  | |__| __| _/____ _/  |_  ____  
\  \/ /\__  \ |  | |  |/ __ |\__  \\   __\/ __ \ 
 \   /  / __ \|  |_|  / /_/ | / __ \|  | \  ___/ 
  \_/  (____  /____/__\____ |(____  /__|  \___  >
            \/             \/     \/          \/
```

Validate a specific paymail address (user@domain.tld) or validate a domain for required paymail capabilities. 

By default, this will check for a SRV record, DNSSEC and SSL for the domain. 

This will also check for required capabilities that all paymail services are required to support.

All these validations are suggestions/requirements from bsvalias spec.

Read more at: http://bsvalias.org/index.html

```
paymail validate [flags]
```

### Examples

```
paymail validate moneybutton.com
paymail v moneybutton.com
```

### Options

```
  -h, --help                help for validate
  -n, --nameserver string   DNS name server for resolving records (default "8.8.8.8")
  -p, --port uint16         Port that is found in the SRV record (default 443)
      --priority uint16     Priority value that is found in the SRV record (default 10)
      --protocol string     Protocol in the SRV record (default "tcp")
  -s, --service string      Service name in the SRV record (default "bsvalias")
  -d, --skip-dnssec         Skip checking DNSSEC of the target domain
      --skip-srv            Skip checking SRV record of the main domain
      --skip-ssl            Skip checking SSL of the target domain
  -w, --weight uint16       Weight value that is found in the SRV record (default 10)
```

### Options inherited from parent commands

```
      --bsvalias string   The bsvalias version (default "1.0")
      --config string     Custom config file (default is $HOME/paymail/config.yaml)
      --docs              Generate docs from all commands (./docs/commands)
      --flush-cache       Flushes ALL cache, empties local database
      --no-cache          Turn off caching for this specific command
  -t, --skip-tracing      Turn off request tracing information
```

### SEE ALSO

* [paymail](paymail.md)	 - Inspect, validate domains or resolve paymail addresses

