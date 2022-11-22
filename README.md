# Bandwidth Proxy
This is a simple HTTP proxy server that can be easily configured at runtime to throttle traffic at set bitrates.  It presents both a command line interface and a _very simple_ HTML interface.

It was written to help perform manual smoke tests of bitrate switching of a DASH client.

It was written by Cognitive-i Ltd and distributed under 3-clause BSD.  Suggestions and improvements are most welcome.

# Installation

```
go install github.com/cognitive-i/bandwidth-proxy/cmd/bandwidth-proxy@v1

```

# Build and Test
The tool is built with the standard Go manner:

```
go mod download
go build ./cmd/bandwidth-proxy
```

The test cases can be run with either `go test` or [Ginkgo](https://onsi.github.io/ginkgo/#getting-started).  They take about 1.5 minutes to run because they use the wallclock and perform a number of samples to ensure things are stable.

```
# either
ginkgo -r .
# or
go test ./proxy ./ioshaper
```

# Usage
Once the service is running you can connect to its Control Panel over HTTP (defined by `--control-address`):

```
# Start Proxy service.  Proxy on :8080 and Control UI on :8081
bandwidth-proxy --service --proxy-address :8080 --control address :8081

# Set bitrate of a service from cli (this communicates over HTTP to service):
bandwidth-proxy --max-bitrate 8388608
```


