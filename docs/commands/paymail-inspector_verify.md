## paymail-inspector verify

Verifies if a paymail is associated to a pubkey

### Synopsis

```
                   .__  _____       
___  __ ___________|__|/ ____\__.__.
\  \/ // __ \_  __ \  \   __<   |  |
 \   /\  ___/|  | \/  ||  |  \___  |
  \_/  \___  >__|  |__||__|  / ____|
           \/                \/
```

Verify will check the paymail address against a given pubkey using the provider domain (if capability is supported).

This capability allows clients to verify if a given public key is a valid identity key for a given paymail handle.

The public key returned by pki flow for a given paymail handle may change over time. 
This situation may produce troubles to verify data signed using old keys, because even having the keys, 
the verifier doesn't know if the public key actually belongs to the right user.

Read more at: http://bsvalias.org/05-verify-public-key-owner.html

```
paymail-inspector verify [flags]
```

### Examples

```
paymail-inspector verify mrz@simply.cash 022d613a707aeb7b0e2ed73157d401d7157bff7b6c692733caa656e8e4ed5570ec
```

### Options

```
  -h, --help   help for verify
```

### Options inherited from parent commands

```
      --bsvalias string   The bsvalias version (default "1.0")
      --config string     config file (default is $HOME/.paymail-inspector.yaml)
      --docs              Generate docs from all commands (./docs/commands)
```

### SEE ALSO

* [paymail-inspector](paymail-inspector.md)	 - Inspect, validate or resolve paymail domains and addresses

