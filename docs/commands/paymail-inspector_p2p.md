## paymail-inspector p2p

Starts a new p2p payment request

### Synopsis


       ________         
______ \_____  \______  
\____ \ /  ____/\____ \ 
|  |_> >       \|  |_> >
|   __/\_______ \   __/ 
|__|           \/__|

This command will start a new p2p request with the receiver and optional amount expected (in Satoshis).

This protocol is an alternative protocol to basic address resolution. 
Instead of returning one address, it returns a list of outputs with a reference number. 
It is only intended to be used with P2P Transactions and will continue to function even 
after basic address resolution is deprecated.

Read more at: https://docs.moneybutton.com/docs/paymail-07-p2p-payment-destination.html

```
paymail-inspector p2p [flags]
```

### Examples

```
paymail-inspector p2p this@address.com
```

### Options

```
  -h, --help            help for p2p
      --satoshis uint   Amount in satoshis for the the incoming transaction(s)
```

### Options inherited from parent commands

```
      --bsvalias string   The bsvalias version (default "1.0")
      --config string     config file (default is $HOME/.paymail-inspector.yaml)
      --docs              Generate docs from all commands (./docs/commands)
```

### SEE ALSO

* [paymail-inspector](paymail-inspector.md)	 - Inspect, validate or resolve paymail domains and addresses

