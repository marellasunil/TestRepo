# Enterprise one-agent validation checklist

Use this checklist only in an approved **non-production** environment. Record results in the enterprise change system. Do not commit screenshots or output containing internal names, addresses, identifiers, tokens, certificates, or configuration secrets.

## 1. Entry criteria

- [ ] A named technical owner and rollback owner are available.
- [ ] FleetAMP has passed the repository test, race, vet, vulnerability, and CodeQL checks for the selected revision.
- [ ] FleetAMP, OpAMP Supervisor, and OTel Collector versions are explicitly pinned.
- [ ] The test host, one Collector, test duration, and maintenance window are approved.
- [ ] The enterprise PKI certificate covers the FleetAMP DNS name and is trusted by the browser and Supervisor host.
- [ ] HTTPS/WSS ingress is restricted to approved networks.
- [ ] FleetAMP backend listeners are reachable only from the approved proxy/load balancer or private network.
- [ ] A unique bearer token is stored in the enterprise secret manager.
- [ ] Persistent FleetAMP and Supervisor storage paths exist and have least-privilege ownership.
- [ ] A FleetAMP database backup and restoration location are prepared.

## 2. Prepare files outside Git

Install only reviewed binaries obtained through the approved artifact channel. Verify checksums/signatures according to enterprise policy.

Suggested protected locations:

```text
/opt/fleetamp/fleetamp
/opt/opampsupervisor/opampsupervisor
/opt/otelcol-contrib/otelcol-contrib
/etc/fleetamp/fleetamp.env
/etc/fleetamp/tls/
/etc/opampsupervisor/supervisor.yaml
/etc/otelcol-contrib/collector-base.yaml
/var/lib/fleetamp/
/var/lib/opampsupervisor/
```

Create local configurations:

```bash
sudo install -d -o root -g otel -m 0750 /etc/opampsupervisor /etc/otelcol-contrib
sudo install -o root -g otel -m 0640 config/supervisor.example.yaml /etc/opampsupervisor/supervisor.yaml
sudo install -o root -g otel -m 0640 config/collector-base.example.yaml /etc/otelcol-contrib/collector-base.yaml
```

On the host, replace:

- `CHANGE_ME_FLEETAMP_HOST` with the approved FleetAMP DNS name;
- the bearer placeholder with a unique secret obtained from the approved secret store;
- `enterprise-test-agent-01` only if the identity is already used;
- executable/configuration paths when the enterprise layout differs.

Never set `insecure_skip_verify: true`.

## 3. Preflight validation

```bash
getent hosts <fleetamp-dns-name>
openssl s_client -connect <fleetamp-dns-name>:443 -servername <fleetamp-dns-name> </dev/null
curl --fail --silent --show-error https://<fleetamp-dns-name>/health
sudo -u otel test -r /etc/opampsupervisor/supervisor.yaml
sudo -u otel test -x /opt/opampsupervisor/opampsupervisor
sudo -u otel test -x /opt/otelcol-contrib/otelcol-contrib
sudo /opt/otelcol-contrib/otelcol-contrib --config /etc/otelcol-contrib/collector-base.yaml --dry-run
sudo systemd-analyze verify systemd/opampsupervisor.service
```

Expected:

- [ ] DNS resolves only to the approved endpoint.
- [ ] The certificate chain, hostname, and validity period pass.
- [ ] FleetAMP health returns success over HTTPS.
- [ ] Both executables and configurations are accessible to the service account.
- [ ] The Collector configuration and systemd unit validate successfully.

## 4. Install and start the Supervisor

```bash
sudo install -o root -g root -m 0644 systemd/opampsupervisor.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now opampsupervisor
sudo systemctl status opampsupervisor
sudo journalctl -u opampsupervisor --since "10 minutes ago"
```

Expected:

- [ ] The Supervisor remains active.
- [ ] The Collector child process is running.
- [ ] Logs show a successful WSS OpAMP connection.
- [ ] Logs contain no token, private key, or full remote configuration.
- [ ] The Collector health endpoint responds locally.

## 5. FleetAMP inventory and identity

In FleetAMP:

- [ ] Exactly one new agent appears.
- [ ] The agent reports **Active**, connected, and healthy.
- [ ] Version, platform, capabilities, and effective configuration are visible.
- [ ] The logical identifier corresponds to `fleetamp.agent.id`.
- [ ] No duplicate agent is created after restarting the Collector child process.
- [ ] No duplicate agent is created after restarting `opampsupervisor`.
- [ ] No duplicate agent is created after rebooting the test host.
- [ ] The original group assignment and labels remain after each restart.

A changed runtime InstanceUID must not create a second logical agent when the same unique stable FleetAMP agent ID and Supervisor storage are retained.

## 6. Group and deployment test

Create one test group, for example `enterprise-pilot`, and assign only the pilot agent.

Use a harmless configuration change such as adjusting the batch timeout. Do not add production exporters or credentials.

- [ ] FleetAMP validates the configuration before deployment.
- [ ] Deployment transitions through the expected states, such as sent, applying, and applied.
- [ ] Desired and effective configuration are displayed.
- [ ] The agent remains healthy.
- [ ] The assignment is still present after FleetAMP restart.
- [ ] The agent page shows the current or last-connected group.
- [ ] Audit/event history identifies the action without revealing secrets.

## 7. Lifecycle test

Perform controlled disconnections:

1. Stop the Supervisor.
2. Confirm the agent becomes **Offline** after the configured threshold.
3. Start the Supervisor and confirm the same logical agent returns to **Active**.
4. Test **Retired** only with a short, approved pilot retirement threshold.
5. Reconnect and verify the expected policy for a retired identity.

- [ ] Active means recently connected.
- [ ] Offline means known but disconnected.
- [ ] Retired means disconnected beyond the configured retirement policy.
- [ ] No group or label is silently assigned to a different logical Collector.

## 8. Rollback test

Before changing configuration, capture the sanitized version identifier—not secret-bearing content.

- [ ] Deploy a known-good pilot configuration.
- [ ] Deploy a deliberately invalid but non-destructive change and confirm validation blocks it.
- [ ] Deploy a valid harmless change.
- [ ] Roll back to the previous known-good revision.
- [ ] Confirm the deployment status and effective configuration report the rollback.
- [ ] Confirm the Collector stays healthy.

## 9. FleetAMP restart and database persistence

```bash
sudo systemctl restart fleetamp
sudo systemctl status fleetamp
curl --fail --silent --show-error https://<fleetamp-dns-name>/health
```

Verify:

- [ ] The administrator account remains available.
- [ ] The agent inventory remains intact.
- [ ] Group membership, labels, desired configuration, deployment history, and lifecycle state remain intact.
- [ ] The Supervisor reconnects without creating a duplicate.
- [ ] The SQLite database remains on persistent local storage.

## 10. Exit criteria

The pilot is successful only when:

- [ ] all required checks above pass;
- [ ] no secret or private enterprise detail entered Git;
- [ ] no unexplained duplicate agents exist;
- [ ] group assignment survives all approved restarts;
- [ ] configuration deployment and rollback are observable;
- [ ] the Collector remains within agreed CPU and memory guardrails;
- [ ] findings, limitations, and follow-up actions are recorded.

Do not call this production-ready based on a one-agent pilot. HA, scale, OIDC/RBAC, approval separation, certificate-to-agent authorization, disaster recovery, performance, and security testing need separate work.
