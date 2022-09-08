# Merak Test

Merak Test is the service that will be responsible for automatically running network tests across all VMs in the cluster.

![merak test design diagram](../images/merak_compute_design_diagram.png)

## Services

The following services are provided over gRPC.

- INFO
  - Returns information about the current status of the tests.
- START
  - Starts a new tests.

## Components


### Merak Test Controller

The Merak Test Controller will be responsible for receiving, parsing, and acting on requests sent from the scenario manager. It also be responsible for registering the various
workflows and activities with their corresponding workers.
Based on the requests, it will invoke workers via the temporal client to run the workflows.

### Test Workers

The Test Worker will be responsible for making calls to the Merak Agent to start various network tests.

#### Test Worklfows

The Test workers will be responsible for running the following workflows

- Test Info
- Test Start

## Datamodel

- Test
  - ID
  - Test-Type
  - VMs
    - VM Name
    - VM ID
    - VM Host
    - Results

## Example
```
{
    test:
    {
        "ID": "1",
        "Test-Type": "Ping-All"
        "VMs":
        [
            {
                "Name": "vm1",
                "ID":   "1",
                "Host": "pod1"
                "Results": "Pass"
            },
            {
                "Name": "vm2",
                "ID":   "2",
                "Host": "pod2"
                "Results": "Failed"
            },
        ]
    }
}
```