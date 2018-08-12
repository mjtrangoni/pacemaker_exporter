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
	testCrmStatusOk     = "fixtures/crm_status.xml"
	testCrmStatusFailed = "fixtures/crm_status_failed.xml"
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
