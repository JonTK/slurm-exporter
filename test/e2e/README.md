# End-to-End Tests

## Status: 🚧 Infrastructure Not Yet Implemented

The E2E test infrastructure for this project is currently under development.

## What's Needed

To enable E2E tests, the following components need to be created:

### 1. Test Manifests (`manifests/`)

- **`slurm-simulator.yaml`** - Kubernetes manifest for a mock SLURM cluster
  - Should expose SLURM REST API on port 6820
  - Should simulate basic SLURM functionality (jobs, nodes, partitions)
  - Should have label `app=slurm-simulator` for pod selection

### 2. Helm Chart

The workflow expects a Helm chart at `charts/slurm-exporter/` for deployment.

### 3. E2E Test Suite (`*.go`)

Test files tagged with `// +build e2e` that:
- Deploy the exporter
- Submit test jobs to the SLURM simulator
- Verify metrics are collected and exposed correctly
- Test scraping by Prometheus
- Validate metric cardinality and labels

## Enabling E2E Tests

E2E tests are **disabled by default** and can be enabled in two ways:

### Option 1: PR Label
Add the `e2e` label to a pull request to run E2E tests in CI.

### Option 2: Manual Trigger
Run the "Comprehensive Testing" workflow manually with the `run_e2e` input set to `true`.

## Current Workflow Behavior

Without the E2E infrastructure:
- ✅ E2E tests are **skipped** (not failed)
- ✅ Other tests continue to run normally
- ✅ CI pipeline succeeds

## Development Plan

1. **Create SLURM Simulator**
   - Mock SLURM REST API server
   - Container image for Kubernetes deployment
   - Basic endpoint responses for jobs/nodes/partitions

2. **Create Kubernetes Manifests**
   - Deployment for SLURM simulator
   - Service exposure on port 6820
   - Optional: ConfigMaps for test data

3. **Implement E2E Test Cases**
   - Exporter deployment and health checks
   - Metric collection validation
   - Prometheus integration tests
   - Load and performance tests

4. **CI/CD Integration**
   - Verify Kind cluster setup
   - Validate Helm deployments
   - Automated log collection on failures

## Contributing

If you'd like to help implement E2E tests:

1. Start with the SLURM simulator (simplest approach: HTTP mock server)
2. Create basic Kubernetes manifests
3. Write initial test cases
4. Submit PR with the `e2e` label to validate

For questions, see [CONTRIBUTING.md](../../CONTRIBUTING.md).
