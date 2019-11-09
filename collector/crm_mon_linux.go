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
	outBytes, err := crmMonExec("-Xr")
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
		c.exposeNodes(ch, crmMonStruct.Nodes)
	}

	// Node attribute section metrics
	if stringInSlice("nodes", elemEnabledSlice) {
		c.exposeNodeAttributes(ch, crmMonStruct.NodeAttributes)
	}

	// Resources section metrics
	if stringInSlice("clones", elemEnabledSlice) {
		c.exposeResourcesClone(ch, crmMonStruct.Resources)
	}

	if stringInSlice("resources", elemEnabledSlice) {
		c.exposeResources(ch, crmMonStruct.Resources)
	}

	if stringInSlice("resources_group", elemEnabledSlice) {
		c.exposeResourcesGroup(ch, crmMonStruct.Resources)
	}

	if stringInSlice("failures", elemEnabledSlice) {
		c.exposeFailures(ch, crmMonStruct)
	}

	if stringInSlice("bans", elemEnabledSlice) {
		c.exposeBans(ch, crmMonStruct)
	}

	return nil
}

// HTMLHandler returns crm_mon -wr
func HTMLHandler(w http.ResponseWriter, r *http.Request) {
	outBytes, err := crmMonExec("-wr")
	if err != nil {
		log.Warnln("Error running `crm_mon -wr`", err)
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

// XMLHandler returns crm_mon -Xr
func XMLHandler(w http.ResponseWriter, r *http.Request) {
	outBytes, err := crmMonExec("-Xr")
	if err != nil {
		log.Warnln("Error running `crm_mon -Xr`", err)
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
		return err
	}
	ch <- prometheus.MustNewConstMetric(c.crmMonLastUpdate,
		prometheus.GaugeValue, float64(lastUpdateTime.Unix()),
		summaryStruct.Stack.Type)

	lastChangeTime, err := time.Parse("Mon Jan _2 15:04:05 2006",
		summaryStruct.LastChange.Time)
	if err != nil {
		log.Errorln(err)
		return err
	}
	ch <- prometheus.MustNewConstMetric(c.crmMonLastChange,
		prometheus.GaugeValue, float64(lastChangeTime.Unix()),
		summaryStruct.LastChange.User, summaryStruct.LastChange.Client,
		summaryStruct.LastChange.Origin)

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
func (c *crmMonCollector) exposeNodes(ch chan<- prometheus.Metric, nodesStruct NodesStruct) {
	for _, node := range nodesStruct.Node {
		ch <- prometheus.MustNewConstMetric(c.crmMonNodeID,
			prometheus.GaugeValue, 1.0, node.Name, node.Type, node.ID)

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
}

// expose Node Attribute metrics
func (c *crmMonCollector) exposeNodeAttributes(ch chan<- prometheus.Metric, nodeAttrStruct NodeAttrStruct) {
	for _, node := range nodeAttrStruct.Node {
		for _, attribute := range node.Attribute {
			ch <- prometheus.MustNewConstMetric(c.crmMonNodeAttribute,
				prometheus.GaugeValue, 1.0, node.Name,
				attribute.Name, attribute.Value)
		}
	}
}

// expose Resources metrics
func (c *crmMonCollector) exposeResources(ch chan<- prometheus.Metric, resourcesStruct ResourcesStruct) {
	for _, resource := range resourcesStruct.Resource {
		for _, nodeName := range resource.Node {
			if resource.Active {
				ch <- prometheus.MustNewConstMetric(c.crmMonResourceActive,
					prometheus.GaugeValue, 1.0, resource.ID, nodeName.Name,
					resource.ResourceAgent, resource.Role, resource.TargetRole)
			} else {
				ch <- prometheus.MustNewConstMetric(c.crmMonResourceActive,
					prometheus.GaugeValue, 0.0, resource.ID, nodeName.Name,
					resource.ResourceAgent, resource.Role, resource.TargetRole)
			}

			if resource.Orphaned {
				ch <- prometheus.MustNewConstMetric(c.crmMonResourceOrphaned,
					prometheus.GaugeValue, 1.0, resource.ID, nodeName.Name,
					resource.ResourceAgent, resource.Role, resource.TargetRole)
			} else {
				ch <- prometheus.MustNewConstMetric(c.crmMonResourceOrphaned,
					prometheus.GaugeValue, 0.0, resource.ID, nodeName.Name,
					resource.ResourceAgent, resource.Role, resource.TargetRole)
			}

			if resource.Blocked {
				ch <- prometheus.MustNewConstMetric(c.crmMonResourceBlocked,
					prometheus.GaugeValue, 1.0, resource.ID, nodeName.Name,
					resource.ResourceAgent, resource.Role, resource.TargetRole)
			} else {
				ch <- prometheus.MustNewConstMetric(c.crmMonResourceBlocked,
					prometheus.GaugeValue, 0.0, resource.ID, nodeName.Name,
					resource.ResourceAgent, resource.Role, resource.TargetRole)
			}

			if resource.Managed {
				ch <- prometheus.MustNewConstMetric(c.crmMonResourceManaged,
					prometheus.GaugeValue, 1.0, resource.ID, nodeName.Name,
					resource.ResourceAgent, resource.Role, resource.TargetRole)
			} else {
				ch <- prometheus.MustNewConstMetric(c.crmMonResourceManaged,
					prometheus.GaugeValue, 0.0, resource.ID, nodeName.Name,
					resource.ResourceAgent, resource.Role, resource.TargetRole)
			}

			if resource.Failed {
				ch <- prometheus.MustNewConstMetric(c.crmMonResourceFailed,
					prometheus.GaugeValue, 1.0, resource.ID, nodeName.Name,
					resource.ResourceAgent, resource.Role, resource.TargetRole)
			} else {
				ch <- prometheus.MustNewConstMetric(c.crmMonResourceFailed,
					prometheus.GaugeValue, 0.0, resource.ID, nodeName.Name,
					resource.ResourceAgent, resource.Role, resource.TargetRole)
			}

			if resource.FailureIgnored {
				ch <- prometheus.MustNewConstMetric(c.crmMonResourceFailureIgnored,
					prometheus.GaugeValue, 1.0, resource.ID, nodeName.Name,
					resource.ResourceAgent, resource.Role, resource.TargetRole)
			} else {
				ch <- prometheus.MustNewConstMetric(c.crmMonResourceFailureIgnored,
					prometheus.GaugeValue, 0.0, resource.ID, nodeName.Name,
					resource.ResourceAgent, resource.Role, resource.TargetRole)
			}
		}
	}
}

// expose Resources by Group metrics
func (c *crmMonCollector) exposeResourcesGroup(ch chan<- prometheus.Metric, resourcesStruct ResourcesStruct) {
	for _, group := range resourcesStruct.Group {
		ch <- prometheus.MustNewConstMetric(c.crmMonResourcesGroup,
			prometheus.GaugeValue, group.NumberResources, group.ID)

		for _, resource := range group.Resource {
			for _, nodeName := range resource.Node {
				if resource.Active {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceGroupActive,
						prometheus.GaugeValue, 1.0, resource.ID, group.ID,
						nodeName.Name, resource.ResourceAgent, resource.Role,
						resource.TargetRole)
				} else {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceGroupActive,
						prometheus.GaugeValue, 0.0, resource.ID, group.ID,
						nodeName.Name, resource.ResourceAgent, resource.Role,
						resource.TargetRole)
				}

				if resource.Orphaned {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceGroupOrphaned,
						prometheus.GaugeValue, 1.0, resource.ID, group.ID,
						nodeName.Name, resource.ResourceAgent, resource.Role,
						resource.TargetRole)
				} else {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceGroupOrphaned,
						prometheus.GaugeValue, 0.0, resource.ID, group.ID,
						nodeName.Name, resource.ResourceAgent, resource.Role,
						resource.TargetRole)
				}

				if resource.Blocked {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceGroupBlocked,
						prometheus.GaugeValue, 1.0, resource.ID, group.ID,
						nodeName.Name, resource.ResourceAgent, resource.Role,
						resource.TargetRole)
				} else {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceGroupBlocked,
						prometheus.GaugeValue, 0.0, resource.ID, group.ID,
						nodeName.Name, resource.ResourceAgent, resource.Role,
						resource.TargetRole)
				}

				if resource.Managed {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceGroupManaged,
						prometheus.GaugeValue, 1.0, resource.ID, group.ID,
						nodeName.Name, resource.ResourceAgent, resource.Role,
						resource.TargetRole)
				} else {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceGroupManaged,
						prometheus.GaugeValue, 0.0, resource.ID, group.ID,
						nodeName.Name, resource.ResourceAgent, resource.Role,
						resource.TargetRole)
				}

				if resource.Failed {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceGroupFailed,
						prometheus.GaugeValue, 1.0, resource.ID, group.ID,
						nodeName.Name, resource.ResourceAgent, resource.Role,
						resource.TargetRole)
				} else {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceGroupFailed,
						prometheus.GaugeValue, 0.0, resource.ID, group.ID,
						nodeName.Name, resource.ResourceAgent, resource.Role,
						resource.TargetRole)
				}

				if resource.FailureIgnored {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceGroupFailureIgnored,
						prometheus.GaugeValue, 1.0, resource.ID, group.ID,
						nodeName.Name, resource.ResourceAgent, resource.Role,
						resource.TargetRole)
				} else {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceGroupFailureIgnored,
						prometheus.GaugeValue, 0.0, resource.ID, group.ID,
						nodeName.Name, resource.ResourceAgent, resource.Role,
						resource.TargetRole)
				}
			}
		}
	}
}

// expose Resources by Clone metrics
func (c *crmMonCollector) exposeResourcesClone(ch chan<- prometheus.Metric, resourcesStruct ResourcesStruct) {
	for _, clone := range resourcesStruct.Clone {
		numActive := 0
		numPromoted := 0

		if clone.MultiState {
			ch <- prometheus.MustNewConstMetric(c.crmMonResourceCloneMultistate,
				prometheus.GaugeValue, 1.0, clone.ID)
		} else {
			ch <- prometheus.MustNewConstMetric(c.crmMonResourceCloneMultistate,
				prometheus.GaugeValue, 0.0, clone.ID)
		}

		for _, resource := range clone.Resource {
			for _, nodeName := range resource.Node {
				if clone.MultiState {
					if resource.Role == "Master" {
						ch <- prometheus.MustNewConstMetric(c.crmMonResourceClonePromoted,
							prometheus.GaugeValue, 1.0, resource.ID, clone.ID,
							nodeName.Name, resource.ResourceAgent, resource.Role,
							resource.TargetRole)
						numPromoted++
					} else {
						ch <- prometheus.MustNewConstMetric(c.crmMonResourceClonePromoted,
							prometheus.GaugeValue, 0.0, resource.ID, clone.ID,
							nodeName.Name, resource.ResourceAgent, resource.Role,
							resource.TargetRole)
					}
				}

				if resource.Active {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceCloneActive,
						prometheus.GaugeValue, 1.0, resource.ID, clone.ID,
						nodeName.Name, resource.ResourceAgent, resource.Role,
						resource.TargetRole)
					numActive++
				} else {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceCloneActive,
						prometheus.GaugeValue, 0.0, resource.ID, clone.ID,
						nodeName.Name, resource.ResourceAgent, resource.Role,
						resource.TargetRole)
				}

				if resource.Orphaned {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceCloneOrphaned,
						prometheus.GaugeValue, 1.0, resource.ID, clone.ID,
						nodeName.Name, resource.ResourceAgent, resource.Role,
						resource.TargetRole)
				} else {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceCloneOrphaned,
						prometheus.GaugeValue, 0.0, resource.ID, clone.ID,
						nodeName.Name, resource.ResourceAgent, resource.Role,
						resource.TargetRole)
				}

				if resource.Blocked {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceCloneBlocked,
						prometheus.GaugeValue, 1.0, resource.ID, clone.ID,
						nodeName.Name, resource.ResourceAgent, resource.Role,
						resource.TargetRole)
				} else {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceCloneBlocked,
						prometheus.GaugeValue, 0.0, resource.ID, clone.ID,
						nodeName.Name, resource.ResourceAgent, resource.Role,
						resource.TargetRole)
				}

				if resource.Managed {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceCloneManaged,
						prometheus.GaugeValue, 1.0, resource.ID, clone.ID,
						nodeName.Name, resource.ResourceAgent, resource.Role,
						resource.TargetRole)
				} else {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceCloneManaged,
						prometheus.GaugeValue, 0.0, resource.ID, clone.ID,
						nodeName.Name, resource.ResourceAgent, resource.Role,
						resource.TargetRole)
				}

				if resource.Failed {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceCloneFailed,
						prometheus.GaugeValue, 1.0, resource.ID, clone.ID,
						nodeName.Name, resource.ResourceAgent, resource.Role,
						resource.TargetRole)
				} else {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceCloneFailed,
						prometheus.GaugeValue, 0.0, resource.ID, clone.ID,
						nodeName.Name, resource.ResourceAgent, resource.Role,
						resource.TargetRole)
				}

				if resource.FailureIgnored {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceCloneFailureIgnored,
						prometheus.GaugeValue, 1.0, resource.ID, clone.ID,
						nodeName.Name, resource.ResourceAgent, resource.Role,
						resource.TargetRole)
				} else {
					ch <- prometheus.MustNewConstMetric(c.crmMonResourceCloneFailureIgnored,
						prometheus.GaugeValue, 0.0, resource.ID, clone.ID,
						nodeName.Name, resource.ResourceAgent, resource.Role,
						resource.TargetRole)
				}
			}
		}
		ch <- prometheus.MustNewConstMetric(c.crmMonResourceCloneNumActive,
			prometheus.GaugeValue, float64(numActive), clone.ID)

		if clone.MultiState {
			ch <- prometheus.MustNewConstMetric(c.crmMonResourceCloneNumPromoted,
				prometheus.GaugeValue, float64(numPromoted), clone.ID)
		}
	}
}

// expose Failures metrics
func (c *crmMonCollector) exposeFailures(ch chan<- prometheus.Metric, crmMonStruct CrmMonStruct) {
	ch <- prometheus.MustNewConstMetric(c.crmMonFailuresCount,
		prometheus.GaugeValue, float64(len(crmMonStruct.Failures.Failure)),
		crmMonStruct.Summary.CurrentDC.Name)

	for idx := range crmMonStruct.Failures.Failure {
		ch <- prometheus.MustNewConstMetric(c.crmMonFailureDescription,
			prometheus.GaugeValue, 1.0,
			crmMonStruct.Failures.Failure[idx].Node,
			crmMonStruct.Failures.Failure[idx].OpKey,
			crmMonStruct.Failures.Failure[idx].Status,
			crmMonStruct.Failures.Failure[idx].Task)
	}
}

// expose Bans metrics
func (c *crmMonCollector) exposeBans(ch chan<- prometheus.Metric, crmMonStruct CrmMonStruct) {
	ch <- prometheus.MustNewConstMetric(c.crmMonBansCount,
		prometheus.GaugeValue, float64(len(crmMonStruct.Bans.Ban)),
		crmMonStruct.Summary.CurrentDC.Name)

	for _, ban := range crmMonStruct.Bans.Ban {
		ch <- prometheus.MustNewConstMetric(c.crmMonBanDescription,
			prometheus.GaugeValue, 1.0,
			ban.ID, ban.Resource, ban.Node, ban.Weight, ban.MasterOnly)
	}
}
