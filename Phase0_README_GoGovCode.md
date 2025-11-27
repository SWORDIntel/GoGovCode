# Phase 0 – How to Use These GoGovCode Phase Specs with an LLM

**Project:** GoGovCode – DSMIL Control Plane Framework  
**Purpose:** Guide an LLM on how to apply the phase documents to the existing GoGovCode repository.

---

## 1. Repository Context

The GoGovCode repo already contains:

- `cmd/gogovcode/` – main entrypoint  
- `api/handlers`, `api/middleware`, `api/routes` – HTTP API logic  
- `config/` – configuration loading and defaults  
- `internal/` – core subsystems (`auth`, `audit`, `certs`, `errors`, `health`, `logging`, `policy`, `server`, `util`)  
- `pkg/` – shared `client`, `models`, `schema` packages  
- `scripts/` – helper tools for generation/overrides  

These phases describe **incremental upgrades** that must respect this structure.

---

## 2. Phase Overview

- **Phase 1 – Foundation & Baseline Integration**  
  Clean up structure, config, logging, health, and the base HTTP server so GoGovCode is a solid DSMIL-ready scaffold.

- **Phase 2 – Clearance, Policy, and Audit Integration**  
  Add DSMIL-specific clearance enforcement, device-token semantics, policy engine integration, and a unified audit event model.

- **Phase 3 – PQC Transport & Event Bus Integration**  
  Introduce post-quantum–ready transport primitives and a Redis Streams-based event bus client following DSMIL stream naming.

- **Phase 4 – QUIC Binary Protocols & Immutable MinIO Audit**  
  Implement QUIC-based binary protocols for SHRINK/MEMSHADOW/HURRICANE and a MinIO-backed immutable audit log with hash chains.

- **Phase 5 – SDKs, Device Templates & Identity Integration**  
  Provide auto-generated SDKs, device daemon templates, SPIFFE/SPIRE identity integration, and constant-time helpers.

Each phase is **additive and incremental**; later phases depend on earlier ones.

---

## 3. Instructions to the LLM

When processing these `.md` specs, the LLM should:

1. **Honor Existing Layout**  
   Use current directories (`cmd/`, `api/`, `internal/`, `pkg/`, `scripts/`) instead of inventing new structures unless explicitly requested.

2. **Implement One Phase at a Time**  
   - Finish Phase 1 changes (including tests) before applying Phase 2.  
   - Avoid mixing concerns from different phases in a single step unless necessary.

3. **Prefer Small, Focused PR-Style Changes**  
   For each phase:
   - Propose new files with full contents.  
   - For existing files, show full updated versions when changes are large, or clearly marked sections when minimal.

4. **Maintain Backwards Compatibility Where Possible**  
   - Existing public APIs and behavior should not break without a clear reason.  
   - New features should be gated by config flags where possible.

5. **Add Tests with Each Feature**  
   - For new subsystems (policy, audit, bus, PQC, QUIC), include unit tests.  
   - Prefer clear, deterministic tests with no external network calls unless mocked.

6. **Document New Features**  
   - Update `README.md` with a short overview of the new phase’s features.  
   - For complex subsystems (PQC, QUIC, MinIO audit, SPIFFE), add a short section explaining usage and configuration keys.

7. **Security & Performance Defaults**  
   - Choose secure defaults (TLS on, clearance checks enabled, audit enabled).  
   - Provide configuration options to relax behaviors explicitly, but keep hardened behavior by default.

---

## 4. Suggested Application Order

1. Read **Phase 1** and generate code changes.  
2. Run `go build ./...` and `go test ./...`. Fix issues.  
3. Proceed to **Phase 2**, apply changes, and extend tests.  
4. Repeat for **Phase 3**, **Phase 4**, and **Phase 5**.

At the end, GoGovCode should be a **DSMIL-aligned, PQC-ready, QUIC-capable, policy- and audit-driven control-plane framework** usable across all 104 DSMIL devices.

