# GoGovCode – Phase 3: PQC Transport & Event Bus Integration

**Version:** 1.0  
**Status:** Draft – Ready for LLM refinement  
**Project:** GoGovCode – DSMIL Control Plane Framework  
**Dependencies:** Phase 1 (Foundation), Phase 2 (Clearance/Policy/Audit)

---

## 1. Objectives

- Add **post-quantum–safe transport** primitives to GoGovCode.
- Integrate a **Redis Streams event bus client** for cross-layer messaging.
- Ensure all core control-plane traffic can move onto PQC-ready channels and structured streams.

---

## 2. Scope

This phase adds:

- PQC handshake and hybrid TLS support.  
- Redis Streams client with DSMIL topic conventions.  
- Wiring PQC and streams into the existing `internal/server` and `pkg/bus` layers.

---

## 3. Deliverables

1. **PQC Transport Layer**
   - `internal/auth/pqc.go`:
     - Interfaces for:
       - KEM (e.g., ML-KEM-1024)
       - Signature (e.g., ML-DSA-87)
     - Hybrid key exchange:
       - classical ECDHE + PQ KEM
     - Session key derivation for AES-256-GCM.
   - `internal/server/tls_config.go`:
     - Ability to spin up:
       - classical TLS
       - hybrid PQC TLS (feature flag)
     - Configuration keys:
       - `pqc.enabled`
       - `pqc.kem_algorithm`
       - `pqc.signature_algorithm`

2. **Redis Streams Bus**
   - `pkg/bus/redis_streams.go`:
     - Connects to Redis with:
       - TLS support (optionally PQC offload later)
     - Implements:
       - producer (`PublishEvent`)
       - consumer (`ConsumeLoop`, with handler callback)
       - consumer groups
       - backpressure logic
     - DSMIL standard streams:
       - `L3_IN`, `L3_OUT`
       - `L4_IN`, `L4_OUT`
       - `L7_FUSION`
       - `SECURITY_ALERTS`
       - `EXEC_IN`

   - Event model:
     - `BusEvent` with:
       - `id`
       - `stream`
       - `device_id`
       - `layer`
       - `payload` (JSON/binary)
       - `clearance` (copied from request context when applicable)

3. **Server Wiring**
   - Extend `cmd/gogovcode/main.go` to:
     - optionally start bus consumers for configured streams
     - expose simple test endpoints that:
       - publish events to streams
       - show event consumption logs

---

## 4. High-Level Tasks

- Implement PQC KEM/signature abstraction.
- Integrate hybrid key exchange into TLS config module.
- Implement Redis Streams client and event models.
- Add configuration toggles for:
  - PQC transport on/off
  - Redis streams endpoints
- Add basic example service:
  - HTTP POST → validated → clearance/policy → publish to `L3_IN`.

---

## 5. Exit Criteria

- PQC transport mode can be enabled via config and server successfully negotiates hybrid TLS.
- Redis Streams integration tested with:
  - `PublishEvent` called from API handlers.
  - `ConsumeLoop` running in a background goroutine.
- Events carry device_id, clearance, and layer fields.
- Test service demonstrates end-to-end flow:
  - HTTP request → clearance/policy → audit log → Redis stream event.
