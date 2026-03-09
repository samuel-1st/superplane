# DigitalOcean Integration Skills

Use this guidance when planning or configuring DigitalOcean components.

## Available Actions

### `digitalocean.createDroplet`
Creates a new DigitalOcean Droplet (virtual machine). Polls until the droplet reaches **active** status.

**Required:** `name`, `region`, `size`, `image`
**Optional:** `sshKeys`, `tags`, `userData`
**Output:** `digitalocean.droplet.created` — the fully provisioned droplet object including `id`, `name`, `status`, `region`, `networks`, `size_slug`

### `digitalocean.getDroplet`
Retrieves an existing Droplet by its numeric ID.

**Required:** `dropletID` (string, supports expressions)
**Output:** `digitalocean.droplet.retrieved` — the droplet object

### `digitalocean.deleteDroplet`
Deletes a Droplet by its numeric ID.

**Required:** `dropletID` (string, supports expressions)
**Output:** `digitalocean.droplet.deleted` — `{"deleted": true, "dropletID": <id>}`

### `digitalocean.manageDropletPower`
Changes the power state of a Droplet. Polls the action until it **completes**.

**Required:** `dropletID`, `operation` (`power_on`, `shutdown`, `reboot`, `power_cycle`, `power_off`)
**Output:** `digitalocean.droplet.power.managed` — the completed DigitalOcean action object

### `digitalocean.createSnapshot`
Creates a snapshot of a Droplet. Polls the action until it **completes**.

**Required:** `dropletID`, `snapshotName`
**Output:** `digitalocean.droplet.snapshot.created` — the completed DigitalOcean action object

### `digitalocean.deleteSnapshot`
Deletes a snapshot by its ID.

**Required:** `snapshotID`
**Output:** `digitalocean.snapshot.deleted` — `{"deleted": true, "snapshotID": "<id>"}`

### `digitalocean.createDNSRecord`
Creates a DNS record in a DigitalOcean domain.

**Required:** `domain`, `type` (A/AAAA/CNAME/MX/TXT/NS/SRV/CAA), `name`, `data`
**Optional:** `ttl` (default 1800), `priority` (for MX/SRV)
**Output:** `digitalocean.dns.record.created` — the created DNS record object

### `digitalocean.deleteDNSRecord`
Deletes a DNS record from a DigitalOcean domain.

**Required:** `domain`, `recordID`
**Output:** `digitalocean.dns.record.deleted` — `{"deleted": true, "domain": "<domain>", "recordID": <id>}`

### `digitalocean.upsertDNSRecord`
Creates or updates a DNS record idempotently. If a record with the same `name` and `type` exists it is updated; otherwise a new one is created.

**Required:** `domain`, `type`, `name`, `data`
**Optional:** `ttl`, `priority`
**Output:** `digitalocean.dns.record.upserted` — the created or updated DNS record object

### `digitalocean.createLoadBalancer`
Creates a DigitalOcean Load Balancer with forwarding rules.

**Required:** `name`, `region`, `entryProtocol`, `entryPort`, `targetProtocol`, `targetPort`
**Optional:** `size` (lb-small/lb-medium/lb-large), `dropletIDs`, `tags`
**Output:** `digitalocean.load_balancer.created` — the load balancer object

### `digitalocean.deleteLoadBalancer`
Deletes a Load Balancer by its ID.

**Required:** `loadBalancerID`
**Output:** `digitalocean.load_balancer.deleted` — `{"deleted": true, "loadBalancerID": "<id>"}`

### `digitalocean.assignReservedIP`
Assigns or unassigns a Reserved IP to/from a Droplet.

**Required:** `reservedIP`, `action` (`assign` or `unassign`)
**Optional:** `dropletID` (required when `action` is `assign`)
**Output:** `digitalocean.reserved_ip.action` — the DigitalOcean action object

## Planning Rules

1. Use `digitalocean.createDroplet` when the workflow needs to provision a new VM. Always provide all four required fields: `name`, `region`, `size`, `image`.
2. Use `digitalocean.getDroplet` to fetch current droplet info without making changes.
3. Use `digitalocean.manageDropletPower` for controlled shutdowns/reboots; prefer `shutdown` over `power_off` for graceful shutdown.
4. Use `digitalocean.createSnapshot` before destructive operations or deployments.
5. Use `digitalocean.upsertDNSRecord` instead of `createDNSRecord` when the workflow may run multiple times to avoid creating duplicate records.
6. For `assignReservedIP` with `action: assign`, `dropletID` is required — always include it.
7. When chaining actions that depend on a newly created droplet, reference `data.id` from the `createDroplet` output.

## Output Field References

| Component | Output Field Path | Description |
|-----------|-------------------|-------------|
| createDroplet | `data.id` | Droplet numeric ID |
| createDroplet | `data.networks.v4[0].ip_address` | Public IP address |
| getDroplet | `data.id` | Droplet numeric ID |
| manageDropletPower | `data.id` | Action ID |
| manageDropletPower | `data.status` | Action status (`completed`) |
| createSnapshot | `data.id` | Action ID |
| createDNSRecord | `data.id` | DNS record numeric ID |
| upsertDNSRecord | `data.id` | DNS record numeric ID |
| createLoadBalancer | `data.id` | Load balancer UUID |
| createLoadBalancer | `data.ip` | Load balancer IP |
| assignReservedIP | `data.id` | Reserved IP action ID |
