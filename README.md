# Microservice System Testing Tool

This project is a tool designed to test a microservice system. The most novel part of this tool is its ability to utilize traces collected from the system to guide the testing process.

## Table of Contents

- [Introduction](#introduction)
- [Features](#features)
- [Architecture](#architecture)
- [Installation](#installation)
- [Prepration](#prepration)
- [Usage](#usage)
- [Configuration](#configuration)
- [License](#license)

## Introduction

Microservice systems are complex and require thorough testing to ensure reliability and performance. This tool leverages system traces to guide the testing process, providing a more efficient and effective way to identify issues and ensure system robustness.

## Features

- **Trace-Guided Testing**: Utilizes system traces to guide the testing process.
- **Comprehensive Coverage**: Ensures thorough testing of all microservice endpoints.
- **Automated Test Case Generation**: Automatically generates test cases based on system traces.
- **Detailed Reporting**: Provides detailed reports on test coverage and results.

## Architecture

You can get details of architecture design [here](docs/design.md)

## Installation

To install the tool, follow these steps:

1. Clone the repository:

2. Install dependencies:
    ```sh
    go mod tidy
    ```

## Prepration

1. Prepare the OpenAPI specification file for the system under test. The tool supports OpenAPI 3 by default.
2. Prepare a protobuf file for all RPC which internal services use.
  - We use [protoc-gen-openapi](https://github.com/google/gnostic/tree/main/cmd/protoc-gen-openapi) to convert protobuf to openapi.
  - You should annotate the proto file, and you can refer to this [issue](https://github.com/google/gnostic/issues/412).

## Usage

To use the tool, follow these steps:

1. Build the project:
    ```sh
    make build
    ```

2. Run the tool:
    ```sh
    make run
    ```

3. Clean the project:
    ```sh
    make clean
    ```

In addition, we provide vscode tasks to use the tool. You can build, run and clean the project by selecting the task `Run` in vscode. See the [.vscode/tasks.json](.vscode/tasks.json) file for more details.

## Configuration

The tool can be configured using command-line arguments and environment variables. The following options are available:

- `--openapi-spec`: Path to the OpenAPI specification file.
- `--server-base-url`: Base URL of the server to test.
- `--internal-service-openapi-spec`: Path to the internal service OpenAPI specification file.
- `--trace-backend-url`: URL of the trace backend.
- `--trace-backend-type`: Type of the trace backend. Currently only supports 'Jaeger'.
- `--fuzzer-type`: Type of fuzzer to use (e.g., Basic).
- `--fuzzer-budget`: Time budget for fuzzing (e.g., 30s).
- `--log-level`: Log level: debug, info, warn, error, fatal, panic.
- `--output-dir`: Directory to save the output reports.

## License

This project is licensed under the GPL-3.0 License - see the [LICENSE](LICENSE) file for details.
