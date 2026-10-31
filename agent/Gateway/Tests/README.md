# Gateway Noise transport checks

This executable tests the dependency extensions in `ThirdParty/Noise` and the
public key/Noise layer in `Ensemble.Gateway`. It does not yet test a Gateway
router, WebSocket client, queues, adapter, or deployment.
The only suite exercised here is `Noise_KK_25519_ChaChaPoly_SHA256`.

The first three groups test real Noise.NET/libsodium operations:

- Both directions, legacy/new empty-AD interoperability, nonempty and binary AD,
  empty/Unicode/maximum-size payloads, buffer limits, and disposal.
- Wrong or omitted AD, altered authentication tag and replay refusal. Rejected
  authentication does not consume the receive nonce; subsequent correct input
  authenticates. These are low-level cipher tests, **not permission to continue a
  Gateway connection after an authentication failure**. The existing Gateway
  policy closes that connection.
- Incorrect known static keys and a mismatched application prologue refuse the KK
  handshake.

Three additional local groups test the public layer:

- Immutable identity values, existing-private-key import, public encoding and
  malformed/incorrect-length refusal.
- Private-only key files, twelve concurrent creators, fresh owner-only modes,
  restoration, bounded reads and refusal without overwriting damaged files. These
  use generated test identities and owned temporary directories, not credentials.
- Public session lifecycle, maximum-size payloads, idempotent disposal and terminal
  authentication failure. Unlike the low-level cipher, a failed public session
  cannot resume. The caller must close its socket and make a new session.

With `--interop`, two more groups exchange actual bytes through the public C#
wrapper and existing Go Gateway wrapper, in both role arrangements. C# imports its
private identity through the new factory. Each side sends empty, Unicode and
4096-byte payloads with empty, textual and binary associated data. The peer owns
its throwaway keys, communicates over pipes, and exits on EOF. Tests bound response
waits and join/terminate only their own process.

Private-file publication uses an exclusive hard link on macOS/Linux; it refuses
on other platforms or unsupported filesystems rather than weaken privacy or
atomicity. A retained baseline found that `File.Move(..., overwrite:false)` on this
Mac could overwrite a concurrent creator. The Unix runtime's check-then-rename
path explains why that API did not supply the exclusive publication needed here.
Existing parent permissions are not changed, and no directory-fsync/power-loss
claim is made.

## Reproduce locally

Use an explicitly selected installed .NET10 SDK and Go toolchain. Keep caches,
HOME and temporary files in an owned scratch directory as in the chapter process.
The repository disables package feeds by default. The libsodium1.0.22 package must
already be cached, or restore must explicitly name an approved package source.
No system library installation is needed.

Build `NoiseTests.csproj`, then run its DLL without arguments for the six local
groups. To add interoperability, prepare the peer in a temporary workspace from
the repository root (the original Go wrapper remains unchanged):

```sh
set -eu
repo="$PWD"
scratch=$(mktemp -d)
cp "$repo/agent/Gateway/Tests/Interop/"* "$scratch/"
mkdir "$scratch/gateway"
cp "$repo/solutions/ch26/gateway/noise.go" "$scratch/gateway/"
(cd "$scratch" && GOWORK=off go build -mod=readonly -o "$scratch/go-peer" .)
dotnet "$repo/agent/Gateway/Tests/bin/Release/net10.0/NoiseTests.dll" \
  --interop "$scratch/go-peer"
```

This is an independent implementation cross-check, not a cryptographic audit.
Do not generalize it to other Noise patterns, AES-GCM, or untested upstream APIs.
Cross-publish success is packaging evidence; runtime results must be obtained on
the actual destination. Remote execution requires explicit authorization and must
use owned temporary paths, not an installed Gateway or other running services.
