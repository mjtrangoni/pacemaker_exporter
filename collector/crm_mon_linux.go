// Copyright 2018 Mario Trangoni
// Copyright 2015 The Prometheus Authors
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

package collector

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

// execute crm_mon utility.
func crmMonExec(args ...string) ([]byte, error) {
	cmd := exec.Command(*crmMonPath, args...)
	// Disable localization for parsing.
	cmd.Env = append(os.Environ(), "LANG=C")
	out, err := cmd.Output()
	if err != nil {
		log.Errorf("error while calling '%s %s': %v", *crmMonPath,
			strings.Join(args, " "), err)
	}
	return out, err
}

// parseCrmMonXML returns an XML structs.
func parseCrmMonXML(data []byte) (CrmMonStruct, error) {
	var crmMonOut CrmMonStruct
	err := xml.Unmarshal(data, &crmMonOut)
	if err != nil {
		log.Errorln(err)
		return crmMonOut, err
	}
	return crmMonOut, nil
}

// getCrmMonInfo returns crm_mon information
func (c *crmMonCollector) getCrmMonInfo(ch chan<- prometheus.Metric) error {
	outBytes, err := crmMonExec("-X")
	if err != nil {
		log.Errorln(err)
		return err
	}

	crmMonStruct, err := parseCrmMonXML(outBytes)
	if err != nil {
		log.Errorln(err)
		return err
	}

	elemEnabledSlice := strings.Split(*crmMonElemEnabled, ",")

	// Summary metrics
	if stringInSlice("summary", elemEnabledSlice) {
		err = c.exposeSummary(ch, crmMonStruct.Summary)
		if err != nil {
			log.Errorln(err)
		}
	}

	// Nodes section metrics
	if stringInSlice("nodes", elemEnabledSlice) {
		err = c.exposeNodes(ch, crmMonStruct.Nodes)
		if err != nil {
			log.Errorln(err)
		}
	}

	// Node attribute section metrics
	if stringInSlice("nodes", elemEnabledSlice) {
		err = c.exposeNodeAttributes(ch, crmMonStruct.NodeAttributes)
		if err != nil {
			log.Errorln(err)
		}
	}

	// Resources section metrics
	if stringInSlice("resources", elemEnabledSlice) {
		err = c.exposeResources(ch, crmMonStruct.Resources)
		if err != nil {
			log.Errorln(err)
		}
	}

	if stringInSlice("resources_group", elemEnabledSlice) {
		err = c.exposeResourcesGroup(ch, crmMonStruct.Resources)
		if err != nil {
			log.Errorln(err)
		}
	}

	if stringInSlice("failures", elemEnabledSlice) {
		err = c.exposeFailures(ch, crmMonStruct)
		if err != nil {
			log.Errorln(err)
		}
	}

	return nil
}

// HTMLHandler returns crm_mon -w
func HTMLHandler(w http.ResponseWriter, r *http.Request) {
	outBytes, err := crmMonExec("-w")
	if err != nil {
		log.Warnln("Error running `crm_mon -w`", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		_, err = w.Write([]byte(fmt.Sprintf("Couldn't create %s", err)))
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err = w.Write(outBytes)
	if err != nil {
		log.Fatal(err)
	}
}

// XMLHandler returns crm_mon -X
func XMLHandler(w http.ResponseWriter, r *http.Request) {
	outBytes, err := crmMonExec("-X")
	if err != nil {
		log.Warnln("Error running `crm_mon -X`", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		_, err = w.Write([]byte(fmt.Sprintf("Couldn't create %s", err)))
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	_, err = w.Write(outBytes)
	if err != nil {
		log.Fatal(err)
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// expose Summary metrics
func (c *crmMonCollector) exposeSummary(ch chan<- prometheus.Metric, summaryStruct SummaryStruct) error {
	ch <- prometheus.MustNewConstMetric(c.crmMonInfo, prometheus.GaugeValue,
		1.0, summaryStruct.CurrentDC.Version)

	lastUpdateTime, err := time.Parse("Mon Jan _2 15:04:05 2006",
		summaryStruct.LastUpdate.Time)
	if err != nil {
		log.Errorln(err)
	} else {
		ch <- prometheus.MustNewConstMetric(c.crmMonLastUpdate,
			prometheus.GaugeValue, float64(lastUpdateTime.Unix()),
			summaryStruct.Stack.Type)
	}

	lastChangeTime, err := time.Parse("Mon Jan _2 15:04:05 2006",
		summaryStruct.LastChange.Time)
	if err != nil {
		log.Errorln(err)
	} else {
		ch <- prometheus.MustNewConstMetric(c.crmMonLastChange,
			prometheus.GaugeValue, float64(lastChangeTime.Unix()),
			summaryStruct.LastChange.User,
			summaryStruct.LastChange.Client,
			summaryStruct.LastChange.Origin)
	}

	if summaryStruct.CurrentDC.Present {
		ch <- prometheus.MustNewConstMetric(c.crmMonDCPresent,
			prometheus.GaugeValue, 1.0, summaryStruct.CurrentDC.Name)
	} else {
		ch <- prometheus.MustNewConstMetric(c.crmMonDCPresent,
			prometheus.GaugeValue, 0.0, summaryStruct.CurrentDC.Name)
	}

	if summaryStruct.CurrentDC.Quorum {
		ch <- prometheus.MustNewConstMetric(c.crmMonDCQuorum,
			prometheus.GaugeValue, 1.0, summaryStruct.CurrentDC.Name)
	} else {
		ch <- prometheus.MustNewConstMetric(c.crmMonDCQuorum,
			prometheus.GaugeValue, 0.0, summaryStruct.CurrentDC.Name)
	}

	ch <- prometheus.MustNewConstMetric(c.crmMonNodesConfigured,
		prometheus.GaugeValue, summaryStruct.NodesConfigured.Number,
		summaryStruct.NodesConfigured.ExpectedVotes)

	ch <- prometheus.MustNewConstMetric(c.crmMonResourcesConfigured,
		prometheus.GaugeValue, summaryStruct.ResourcesConfigured.Number,
		summaryStruct.CurrentDC.Name)

	ch <- prometheus.MustNewConstMetric(c.crmMonResourcesDisabled,
		prometheus.GaugeValue, summaryStruct.ResourcesConfigured.Disabled,
		summaryStruct.CurrentDC.Name)

	ch <- prometheus.MustNewConstMetric(c.crmMonResourcesBlocked,
		prometheus.GaugeValue, summaryStruct.ResourcesConfigured.Blocked,
		summaryStruct.CurrentDC.Name)

	if summaryStruct.ClusterOptions.StonithEnabled {
		ch <- prometheus.MustNewConstMetric(c.crmMonStonith,
			prometheus.GaugeValue, 1.0, summaryStruct.CurrentDC.Name)
	} else {
		ch <- prometheus.MustNewConstMetric(c.crmMonStonith,
			prometheus.GaugeValue, 0.0, summaryStruct.CurrentDC.Name)
	}

	if summaryStruct.ClusterOptions.SymmetricCluster {
		ch <- prometheus.MustNewConstMetric(c.crmMonSymmetricCluster,
			prometheus.GaugeValue, 1.0, summaryStruct.CurrentDC.Name)
	} else {
		ch <- prometheus.MustNewConstMetric(c.crmMonSymmetricCluster,
			prometheus.GaugeValue, 0.0, summaryStruct.CurrentDC.Name)
	}

	if summaryStruct.ClusterOptions.MaintenanceMode {
		ch <- prometheus.MustNewConstMetric(c.crmMonMaintenanceMode,
			prometheus.GaugeValue, 1.0, summaryStruct.CurrentDC.Name)
	} else {
		ch <- prometheus.MustNewConstMetric(c.crmMonMaintenanceMode,
			prometheus.GaugeValue, 0.0, summaryStruct.CurrentDC.Name)
	}

	return nil
}

// expose Nodes metrics
func (c *crmMonCollector) exposeNodes(ch chan<- prometheus.Metric, nodesStruct NodesStruct) error {
	for _, node := range nodesStruct.Node {
		ch <- prometheus.MustNewConstMetric(c.crmMonNodeID,
			prometheus.GaugeValue, node.ID, node.Name, node.Type)
		if node.Online {
			ch <- prometheus.MustNewConstMetric(c.crmMonNodeOnline,
				prometheus.GaugeValue, 1.0, node.Name)
		} else {
			ch <- prometheus.MustNewConstMetric(c.crmMonNodeOnline,
				prometheus.GaugeValue, 0.0, node.Name)
		}
		if node.Standby {
			ch <- prometheus.MustNewConstMetric(c.crmMonNodeStandby,
				prometheus.GaugeValue, 1.0, node.Name)
		} else {
			ch <- prometheus.MustNewConstMetric(c.crmMonNodeStandby,
				prometheus.GaugeValue, 0.0, node.Name)
		}
		if node.StandbyOnFail {
			ch <- prometheus.MustNewConstMetric(c.crmMonNodeStandbyOnFail,
				prometheus.GaugeValue, 1.0, node.Name)
		} else {
			ch <- prometheus.MustNewConstMetric(c.crmMonNodeStandbyOnFail,
				prometheus.GaugeValue, 0.0, node.Name)
		}
		if node.Maintenance {
			ch <- prometheus.MustNewConstMetric(c.crmMonNodeMaintenance,
				prometheus.GaugeValue, 1.0, node.Name)
		} else {
			ch <- prometheus.MustNewConstMetric(c.crmMonNodeMaintenance,
				prometheus.GaugeValue, 0.0, node.Name)
		}
		if node.Pending {
			ch <- prometheus.MustNewConstMetric(c.crmMonNodePending,
				prometheus.GaugeValue, 1.0, node.Name)
		} else {
			ch <- prometheus.MustNewConstMetric(c.crmMonNodePending,
				prometheus.GaugeValue, 0.0, node.Name)
		}
		if node.Unclean {
			ch <- prometheus.MustNewConstMetric(c.crmMonNodeUnclean,
				prometheus.GaugeValue, 1.0, node.Name)
		} else {
			ch <- prometheus.MustNewConstMetric(c.crmMonNodeUnclean,
				prometheus.GaugeValue, 0.0, node.Name)
		}
		if node.Shutdown {
			ch <- prometheus.MustNewConstMetric(c.crmMonNodeShutdown,
				prometheus.GaugeValue, 1.0, node.Name)
		} else {
			ch <- prometheus.MustNewConstMetric(c.crmMonNodeShutdown,
				prometheus.GaugeValue, 0.0, node.Name)
		}
		if node.ExpectedUp {
			ch <- prometheus.MustNewConstMetric(c.crmMonNodeExpectedUp,
				prometheus.GaugeValue, 1.0, node.Name)
		} else {
			ch <- prometheus.MustNewConstMetric(c.crmMonNodeExpectedUp,
				prometheus.GaugeValue, 0.0, node.Name)
		}
		if node.IsDC {
			ch <- prometheus.MustNewConstMetric(c.crmMonNodeIsDC,
				prometheus.GaugeValue, 1.0, node.Name)
		} else {
			ch <- prometheus.MustNewConstMetric(c.crmMonNodeIsDC,
				prometheus.GaugeValue, 0.0, node.Name)
		}
		ch <- prometheus.MustNewConstMetric(c.crmMonNodeResourcesRunning,
			prometheus.GaugeValue, node.ResourcesRunning, node.Name)
	}
	return nil
}

// expose Node Attribute metrics
func (c *crmMonCollector) exposeNodeAttributes(ch chan<- prometheus.Metric, nodeAttrStruct NodeAttrStruct) error {
	for _, node := range nodeAttrStruct.Node {
		for _, attribute := range node.Attribute {
			ch <- prometheus.MustNewConstMetric(c.crmMonNodeAttribute,
				prometheus.GaugeValue, attribute.Value, node.Name,
				attribute.Name)
		}
	}
	return nil
}

// expose Resources metrics
func (c *crmMonCollector) exposeResources(ch chan<- prometheus.Metric, resourcesStruct ResourcesStruct) error {
	for _, node := range resourcesStruct.Resource {
		for _, nodeName := range node.Node {
			if node.Active {
				ch <- prometheus.MustNewConstMetric(c.crmMonResourceActive,
					prometheus.GaugeValue, 1.0, node.ID, nodeName.Name,
					node.ResourceAgent, node.Role, node.TargetRole)
			} else {
				ch <- prometheus.MustNewConstMetric(c.crmMonResourceActive,
					prometheus.GaugeValue, 0.0, node.ID, nodeName.Name,
					node.ResourceAgent, node.Role, node.TargetRole)
			}
			if node.Orphaned {
				ch <- prometheus.MustNewConstMetric(c.crmMonResourceOrphaned,
					prometheus.GaugeValue, 1.0, node.ID, nodeName.Name,
					node.ResourceAgent, node.Role, node.TargetRole)
			} else {
				ch <- prometheus.MustNewConstMetric(c.crmMonResourceOrphaned,
					prometheus.GaugeValue, 0.0, node.ID, nodeName.Name,
					node.ResourceAgent, node.Role, node.TargetRole)
			}
			if node.Blocked {
				ch <- prometheus.MustNewConstMetric(c.crmMonResourceBlocked,
					prometheus.GaugeValue, 1.0, node.ID, nodeName.Name,
					node.ResourceAgent, node.Role, node.TargetRole)
			} else {
				ch <- prometheus.MustNewConstMetric(c.crmMonResourceBlocked,
					prometheus.GaugeValue, 0.0, node.ID, nodeName.Name,
					node.ResourceAgent, node.Role, node.TargetRole)
			}
			if node.Managed {
				ch <- prometheus.MustNewConstMetric(c.crmMonResourceManaged,
					prometheus.GaugeValue, 1.0, node.ID, nodeName.Name,
					node.ResourceAgent, node.Role, node.TargetRole)
			} else {
				ch <- prometheus.MustNewConstMetric(c.crmMonResourceManaged,
					prometheus.GaugeValue, 0.0, node.ID, nodeName.Name,
					node.ResourceAgent, node.Role, node.TargetRole)
			}
			if node.Failed {
				ch <- prometheus.MustNewConstMetric(c.crmMonResourceFailed,
					prometheus.GaugeValue, 1.0, node.ID, nodeName.Name,
					node.ResourceAgent, node.Role, node.TargetRole)
			} else {
				ch <- prometheus.MustNewConstMetric(c.crmMonResourceFailed,
					prometheus.GaugeValue, 0.0, node.ID, nodeName.Name,
					node.ResourceAgent, node.Role, node.TargetRole)
			}
			if node.FailureIgnored {
				ch <- prometheus.MustNewConstMetric(c.crmMonResourceFailureIgnored,
					prometheus.GaugeValue, 1.0, node.ID, nodeName.Name,
					node.ResourceAgent, node.Role, node.TargetRole)
			} else {
				ch <- prometheus.MustNewConstMetric(c.crmMonResourceFailureIgnored,
					prometheus.GaugeValue, 0.0, node.ID, nodeName.Name,
					node.ResourceAgent, node.Role, node.TargetRole)
			}
		}
		ch <- prometheus.MustNewConstMetric(c.crmMonResourceRunningOn,
			prometheus.GaugeValue, node.NodesRunningOn, node.ID,
			node.ResourceAgent, node.Role, node.TargetRole)
	}
	return nil
}

// expose Resources by Group metrics
func (c *crmMonCollector) exposeResourcesGroup(ch chan<- prometheus.Metric, resourcesStruct ResourcesStruct) error {
	for _, group := range resourcesStruct.Group {
		ch <- prometheus.MustNewConstMetric(c.crmMonResourcesGroup,
			prometheus.GaugeValue, group.NumberResources, group.ID)
		for _, node := range group.Resource {
			for _, nodeName := range node.Node {
				if node.Active {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceGroupActive,
						prometheus.GaugeValue, 1.0, node.ID, group.ID,
						nodeName.Name, node.ResourceAgent, node.Role,
						node.TargetRole)
				} else {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceGroupActive,
						prometheus.GaugeValue, 0.0, node.ID, group.ID,
						nodeName.Name, node.ResourceAgent, node.Role,
						node.TargetRole)
				}
				if node.Orphaned {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceGroupOrphaned,
						prometheus.GaugeValue, 1.0, node.ID, group.ID,
						nodeName.Name, node.ResourceAgent, node.Role,
						node.TargetRole)
				} else {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceGroupOrphaned,
						prometheus.GaugeValue, 0.0, node.ID, group.ID,
						nodeName.Name, node.ResourceAgent, node.Role,
						node.TargetRole)
				}
				if node.Blocked {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceGroupBlocked,
						prometheus.GaugeValue, 1.0, node.ID, group.ID,
						nodeName.Name, node.ResourceAgent, node.Role,
						node.TargetRole)
				} else {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceGroupBlocked,
						prometheus.GaugeValue, 0.0, node.ID, group.ID,
						nodeName.Name, node.ResourceAgent, node.Role,
						node.TargetRole)
				}
				if node.Managed {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceGroupManaged,
						prometheus.GaugeValue, 1.0, node.ID, group.ID,
						nodeName.Name, node.ResourceAgent, node.Role,
						node.TargetRole)
				} else {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceGroupManaged,
						prometheus.GaugeValue, 0.0, node.ID, group.ID,
						nodeName.Name, node.ResourceAgent, node.Role,
						node.TargetRole)
				}
				if node.Failed {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceGroupFailed,
						prometheus.GaugeValue, 1.0, node.ID, group.ID,
						nodeName.Name, node.ResourceAgent, node.Role,
						node.TargetRole)
				} else {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceGroupFailed,
						prometheus.GaugeValue, 0.0, node.ID, group.ID,
						nodeName.Name, node.ResourceAgent, node.Role,
						node.TargetRole)
				}
				if node.FailureIgnored {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceGroupFailureIgnored,
						prometheus.GaugeValue, 1.0, node.ID, group.ID,
						nodeName.Name, node.ResourceAgent, node.Role,
						node.TargetRole)
				} else {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceGroupFailureIgnored,
						prometheus.GaugeValue, 0.0, node.ID, group.ID,
						nodeName.Name, node.ResourceAgent, node.Role,
						node.TargetRole)
				}
			}
			ch <- prometheus.MustNewConstMetric(c.crmMonResourceGroupRunningOn,
				prometheus.GaugeValue, node.NodesRunningOn, node.ID, group.ID,
				node.ResourceAgent, node.Role, node.TargetRole)
		}
	}
	return nil
}

// expose Failures metrics
func (c *crmMonCollector) exposeFailures(ch chan<- prometheus.Metric, crmMonStruct CrmMonStruct) error {
	ch <- prometheus.MustNewConstMetric(c.crmMonFailuresCount,
		prometheus.GaugeValue, float64(len(crmMonStruct.Failures.Failure)),
		crmMonStruct.Summary.CurrentDC.Name)

	for _, failure := range crmMonStruct.Failures.Failure {
		ch <- prometheus.MustNewConstMetric(c.crmMonFailureDescription,
			prometheus.GaugeValue, 1.0,
			failure.Node, failure.OpKey, failure.Status, failure.Task)
	}
	return nil
}
