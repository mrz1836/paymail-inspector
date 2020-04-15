## paymail p2p

Starts a new P2P payment request

### Synopsis

```
       ________         
______ \_____  \______  
\____ \ /  ____/\____ \ 
|  |_> >       \|  |_> >
|   __/\_______ \   __/ 
|__|           \/__|
```

This command will start a new P2P request with the receiver and optional amount expected (in Satoshis).

This protocol is an alternative protocol to basic address resolution. 
Instead of returning one address, it returns a list of outputs with a reference number. 
It is only intended to be used with P2P Transactions and will continue to function even 
after basic address resolution is deprecated.

Read more at: https://docs.moneybutton.com/docs/paymail-07-p2p-payment-destination.html

```
paymail p2p [flags]
```

### Examples

```
paymail p2p mrz@moneybutton.com
```

### Options

```
  -h, --help            help for p2p
      --satoshis uint   Amount in satoshis for the the incoming transaction(s)
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

