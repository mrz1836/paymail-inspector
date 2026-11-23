# 🔎&nbsp;&nbsp;Paymail Inspector: Examples

Real output from live paymail providers, captured with color disabled (`NO_COLOR=1`).
Every command is shown as `paymail`, the short name installed next to `paymail-inspector`;
both names run the same binary.

> Request tracing (DNS, TLS, and time-to-first-byte per request) is on by default, as in
> the [`resolve`](#-resolve-an-address) example. The other examples use `-t` / `--skip-tracing`
> to keep the output short. Responses are cached for an hour, which the output marks with `(from cache)`.

<br/>

| Example | Command |
|---------|---------|
| [Capability discovery](#-capability-discovery) | `paymail capabilities handcash.io` |
| [Resolve an address](#-resolve-an-address) | `paymail resolve mrz@relayx.io` |
| [P2P payment destination](#-p2p-payment-destination) | `paymail p2p \$handcash` |
| [Validate a domain](#-validate-a-domain) | `paymail validate relayx.io` |
| [Verify a public key](#-verify-a-public-key) | `paymail verify mrz@simply.cash 022d61…` |
| [Whois across providers](#-whois-across-providers) | `paymail whois mrz` |
| [BRFC specifications](#-brfc-specifications) | `paymail brfc list` / `search` / `generate` |

The complete flag reference for every command is in [`docs/commands`](commands/paymail.md).

<br/>

## 🧭 Capability discovery

Find the host that serves a domain's paymail (its `_bsvalias._tcp` SRV record), then list
every capability it publishes, named after the BRFC specification it implements.
([command reference](commands/paymail_capabilities.md))

```bash
paymail capabilities handcash.io -t
```

<details>
<summary><strong><code>Show output</code></strong></summary>

```text
==========| Retrieving SRV record for handcash.io...
SRV target: cloud.handcash.io:443 --weight 100 --priority 10

==========| Retrieving capabilities for cloud.handcash.io:443...
Found 8 capabilities

==========| Listing 8 capabilities (bsvalias 1.0)...
2a40af698840                 https://cloud.handcash.io/api/bsvalias/p2p-payment-destination/{alias}@{domain.tld}  (P2P Payment Destination)
5c55a7fdb7bb                 https://cloud.handcash.io/api/bsvalias/receive-beef/{alias}@{domain.tld}  (Background Evaluation Extended Format Transaction)
5f1323cddf31                 https://cloud.handcash.io/api/bsvalias/receive-transaction/{alias}@{domain.tld}  (P2P Transactions)
6745385c3fc0                 Disabled  (bsvalias Payment Addressing (Payer Validation))
a9f510c16bde                 https://cloud.handcash.io/api/bsvalias/verifypubkey/{alias}@{domain.tld}/{pubkey}  (bsvalias public key verify (Verify Public Key Owner))
f12f968c92d6                 https://cloud.handcash.io/api/bsvalias/public-profile/{alias}@{domain.tld}  (Public Profile (Name & Avatar))
paymentDestination           https://cloud.handcash.io/api/bsvalias/address/{alias}@{domain.tld}  (bsvalias Payment Addressing (Basic Address ResolutionResponse))
pki                          https://cloud.handcash.io/api/bsvalias/id/{alias}@{domain.tld}  (bsvalias Public Key Infrastructure)
```

</details>

<br/>

## 📬 Resolve an address

Resolve an address to its public key, output script, and address, plus its public profile when
the provider publishes one. `1handle` and `$handle` are shortcuts for RelayX and HandCash
addresses. ([command reference](commands/paymail_resolve.md))

```bash
paymail resolve mrz@relayx.io
```

<details>
<summary><strong><code>Show output</code></strong></summary>

```text
--sender-handle is not set; using the receiver's address mrz@relayx.io

==========| Retrieving SRV record for relayx.io...
SRV target: relayx.io:443 --weight 10 --priority 10

==========| Retrieving capabilities for relayx.io:443...
DNSLookup    | 4.291083ms  | TTFB      | 91.634375ms
TLSHandshake | 64.598542ms | ConnTime  | 95.805416ms
FB to Close  | 111.959µs   | TotalTime | 187.482625ms [200]
Found 3 capabilities

==========| Retrieving public key information for mrz@relayx.io...
DNSLookup     | 0s         | TTFB         | 381.1205ms
TLSHandshake  | 0s         | ConnTime     | 333ns
FB to Close   | 2.257709ms | TotalTime    | 383.378542ms [200]
IsConnWasIdle | true       | ConnIdleTime | 323.75µs
IsConnReused  | true
Found pubkey 0352530c30...

==========| Resolving address for mrz@relayx.io...
DNSLookup     | 0s         | TTFB         | 178.402166ms
TLSHandshake  | 0s         | ConnTime     | 292ns
FB to Close   | 1.156459ms | TotalTime    | 179.558917ms [200]
IsConnWasIdle | true       | ConnIdleTime | 164.625µs
IsConnReused  | true
Resolved output script 76a914da31...

==========| Results for mrz@relayx.io
PubKey       : 0352530c305378fd9dfd99f8c8c44e9092efa7c1674b61d4e9be65f92aa7a77bbe
Output Script: 76a914da31dec79fe8e1d9af38ee883540f286ac99520f88ac
Address      : 1Lti3s6AQNKTSgxnTyBREMa6XdHLBnPSKa
```

</details>

When a provider enforces sender validation (BRFC `6745385c3fc0`), pass the sender with
`--sender-handle`, `--sender-name`, and `--signature`. You can also set the sender once in your
[config file](../config-example.yaml).

<br/>

## 💸 P2P payment destination

Ask the receiver for a set of P2P payment outputs and a reference. `--satoshis` sets the amount
(default 1000). ([command reference](commands/paymail_p2p.md))

```bash
paymail p2p \$handcash -t
```

<details>
<summary><strong><code>Show output</code></strong></summary>

```text
==========| Retrieving SRV record for handcash.io...
SRV target: cloud.handcash.io:443 --weight 100 --priority 10 (from cache)

==========| Retrieving capabilities for cloud.handcash.io:443...
Found 8 capabilities (from cache)

==========| Starting a P2P payment request for handcash@handcash.io...
Found 1 payment output(s)

==========| Retrieving public profile for handcash@handcash.io...
Found public profile [name, avatar]

==========| P2P information for handcash@handcash.io
Name         : HandCash
Avatar       : https://cloud.handcash.io/v3/users/profilePicture/handcash
Reference    : eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzUxMiJ9.eyJ2ZXJzaW…

==========| Output #1
Script       : 76a91408be720e72373edb194d5acb6e448b53ba0bbd9d88ac
Satoshis     : 1000
Address      : 1oEYvVTBh6hkw79ssjKxn1hvKULoa9b2p
```

The reference is a signed token issued by the provider for this request; it is shortened here.

</details>

<br/>

## ✅ Validate a domain

Check a domain, or a single address, against the bsvalias specification: SRV record, DNSSEC,
SSL, and the required capabilities. For an address it also checks the public key. Missing DNSSEC
is a warning. A failed required check makes the command exit non-zero, so you can use it as a CI
gate. ([command reference](commands/paymail_validate.md))

```bash
paymail validate relayx.io -t
```

<details>
<summary><strong><code>Show output</code></strong></summary>

```text
==========| Detecting validation type...
Domain detected: relayx.io

==========| Retrieving SRV record for relayx.io...
SRV record passed all validations (target, port, priority, weight)
SRV target: relayx.io:443 --weight 10 --priority 10

==========| Checking relayx.io for DNSSEC...
DNSSEC not found or invalid for relayx.io; check manually: https://dnsviz.net/d/relayx.io/dnssec/
DNSSEC detail: failed in resolveDomainDS: read udp 192.168.168.222:51072->65.22.161.17:53: i/o timeout

==========| Checking relayx.io for SSL...
SSL certificate is valid for relayx.io

==========| Retrieving capabilities for relayx.io:443...
Found 3 capabilities
Found required capabilities: [pki] [paymentDestination]

==========| relayx.io passed validation (1 warning(s))
```

</details>

Change what the SRV record must contain with `--port`, `--priority`, `--weight`, `--service`, and
`--protocol`. Pick the DNS server with `--nameserver`, and turn off individual checks with
`--skip-srv`, `--skip-dnssec`, or `--skip-ssl`.

<br/>

## 🔑 Verify a public key

Ask the provider whether a public key belongs to an address. The arguments can come in either
order. If the provider reports no match, the command exits non-zero.
([command reference](commands/paymail_verify.md))

```bash
paymail verify mrz@simply.cash 022d613a707aeb7b0e2ed73157d401d7157bff7b6c692733caa656e8e4ed5570ec -t
```

<details>
<summary><strong><code>Show output</code></strong></summary>

```text
==========| Retrieving SRV record for simply.cash...
SRV target: simply.cash:443 --weight 10 --priority 10

==========| Retrieving capabilities for simply.cash:443...
Found 4 capabilities

==========| Verifying pubkey for mrz@simply.cash...

==========| Verify response for mrz@simply.cash
Paymail      : mrz@simply.cash
PubKey       : 022d613a707aeb7b0e2ed73157d401d7157bff7b6c692733caa656e8e4ed5570ec
Paymail & PubKey match! (service responded: match=true)
```

</details>

<br/>

## 🕵️ Whois across providers

Search every public paymail provider in parallel for a handle. The results show who holds it at each
provider, and where it is still free. ([command reference](commands/paymail_whois.md))

```bash
paymail whois mrz
```

<details>
<summary><strong><code>Show output</code></strong></summary>

```text
==========| Searching 5 providers for mrz...

==========| mrz@handcash.io
The handle mrz might be available! Reserve it now: https://handcash.io

==========| Results for mrz@relayx.io
PubKey       : 0352530c305378fd9dfd99f8c8c44e9092efa7c1674b61d4e9be65f92aa7a77bbe

==========| Results for mrz@simply.cash
PubKey       : 022d613a707aeb7b0e2ed73157d401d7157bff7b6c692733caa656e8e4ed5570ec

==========| Results for mrz@paymail.us
Name         : mrz
Avatar       : https://paymail.us/apple-touch-icon.png
PubKey       : 0279be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798

==========| mrz@1sat.app
The handle mrz might be available! Reserve it now: https://1sat.app
```

</details>

<br/>

## 📜 BRFC specifications

BRFC IDs identify paymail extensions in a provider's capabilities. `list` shows every known
specification and checks each ID against the one its title, author, and version generate.
`search` matches on ID, title, or author. `generate` computes the ID for a new specification.
([command reference](commands/paymail_brfc.md))

```bash
paymail brfc list
paymail brfc search nChain
paymail brfc generate --title "BRFC Specifications" --author "andy (nChain)" --version 1
```

<details>
<summary><strong><code>Show output: brfc search nChain</code></strong></summary>

```text
==========| bsvalias Service Discovery v1
ID        : b2aa66e26b43 (invalid ID, expected fe5064beaf74)
Author(s) : andy (nChain), Ryan X. Charles (Money Button)
URL       : https://bitcoin-sv-specs.github.io/brfc-paymail/02-service-discovery.html

==========| bsvalias Public Key Infrastructure v1
ID        : 0c4339ef99c2 (invalid ID, expected 0f9681ab58f0)
Author(s) : andy (nChain), Ryan X. Charles (Money Button)
Alias     : pki
URL       : https://bitcoin-sv-specs.github.io/brfc-paymail/03-public-key-infrastructure.html

==========| bsvalias Payment Addressing (Basic Address ResolutionResponse) v1
ID        : 759684b1a19a (invalid ID, expected 7efcf51b3a1d)
Author(s) : andy (nChain), Ryan X. Charles (Money Button)

...

Total BRFC specifications found: 16 searching: nChain
```

</details>

<details>
<summary><strong><code>Show output: brfc generate</code></strong></summary>

```text
==========| Generated BRFC ID
ID           : 57dd1f54fc67
Title        : BRFC Specifications
Author       : andy (nChain)
Version      : 1
```

`57dd1f54fc67` is the ID that the [BRFC ID assignment spec](https://bitcoin-sv-specs.github.io/brfc-paymail/01-02-brfc-id-assignment.html)
gives for this example.

</details>

> **Why do some original bsvalias IDs show as invalid?** Several early specifications were assigned IDs
> before their final titles and authors were set. `list` reports the ID that the current metadata would
> generate, and keeps the published ID, which providers still advertise. Use `--skip-validation` to hide
> the check.
