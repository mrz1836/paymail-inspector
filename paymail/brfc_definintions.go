package paymail

// All BRFC IDs that have been used/referenced in the application
const (
	BRFCBasicAddressResolution = "759684b1a19a"       // more info: http://bsvalias.org/04-01-basic-address-resolution.html
	BRFCP2PPaymentDestination  = "2a40af698840"       // more info: https://docs.moneybutton.com/docs/paymail-07-p2p-payment-destination.html
	BRFCP2PTransactions        = "5f1323cddf31"       // more info: https://docs.moneybutton.com/docs/paymail-06-p2p-transactions.html
	BRFCPaymentDestination     = "paymentDestination" // more info: http://bsvalias.org/04-01-basic-address-resolution.html
	BRFCPayToProtocolPrefix    = "7bd25e5a1fc6"       // more info: http://bsvalias.org/04-04-payto-protocol-prefix.html
	BRFCPki                    = "pki"                // more info: http://bsvalias.org/03-public-key-infrastructure.html
	BRFCPkiAlternate           = "0c4339ef99c2"       // more info: http://bsvalias.org/03-public-key-infrastructure.html
	BRFCPublicProfile          = "f12f968c92d6"       // more info: https://github.com/bitcoin-sv-specs/brfc-paymail/pull/7/files
	BRFCReceiverApprovals      = "3d7c2ca83a46"       // more info: http://bsvalias.org/04-03-receiver-approvals.html
	BRFCSenderValidation       = "6745385c3fc0"       // more info: http://bsvalias.org/04-02-sender-validation.html
	BRFCVerifyPublicKeyOwner   = "a9f510c16bde"       // more info: http://bsvalias.org/05-verify-public-key-owner.html
)

// BRFCKnownSpecifications is a running list of all known BRFC specifications
// JSON file was converted into a go:var for binary shipment (todo: use a static file pkg app)
// Add your spec: https://github.com/mrz1836/paymail-inspector/issues/new/choose
var BRFCKnownSpecifications = `
[
  {
   "author": "andy (nChain), Ryan X. Charles (Money Button)",
   "id": "b2aa66e26b43",
   "title": "bsvalias Service Discovery",
   "url": "http://bsvalias.org/02-service-discovery.html",
   "version": "1"
  },
  {
   "alias": "pki",
   "author": "andy (nChain), Ryan X. Charles (Money Button)",
   "id": "0c4339ef99c2",
   "title": "bsvalias Public Key Infrastructure",
   "url": "http://bsvalias.org/03-public-key-infrastructure.html",
   "version": "1"
  },
  {
   "alias": "paymentDestination",
   "author": "andy (nChain), Ryan X. Charles (Money Button)",
   "id": "759684b1a19a",
   "title": "bsvalias Payment Addressing (Basic Address Resolution)",
   "url": "http://bsvalias.org/04-01-basic-address-resolution.html",
   "version": "1"
  },
  {
   "author": "andy (nChain)",
   "id": "6745385c3fc0",
   "title": "bsvalias Payment Addressing (Payer Validation)",
   "url": "http://bsvalias.org/04-02-sender-validation.html",
   "version": "1"
  },
  {
   "author": "andy (nChain)",
   "id": "3d7c2ca83a46",
   "title": "bsvalias Payment Addressing (Payee Approvals)",
   "url": "http://bsvalias.org/04-03-receiver-approvals.html",
   "version": "1"
  },
  {
   "author": "andy (nChain)",
   "id": "7bd25e5a1fc6",
   "title": "bsvalias Payment Addressing (PayTo Protocol Prefix)",
   "url": "http://bsvalias.org/04-04-payto-protocol-prefix.html",
   "version": "1"
  },
  {
   "author": "andy (nChain), Ryan X. Charles (Money Button), Miguel Duarte (Money Button)",
   "id": "a9f510c16bde",
   "title": "bsvalias public key verify (Verify Public Key Owner)",
   "url": "http://bsvalias.org/05-verify-public-key-owner.html",
   "version": "1"
  },
  {
   "author": "Ryan X. Charles (Money Button), Miguel Duarte (Money Button), Rafa Jimenez Seibane (Handcash), Ivan Mlinarić  (Handcash)",
   "id": "5f1323cddf31",
   "title": "P2P Transactions",
   "url": "https://docs.moneybutton.com/docs/paymail-06-p2p-transactions.html",
   "version": "1"
  },
  {
   "author": "Ryan X. Charles (Money Button), Miguel Duarte (Money Button), Rafa Jimenez Seibane (Handcash), Ivan Mlinarić  (Handcash)",
   "id": "2a40af698840",
   "title": "P2P Payment Destination",
   "url": "https://docs.moneybutton.com/docs/paymail-07-p2p-payment-destination.html",
   "version": "1"
  },
  {
   "author": "Ryan X. Charles (Money Button)",
   "id": "f12f968c92d6",
   "title": "Public Profile (Name & Avatar)",
   "url": "https://github.com/bitcoin-sv-specs/brfc-paymail/pull/7/files",
   "version": "1"
  },
  {
   "author": "nChain",
   "id": "ce852c4c2cd1",
   "title": "merchant_api",
   "url": "https://github.com/bitcoin-sv-specs/brfc-merchantapi",
   "version": "0.1"
  },
  {
   "author": "nChain",
   "id": "07f0786cdab6",
   "title": "minerId",
   "url": "https://github.com/bitcoin-sv-specs/brfc-minerid",
   "version": "0.1"
  },
  {
   "author": "nChain",
   "id": "fb567267440a",
   "title": "feeSpec",
   "url": "https://github.com/bitcoin-sv-specs/brfc-misc/tree/master/feespec",
   "version": "0.1"
  },
  {
   "author": "nChain",
   "id": "62b21572ca46",
   "title": "minerIdExt-feeSpec",
   "url": "https://github.com/bitcoin-sv-specs/brfc-minerid/tree/master/extensions/feespec",
   "version": "0.1"
  },
  {
   "author": "nChain",
   "id": "298e080a4598",
   "title": "jsonEnvelope",
   "url": "https://github.com/bitcoin-sv-specs/brfc-misc/tree/master/jsonenvelope",
   "version": "0.1"
  },
  {
   "author": "nChain",
   "id": "1b1d980b5b72",
   "title": "minerIdExt-minerParams",
   "url": "https://github.com/bitcoin-sv-specs/brfc-minerid/tree/master/extensions/minerparams",
   "version": "0.1"
  },
  {
   "author": "nChain",
   "id": "a224052ad433",
   "title": "minerIdExt-blockInfo",
   "url": "https://github.com/bitcoin-sv-specs/brfc-minerid/tree/master/extensions/blockinfo",
   "version": "0.1"
  },
  {
   "author": "nChain",
   "id": "b8930c2bbf5d",
   "title": "minerIdExt-blockBind",
   "url": "https://github.com/bitcoin-sv-specs/brfc-minerid/tree/master/extensions/blockbind",
   "version": "0.1"
  }
]
`
