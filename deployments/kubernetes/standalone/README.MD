# Kustomize Merak-Compute test overlay
---
This folder contains the base kustomize file and yamls to deploy the following components for integration testing between Merak-Compute and the Merak-Agent.

- Merak Compute
- One vhost pod with the "Standalone" (No Alcor) Merak Agent
- Temporalite (Single pod Temporal)
- A test driver pod to send a gRPC deploy message to Merak-Compute

