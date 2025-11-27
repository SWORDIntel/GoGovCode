# GoGovCode – Phase 4: QUIC Binary Protocols & Immutable MinIO Audit

**Version:** 1.0  
**Status:** Draft – Ready for LLM refinement  
**Project:** GoGovCode – DSMIL Control Plane Framework  
**Dependencies:** Phase 1–3 complete

---

## 1. Objectives

- Provide a **QUIC-based binary protocol server** for SHRINK, MEMSHADOW, HURRICANE and other telemetry-heavy paths.
- Finalize **immutable MinIO-backed audit logging** with hash chaining.
- Move critical DSMIL feeds onto hardened binary channels with strong audit trails.

---

## 2. Scope

- QUIC server implementation.  
- Binary frame format design.  
- MinIO audit driver.  
- Integration with policy/clearance where relevant.

---

## 3. Deliverables

1. **QUIC Server**
   - `internal/server/quic.go`:
     - Bootstraps a QUIC listener (UDP) with:
       - TLS + PQC hybrid keys (from Phase 3)
     - Accepts incoming sessions and dispatches to protocol handlers.
   - `pkg/protocol/frame.go`:
     - Binary frame format:
       - header: version, type, flags, length
       - metadata: device_id, layer, clearance, correlation_id
       - payload: compressed or raw body
     - Encode/decode helpers with:
       - bounds checking
       - constant-time header parsing for sensitive fields.

2. **Binary Protocol Handlers**
   - Initial handlers for:
     - `SHRINK_TELEMETRY`
     - `MEMSHADOW_SYNC`
     - `HURRICANE_EVENTS`
   - Each handler:
     - validates header/device/clearance via Phase 2 engines
     - optionally publishes to Redis Streams
     - writes audit records

3. **MinIO Audit Driver**
   - `internal/audit/minio_writer.go`:
     - configuration for:
       - endpoint
       - bucket
       - access key / secret (from Vault or env)
     - Writes append-only log objects (e.g., daily files).
     - Uses SHA3-512 hash chain:
       - each record references previous record hash.
     - Optional PQC signatures on batches.

4. **Audit Schema**
   - Documented JSON schema for `AuditEvent` as stored in MinIO.
   - Versioning strategy for audit schema.

---

## 4. High-Level Tasks

- Implement QUIC server wrapper and configuration.
- Define binary frame spec and implement encoding/decoding.
- Implement proto handlers for SHRINK/MEMSHADOW/HURRICANE.
- Implement MinIO audit writer using hash chaining.
- Integrate QUIC session events with audit logs.

---

## 5. Exit Criteria

- QUIC server starts and accepts test connections.
- Binary frames can be sent, parsed, and validated with:
  - correct device_id, layer, and clearance.
- SHRINK/MEMSHADOW/HURRICANE handlers wired and tested with sample frames.
- MinIO audit writer:
  - creates bucket and objects successfully
  - produces verifiable hash chain across records.
- End-to-end: QUIC frame → clearance/policy → audit + Redis event.
