// Package cli implements the paymail-inspector command-line interface.
//
// The command tree is built fresh for every run by newRootCmd and holds no
// mutable package-level state: everything a command needs (output streams,
// configuration, the response cache, and the paymail client factory) lives on
// an *app value that is passed down explicitly. Production wiring happens in
// Execute; tests build their own app with fakes, so no test ever reaches the
// network.
//
// The binary ships as paymail-inspector and also answers to "paymail" when it
// is invoked through a symlink of that name; help text and examples follow
// the name it was called by.
package cli
