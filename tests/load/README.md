# Load Tests

The k6 suite exercises the API against an isolated Docker Compose stack. Each
run builds the API image, starts disposable MongoDB and Redis services, waits
for API readiness, runs one scenario, and removes the stack and its data.

Rate limiting is disabled inside this stack so the scenarios measure service
capacity instead of the abuse-control policy. Redis redirect caching remains
enabled.

## Scenarios

| Scenario | Command | Default workload | Primary thresholds |
| --- | --- | --- | --- |
| Probe smoke | `make load-smoke` | 1 VU for 15 seconds | All checks pass, no HTTP failures, p95 below 250 ms |
| Authenticated management | `make load-management` | 2 VUs for 30 seconds | Create, statistics, and delete p95 below 500 ms |
| Redirect throughput | `make load-redirects` | 50 iterations/s for 30 seconds across 20 links | No dropped iterations or HTTP failures, redirect p95 below 200 ms |

The management scenario registers one test user during setup. Every iteration
creates a unique short link, reads its statistics, and deletes it. The redirect
scenario creates a bounded link pool, requests links without following their
external destinations, and deletes the pool during teardown.

## Running Locally

Docker with Docker Compose is the only local prerequisite. Run scenarios from
the repository root:

```bash
make load-smoke
make load-management
make load-redirects
```

The runner preserves the k6 exit status and cleans up after success, threshold
failure, interruption, or startup failure. To remove a stack left by an
external process termination, run:

```bash
make load-down
```

The API is exposed on `127.0.0.1:18081` while a scenario is running. Set
`LOAD_API_PORT` to use another host port.

## Configuration

Set scenario variables before the Make command to override the defaults:

| Variable | Default | Purpose |
| --- | --- | --- |
| `SMOKE_VUS` | `1` | Concurrent smoke-test virtual users |
| `SMOKE_DURATION` | `15s` | Smoke-test duration |
| `SMOKE_P95_MS` | `250` | Probe p95 latency limit in milliseconds |
| `MANAGEMENT_VUS` | `2` | Concurrent management virtual users |
| `MANAGEMENT_DURATION` | `30s` | Management-test duration |
| `MANAGEMENT_P95_MS` | `500` | Per-operation p95 latency limit in milliseconds |
| `REDIRECT_LINKS` | `20` | Seeded links shared by redirect iterations |
| `REDIRECT_RATE` | `50` | Redirect iterations started per second |
| `REDIRECT_DURATION` | `30s` | Redirect-test duration |
| `REDIRECT_PREALLOCATED_VUS` | `10` | VUs reserved before redirect traffic starts |
| `REDIRECT_MAX_VUS` | `50` | Maximum VUs available to sustain the arrival rate |
| `REDIRECT_P95_MS` | `200` | Redirect p95 latency limit in milliseconds |
| `LOAD_API_PORT` | `18081` | API port exposed to the host during a run |

Example:

```bash
REDIRECT_RATE=200 \
REDIRECT_DURATION=2m \
REDIRECT_PREALLOCATED_VUS=25 \
REDIRECT_MAX_VUS=100 \
make load-redirects
```

These defaults are regression thresholds for the local container stack, not a
claim of production capacity. Run sustained tests on representative dedicated
infrastructure before using results for sizing or service-level objectives.

## CI

Relevant branch pushes and pull requests run the probe smoke scenario. The
GitHub Actions `Load Tests` workflow can also be started manually with any of
the three scenarios. Manual management and redirect runs use the same defaults
and cleanup path as local execution.
