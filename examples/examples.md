## Paymail Inspector Examples
Below are some examples using **paymail-inspector**

#### View All Commands (Help)
```bash
$ paymail-inspector -h
```
<img src="../.github/IMAGES/help-command.gif?raw=true" height="350" width="500" alt="Help Command">

___

#### Get Capabilities (by Domain)
```bash
$ paymail-inspector capabilities moneybutton.com
```
<img src="../.github/IMAGES/capabilities-command.gif?raw=true" height="350" width="500" alt="Capabilities Command">

___

#### Resolve Paymail Address (by Paymail)
```bash
$ paymail-inspector resolve this@address.com --sender-handle you@yourdomain.com
```
<img src="../.github/IMAGES/resolve-command.gif?raw=true" height="350" width="500" alt="Resolve Command">

Custom flags for creating the "sender request":
```
  -a, --amount uint            Amount in satoshis for the payment request
  -h, --help                   help for resolve
  -p, --purpose string         Purpose for the transaction
      --sender-handle string   (Required) The sender's paymail handle
  -n, --sender-name string     The sender's name
  -s, --signature string       The signature of the entire request
```

___

#### Validate Paymail Setup (by Paymail or Domain)
```bash
$ paymail-inspector validate moneybutton.com --priority 1 --skip-dnssec
```
<img src="../.github/IMAGES/validate-command.gif?raw=true" height="350" width="500" alt="Validate Command">

Custom flags for configuring the validation (enable/disable checks)
```
  -h, --help                help for validate
  -n, --nameserver string   DNS name server for resolving records (default "8.8.8.8")
  -p, --port int            Port that is found in the SRV record (default 443)
      --priority int        Priority value that is found in the SRV record (default 10)
      --protocol string     Protocol in the SRV record (default "tcp")
  -s, --service string      Service name in the SRV record (default "bsvalias")
  -d, --skip-dnssec         Skip checking DNSSEC of the target
      --skip-ssl            Skip checking SSL of the target
  -w, --weight int          Weight value that is found in the SRV record (default 10)
```