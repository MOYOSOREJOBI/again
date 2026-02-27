# Dead code and dependency pruning attempt

- Attempted: `go install golang.org/x/tools/cmd/deadcode@latest`
- Result: blocked by network/proxy (`Forbidden` from proxy.golang.org).
- Attempted: `go mod tidy`
- Result: blocked by network/proxy while fetching modules.

Given the environment limitations, no provably unreachable code was deleted in this pass.
