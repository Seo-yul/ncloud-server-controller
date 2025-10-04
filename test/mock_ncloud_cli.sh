#!/bin/bash

# Mock NCloud CLI for testing purposes
# This script simulates the behavior of the actual NCloud CLI

# Check if the command is for server creation
if [[ "$1" == "vserver" && "$2" == "createServerInstances" ]]; then
    # Return a mock JSON response for server creation
    echo '{"serverInstanceList":[{"serverInstanceNo":"test-instance-123","serverName":"test-server","serverInstanceStatus":{"code":"INIT","codeName":"Initializing"}}]}'
    exit 0
fi

# Check if the command is for server detail
if [[ "$1" == "vserver" && "$2" == "getServerInstanceDetail" ]]; then
    # Return a mock JSON response for server detail
    echo '{"serverInstanceNo":"test-instance-123","serverName":"test-server","serverInstanceStatus":{"code":"RUN","codeName":"Running"},"publicIp":"1.2.3.4","privateIp":"10.0.0.1","regionCode":"KR","zoneCode":"KR-1"}'
    exit 0
fi

# Check if the command is for server termination
if [[ "$1" == "vserver" && "$2" == "terminateServerInstances" ]]; then
    # Return a mock JSON response for server termination
    echo '{"returnCode":"0","returnMessage":"Success"}'
    exit 0
fi

# For any other commands, return an error
echo "Unknown command: $*"
exit 1
