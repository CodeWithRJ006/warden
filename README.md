# Warden

> "Warden is a production-shaped security control-plane prototype, acting as a deterministic policy boundary between AI agents and financial tool access."

## Overview
Warden intercepts proposed actions, tokenizes sensitive data (PII/PAN), enforces deterministic authorization rules, routes to AI models safely, and maintains a strict audit trail.

*This project is a production-shaped prototype built to demonstrate architectural discipline and secure system design. It is not production-ready financial infrastructure.*

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
Run `docker compose up` to start the infrastructure (which includes real Redis, Postgres, and Presidio containers).
1. **Normal request**: Allowed based on explicit role bounds (`finance-operator`, `≤ ₹10k`).
2. **Sensitive request**: PII automatically replaced with deterministic tokens (`[CARD_001]`).
3. **Unauthorized refund**: Denied deterministically before touching the tool.
4. **Local model failure + Sensitive**: **FAIL CLOSED**. No fallback to external providers for sensitive data.

## Failure Semantics Matrix

| Failure                                  | Expected behavior                              |
| ---------------------------------------- | ---------------------------------------------- |
| Redis unavailable                        | Fail closed for rate-limit protected operation |
| Presidio unavailable                     | Restricted request rejected                    |
| PostgreSQL unavailable                   | Operation requiring vault rejected             |
| Local model unavailable + sensitive data | Fail closed                                    |
| Local model unavailable + non-sensitive  | External fallback if policy allows             |
| External model unavailable               | Deterministic error                            |
| Policy engine error                      | Fail closed                                    |
| Unknown tool                             | Deny                                           |
| Invalid tool arguments                   | Deny                                           |

## Threat Model (Scope)
The current implementation addresses the threats defined in the primary threat model (unauthorized AI tool execution, sensitive data leaking to external LLMs). 

**Several production threats remain explicitly documented as out of scope for this prototype:**
- Compromised gateway host
- Malicious/compromised dependency
- Stolen credentials / Insider access
- Vault compromise / Secret rotation
- Multi-tenant isolation
- Supply-chain attacks

## Production Gaps
To move this prototype to real production infrastructure, the following are required:
- HA deployment not implemented
- Secrets manager not implemented
- KMS/HSM integration not implemented (for the token vault)
- Multi-region failover not implemented
- Managed Redis/Postgres not implemented
- Formal compliance certification not performed
- Live Razorpay production tools not connected
- **Idempotency-Keys**: Idempotency is designed for (interface exists) but not yet wired into the gateway. Rate limiting is fully implemented and verified, but idempotency is not.
- **Detokenization**: The gateway tokenizes raw inputs (e.g., "John Doe" -> `[PERSON_001]`) and passes the raw input to the tool for simplicity in this prototype. In production, the model would output the tool execution instruction using the token, and the gateway would detokenize it via the Vault before hitting the financial backend.
- **Real LLM Integrations**: The `external-model` in `docker-compose` is just an HTTP 200 nginx stub used to test routing/failover logic only, not a real LLM.

## Design Decisions (ADRs)
- **ADR-001 Why deterministic policy?**: LLMs are non-deterministic; security boundaries cannot be.
- **ADR-002 Why tokenization?**: Tokenizing into a vault allows authorized services to reverse the token later if policy permits, whereas redaction destroys the data permanently.
- **ADR-003 Why local/external routing?**: Sensitive payloads must never leave the VPC; they route strictly to local models.
- **ADR-004 Why Redis?**: Shared state for rate-limiting across gateway instances ensures consistent enforcement.
- **ADR-005 Why PostgreSQL?**: Strict schema enforcement and atomic sequences for the token vault and audit trail.
- **ADR-006 Why no Kafka?**: Reduces operational complexity for synchronous HTTP gateways.
- **ADR-007 Why fail closed?**: In financial contexts, unavailability is preferable to an unsafe execution or PII leak.
- **ADR-008 How identity is verified**: Identity (`X-Warden-Actor`) must be injected by a trusted upstream context (e.g. an API Gateway validating an OIDC session), never from the untrusted LLM JSON payload.

---
*Built with AI pair-programming tools; every architectural decision and the security model were designed and validated by me.*