# CCCLEO Task Orchestrator

**Lightweight Container-Based Task Orchestration for CCCLEO**

A lightweight, Go-based container orchestration system designed for heterogeneous edge devices and unreliable network environments.

**CSC 579 — Overlay and Peer-to-Peer Networking**  
University of Victoria | Fall 2026  
**Author:** Yi-Ze Chen  
**Project Type:** Individual Research Project

## Overview

Traditional container orchestration platforms such as Kubernetes and K3s provide powerful workload management capabilities, but their resource requirements and operational complexity may be unnecessary for small-scale research testbeds.

The **Coast-to-Coast-to-Coast Low-Earth-Orbit Satellite Network (CCCLEO)** testbed includes heterogeneous computing environments, ranging from virtual machines and mini PCs to resource-constrained ARM devices.

Satellite network conditions introduce additional challenges, including high latency, limited bandwidth, and intermittent connectivity.

This project investigates whether a lightweight, task-oriented orchestration system can provide reliable container execution with lower resource overhead under these conditions.

The system will be implemented in **Go**, using a controller-worker architecture and Docker-based task execution.

## Research Objectives

The project focuses on four research questions:

1. **Push vs. Pull Scheduling:** How do different scheduling approaches affect task completion, scheduling latency, and recovery after network disconnections?
2. **Resource Efficiency:** How much CPU and memory does a lightweight Go-based orchestrator consume compared with K3s?
3. **Configuration Simplicity:** Can a minimal YAML/JSON job specification replace much of the configuration complexity required by Kubernetes-style workloads?
4. **Container Image Caching:** How significantly does image caching improve task execution over bandwidth-constrained networks?

## Proposed Architecture

The system consists of two primary components.

### Controller

The controller manages task submission, scheduling, worker registration, and job state.

Planned responsibilities include:

- Exposing REST APIs for task submission and monitoring
- Maintaining a persistent job store
- Tracking worker availability
- Supporting push-based and pull-based scheduling
- Recording execution metrics
- Handling task retries and recovery

### Worker Agent

Each worker runs a lightweight Go agent responsible for:

- Registering with the controller
- Retrieving or receiving assigned tasks
- Executing tasks using Docker containers
- Reporting task status and results
- Recovering from temporary network disconnections
- Preventing duplicate execution through unique job identifiers and local state

### Scheduling Strategies

| Strategy | Description |
|---|---|
| Push | Controller initiates task assignment to workers |
| Pull | Workers request available tasks from the controller |
| Hybrid | Combines both approaches (optional extension) |

The primary research comparison will focus on **push versus pull scheduling**.

## Technology Stack

| Component | Technology |
|---|---|
| Programming Language | Go |
| Container Runtime | Docker |
| Communication | REST API / HTTP |
| Job Configuration | YAML / JSON |
| Network Emulation | Linux `tc/netem` |
| Lightweight Orchestration Baseline | K3s |
| Device Management | Azure IoT Hub (planned integration) |
| Task Tracking | Jira |
| Data Visualization | Python / Power BI (optional) |
| CI/CD | GitHub Actions |
| Documentation | GitHub Pages |

## Experimental Methodology

The project will use Docker-enabled virtual machines and Linux network emulation to evaluate orchestration performance under controlled conditions.

### Network Conditions

Experiments will vary:

- Network latency
- Packet loss
- Available bandwidth
- Temporary and extended disconnections
- Container image cache availability

### Evaluation Metrics

| Metric | Description |
|---|---|
| Task Completion Rate | Percentage of submitted tasks successfully completed |
| Scheduling Latency | Time between task submission and assignment |
| Recovery Time | Time required to resume operation after reconnection |
| CPU Usage | Controller and worker CPU consumption |
| Memory Usage | Controller and worker memory consumption |
| Duplicate Executions | Number of unintended repeated task executions |
| Image Transfer Time | Time required to retrieve container images |

Experiments will use repeatable workloads to compare push and pull scheduling under equivalent network conditions.

K3s comparisons will be performed only on compatible hardware.

## Project Roadmap

| Period | Milestone |
|---|---|
| Oct 13–19 | Finalize job schema, REST API, metrics, website, and test VMs |
| Oct 20–26 | Implement Go controller, job store, and worker registration |
| Oct 27–Nov 9 | Implement worker agent and push/pull scheduling; prepare midterm demonstration |
| Nov 10–16 | Integrate Azure IoT Hub and evaluate reconnection recovery |
| Nov 17–23 | Run network emulation and container image caching experiments |
| Nov 24–End of Term | Analyze results, finalize presentation, report, and project website |

## Project Structure

The planned repository organization is:

```text
cccleo-task-orchestrator/
├── cmd/
│   ├── controller/
│   └── worker/
├── internal/
│   ├── api/
│   ├── scheduler/
│   ├── store/
│   ├── executor/
│   └── metrics/
├── configs/
├── experiments/
│   ├── netem/
│   ├── workloads/
│   └── results/
├── docs/
├── .github/
│   └── workflows/
├── go.mod
└── README.md
```

This structure represents the planned implementation and may evolve during development.

## Expected Deliverables

- Working Go-based controller and worker prototype
- YAML/JSON task specification
- Push and pull scheduling implementations
- Persistent job tracking and reconnection recovery
- Reproducible network emulation experiments
- Quantitative performance comparisons
- Research report and final presentation
- Public project documentation

## Scope and Limitations

This project does not aim to replace Kubernetes or provide a production-ready orchestration platform.

Instead, it evaluates the trade-offs between scheduling reliability, resource efficiency, and implementation simplicity in constrained network environments.

The primary priority is to produce a working prototype and reproducible push/pull scheduling experiments.

Hybrid scheduling, a web frontend, and Power BI dashboards are considered optional extensions.

If Azure IoT Hub is unavailable, the system will continue operating through its local REST-based control path.

## References

- [CCCLEO Research Testbed](https://onlineacademiccommunity.uvic.ca/cccleo/)
- [Kubernetes Documentation](https://kubernetes.io/docs/concepts/overview/)
- [K3s Documentation](https://docs.k3s.io/)
- [K3s Resource Profiling](https://docs.k3s.io/reference/resource-profiling)
- [Linux tc-netem Manual](https://man7.org/linux/man-pages/man8/tc-netem.8.html)
- [Azure IoT Hub Device Twins](https://learn.microsoft.com/en-us/azure/iot-hub/iot-hub-devguide-device-twins)

---

**University of Victoria — CSC 579, Fall 2026**

*Research project by Yi-Ze Chen.*
