// Package demo holds the generator for the throwaway tree of git repositories
// the README's VHS demo films against. The generator is a test — it needs a real
// *testing.T to drive the gitest helpers — gated behind the `fixture` build tag
// so it never runs in the normal suite. Drive it via `just gen-demo-fixture`.
package demo
