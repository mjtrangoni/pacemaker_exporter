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
	"io/ioutil"
	"testing"
)

const (
	testCrmStatusOk       = "fixtures/crm_status.xml"
	testCrmStatusDockerOk = "fixtures/crm_status_docker.xml"
	testCrmStatusFailed   = "fixtures/crm_status_failed.xml"
)

func TestParseCrmMonXML(t *testing.T) {
	dataByte, err := ioutil.ReadFile(testCrmStatusOk)
	if err != nil {
		t.Fatal(err)
	}

	dataStr, err := parseCrmMonXML(dataByte)
	if err != nil {
		t.Fatal(err)
	}

	if dataStr.Summary.Stack.Type != "corosync" {
		t.Fatalf("summary stack type : %v!=corosync",
			dataStr.Summary.Stack.Type)
	}

	// Check some node attributes values.
	for _, node := range dataStr.NodeAttributes.Node {
		if node.Name == "lustre-mds1" {
			for _, attr := range node.Attribute {
				if attr.Name == "ping-lnet" {
					if attr.Value != "3360" {
						t.Fatalf("node-attribute '%s' value: %v!=3360",
							attr.Name, attr.Value)
					}
				} else if attr.Name == "hana_prd_roles" {
					if attr.Value != "4:P:master1:master:worker:master" {
						t.Fatalf("node-attribute '%s' value: %v!=4:P:master1:master:worker:master",
							attr.Name, attr.Value)
					}
				}
			}
		}
	}
}

func TestParseCrmMonXMLDocker(t *testing.T) {
	dataByte, err := ioutil.ReadFile(testCrmStatusDockerOk)
	if err != nil {
		t.Fatal(err)
	}

	dataStr, err := parseCrmMonXML(dataByte)
	if err != nil {
		t.Fatal(err)
	}

	if dataStr.Bans.Ban[0].Node != " host01" {
		t.Fatalf("Failure name  : %v!=host01",
			dataStr.Bans.Ban[0].Node)
	}
}

func TestParseCrmMonXMLFailed(t *testing.T) {
	dataByte, err := ioutil.ReadFile(testCrmStatusFailed)
	if err != nil {
		t.Fatal(err)
	}

	dataStr, err := parseCrmMonXML(dataByte)
	if err != nil {
		t.Fatal(err)
	}

	if dataStr.Failures.Failure[0].Node != "lustre-oss1" {
		t.Fatalf("Failure name  : %v!=lustre-oss1",
			dataStr.Failures.Failure[0].Node)
	}
}
