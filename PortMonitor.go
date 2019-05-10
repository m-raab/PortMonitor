// +build linux darwin
// +build amd64

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
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/int128/slack"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type PortMonitor struct {
	Ips      []string
	hostname string
	url      string

	start int64
	end   int64

	rstart int64
	rend   int64

	list []int64

	debug bool
}

type ConfigProperties map[string]string

func ReadPropertiesFile(filename string) (ConfigProperties, error) {

	config := ConfigProperties{}

	if len(filename) == 0 {
		return config, nil
	}

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// read line
		line := strings.TrimSpace(scanner.Text())
		// is no comment
		if !strings.HasPrefix(line, "#") {
			// is a key value pair
			if equal := strings.Index(line, "="); equal >= 0 {
				// a key is available
				if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
					value := ""
					if len(line) > equal {
						value = strings.TrimSpace(line[equal+1:])
					}
					config[key] = value
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
		return nil, err
	}

	return config, nil
}

func PrintUsage() {
	fmt.Printf("usage: %s <command> [<args>] \n", os.Args[0])
	fmt.Println("This are the optional commands: ")
	fmt.Println("   params      Configuration over params")
	fmt.Println("   properties  Configuration for properties file")
}

func PortOpen(ip string, port int64) bool {
	portStr := strconv.FormatInt(port, 10)
	if conn, err := net.Dial("tcp", ip+":"+portStr); err == nil {
		conn.Close()
		return true
	} else {
		return false
	}
}

func (pm *PortMonitor) ParseCommandLine() {
	paramSet := flag.NewFlagSet("params", flag.ExitOnError)
	propertiesSet := flag.NewFlagSet("properties", flag.ExitOnError)

	paramStartPtr := paramSet.String("start", "", "Start Port")
	paramEndPtr := paramSet.String("end", "", "End Port")
	paramRangePtr := paramSet.String("range", "", "Port Range")
	paramListPtr := paramSet.String("list", "", "Port List")
	paramWebhookUrlPtr := paramSet.String("webhook", "", "Webhook Url for Message")
	paramDebugPtr := paramSet.Bool("debug", false, "Activates Debug Output")

	propsFilePtr := propertiesSet.String("file", "", "Properties File (Required)")
	propsStartPtr := propertiesSet.String("start", "", "Property Start Port")
	propsEndPtr := propertiesSet.String("end", "", "Property End Port")
	propsRangePtr := propertiesSet.String("range", "", "Property Range Port")
	propsListPtr := propertiesSet.String("list", "", "Property Port List")
	propsWebhookUrlPtr := propertiesSet.String("webhook", "", "Webhook Url for Message")
	propsDebugPtr := propertiesSet.Bool("debug", false, "Activates Debug Output")

	if len(os.Args) > 1 {
		var err error = nil
		if len(os.Args) == 2 && strings.Contains(os.Args[1], "help") {
			PrintUsage()
		} else {
			switch os.Args[1] {
			case paramSet.Name():
				err = paramSet.Parse(os.Args[2:])
				if err == nil {
					if *paramWebhookUrlPtr != "" {
						pm.url = *paramWebhookUrlPtr
					}

					err = pm.ReadParameters(paramRangePtr, paramListPtr, paramStartPtr, paramEndPtr)
				}
				if err != nil {
					log.Println(err)
					fmt.Fprintf(os.Stderr, "Usage of %s params :\n", os.Args[0])
					paramSet.PrintDefaults()
					os.Exit(1)
				}
				pm.debug = *paramDebugPtr
			case propertiesSet.Name():
				err = propertiesSet.Parse(os.Args[2:])
				if err == nil {
					if *propsFilePtr == "" {
						if *propsWebhookUrlPtr != "" {
							pm.url = *propsWebhookUrlPtr
						}

						err = errors.New("The properties file must be specified for properties configuration.")
					}
					if err == nil {
						properties, err := ReadPropertiesFile(*propsFilePtr)
						if err != nil {
							err = errors.New("The properties file is not readable.")
						} else {
							err = pm.ReadProperties(properties, propsRangePtr, propsListPtr, propsStartPtr, propsEndPtr)
						}
					}

					pm.debug = *propsDebugPtr
				}
				if err != nil {
					log.Println(err)
					fmt.Fprintf(os.Stderr, "Usage of %s properties :\n", os.Args[0])
					propertiesSet.PrintDefaults()
					os.Exit(1)
				}
			default:
				fmt.Fprintf(os.Stdout, "unknown parameters: %s \n", os.Args[1:])
				PrintUsage()
				os.Exit(2)
			}
		}
	} else {
		PrintUsage()
		os.Exit(2)
	}
}

func (pm *PortMonitor) ReadParameters(portRange *string, portList *string, startPort *string, endPort *string) error {
	switch {
	case *portRange != "":
		if strings.Contains(*portRange, "-") {
			ps := strings.Split(*portRange, "-")
			if len(ps) < 2 {
				return errors.New(fmt.Sprintf("Configured port range '%s' is not range!", *portRange))
			} else {
				if p, err := strconv.ParseInt(ps[0], 10, 0); err == nil {
					pm.rstart = p
				} else {
					return errors.New(fmt.Sprintf("Start port '%s' of '%s' is not an integer (%s).\n", ps[0], *portRange, err))
				}

				if p, err := strconv.ParseInt(ps[1], 10, 0); err == nil {
					pm.rend = p
				} else {
					return errors.New(fmt.Sprintf("End port '%s' of '%s' is not an integer (%s).\n", ps[1], *portRange, err))
				}
			}
		} else {
			return errors.New(fmt.Sprintf("There is no port range configured %s.", *portRange))
		}
	case *portList != "":
		pl := strings.Split(*portList, ",")
		for _, ps := range pl {
			if pi, err := strconv.ParseInt(ps, 10, 0); err == nil {
				pm.list = append(pm.list, pi)
			} else {
				log.Println(fmt.Sprintf("The start port '%s' of '%s' is not an integer.", ps, *portList))
			}
		}
		if len(pl) < 1 {
			return errors.New(fmt.Sprintf("There is no port list configured for '%s'.", *portList))
		}
	case *startPort != "" && *endPort != "":
		if pi, err := strconv.ParseInt(*startPort, 10, 0); err == nil {
			pm.start = pi
		} else {
			return errors.New(fmt.Sprintf("The start port '%s' is not an integer.", *startPort))
		}

		if pi, err := strconv.ParseInt(*endPort, 10, 0); err == nil {
			pm.end = pi
		} else {
			return errors.New(fmt.Sprintf("The end port '%s' is not an integer.", *endPort))
		}
	default:
		return errors.New("It is necessary to specify a port range, a port list or a start and an end port.")
	}
	return nil
}

func (pm *PortMonitor) ReadProperties(props map[string]string, portRange *string, portList *string, startPort *string, endPort *string) error {
	if *portRange != "" {
		prValue, ok := props[*portRange]
		if ok {
			if strings.Contains(prValue, "-") {
				ps := strings.Split(prValue, "-")
				if len(ps) < 2 {
					return errors.New(fmt.Sprintf("Configured port range '%s' of '%s' is not range!", prValue, *portRange))
				} else {
					if p, err := strconv.ParseInt(ps[0], 10, 0); err == nil {
						pm.start = p
					} else {
						return errors.New(fmt.Sprintf("Start port '%s' of '%s' is not an integer (%s).\n", ps[0], *portRange, err))
					}

					if p, err := strconv.ParseInt(ps[1], 10, 0); err == nil {
						pm.end = p
					} else {
						return errors.New(fmt.Sprintf("End port '%s' of '%s' is not an integer (%s).\n", ps[1], *portRange, err))
					}
				}
			} else {
				return errors.New(fmt.Sprintf("There is no port range configured for '%s' in (%s).", *portRange, prValue))
			}
		} else {
			return errors.New(fmt.Sprintf("There is no port range configured for '%s' in properties file.", *portRange))
		}
	}
	if *portList != "" {
		if strings.Contains(*portList, ",") {
			plp := strings.Split(*portList, ",")
			for _, p := range plp {
				sp, ok := props[p]
				if ok {
					if spi, err := strconv.ParseInt(sp, 10, 0); err == nil {
						pm.list = append(pm.list, spi)
					} else {
						log.Fatal(fmt.Sprintf("Port list member element '%s' is not a string in '%s' of '%s'!", sp, plp, *portList))
					}
				} else {
					return errors.New(fmt.Sprintf("There is no port configured for '%s' in properties file.", p))
				}
			}
		} else {
			prList, ok := props[*portList]
			if ok {
				pl := strings.Split(prList, ",")
				for _, ps := range pl {
					if pi, err := strconv.ParseInt(ps, 10, 0); err == nil {
						pm.list = append(pm.list, pi)
					} else {
						log.Fatal(fmt.Sprintf("Port list element '%s' is not a string in '%s' of '%s'!", ps, pl, *portList))
					}
				}
				if len(pl) < 1 {
					return errors.New(fmt.Sprintf("There is no port list configured for '%s' of '%s'.", *portList, prList))
				}
			} else {
				return errors.New(fmt.Sprintf("There is no port list configured for '%s' in properties file.", *portList))
			}
		}
	}
	if *startPort != "" && *endPort != "" {
		p, ok := props[*startPort]
		if ok {
			if pi, err := strconv.ParseInt(p, 10, 0); err == nil {
				pm.start = pi
			} else {
				return errors.New(fmt.Sprintf("The start port '%s' of '%s' is not an integer.", p, *startPort))
			}
		} else {
			return errors.New(fmt.Sprintf("There is no start port configured for '%s' in properties file.", *startPort))
		}
		p, ok = props[*endPort]
		if ok {
			if pi, err := strconv.ParseInt(p, 10, 0); err == nil {
				pm.end = pi
			} else {
				return errors.New(fmt.Sprintf("The end port '%s' of '%s' is not an integer.", p, *endPort))
			}
		} else {
			return errors.New(fmt.Sprintf("There is no start port configured for '%s' in properties file.", *endPort))
		}
	}
	return nil
}

func (pm *PortMonitor) CalculateIPConfig() {

	var err error
	pm.hostname, err = os.Hostname()

	if err != nil {
		log.Fatalf("It was not possible to calculate the hostname. (%s)", err)
	}

	if addrs, err := net.InterfaceAddrs(); err == nil {
		for _, address := range addrs {
			// check the address type and if it is not a loopback the display it
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					pm.Ips = append(pm.Ips, ipnet.IP.String())
				}
			}
		}
	} else {
		log.Printf("It was not possible to identify all interfaces. (%s)", err)
	}
}

func (pm *PortMonitor) sendMessage(monitormsg string) {
	if pm.url == "" {
		log.Fatalf("Run with parameter URL for webhook configuration.")
	}

	message := slack.Message{
		Username:  "portmonitor",
		IconEmoji: ":star:",
		Attachments: []slack.Attachment{
			{
				Title:      fmt.Sprintf("Ports is still open on %s", pm.hostname),
				Text:       monitormsg,
				AuthorName: "@portminitor",
				Footer:     "Port Monitor Message",
				Color:      "danger",
				Timestamp:  time.Now().Unix(),
			},
		},
	}
	err := slack.Send(pm.url, &message)
	if err != nil {
		log.Fatalf("Could not send the message to Slack: %s", err)
	}
	log.Printf("Sent the message %+v", message)
}

func main() {
	m := &PortMonitor{}
	m.ParseCommandLine()
	m.CalculateIPConfig()

	message := "Port Monitor \n"
	portIsOpen := false

	for _, ip := range m.Ips {
		if m.start > 0 && m.end > 0 {
			for p := m.start; p <= m.end; p++ {
				if PortOpen(ip, p) {
					log.Println(fmt.Sprintf("Port %d for %s is open.", p, ip))
					message += fmt.Sprintf("Port %d for %s is open. \n", p, ip)
					portIsOpen = true
				} else {
					if m.debug == true {
						log.Println(fmt.Sprintf("Port %d for %s is not open.", p, ip))
					}
				}
			}
		}

		if m.rstart > 0 && m.rend > 0 {
			for p := m.rstart; p <= m.rend; p++ {
				if PortOpen(ip, p) {
					log.Println(fmt.Sprintf("Port %d for %s is open.", p, ip))
					message += fmt.Sprintf("Port %d for %s is open. \n", p, ip)
					portIsOpen = true
				} else {
					if m.debug == true {
						log.Println(fmt.Sprintf("Port %d for %s is not open.", p, ip))
					}
				}
			}
		}

		if len(m.list) > 0 {
			for _, port := range m.list {
				if PortOpen(ip, port) {
					log.Println(fmt.Sprintf("Port %d for %s is open.", port, ip))
					message += fmt.Sprintf("Port %d for %s is open. \n", port, ip)
					portIsOpen = true
				} else {
					if m.debug == true {
						log.Println(fmt.Sprintf("Port %d for %s is not open.", port, ip))
					}
				}
			}
		}
	}

	if portIsOpen && m.url != "" {
		log.Println("Send message to :", m.url)
		m.sendMessage(message)
	}

	if portIsOpen {
		os.Exit(10)
	} else {
		os.Exit(0)
	}
}
