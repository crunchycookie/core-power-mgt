# core-power-mgt

![Go Build](https://github.com/crunchycookie/core-power-mgt/actions/workflows/go.yml/badge.svg)

A CPU-core power management microservice.

This service wraps the [Intel Power Optimization Library](https://github.com/intel/power-optimization-library) to 
manage per-core power features of intel processors. However, APIs are designed to be generic enough, such that future
support for different vendors would be seamless for the end-user. Current APIs supports,

- Grouping cores into high-performance and dynamic pools
- Changing status of dynamic pool to idle (minimum power consumption) to high performance, and vice-versa.

###### Limitations

Current implementation expects below requirements.

- A Linux-based operating system
- Intel processor with core idle-states and dynamic frequency scaling support.
- Linux idle driver must be `intel_idle`

...and supports followings.
- Creates two core groups: Stable and Dynamic.
- Supports assigning power profiles for each group: core idle state + clock frequency.
- Upon termination (`^C`), safely handovers power management back to the operating system.

###### Project goals

Evolve towards supporting fine-grained core power management through APIs; core-grouping, setting power profiles, per-
core power management, power monitoring, etc.

### Build & run

Service can be started by building the binary for the target env. and running it.

Execute below. Make sure to replace placeholders values for the target environment.

`GOOS=<target-operating-system> GOARCH=<target-cpu-architecture> go build -gcflags="-N -l" -o gc-controller`

For example, below will build the service for linux env. with `amd64` cpu.

`GOOS=linux GOARCH=amd64 go build -gcflags="-N -l" -o gc-controller`

Run the service with `./gc-controller <config-file>` command (tested on linux - might change for other envs).

An example `<config-file>` is `conf.yaml`.
```yaml
host:
  name: localhost
  port: 3000
topology:
  stable-core-count: 3
  dynamic-core-count: 1
power-profile:
  sleep-idle-state: C3_ACPI
  sleep-frq: 400
  perf-idle-state: POLL
  perf-frq: 2600
```
Note: Total core count must exceed stable and dynamic core sum. Available total cores can be obtained via `lscpu` in 
linux to check `Core(s) per socket` attribute. Available idle states can be obtained via `cpupower idle-info` command 
and observing attribute `Available idle states:`. Frequency (`frq`) values can be set by reading cpu spec sheet. Notice 
per-core max frequency might be lower than cpu max frequency. Overcommitment values will be capped at upper and lower bounds.


### Supported APIs

- `/gc-controller/sleep`
    - Set dynamic cores to sleep mode.
    - ```
      curl --location --request PUT 'http://<host.ip>:<host.port>/gc-controller/sleep'
      ``` 
- `/gc-controller/wake`
    - Set dynamic cores to perf mode.
    - ```
      curl --location --request PUT 'http://<host.ip>:<host.port>/gc-controller/wake'
      ```
- `/gc-controller/dev/perf`
    - Change clock frequency of dynamic cores.
    - ```
      curl --location --request PUT 'http://<host.ip>:<host.port>/gc-controller/dev/perf' \
      --header 'Content-Type: application/json' \
      --data '{
      "f-mhz": 2600
      }'
      ```

### Tested on
- Development was done in MacOS, and tested on Lenovo ThinkPad X1 Carbon X1 Gen 9 with Intel Core i7-1165G7
  (4 cores - hyper-threading disabled).
- Tested env was Ubuntu 23.04 with Linux kernel 6.2.0-39-generic

### Verification

The following sections explain the verification steps carried out.

#### Deployment

Execute [build-debug.sh](../build-debug.sh) script. It builds the service for linux and amd64 cpu architecture, tests,
and deploy the service binary to a remote dev
environment via SSH. The authentication details to perform this is injected via inline environment variables. An
example execution looks like below.

- Template:`REMOTE_IP=<remote-ip> REMOTE_USER=<remote-user> sh build-debug.sh`
- Say remote ip is 10.13.13.5, and user is ubuntu,
    - `REMOTE_IP=10.13.13.5 REMOTE_USER=ubuntu sh build-debug.sh`

ps: For any other environment, the script needs to be modified (ex: for linux with arm64, the
env vars of app building command needs to be changed appropriately)

#### Remote debug

The same script also deploys the script: [run-at-remote-for-debug.sh](../run-at-remote-for-debug.sh), which
uses delve to support remote debugging. Remote debugging is tested for Goland IDE by following the steps from
https://www.jetbrains.com/help/go/attach-to-running-go-processes-with-debugger.html. Might want to create an SSH tunnel
to avoid configuring the network for ip discovery (`ssh -L 3000:localhost:3000 <remote-ip>`).

Since the service binary involves os userspace modifications, it needs sudo privileges. Therefore, if the debugging
script is run with sudo, then PATH env. var of sudo session must point to the delve binary. One way to achieve this is
install golang and then install delve via go tool and then modify PATH var in sudo session to point the delve binary
in `/home/<user>/go/bin`.

#### Execution

Intel power library is the core driver of this service. It requires sudo privileges to modify `sysfs`, thus gc-controller
binary needs to run with sudo.

`sudo ./gc-controller ./sample-conf.yaml`

#### Verification

1. **Configure remote env for monitoring:** We use multiple different tools because i7z reads directly from CPU, thus ACPI related c-state
   information are not visible.
    1. **Frequency scaling:**
        1. Install i7z via `sudo apt-get install i7z`
        2. Start monitoring `sudo i7z`
    2. **Core sleep:**
        1. Install turbostat via `sudo apt-get install turbostat`
        2. Start monitoring `turbostat --show Core,POLL%,C1ACPI%,C2ACPI%,C3ACPI%,CPU%c1,CPU%c6,CPU%c7,PkgWatt,CorWatt,Busy%`
    3. **CPU power:**
        1. Install powerstat via `sudo apt-get install powerstat`
        2. We will monitor the effect later.
2. Perform test:
    1. Start the service `sudo ./gc-controller`.
        1. **What is does:** pool first three cores (`0,1,2`) as fully awake (`POLL` state) stables and set its frequency to `2.8Ghz` -> Higher
           performance pool of cores supporting most workloads including latency critical tasks. Last core is set as a
           dynamic Core and initialized to the deepest possible sleep state (`C3_ACPI`) and its performance is degraded to
           a low value (`core frequency is less than 500 Mhz`).
        2. **Verification:** i7z shows actual frequency values as expected, and turbostat verifies sleep states.
           ![perf-states-verification.png](docs/perf-states-verification.png)
           ![c-states-verification.png](docs/c-states-verification.png)
    2. Run powerstat tool to collect CPU power through RAPL `sudo powerstat -R`
    3. Wake up the dynamic pool via `curl --location --request PUT 'http://localhost:3000/gc-controller/wake' \
       --header 'Content-Type: application/json' \
       --data '{
       "count": 2
       }'` and make it high perf via `curl --location --request PUT 'http://localhost:3000/gc-controller/perf' \
       --header 'Content-Type: application/json' \
       --data '{
       "f-mhz": 2600
       }'`
    4. Re-run powerstat tool to collect CPU power through RAPL `sudo powerstat -R`
    5. Compare two power results. With the core turned off, it should save about ~1 watt of power.
       ![power-verification-pre.png](docs/power-verification-pre.png)
       ![power-verification-post.png](docs/power-verification-post.png)
#### Post-cleanup

Upon successful startup, terminating service via `^c` (cntrl + c or cmd + c in mac) will safely close the program.

For any other case, restart the system. This will reset any changes done.
