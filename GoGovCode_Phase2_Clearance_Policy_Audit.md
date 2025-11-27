# GoGovCode – Phase 2: Clearance, Policy, and Audit Integration

**Version:** 1.0  
**Status:** Draft – Ready for LLM refinement  
**Project:** GoGovCode – DSMIL Control Plane Framework  
**Dependencies:** Phase 1 (Foundation) complete

---

## 1. Objectives

- Embed **DSMIL-native clearance and device semantics** into GoGovCode.
- Introduce a **policy engine** for runtime decisions.
- Integrate **immutable audit logging** suitable for MinIO / object storage backends.

---

## 2. Scope

Phase 2 focuses on **who can do what** and **how it is recorded**:

- Clearance enforcement middleware  
- Device token and layer awareness  
- Policy evaluation (allow/deny with reasons)  
- Immutable audit log model and interface  

No PQC or message bus yet — those arrive in later phases.

---

## 3. Deliverables

1. **Clearance Middleware**
   - New middleware in `api/middleware/clearance.go`:
     - Extracts:
       - `device_id`
       - `layer`
       - `clearance` (e.g., `0x02020202`–`0x09090909`)
     - Validates:
       - device-specific rules
       - minimum clearance requirements per route
     - Enforces **upward-only** data flows (lower → higher layers only).
   - Configurable via route metadata or policy rules:
     - e.g., `RequiredClearance`, `AllowedLayers`, `AllowedMethods`.

2. **Device & Token Model**
   - `pkg/models/device.go`:
     - `DeviceID`, `Layer`, `Class`, `TokenBase`.
   - Helper:
     - token format: `0x8000 + (device_id * 3) + offset`
       - `offset = 0` → STATUS
       - `offset = 1` → CONFIG
       - `offset = 2` → DATA
   - Utilities to:
     - compute token IDs
     - reverse-map from token → device info (for logging and policy).

3. **Policy Engine Integration**
   - `internal/policy`:
     - policy type definitions (YAML/JSON-driven)
     - JSON Schema validation that ensures:
       - correct field types
       - known device IDs/layers
     - basic conflict detection:
       - conflicting allow/deny on same route/device
     - evaluation API:
       - `Evaluate(reqContext) -> (decision, reason)`

4. **Audit Logging**
   - `internal/audit`:
     - unified `AuditEvent` struct:
       - event_id
       - timestamp
       - actor (user/device)
       - clearance
       - device_id
       - action (route, method, resource)
       - policy_decision (allow/deny + reason)
     - interface:
       - `Writer.Write(event AuditEvent) error`
     - stub drivers:
       - `stdout` driver
       - `file` driver
       - `minio` driver (signature: defined but can be mocked until Phase 4)

---

## 4. High-Level Tasks

- Define clearance and device models.
- Implement clearance middleware and route integration.
- Implement policy loading, validation, and evaluation.
- Implement audit event model and writer interfaces.
- Add route examples demonstrating:
  - open endpoint
  - restricted endpoint
  - device-only endpoint

---

## 5. Exit Criteria

- A protected test endpoint returns:
  - 403 when clearance/device is insufficient
  - 200 when clearance/device is valid
- Policy changes can be hot-reloaded (or reloaded via signal/endpoint).
- Audit events are emitted per protected request with:
  - decision (allow/deny)
  - device_id, clearance, route
- `go test ./internal/policy ./internal/audit` passes with coverage for:
  - policy validation
  - basic conflict detection
  - audit writer behavior (stdout/file)
