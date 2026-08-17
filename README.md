# Kubernetes System Labs

Hands-on failure labs for learning how Kubernetes and cloud infrastructure behave when things go wrong.

The goal is not to collect Kubernetes examples or copy troubleshooting commands. Each lab should create a realistic failure, expose enough evidence to investigate it, and require diagnosis, mitigation, verification, and prevention.

## Learning model

Each scenario follows the production-engineering loop:

```text
symptom
  ↓
form hypotheses
  ↓
inspect signals
  ↓
narrow the failure domain
  ↓
identify root cause
  ↓
mitigate
  ↓
verify recovery
  ↓
explain the failure
  ↓
design prevention
```

Labs should reward understanding of system behavior rather than memorizing commands.

## Scope

The repository will grow in stages.

### 1. Local Kubernetes failure labs

Start with small, reproducible scenarios on a local Kubernetes cluster. Planned areas include:

- scheduling and resource constraints
- CPU and memory pressure
- OOM kills and evictions
- readiness, liveness, and startup probe failures
- broken deployments and rollout behavior
- service discovery and DNS
- networking and NetworkPolicy
- persistent storage
- stateful workloads and database availability
- observability and incident investigation

### 2. Terraform as the infrastructure layer

Terraform will be used to provision infrastructure for later cloud-based labs. This provides a practical setting for learning Infrastructure as Code rather than treating Terraform as an isolated syntax exercise.

Topics will include:

- providers, resources, data sources, variables, and outputs
- dependency graphs and resource lifecycle
- modules and reusable infrastructure components
- `for_each` and resource identity
- state and remote state
- drift detection and reconciliation
- imports and bringing existing resources under management
- plan review and replacement risk
- partial apply failures and recovery
- CI/CD workflows and controlled applies

Terraform itself may also become part of a failure scenario—for example drift, unsafe replacement, excessive blast radius, or state inconsistencies.

### 3. AWS and EKS labs

Later labs will move selected scenarios into AWS and introduce failures that cross the Kubernetes/cloud boundary.

Potential infrastructure includes:

- VPCs, subnets, routing, and security groups
- EKS clusters and node groups
- IAM integration
- RDS PostgreSQL
- S3
- cloud and Kubernetes observability

This enables scenarios where the same symptom can originate from different layers—for example Kubernetes scheduling, AWS networking, IAM, storage, or database configuration.

## Example scenario

A database workload remains `Pending` in EKS.

The root cause might involve:

- node taints and tolerations
- affinity rules
- resource requests
- node-group capacity
- Terraform configuration
- AWS infrastructure constraints

The objective is not just to make the Pod run. A completed investigation should explain:

1. What failed?
2. Why did it fail?
3. Which evidence proved the diagnosis?
4. What was the safest mitigation?
5. How was recovery verified?
6. What change would prevent recurrence?

## AI guardrails

AI assistance should act as an incident facilitator, not a solution generator.

During an active lab, an assistant should avoid revealing the root cause or prescribing the next debugging command unless explicitly requested by the lab. It may provide requested evidence, clarify concepts, challenge hypotheses, and evaluate the final diagnosis.

The intent is to preserve the troubleshooting work for the learner.

## Principles

- Prefer small, reproducible failures over large demo environments.
- Make symptoms observable before making solutions obvious.
- Treat partial failure and recovery as first-class behavior.
- Keep infrastructure explicit until repeated patterns justify abstraction.
- Prefer realistic operational trade-offs over toy puzzles.
- A successful fix is not complete until recovery is verified and prevention is considered.

## Status

Early-stage project. The initial focus is local Kubernetes failure labs; Terraform and AWS/EKS are planned extensions as the lab collection grows.
