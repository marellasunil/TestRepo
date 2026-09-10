# FleetAMP enterprise one-agent pilot

This repository contains **sanitized deployment examples and a validation plan** for a limited FleetAMP evaluation in an enterprise test environment.

It does not contain FleetAMP application source code. Use a reviewed FleetAMP release from [marellasunil/FleetAMP](https://github.com/marellasunil/FleetAMP) and follow the maintained documentation at [fleetamp.marellasunil.com](https://fleetamp.marellasunil.com).

## Pilot scope

The pilot validates one management path:

```text
Administrator browser
        |
        | HTTPS
        v
FleetAMP (one instance, persistent SQLite)
        |
        | WSS OpAMP + bearer token
        v
OpAMP Supervisor (persistent identity storage)
        |
        | starts and manages
        v
One OpenTelemetry Collector
```

The pilot verifies:

- secure web access and administrator setup;
- secure OpAMP connectivity for one Supervisor;
- Collector inventory and details;
- a stable logical agent identity;
- Active, Offline, and Retired lifecycle states;
- group assignment and last-connected group;
- remote configuration deployment and status;
- persistence across Supervisor, Collector, FleetAMP, and host restarts;
- rollback to the last known-good configuration.

## Explicitly out of scope

This is not a production certification. The following remain outside this pilot:

- Kubernetes, Grafana Alloy, multiple Collectors, gateways, or application onboarding;
- FleetAMP high availability, autoscaling, or multi-replica SQLite;
- OIDC, enterprise RBAC, certificate-to-agent authorization, or approval separation;
- automated Collector binary upgrades;
- external Git/config providers;
- production telemetry backends and their credentials;
- performance, scale, disaster-recovery, and penetration testing.

## Repository layout

```text
.
├── README.md
├── .gitignore
├── config/
│   ├── collector-base.example.yaml
│   └── supervisor.example.yaml
├── docs/
│   └── validation-checklist.md
└── systemd/
    └── opampsupervisor.service
```

Files ending in `.example.yaml` are templates. Copy them to protected host paths and replace placeholders locally. Do not rename a file containing real credentials back into a tracked example filename.

## Security rules

Never commit:

- bearer tokens, passwords, server peppers, API keys, or session material;
- private keys, client certificates, PKCS#12/JKS files, or internal CA bundles;
- `fleetamp.db`, Supervisor identity/state, backups, snapshots, logs, or PID files;
- enterprise DNS names, IP addresses, hostnames, proxy details, or unredacted command output;
- Collector configurations containing backend credentials.

Use the enterprise secret manager and approved PKI. Keep FleetAMP and OpAMP endpoints on private networks where possible. Use HTTPS/WSS, validate the server certificate, and retain the OpAMP bearer token even when mTLS is enabled.

## Start here

1. Obtain security and change approval for a non-production host and one test Collector.
2. Pin reviewed FleetAMP, OpAMP Supervisor, and Collector versions.
3. Install FleetAMP as a single instance with persistent local storage.
4. Complete the one-time administrator setup over HTTPS.
5. Copy the templates from `config/` to protected paths outside this repository.
6. Replace the endpoint, bearer-token placeholder, executable path, and stable agent ID locally.
7. Install `systemd/opampsupervisor.service`, adjusting paths and the service account for the approved host.
8. Execute [the validation checklist](docs/validation-checklist.md).
9. Record only sanitized evidence in the enterprise change record, not in this repository.

## Reference documentation

- [FleetAMP installation](https://fleetamp.marellasunil.com/docs/getting-started/installation)
- [Production deployment](https://fleetamp.marellasunil.com/docs/operations/production-deployment)
- [TLS and mutual TLS](https://fleetamp.marellasunil.com/docs/operations/transport-tls)
- [Security hardening](https://fleetamp.marellasunil.com/docs/operations/security-hardening)
- [Fleet management](https://fleetamp.marellasunil.com/docs/fleet-management/inventory)
