#!/bin/bash

PYTHON_SCRIPT_PATH="./internal/config/arg_config_generate.py"
MODULE_PATH="./internal/config"

# Call the Python script to generate the argument configuration
python3 $PYTHON_SCRIPT_PATH $MODULE_PATH

# Format the Go code in the specified directory
go fmt $MODULE_PATH
