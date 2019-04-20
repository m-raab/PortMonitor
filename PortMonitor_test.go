/*
 * Copyright (c) 2019.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestReadPropertiesFile(t *testing.T) {
	props, err := ReadPropertiesFile("testprops.properties")

	if err != nil {
		t.Errorf("File is not readable.")
	}

	tables := []struct {
		key   string
		value string
	}{
		{"test.startproperty", "80"},
		{"test.endproperty", "1020"},
		{"portrange.test", "90-1030"},
		{"portlist.test", "81,91,1040"},
	}

	for _, table := range tables {
		rv, ok := props[table.key]
		if !ok {
			t.Errorf("Properties are not correct parsed.")
		}
		if table.value != rv {
			t.Errorf("Values is not correct (key: %s, value: %s, value was: %s", table.key, table.value, rv)
		}
	}
}

func TestReadPropertiesRange(t *testing.T) {
	props, err := ReadPropertiesFile("testprops.properties")

	if err != nil {
		t.Errorf("File is not readable.")
	}

	m := &PortMonitor{}

	kRange := "portrange.test"
	kpRange := &kRange
	kList := ""
	kpList := &kList
	kSPort := ""
	kpSPort := &kSPort
	kEPort := ""
	kESPort := &kEPort

	m.ReadProperties(props, kpRange, kpList, kpSPort, kESPort)

	if m.start != 90 {
		t.Errorf("Port is not correct. It is %d and should be %d", m.start, 90)
	}
	if m.end != 1030 {
		t.Errorf("Port is not correct. It is %d and should be %d", m.end, 1030)
	}
}

func TestReadPropertiesStartEnd(t *testing.T) {
	props, err := ReadPropertiesFile("testprops.properties")

	if err != nil {
		t.Errorf("File is not readable.")
	}

	m := &PortMonitor{}

	kRange := ""
	kpRange := &kRange
	kList := ""
	kpList := &kList
	kSPort := "test.startproperty"
	kpSPort := &kSPort
	kEPort := "test.endproperty"
	kESPort := &kEPort

	m.ReadProperties(props, kpRange, kpList, kpSPort, kESPort)

	if m.start != 80 {
		t.Errorf("Port is not correct. It is %d and should be %d", m.start, 80)
	}
	if m.end != 1020 {
		t.Errorf("Port is not correct. It is %d and should be %d", m.end, 1020)
	}
}

func TestReadPropertiesList(t *testing.T) {
	props, err := ReadPropertiesFile("testprops.properties")

	if err != nil {
		t.Errorf("File is not readable.")
	}

	m := &PortMonitor{}

	kRange := ""
	kpRange := &kRange
	kList := "portlist.test"
	kpList := &kList
	kSPort := ""
	kpSPort := &kSPort
	kEPort := ""
	kESPort := &kEPort

	m.ReadProperties(props, kpRange, kpList, kpSPort, kESPort)

	if len(m.list) != 3 {
		t.Errorf("Port list is not correct. It is %d elements should have %d", len(m.list), 3)
	} else {
		if p := m.list[0]; p != 81 {
			t.Errorf("Port is not correct. It is %d and should be %d", m.list[0], 81)
		}
		if p := m.list[1]; p != 91 {
			t.Errorf("Port is not correct. It is %d and should be %d", m.list[0], 91)
		}
		if p := m.list[2]; p != 1040 {
			t.Errorf("Port is not correct. It is %d and should be %d", m.list[0], 1040)
		}
	}
}

func TestParseCommandLineStartEnd(t *testing.T) {
	os.Args = []string{"command", "params", "--start=82", "--end=1022"}
	m := &PortMonitor{}
	m.ParseCommandLine()

	if m.start != 82 {
		t.Errorf("Port is not correct. It is %d and should be %d", m.start, 82)
	}
	if m.end != 1022 {
		t.Errorf("Port is not correct. It is %d and should be %d", m.end, 1022)
	}
}

func TestParseCommandLineRange(t *testing.T) {
	os.Args = []string{"command", "params", "--range=83-1023"}
	m := &PortMonitor{}
	m.ParseCommandLine()

	if m.start != 83 {
		t.Errorf("Port is not correct. It is %d and should be %d", m.start, 83)
	}
	if m.end != 1023 {
		t.Errorf("Port is not correct. It is %d and should be %d", m.end, 1023)
	}
}

func TestParseCommandLineList(t *testing.T) {
	os.Args = []string{"command", "params", "--list=83,94,122"}
	m := &PortMonitor{}
	m.ParseCommandLine()

	if len(m.list) != 3 {
		t.Errorf("Port list is not correct. It is %d elements should have %d", len(m.list), 3)
	} else {
		if p := m.list[0]; p != 83 {
			t.Errorf("Port is not correct. It is %d and should be %d", m.list[0], 83)
		}
		if p := m.list[1]; p != 94 {
			t.Errorf("Port is not correct. It is %d and should be %d", m.list[0], 94)
		}
		if p := m.list[2]; p != 122 {
			t.Errorf("Port is not correct. It is %d and should be %d", m.list[0], 122)
		}
	}
}

func TestCalculateIPs(t *testing.T) {
	m := &PortMonitor{}
	m.CalculateIPConfig()

	if len(m.Ips) < 1 {
		t.Errorf("IP is not calculated")
	}
}

func TestPortOpen(t *testing.T) {
	m := &PortMonitor{}
	m.CalculateIPConfig()

	if len(m.Ips) < 1 {
		t.Errorf("IP is not calculated")
	}

	handler := func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, "Hello")
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	addrString := server.Listener.Addr().String()
	addrStringParts := strings.Split(addrString, ":")

	port, _ := strconv.ParseInt(addrStringParts[1], 10, 0)
	rv := PortOpen(addrStringParts[0], port)

	if !rv {
		t.Errorf("Port check does not work")
	}
}
