#!/bin/bash

# Mock NCloud CLI for testing
# This script simulates NCloud CLI responses for testing purposes

set -e

# Mock responses based on command
case "$1" in
    "vserver")
        case "$2" in
            "createServerInstances")
                # Mock successful server creation
                echo '{
                    "createServerInstancesResponse": {
                        "requestId": "test-request-id",
                        "returnCode": "0",
                        "returnMessage": "success",
                        "serverInstanceList": [{
                            "serverInstanceNo": "12345",
                            "serverName": "test-server",
                            "serverInstanceStatusName": "init",
                            "serverInstanceOperation": "NULL",
                            "serverInstanceStatus": "INIT",
                            "platformType": "LNX64",
                            "loginKeyName": "test-key",
                            "isFeeChargingMonitoring": false,
                            "publicIp": "",
                            "privateIp": "10.0.0.1",
                            "serverImageName": "CentOS 7.3",
                            "serverInstanceType": "SVR.VSVR.HICPU.C002.M004.NET.SSD.B050.G002",
                            "regionCode": "KR",
                            "zoneCode": "KR-2",
                            "vpcNo": "vpc-12345",
                            "subnetNo": "subnet-12345",
                            "networkInterfaceNoList": ["eni-12345"],
                            "placementGroupName": "",
                            "fabricClusterPoolNo": "",
                            "isProtectServerTermination": false
                        }]
                    }
                }'
                ;;
            "getServerInstanceList")
                # Mock server status check
                echo '{
                    "getServerInstanceListResponse": {
                        "requestId": "test-request-id",
                        "returnCode": "0",
                        "returnMessage": "success",
                        "serverInstanceList": [{
                            "serverInstanceNo": "12345",
                            "serverName": "test-server",
                            "serverInstanceStatusName": "running",
                            "serverInstanceOperation": "NULL",
                            "serverInstanceStatus": "RUN",
                            "platformType": "LNX64",
                            "loginKeyName": "test-key",
                            "isFeeChargingMonitoring": false,
                            "publicIp": "1.2.3.4",
                            "privateIp": "10.0.0.1",
                            "serverImageName": "CentOS 7.3",
                            "serverInstanceType": "SVR.VSVR.HICPU.C002.M004.NET.SSD.B050.G002",
                            "regionCode": "KR",
                            "zoneCode": "KR-2",
                            "vpcNo": "vpc-12345",
                            "subnetNo": "subnet-12345",
                            "networkInterfaceNoList": ["eni-12345"],
                            "placementGroupName": "",
                            "fabricClusterPoolNo": "",
                            "isProtectServerTermination": false
                        }]
                    }
                }'
                ;;
            "terminateServerInstances")
                # Mock server termination
                echo '{
                    "terminateServerInstancesResponse": {
                        "requestId": "test-request-id",
                        "returnCode": "0",
                        "returnMessage": "success",
                        "serverInstanceList": [{
                            "serverInstanceNo": "12345",
                            "serverName": "test-server",
                            "serverInstanceStatusName": "terminating",
                            "serverInstanceOperation": "NULL",
                            "serverInstanceStatus": "TERMT",
                            "platformType": "LNX64",
                            "loginKeyName": "test-key",
                            "isFeeChargingMonitoring": false,
                            "publicIp": "1.2.3.4",
                            "privateIp": "10.0.0.1",
                            "serverImageName": "CentOS 7.3",
                            "serverInstanceType": "SVR.VSVR.HICPU.C002.M004.NET.SSD.B050.G002",
                            "regionCode": "KR",
                            "zoneCode": "KR-2",
                            "vpcNo": "vpc-12345",
                            "subnetNo": "subnet-12345",
                            "networkInterfaceNoList": ["eni-12345"],
                            "placementGroupName": "",
                            "fabricClusterPoolNo": "",
                            "isProtectServerTermination": false
                        }]
                    }
                }'
                ;;
            *)
                echo "Unknown vserver command: $2"
                exit 1
                ;;
        esac
        ;;
    *)
        echo "Unknown command: $1"
        exit 1
        ;;
esac