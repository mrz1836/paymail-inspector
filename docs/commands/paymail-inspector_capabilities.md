## paymail-inspector capabilities

Get the capabilities of the paymail domain

### Synopsis

```
                          ___.   .__.__  .__  __  .__               
  ____ _____  ___________ \_ |__ |__|  | |__|/  |_|__| ____   ______
_/ ___\\__  \ \____ \__  \ | __ \|  |  | |  \   __\  |/ __ \ /  ___/
\  \___ / __ \|  |_> > __ \| \_\ \  |  |_|  ||  | |  \  ___/ \___ \ 
 \___  >____  /   __(____  /___  /__|____/__||__| |__|\___  >____  >
     \/     \/|__|       \/    \/                         \/     \/
```

This command will return the capabilities for a given paymail domain.

Capability Discovery is the process by which a paymail client learns the supported 
features of a paymail service and their respective endpoints and configurations.

Drawing inspiration from RFC 5785 and IANA's Well-Known URIs resource, the Capability Discovery protocol 
dictates that a machine-readable document is placed in a predictable location on a web server.

Read more at: http://bsvalias.org/02-02-capability-discovery.html

```
paymail-inspector capabilities [flags]
```

### Examples

```
paymail-inspector capabilities simply.cash
```

### Options

```
  -h, --help   help for capabilities
```

### Options inherited from parent commands

```
      --bsvalias string   The bsvalias version (default "1.0")
      --config string     config file (default is $HOME/.paymail-inspector.yaml)
      --docs              Generate docs from all commands (./docs/commands)
```

### SEE ALSO

* [paymail-inspector](paymail-inspector.md)	 - Inspect, validate or resolve paymail domains and addresses

