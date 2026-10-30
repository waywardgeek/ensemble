# Noise.NET transport dependency

Source: [Metalnem/noise v1.0.0](https://github.com/Metalnem/noise/tree/4279cfa57df028dec4a4cb3ee9579a77a5cc39fa),
commit `4279cfa57df028dec4a4cb3ee9579a77a5cc39fa`, by Nemanja Mijailovic.
The original MIT license is retained in `LICENSE`. This is third-party dependency
source, not an example of Ensemble's Common/implementation layout.

Bill approved preserving the original Gateway wrapper's associated-data contract.
Two upstream files contain narrow API exposures:

- `Transport.cs`: new three-argument overloads forward associated data to the
  existing `CipherState.EncryptWithAd/DecryptWithAd`. The original two-argument
  methods delegate with empty associated data.
- `KeyPair.cs`: `Generate(ReadOnlySpan<byte>)` validates the private-key length and
  delegates to the existing internal `GenerateKeyPair(privateKey)`. The public
  Gateway key-file helper needs this to restore the historical private-only file
  format and derive, rather than independently trust, the public half.

No cipher, curve operation, nonce logic, handshake, KDF, or hash implementation
is changed.
Gateway's actual wire traffic continues to use empty associated data.

`Noise.csproj` is a local build definition, not the upstream packaging project.
It targets .NET 10, whose framework supplies the original supplementary BCL APIs,
and names the modified assembly `Ensemble.Noise` to distinguish it from the NuGet
release. The native dependency is pinned to libsodium 1.0.22; 1.0.16 lacks Mac ARM64
assets. Building this project requires that package in a cache or an explicitly
selected NuGet source. The repository's default disabled feeds are not changed.

Use `Noise_KK_25519_ChaChaPoly_SHA256` for Ensemble Gateway. The earlier isolated
package probe passed on ARM64 macOS and Debian ARM64. That does not audit this
library, prove other platforms, or verify upstream functionality Gateway does not
use. This source copy retains upstream APIs; Gateway need not expose all of them.

## Unused upstream rekey path: review required before use

Source inspection found that upstream `CipherState.Rekey()` calls
`key.Slice(Aead.KeySize).CopyTo(k)` on a 48-byte encryption result: it copies the
16-byte suffix, whereas flynn/noise1.1.0 copies the first32 bytes. This is a concrete
source discrepancy, not a tested rekey-interoperability result. Gateway's existing
wrapper and wire protocol do not call rekey. This approved patch leaves that code
unchanged; do not expose or rely on it without separate review/correction and tests.
The passing transport tests exercise unchanged keys with advancing nonces, not rekey.
