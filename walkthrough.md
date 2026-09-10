# Day 7: Decision Trace UI
The front-end has been constructed as a Vite React application (`ui/`) demonstrating the Decision Trace capabilities.

## Technical Details
- Uses `React` and `Vite` for speed.
- Hits the Go Gateway's new `GET /v1/audit/logs` API endpoint every 2 seconds to poll for trace events.
- Displays events sorted strictly chronologically (descending).
- Parses the complex JSON metadata:
  1. **Identity**: The calling actor.
  2. **PII Processing**: Categorization (clean vs restricted) and the exact tokenized payload output by Presidio.
  3. **Model Routing**: Whether the AI model invocation was `local_model`, `external_model`, or `Bypassed/Failed`.
  4. **Policy Engine**: Exact rule matched by the deterministic engine.
  5. **Final Decision**: Visual badge indicating ALLOW/DENY/REDACT.

## Verification
1. I populated Postgres with a barrage of normal, adversarial, and edge-case (fail-closed) events by running all 3 PowerShell test scripts.
2. Visit `http://localhost:5173` to view the UI.
3. Every test execution is explicitly captured and rendered with absolute fidelity to the backend state. There are no mocks.
