## Paymail Inspector: Examples & Docs
Below are some examples using the **paymail** cli app

#### View All Commands (Help)
```bash
$ paymail
```
<img src="../.github/IMAGES/help-command.gif?raw=true&v=5" alt="Help Command">

Global flags for the entire application [(view command specs)](commands/paymail.md)
```
  -h, --help              help for paymail inspector
  -v, --version           version for paymail inspector
      --bsvalias string   The bsvalias version (default: 1.0)
      --config string     config file (default is $HOME/.paymail-inspector.yaml)
```

___


#### List BRFC Specifications
```bash
$ paymail brfc list
```
<details>
<summary><strong><code>Show Example</code></strong></summary>

<img src="../.github/IMAGES/brfc-list-command.gif?raw=true&v=5" alt="BRFC List Command">
</details>

Custom flags for the brfc:list command [(view command specs)](commands/paymail_brfc.md)
```
  -h, --help              help for brfc
      --skip-validation   Skip validating the existing BRFC IDs
```

___

#### Generate new BRFC ID
```bash
$ paymail brfc generate --title "BRFC Specifications" --author "andy (nChain)" --version 1
```
<img src="../.github/IMAGES/brfc-generate-command.gif?raw=true&v=5" alt="BRFC Generate Command">

Custom flags for the brfc:generate command [(view command specs)](commands/paymail_brfc.md)
```
      --author string     Author(s) new BRFC specification
  -h, --help              help for brfc
      --title string      Title of the new BRFC specification
      --version string    Version of the new BRFC specification
```

___

#### Search BRFC Specifications
```bash
$ paymail brfc search nChain
```
<img src="../.github/IMAGES/brfc-search-command.gif?raw=true&v=5" alt="BRFC Search Command">

Custom flags for the brfc:search command [(view command specs)](commands/paymail_brfc.md)
```
  -h, --help              help for brfc
      --skip-validation   Skip validating the existing BRFC IDs
```

___

#### Get Capabilities (by Domain)
```bash
$ paymail capabilities moneybutton.com
```
<img src="../.github/IMAGES/capabilities-command.gif?raw=true&v=5" alt="Capabilities Command">

Custom flags for the capabilities request [(view command specs)](commands/paymail_capabilities.md)
```
  -h, --help              help for capabilities
```

___

#### Start P2P Payment Request (by Paymail)
```bash
$ paymail p2p mrz@moneybutton.com
```
<img src="../.github/IMAGES/p2p-command.gif?raw=true&v=5" alt="P2P Command">

Custom flags for the p2p command [(view command specs)](commands/paymail_p2p.md)
```
  -h, --help              help for p2p
      --satoshis uint     Amount in satoshis for the payment
```

___

#### Resolve Paymail Address (by Paymail)
```bash
$ paymail resolve mrz@moneybutton.com
```
<img src="../.github/IMAGES/resolve-command.gif?raw=true&v=5" alt="Resolve Command">

Custom flags for the resolve command [(view command specs)](commands/paymail_resolve.md)
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
$ paymail validate moneybutton.com
```
<img src="../.github/IMAGES/validate-command.gif?raw=true&v=5" alt="Validate Command">

Custom flags for the validation command [(view command specs)](commands/paymail_validate.md)
```
  -h, --help                help for validate
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
$ paymail verify mrz@moneybutton.com 02ead23149a1e33df17325ec7a7ba9e0b20c674c57c630f527d69b866aa9b65b10
```
<img src="../.github/IMAGES/verify-command.gif?raw=true&v=5" alt="Verify Command">

Custom flags for the verify command [(view command specs)](commands/paymail_verify.md)
```
  -h, --help              help for verify
```