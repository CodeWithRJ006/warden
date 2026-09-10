# Warden

> "Warden is a deterministic policy boundary between AI agents and financial tool access."

## Overview
Warden is a policy enforcement gateway that sits between AI agents and financial tools/models. It intercepts proposed actions, tokenizes sensitive data (PII/PAN), enforces deterministic authorization rules, routes to AI models safely, and maintains a strict audit trail.

*This project is a production-shaped, not production-ready prototype built to demonstrate architectural discipline and secure system design.*

## Architecture
**UNTRUSTED**
AI models, agents, external responses
↓
**TRUST BOUNDARY (Warden Gateway)**
1. Identity & Rate Limiting (Redis)
2. PII Detection & Tokenization (Presidio/Vault)
3. Deterministic Policy Engine (Pure Go)
4. Model Routing (Local/External fail-closed)
5. Audit Logging (Structured JSON)
↓
**TRUSTED**
Financial Tools (Razorpay API mock)

## Core Demos
Run `docker compose up` to start the infrastructure.
1. **Normal request**: Allowed based on explicit role bounds (`finance-operator`, `≤ ₹10k`).
2. **Sensitive request**: PII automatically replaced with deterministic tokens (`[CARD_001]`).
3. **Unauthorized refund**: Denied deterministically before touching the tool.
4. **Local model failure + Sensitive**: **FAIL CLOSED**. No fallback to external providers for sensitive data.

## Design Decisions (ADRs)
- **Why Go**: Strict typing, fast startup, pure interface-driven design for the policy engine.
- **Why Redis**: Shared state for rate-limiting and idempotency across gateway instances.
- **Why Tokenize not Redact**: Tokenizing into a vault allows authorized services to reverse the token later if policy permits, whereas redaction destroys the data permanently.
- **Why Fail-Closed**: In financial context, unavailability is preferable to an unsafe execution or PII leak to an external model.

## Evaluation
A 40-case evaluation suite proves the strictness of the boundary across:
- Security (Rate limit bounds)
- Authorization (Role constraints)
- Reliability (Timeout behaviors)
- Routing (Sensitive data isolation)

```json
{
  "total": 40,
  "passed": 40,
  "unsafe_allows": 0,
  "pii_exposures": 0,
  "unauthorized_tool_calls": 0
}
```

## Known Limitations & Production Evolution
- PII detection currently utilizes deterministic fallback regexes to mock the Presidio HTTP boundary.
- Distributed vault requires external KMS integration before handling real PAN data.
- Does not currently implement live Razorpay MCP connections.

---
*Built with AI pair-programming tools; every architectural decision and the security model were designed and validated by me.*