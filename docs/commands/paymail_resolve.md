## paymail resolve

Resolves a paymail address

### Synopsis

```
                            .__               
_______   ____   __________ |  |___  __ ____  
\_  __ \_/ __ \ /  ___/  _ \|  |\  \/ // __ \ 
 |  | \/\  ___/ \___ (  <_> )  |_\   /\  ___/ 
 |__|    \___  >____  >____/|____/\_/  \___  >
             \/     \/                     \/
```

Resolves a paymail address into a hex-encoded Bitcoin script, address and public profile (if found).

Given a sender and a receiver, where the sender knows the receiver's 
paymail handle <alias>@<domain>.<tld>, the sender can perform Service Discovery against 
the receiver and request a payment destination from the receiver's paymail service.

Read more at: http://bsvalias.org/04-01-basic-address-resolution.html

```
paymail resolve [flags]
```

### Examples

```
paymail resolve mrz@moneybutton.com
```

### Options

```
  -a, --amount uint            Amount in satoshis for the payment request
  -h, --help                   help for resolve
  -p, --purpose string         Purpose for the transaction
      --sender-handle string   Sender's paymail handle. Required by bsvalias spec. Receiver paymail used if not specified.
      --sender-name string     The sender's name
  -s, --signature string       The signature of the entire request
      --skip-bitpic            Skip trying to get an associated Bitpic
      --skip-pki               Skip the pki request
      --skip-public-profile    Skip the public profile request
      --skip-roundesk          Skip trying to get an associated Roundesk profile
```

### Options inherited from parent commands

```
      --bsvalias string   The bsvalias version (default "1.0")
      --config string     Config file (default is $HOME/.paymail-inspector.yaml)
      --docs              Generate docs from all commands (./docs/commands)
  -t, --skip-tracing      Turn off request tracing information
```

### SEE ALSO

* [paymail](paymail.md)	 - Inspect, validate domains or resolve paymail addresses

