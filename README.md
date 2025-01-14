# Microservice System Testing Tool

This project is a tool designed to test a microservice system. The most novel part of this tool is its ability to utilize traces collected from the system to guide the testing process.

## Table of Contents

- [Introduction](#introduction)
- [Features](#features)
- [Architecture](#architecture)
- [Installation](#installation)
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

## Usage

To use the tool, follow these steps:

1. Build the project:
    ```sh
    bash build.sh
    ```

2. Run the tool:
    ```sh
    bash run.sh
    ```

## Configuration

The tool can be configured using command-line arguments and environment variables. The following options are available:

- `--openapi-spec`: Path to the OpenAPI specification file.
- `--fuzzer-type`: Type of fuzzer to use (e.g., Basic).
- `--fuzzer-budget`: Time budget for fuzzing (e.g., 30s).
- `--server-base-url`: Base URL of the server to test.
- `--internal-service-openapi-spec-map`: Path to the internal service OpenAPI specification map file.
- `--output-dir`: Directory to save the output reports.

## License

This project is licensed under the GNU License. See the LICENSE file for more details.

