import { ComponentBaseMapper, EventStateRegistry, TriggerRenderer } from "../types";
import { createDropletMapper } from "./create_droplet";
import { getDropletMapper } from "./get_droplet";
import { deleteDropletMapper } from "./delete_droplet";
import { manageDropletPowerMapper } from "./manage_droplet_power";
import { createSnapshotMapper } from "./create_snapshot";
import { deleteSnapshotMapper } from "./delete_snapshot";
import { createDNSRecordMapper } from "./create_dns_record";
import { deleteDNSRecordMapper } from "./delete_dns_record";
import { upsertDNSRecordMapper } from "./upsert_dns_record";
import { createLoadBalancerMapper } from "./create_load_balancer";
import { deleteLoadBalancerMapper } from "./delete_load_balancer";
import { assignReservedIPMapper } from "./assign_reserved_ip";
import { buildActionStateRegistry } from "../utils";

export const componentMappers: Record<string, ComponentBaseMapper> = {
  createDroplet: createDropletMapper,
  getDroplet: getDropletMapper,
  deleteDroplet: deleteDropletMapper,
  manageDropletPower: manageDropletPowerMapper,
  createSnapshot: createSnapshotMapper,
  deleteSnapshot: deleteSnapshotMapper,
  createDNSRecord: createDNSRecordMapper,
  deleteDNSRecord: deleteDNSRecordMapper,
  upsertDNSRecord: upsertDNSRecordMapper,
  createLoadBalancer: createLoadBalancerMapper,
  deleteLoadBalancer: deleteLoadBalancerMapper,
  assignReservedIP: assignReservedIPMapper,
};

export const triggerRenderers: Record<string, TriggerRenderer> = {};

export const eventStateRegistry: Record<string, EventStateRegistry> = {
  createDroplet: buildActionStateRegistry("created"),
  getDroplet: buildActionStateRegistry("retrieved"),
  deleteDroplet: buildActionStateRegistry("deleted"),
  manageDropletPower: buildActionStateRegistry("managed"),
  createSnapshot: buildActionStateRegistry("created"),
  deleteSnapshot: buildActionStateRegistry("deleted"),
  createDNSRecord: buildActionStateRegistry("created"),
  deleteDNSRecord: buildActionStateRegistry("deleted"),
  upsertDNSRecord: buildActionStateRegistry("upserted"),
  createLoadBalancer: buildActionStateRegistry("created"),
  deleteLoadBalancer: buildActionStateRegistry("deleted"),
  assignReservedIP: buildActionStateRegistry("action"),
};
