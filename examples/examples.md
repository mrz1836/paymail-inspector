## Paymail Inspector: Examples & Docs
Below are some examples using **paymail-inspector**

#### View All Commands (Help)
```bash
$ paymail-inspector -h
```
<img src="../.github/IMAGES/help-command.gif?raw=true&v=2" alt="Help Command">

Global flags for the entire application
```
  -h, --help              help for paymail-inspector
  -v, --version           version for paymail-inspector
      --bsvalias string   The bsvalias version (default: 1.0)
      --config string     config file (default is $HOME/.paymail-inspector.yaml)
```

___

#### Get Capabilities (by Domain)
```bash
$ paymail-inspector capabilities simply.cash
```
<img src="../.github/IMAGES/capabilities-command.gif?raw=true&v=2" alt="Capabilities Command">

___

#### Resolve Paymail Address (by Paymail)
```bash
$ paymail-inspector resolve mrz@simply.cash
```
<img src="../.github/IMAGES/resolve-command.gif?raw=true&v=1" alt="Resolve Command">

Custom flags for resolving or starting a payment request
```
  -a, --amount uint            Amount in satoshis for the payment request
  -h, --help                   help for resolve
  -p, --purpose string         Purpose for the transaction
      --sender-handle string   The sender's paymail handle (if not given it will be the receivers address)
  -n, --sender-name string     The sender's name
  -s, --signature string       The signature of the entire request
      --skip-pki               Skip firing pki request and getting the pubkey
      --skip-public-profile    Skip firing public profile request and getting the avatar
```

___

#### Validate Paymail Setup (by Paymail or Domain)
```bash
$ paymail-inspector validate simply.cash --skip-dnssec
```
<img src="../.github/IMAGES/validate-command.gif?raw=true&v=2" alt="Validate Command">

Custom flags for configuring the validation (enable/disable checks)
```
  -n, --nameserver string   DNS name server for resolving records (default "8.8.8.8")
  -p, --port int            Port that is found in the SRV record (default 443)
      --priority int        Priority value that is found in the SRV record (default 10)
      --protocol string     Protocol in the SRV record (default "tcp")
  -s, --service string      Service name in the SRV record (default "bsvalias")
  -d, --skip-dnssec         Skip checking DNSSEC of the target
      --skip-ssl            Skip checking SSL of the target
      --skip-srv            Skip checking SRV record of the main domain
  -w, --weight int          Weight value that is found in the SRV record (default 10)
```

___

#### Verify Public Key Owner
```bash
$ paymail-inspector verify mrz@simply.cash 022d613a707aeb7b0e2ed73157d401d7157bff7b6c692733caa656e8e4ed5570ec
```
<img src="../.github/IMAGES/verify-command.gif?raw=true&v=2" alt="Verify Command">