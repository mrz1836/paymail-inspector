## Paymail Inspector: Examples & Docs
Below are some examples using the **paymail** cli app

### View All Commands (Help)
```shell script
$ paymail
```
<img src="../.github/IMAGES/help-command.gif?raw=true&v=7" alt="Help Command">

Global flags for the entire application [(command specs)](commands/paymail.md)
```text
      --bsvalias string   The bsvalias version (default "1.0")
      --config string     Custom config file (default is $HOME/paymail/config.yaml)
      --docs              Generate docs from all commands (./docs/commands)
      --flush-cache       Flushes ALL cache, empties local database
  -h, --help              help for paymail
      --no-cache          Turn off caching for this specific command
  -t, --skip-tracing      Turn off request tracing information
  -v, --version           version for paymail
```

___


### List BRFC Specifications
```shell script
$ paymail brfc list
```
<details>
<summary><strong><code>Show Example</code></strong></summary>

<img src="../.github/IMAGES/brfc-list-command.gif?raw=true&v=7" alt="BRFC List Command">
</details>

Custom flags for the brfc:list command [(command specs)](commands/paymail_brfc.md)
```text
  -h, --help              help for brfc
      --skip-validation   Skip validating the existing BRFC IDs
```

___

### Generate new BRFC ID
```shell script
$ paymail brfc generate --title "BRFC Specifications" --author "andy (nChain)" --version 1
```
<details>
<summary><strong><code>Show Example</code></strong></summary>

<img src="../.github/IMAGES/brfc-generate-command.gif?raw=true&v=7" alt="BRFC Generate Command">
</details>

<details>
<summary><strong><code>Test Cases</code></strong></summary>

Expected ID: `57dd1f54fc67`
```shell script
$ paymail brfc generate --title "BRFC Specifications" --author "andy (nChain)" --version 1
```

Expected ID: `74524c4d6274`
```shell script
$ paymail brfc generate --title "bsvalias Payment Addressing (PayTo Protocol Prefix)" --author "andy (nChain)" --version 1
```

Expected ID: `0036f9b8860f`
```shell script
$ paymail brfc generate --title "bsvalias Integration with Simplified Payment Protocol" --author "andy (nChain)" --version 1
```

</details>

Custom flags for the brfc:generate command [(command specs)](commands/paymail_brfc.md)
```text
      --author string     Author(s) new BRFC specification
  -h, --help              help for brfc
      --title string      Title of the new BRFC specification
      --version string    Version of the new BRFC specification
```

___

### Search BRFC Specifications
```shell script
$ paymail brfc search nChain
```
<details>
<summary><strong><code>Show Example</code></strong></summary>

<img src="../.github/IMAGES/brfc-search-command.gif?raw=true&v=7" alt="BRFC Search Command">
</details>


Custom flags for the brfc:search command [(command specs)](commands/paymail_brfc.md)
```text
  -h, --help              help for brfc
      --skip-validation   Skip validating the existing BRFC IDs
```

___

### Get Capabilities (by Domain)
```shell script
$ paymail capabilities moneybutton.com
```
<details>
<summary><strong><code>Show Example</code></strong></summary>

<img src="../.github/IMAGES/capabilities-command.gif?raw=true&v=7" alt="Capabilities Command">
</details>

Custom flags for the capabilities request [(command specs)](commands/paymail_capabilities.md)
```text
  -h, --help              help for capabilities
```

___

### Start P2P Payment Request (by Paymail)
```shell script
$ paymail p2p mrz@moneybutton.com
```
<details>
<summary><strong><code>Show Example</code></strong></summary>

<img src="../.github/IMAGES/p2p-command.gif?raw=true&v=7" alt="P2P Command">
</details>

Custom flags for the p2p command [(command specs)](commands/paymail_p2p.md)
```text
  -h, --help              help for p2p
      --satoshis uint     Amount in satoshis for the payment
```

___

### Resolve Paymail Address (by Paymail)
```shell script
$ paymail resolve mrz@moneybutton.com
```
<details>
<summary><strong><code>Show Example</code></strong></summary>

<img src="../.github/IMAGES/resolve-command.gif?raw=true&v=7" alt="Resolve Command">
</details>

Custom flags for the resolve command [(command specs)](commands/paymail_resolve.md)
```text
  -a, --amount uint            Amount in satoshis for the payment request
  -h, --help                   help for resolve
  -p, --purpose string         Purpose for the transaction
      --sender-handle string   The sender's paymail handle (if not given it will be the receivers address)
  -n, --sender-name string     The sender's name
  -s, --signature string       The signature of the entire request
      --skip-bitpic            Skip trying to get an associated Bitpic
      --skip-pki               Skip the pki request
      --skip-public-profile    Skip the public profile request
      --skip-roundesk          Skip trying to get an associated Roundesk profile
```

___

### Validate Paymail Setup (by Paymail or Domain)
```shell script
$ paymail validate moneybutton.com
```
<details>
<summary><strong><code>Show Example</code></strong></summary>

<img src="../.github/IMAGES/validate-command.gif?raw=true&v=7" alt="Validate Command">
</details>

Custom flags for the validation command [(command specs)](commands/paymail_validate.md)
```text
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

### Verify Public Key Owner
```shell script
$ paymail verify mrz@moneybutton.com 02ead23149a1e33df17325ec7a7ba9e0b20c674c57c630f527d69b866aa9b65b10
```
<details>
<summary><strong><code>Show Example</code></strong></summary>

<img src="../.github/IMAGES/verify-command.gif?raw=true&v=7" alt="Verify Command">
</details>

Custom flags for the verify command [(command specs)](commands/paymail_verify.md)
```text
  -h, --help              help for verify
```

___

### Whois For Handles
```shell script
$ paymail whois mrz
```
<details>
<summary><strong><code>Show Example</code></strong></summary>

<img src="../.github/IMAGES/whois-command.gif?raw=true&v=7" alt="Whois Command">
</details>

Custom flags for the whois command [(command specs)](commands/paymail_whois.md)
```text
  -h, --help              help for whois
```