// Copyright 2018 Mario Trangoni
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build linux

package collector

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	crmMonElemEnabled = kingpin.Flag("collector.crm_mon.elements-enabled",
		"Pacemaker `crm_mon` XML elements that will be exported.").Default(
		"summary,nodes,node_attributes,clones,resources,resources_group,failures,bans").String()
)

type crmMonCollector struct {
	crmMonInfo                        *prometheus.Desc
	crmMonLastUpdate                  *prometheus.Desc
	crmMonLastChange                  *prometheus.Desc
	crmMonDCPresent                   *prometheus.Desc
	crmMonDCQuorum                    *prometheus.Desc
	crmMonNodesConfigured             *prometheus.Desc
	crmMonResourcesConfigured         *prometheus.Desc
	crmMonResourcesDisabled           *prometheus.Desc
	crmMonResourcesBlocked            *prometheus.Desc
	crmMonStonith                     *prometheus.Desc
	crmMonSymmetricCluster            *prometheus.Desc
	crmMonMaintenanceMode             *prometheus.Desc
	crmMonNodeID                      *prometheus.Desc
	crmMonNodeOnline                  *prometheus.Desc
	crmMonNodeStandby                 *prometheus.Desc
	crmMonNodeStandbyOnFail           *prometheus.Desc
	crmMonNodeMaintenance             *prometheus.Desc
	crmMonNodePending                 *prometheus.Desc
	crmMonNodeUnclean                 *prometheus.Desc
	crmMonNodeShutdown                *prometheus.Desc
	crmMonNodeExpectedUp              *prometheus.Desc
	crmMonNodeIsDC                    *prometheus.Desc
	crmMonNodeResourcesRunning        *prometheus.Desc
	crmMonNodeAttribute               *prometheus.Desc
	crmMonResourceActive              *prometheus.Desc
	crmMonResourceOrphaned            *prometheus.Desc
	crmMonResourceBlocked             *prometheus.Desc
	crmMonResourceManaged             *prometheus.Desc
	crmMonResourceFailed              *prometheus.Desc
	crmMonResourceFailureIgnored      *prometheus.Desc
	crmMonResourcesGroup              *prometheus.Desc
	crmMonResourceGroupActive         *prometheus.Desc
	crmMonResourceGroupOrphaned       *prometheus.Desc
	crmMonResourceGroupBlocked        *prometheus.Desc
	crmMonResourceGroupManaged        *prometheus.Desc
	crmMonResourceGroupFailed         *prometheus.Desc
	crmMonResourceGroupFailureIgnored *prometheus.Desc
	crmMonResourceCloneMultistate     *prometheus.Desc
	crmMonResourceClonePromoted       *prometheus.Desc
	crmMonResourceCloneActive         *prometheus.Desc
	crmMonResourceCloneOrphaned       *prometheus.Desc
	crmMonResourceCloneBlocked        *prometheus.Desc
	crmMonResourceCloneManaged        *prometheus.Desc
	crmMonResourceCloneFailed         *prometheus.Desc
	crmMonResourceCloneFailureIgnored *prometheus.Desc
	crmMonResourceCloneNumActive      *prometheus.Desc
	crmMonResourceCloneNumPromoted    *prometheus.Desc
	crmMonFailuresCount               *prometheus.Desc
	crmMonFailureDescription          *prometheus.Desc
	crmMonBansCount                   *prometheus.Desc
	crmMonBanDescription              *prometheus.Desc
}

func init() {
	registerCollector("crm_mon", defaultEnabled, NewCrmMonCollector)
}

// NewCrmMonCollector returns a new Collector exposing crm_mon information.
func NewCrmMonCollector() (Collector, error) {
	return &crmMonCollector{
		crmMonInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "crm_mon", "info"),
			"A metric with a constant '1' value labeled by version of crm_mon.",
			[]string{"version"}, nil,
		),
		crmMonLastUpdate: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "last_update_time", "seconds"),
			"Last update time of cluster info since unix epoch in seconds.",
			[]string{"stack"}, nil,
		),
		crmMonLastChange: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "last_change_time", "seconds"),
			"Last Cluster Information Base change time since unix epoch in seconds.",
			[]string{"user", "client", "origin"}, nil,
		),
		crmMonDCPresent: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "dc", "present"),
			"Whether the cluster has an active DC.",
			[]string{"name"}, nil,
		),
		crmMonDCQuorum: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "dc", "quorum"),
			"Whether the cluster has quorum.",
			[]string{"name"}, nil,
		),
		crmMonNodesConfigured: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "nodes", "configured"),
			"Number of nodes configured.",
			[]string{"expected_votes"}, nil,
		),
		crmMonResourcesConfigured: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "resources", "configured"),
			"Number of resources configured.",
			[]string{"name"}, nil,
		),
		crmMonResourcesDisabled: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "resources", "disabled"),
			"Number of resources disabled.",
			[]string{"name"}, nil,
		),
		crmMonResourcesBlocked: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "resources", "blocked"),
			"Number of resources blocked.",
			[]string{"name"}, nil,
		),
		crmMonStonith: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stonith", "enabled"),
			"Whether STONITH is enabled.",
			[]string{"name"}, nil,
		),
		crmMonSymmetricCluster: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "symmetric_cluster", "enabled"),
			"Whether resources run on any node by default.",
			[]string{"name"}, nil,
		),
		crmMonMaintenanceMode: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "maintenance_mode", "enabled"),
			"Whether maintenance mode is enabled.",
			[]string{"name"}, nil,
		),
		// Nodes section metrics
		crmMonNodeID: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "node", "id"),
			"A metric with a constant '1' value labeled by node name, type, and node ID.",
			[]string{"name", "type", "id"}, nil,
		),
		crmMonNodeOnline: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "node", "online"),
			"Node is online.",
			[]string{"name"}, nil,
		),
		crmMonNodeStandby: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "node", "standby"),
			"Node is standby.",
			[]string{"name"}, nil,
		),
		crmMonNodeStandbyOnFail: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "node", "standby_on_fail"),
			"Node is standby on fail.",
			[]string{"name"}, nil,
		),
		crmMonNodeMaintenance: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "node", "maintenance"),
			"Node is in maintenance mode.",
			[]string{"name"}, nil,
		),
		crmMonNodePending: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "node", "pending"),
			"Node is pending.",
			[]string{"name"}, nil,
		),
		crmMonNodeUnclean: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "node", "unclean"),
			"Node is unclean.",
			[]string{"name"}, nil,
		),
		crmMonNodeShutdown: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "node", "shutdown"),
			"Node is shutdown.",
			[]string{"name"}, nil,
		),
		crmMonNodeExpectedUp: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "node", "expected_up"),
			"Node is expected up.",
			[]string{"name"}, nil,
		),
		crmMonNodeIsDC: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "node", "is_dc"),
			"Node is the DC.",
			[]string{"name"}, nil,
		),
		crmMonNodeResourcesRunning: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "node", "resource_running"),
			"Number of resources running on node.",
			[]string{"name"}, nil,
		),
		// Node Attributes section metrics
		crmMonNodeAttribute: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "node", "attribute"),
			"Node attribute with a constant '1' value labeled by name, attribute, and its value.",
			[]string{"name", "attribute", "value"}, nil,
		),
		// Node Resources Resource metrics
		crmMonResourceActive: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "resource", "active"),
			"Resource is active.",
			[]string{"id", "node_name", "resource_agent", "role", "target_role"}, nil,
		),
		crmMonResourceOrphaned: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "resource", "orphaned"),
			"Resource is orphaned.",
			[]string{"id", "node_name", "resource_agent", "role", "target_role"}, nil,
		),
		crmMonResourceBlocked: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "resource", "blocked"),
			"Resource is blocked.",
			[]string{"id", "node_name", "resource_agent", "role", "target_role"}, nil,
		),
		crmMonResourceManaged: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "resource", "managed"),
			"Resource is managed.",
			[]string{"id", "node_name", "resource_agent", "role", "target_role"}, nil,
		),
		crmMonResourceFailed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "resource", "failed"),
			"Resource is failed.",
			[]string{"id", "node_name", "resource_agent", "role", "target_role"}, nil,
		),
		crmMonResourceFailureIgnored: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "resource", "failure_ignored"),
			"Resource failure ignored.",
			[]string{"id", "node_name", "resource_agent", "role", "target_role"}, nil,
		),

		// Node Resources Group metrics
		crmMonResourcesGroup: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "group", "resource_number"),
			"Number of resources configured in a group.",
			[]string{"group"}, nil,
		),
		crmMonResourceGroupActive: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "resource", "active"),
			"Resource is active.",
			[]string{"id", "group", "node_name", "resource_agent", "role", "target_role"}, nil,
		),
		crmMonResourceGroupOrphaned: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "resource", "orphaned"),
			"Resource is orphaned.",
			[]string{"id", "group", "node_name", "resource_agent", "role", "target_role"}, nil,
		),
		crmMonResourceGroupBlocked: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "resource", "blocked"),
			"Resource is blocked.",
			[]string{"id", "group", "node_name", "resource_agent", "role", "target_role"}, nil,
		),
		crmMonResourceGroupManaged: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "resource", "managed"),
			"Resource is managed.",
			[]string{"id", "group", "node_name", "resource_agent", "role", "target_role"}, nil,
		),
		crmMonResourceGroupFailed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "resource", "failed"),
			"Resource is failed.",
			[]string{"id", "group", "node_name", "resource_agent", "role", "target_role"}, nil,
		),
		crmMonResourceGroupFailureIgnored: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "resource", "failure_ignored"),
			"Resource failure ignored.",
			[]string{"id", "group", "node_name", "resource_agent", "role", "target_role"}, nil,
		),

		// Node Resources Clone metrics
		crmMonResourceCloneMultistate: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "clone", "multistate"),
			"Resource is a multi-state one.",
			[]string{"clone_id"}, nil,
		),
		crmMonResourceCloneNumActive: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "clone", "num_active"),
			"Number of running clone instances",
			[]string{"clone_id"}, nil,
		),
		crmMonResourceCloneNumPromoted: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "clone", "num_promoted"),
			"Number of promoted clone instances",
			[]string{"clone_id"}, nil,
		),
		crmMonResourceClonePromoted: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "resource", "promoted"),
			"Resource is promoted.",
			[]string{"id", "clone_id", "node_name", "resource_agent", "role", "target_role"}, nil,
		),
		crmMonResourceCloneActive: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "resource", "active"),
			"Resource is active.",
			[]string{"id", "clone_id", "node_name", "resource_agent", "role", "target_role"}, nil,
		),
		crmMonResourceCloneOrphaned: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "resource", "orphaned"),
			"Resource is orphaned.",
			[]string{"id", "clone_id", "node_name", "resource_agent", "role", "target_role"}, nil,
		),
		crmMonResourceCloneBlocked: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "resource", "blocked"),
			"Resource is blocked.",
			[]string{"id", "clone_id", "node_name", "resource_agent", "role", "target_role"}, nil,
		),
		crmMonResourceCloneManaged: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "resource", "managed"),
			"Resource is managed.",
			[]string{"id", "clone_id", "node_name", "resource_agent", "role", "target_role"}, nil,
		),
		crmMonResourceCloneFailed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "resource", "failed"),
			"Resource is failed.",
			[]string{"id", "clone_id", "node_name", "resource_agent", "role", "target_role"}, nil,
		),
		crmMonResourceCloneFailureIgnored: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "resource", "failure_ignored"),
			"Resource failure ignored.",
			[]string{"id", "clone_id", "node_name", "resource_agent", "role", "target_role"}, nil,
		),

		// Failures metrics
		crmMonFailuresCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "failures", "count"),
			"Cluster failures count.",
			[]string{"name"}, nil,
		),
		crmMonFailureDescription: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "failure", "description"),
			"Metric with a constant '1' value labeled by the failure description.",
			[]string{"node", "op_key", "status", "task"}, nil,
		),
		// Bans metrics
		crmMonBansCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bans", "count"),
			"Cluster bans count.",
			[]string{"name"}, nil,
		),
		crmMonBanDescription: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bans", "description"),
			"Metric with a constant '1' value labeled by the ban description.",
			[]string{"id", "resource", "node", "weight", "master_only"}, nil,
		),
	}, nil
}

// Update calls (*crmMonCollector).getCrmMon to get the platform specific
// memory metrics.
func (c *crmMonCollector) Update(ch chan<- prometheus.Metric) error {
	err := c.getCrmMonInfo(ch)
	if err != nil {
		return fmt.Errorf("couldn't get crm_mon information: %s", err)
	}

	return nil
}
