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

import "encoding/xml"

// CrmMonStruct struct stores the crm_mon XML information
type CrmMonStruct struct {
	XMLName        xml.Name          `xml:"crm_mon"`
	Summary        SummaryStruct     `xml:"summary"`
	Nodes          NodesStruct       `xml:"nodes"`
	Resources      ResourcesStruct   `xml:"resources"`
	NodeAttributes NodeAttrStruct    `xml:"node_attributes"`
	NodeHistory    NodeHistoryStruct `xml:"node_history"`
	Failures       FailuresStruct    `xml:"failures"`
	Bans           BansStruct        `xml:"bans"`
}

// SummaryStruct struct stores the crm_mon XML summary information
type SummaryStruct struct {
	Stack struct {
		Type string `xml:"type,attr"`
	} `xml:"stack"`
	// Current Designated Controller
	CurrentDC struct {
		Present bool   `xml:"present,attr"`
		Quorum  bool   `xml:"with_quorum,attr"`
		Version string `xml:"version,attr"`
		Name    string `xml:"name,attr"`
		ID      string `xml:"id,attr"`
	} `xml:"current_dc"`
	LastUpdate struct {
		Time string `xml:"time,attr"`
	} `xml:"last_update"`
	LastChange struct {
		Time   string `xml:"time,attr"`
		User   string `xml:"user,attr"`
		Client string `xml:"client,attr"`
		Origin string `xml:"origin,attr"`
	} `xml:"last_change"`
	NodesConfigured struct {
		Number        float64 `xml:"number,attr"`
		ExpectedVotes string  `xml:"expected_votes,attr"`
	} `xml:"nodes_configured"`
	ResourcesConfigured struct {
		Number   float64 `xml:"number,attr"`
		Disabled float64 `xml:"disabled,attr"`
		Blocked  float64 `xml:"blocked,attr"`
	} `xml:"resources_configured"`
	ClusterOptions struct {
		StonithEnabled   bool   `xml:"stonith-enabled,attr"`
		SymmetricCluster bool   `xml:"symmetric-cluster,attr"`
		MaintenanceMode  bool   `xml:"maintenance-mode,attr"`
		NoQuorumPolicy   string `xml:"no-quorum-policy,attr"`
	} `xml:"cluster_options"`
}

// NodesStruct struct stores the crm_mon XML nodes information
type NodesStruct struct {
	Node []struct {
		Name             string  `xml:"name,attr"`
		ID               string  `xml:"id,attr"`
		Online           bool    `xml:"online,attr"`
		Standby          bool    `xml:"standby,attr"`
		StandbyOnFail    bool    `xml:"standby_onfail,attr"`
		Maintenance      bool    `xml:"maintenance,attr"`
		Pending          bool    `xml:"pending,attr"`
		Unclean          bool    `xml:"unclean,attr"`
		Shutdown         bool    `xml:"shutdown,attr"`
		ExpectedUp       bool    `xml:"expected_up,attr"`
		IsDC             bool    `xml:"is_dc,attr"`
		ResourcesRunning float64 `xml:"resources_running,attr"`
		Type             string  `xml:"type,attr"`
	} `xml:"node"`
}

// ResourcesStruct struct stores the crm_mon XML resources information
type ResourcesStruct struct {
	Resource []ResourceStruct `xml:"resource"`
	Group    []struct {
		ID              string           `xml:"id,attr"`
		NumberResources float64          `xml:"number_resources,attr"`
		Resource        []ResourceStruct `xml:"resource"`
	} `xml:"group"`
	Clone []struct {
		ID             string           `xml:"id,attr"`
		MultiState     bool             `xml:"multi_state,attr"`
		Unique         bool             `xml:"unique,attr"`
		Managed        bool             `xml:"managed,attr"`
		Failed         bool             `xml:"failed,attr"`
		FailureIgnored bool             `xml:"failure_ignored,attr"`
		Resource       []ResourceStruct `xml:"resource"`
	} `xml:"clone"`
}

// NodeAttrStruct struct stores the crm_mon XML node_attributes information
type NodeAttrStruct struct {
	Node []struct {
		Name      string `xml:"name,attr"`
		Attribute []struct {
			Name string `xml:"name,attr"`
			// Value can be everything, not only a number.
			Value string `xml:"value,attr"`
		} `xml:"attribute"`
	} `xml:"node"`
}

// NodeHistoryStruct struct stores the crm_mon XML node_history information
type NodeHistoryStruct struct {
	Node []struct {
		Name            string `xml:"name,attr"`
		ResourceHistory []struct {
			ID                 string `xml:"id,attr"`
			Orphan             string `xml:"orphan,attr"`
			MigrationThreshold int64  `xml:"migration-threshold,attr"`
			OperationHistory   []struct {
				Call                 string `xml:"call,attr"`
				Task                 string `xml:"task,attr"`
				Interval             string `xml:"interval,attr"`
				LastReturnCodeChange string `xml:"last-rc-change,attr"`
				LastRun              string `xml:"last-run,attr"`
				ExecTime             string `xml:"exec-time,attr"`
				QueueTime            string `xml:"queue-time,attr"`
				ReturnCode           int64  `xml:"rc,attr"`
				ReturnCodeText       string `xml:"rc_text,attr"`
			} `xml:"operation_history"`
		} `xml:"resource_history"`
	} `xml:"node"`
}

// FailuresStruct struct stores the crm_mon XML failures information
type FailuresStruct struct {
	Failure []struct {
		OpKey      string `xml:"op_key,attr"`
		Node       string `xml:"node,attr"`
		ExitStatus string `xml:"exitstatus,attr"`
		ExitReason string `xml:"exitreason,attr"`
		ExitCode   string `xml:"exitcode,attr"`
		Call       string `xml:"call,attr"`
		Status     string `xml:"status,attr"`
		Task       string `xml:"task,attr"`
	} `xml:"failure"`
}

// BansStruct struct stores the crm_mon XML bans information
type BansStruct struct {
	Ban []struct {
		ID         string `xml:"id,attr"`
		Resource   string `xml:"resource,attr"`
		Node       string `xml:"node,attr"`
		Weight     string `xml:"weight,attr"`
		MasterOnly string `xml:"master_only,attr"`
	} `xml:"ban"`
}

// ResourceStruct struct stores the crm_mon XML resource information
type ResourceStruct struct {
	ID             string `xml:"id,attr"`
	ResourceAgent  string `xml:"resource_agent,attr"`
	Role           string `xml:"role,attr"`
	TargetRole     string `xml:"target_role,attr"`
	Active         bool   `xml:"active,attr"`
	Orphaned       bool   `xml:"orphaned,attr"`
	Blocked        bool   `xml:"blocked,attr"`
	Managed        bool   `xml:"managed,attr"`
	Failed         bool   `xml:"failed,attr"`
	FailureIgnored bool   `xml:"failure_ignored,attr"`
	Node           []struct {
		Name   string  `xml:"name,attr"`
		ID     float64 `xml:"id,attr"`
		Cached string  `xml:"cached,attr"`
	} `xml:"node"`
}
