PortMonitor
=========================

This tool checks IPs for open ports. If ports are open a message is sent to Slack / Mattermost.

Usage
-------------------------
It is possible to configure the port range or port list over parameters or with a properties file.

    usage: ./portMonitor <command> [<args>] 
    This are the optional commands: 
       params      Configuration over params
       properties  Configuration for properties file

This are the configuration parameters for the `params` command:

    usage: ./portMonitor params :
       -end string
            End Port
       -list string
            Port List
       -range string
            Port Range
       -start string
            Start Port
       -webhook string
            Webhook Url for Message
    
Example (ports from 80 to 1020 will be checked):
    
    ./portMonitor params --start=80 --end=1020 --webhook=https://mattermost.test.de/hooks/fsadfdsfdsfdsfdsf

This are the configuration parameters for the `properties` command:
    
    Usage of ./portMonitor properties :
       -file string
            Properties File (Required)
       -end string
        	Property End Port
       -list string
        	Property Port List
       -range string
        	Property Range Port
       -start string
        	Property Start Port
       -webhook string
        	Webhook Url for Message    

Example (ports from 80 to 1020 will be checked):
    
    ./portMonitor properties --file=testprops.properties --start=test.startproperty --end=test.endproperty --webhook=https://mattermost.test.de/hooks/fsadfdsfdsfdsfdsf

 - testprops.properties
    
        test.startproperty = 80
        test.endproperty = 1020
        
If a port still open a message is sent to the webhook. This can be used for test preconditions of a test environment.


License
------------
Copyright 2014-2019 Matthias Raab.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
