<h1 align="center">benchttp/cli</h1>

<p align="center">
  <a href="https://github.com/benchttp/cli/actions/workflows/ci.yml?query=branch%3Amain">
  <img alt="Github Worklow Status" src="https://img.shields.io/github/workflow/status/benchttp/cli/Lint%20&%20Test%20&%20Build"></a>
  <a href="https://codecov.io/gh/benchttp/cli">
  <img alt="Code coverage" src="https://img.shields.io/codecov/c/gh/benchttp/cli?label=coverage"></a>
  <a href="https://goreportcard.com/report/github.com/benchttp/cli">
  <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/benchttp/cli" /></a>
  <br />
  <a href="https://pkg.go.dev/github.com/benchttp/cli#section-documentation">
    <img alt="Go package Reference" src="https://img.shields.io/badge/pkg-reference-informational?logo=go" /></a>
  <a href="https://github.com/benchttp/cli/releases">
    <img alt="Latest version" src="https://img.shields.io/github/v/tag/benchttp/cli?label=release"></a>
</p>

## About

`benchttp/cli` is a command-line interface that runs benchmarks on HTTP endpoints.
Highly configurable, it can be used as a CI step as well as
a testing tool at development time.

## Installation

1. Visit https://github.com/benchttp/cli/releases and download the asset
   `benchttp_<os>_<architecture>` matching your OS and CPU architecture.
1. Rename the downloaded asset to `benchttp`, add it to your `PATH`,
   and refresh your terminal if necessary
1. Run `benchttp version` to check it works properly.

## Usage

### Run a benchmark

```sh
benchttp run [options]
```

## Configuration

In this section we dive into the many configuration options provided by the runner.

First, a good way to get started with is via our [configuration generator](https://www.benchttp.app/setup).

By default, the runner uses a default configuration that is valid for use without further tuning, except for `url` that must always be set.

### Configuration Flow

To determine the final configuration of a benchmark, the runner follows that flow:

1. It starts with a [default configuration](./examples/config/default.yml)
1. Then it tries to find a config file and overrides the defaults with the values set in it

   - If flag `-configFile` is set, it resolves its value as a path
   - Else, it tries to find a config file in the working directory, by priority order:
     `.benchttp.yml` > `.benchttp.yaml` > `.benchttp.json`

   The config file is _optional_: if none is found, this step is ignored.
   If a config file has an option `extends`, it resolves config file recursively until the root is reached and overrides the values from parent to child.

1. Then it overrides the current config values with any value set via the CLI
1. Finally, it performs a validation on the resulting config (not before!).
   This allows composed configurations for better granularity.

### Specifications

With rare exceptions, any option can be set either via CLI flags or config file,
and option names always match.

A full config file example is available [here](./examples/config/full.yml).

#### HTTP request options

| CLI flag  | File option           | Description               | Usage example                             |
| --------- | --------------------- | ------------------------- | ----------------------------------------- |
| `-url`    | `request.url`         | Target URL (**Required**) | `-url http://localhost:8080/users?page=3` |
| `-method` | `request.method`      | HTTP Method               | `-method POST`                            |
| -         | `request.queryParams` | Added query params to URL | -                                         |
| `-header` | `request.header`      | Request headers           | `-header 'key0:val0' -header 'key1:val1'` |
| `-body`   | `request.body`        | Raw request body          | `-body 'raw:{"id":"abc"}'`                |

#### Benchmark runner options

| CLI flag          | File option             | Description                                                          | Usage example        |
| ----------------- | ----------------------- | -------------------------------------------------------------------- | -------------------- |
| `-requests`       | `runner.requests`       | Number of requests to run (-1 means infinite, stop on globalTimeout) | `-requests 100`      |
| `-concurrency`    | `runner.concurrency`    | Maximum concurrent requests                                          | `-concurrency 10`    |
| `-interval`       | `runner.interval`       | Minimum duration between two non-concurrent requests                 | `-interval 200ms`    |
| `-requestTimeout` | `runner.requestTimeout` | Timeout for every single request                                     | `-requestTimeout 5s` |
| `-globalTimeout`  | `runner.globalTimeout`  | Timeout for the whole benchmark                                      | `-globalTimeout 30s` |

Note: the expected format for durations is `<int><unit>`, with `unit` being any of `ns`, `µs`, `ms`, `s`, `m`, `h`.

#### Output options

| CLI flag  | File option     | Description               | Usage example               |
| --------- | --------------- | ------------------------- | --------------------------- |
| `-silent` | `output.silent` | Remove convenience prints | `-silent` / `-silent=false` |
