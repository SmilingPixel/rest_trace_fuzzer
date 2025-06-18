"""
This script analyzes test scenario execution logs, extracting relevant data and visualizing it with graphs.

Features:
- Reads logs from a specified file.
- Extracts 'Finish execute current test scenario' logs.
- Parses edge covered count, edge coverage, and status code count.
- Generates three SVG graphs:
  1. Edge Covered Count vs. Test Process
  2. Edge Coverage vs. Test Process
  3. Covered Status Code Count vs. Test Process

Usage:
- Modify `log_file` to point to your actual log file.
- Run the script, and it will save the graphs as SVG files.

Requirements:
- Python 3
- `matplotlib` for plotting

"""

import re
import json
import matplotlib.pyplot as plt

# Read log file
def read_logs(file_path):
    with open(file_path, "r", encoding="utf-8") as file:
        return file.readlines()

# Extract test scenario logs
def extract_test_scenarios(logs):
    pattern = re.compile(
        r'Finish execute current test scenario .*?UUID: ([a-f0-9\-]+).*?Edge covered count: (\d+).*?Edge coverage: ([0-9\.]+).*?covered status code count: (\d+)',
        re.IGNORECASE
    )
    test_data = []
    
    for log in logs:
        match = pattern.search(log)
        if match:
            uuid = match.group(1)
            edge_covered_count = int(match.group(2))
            edge_coverage = float(match.group(3))
            status_code_count = int(match.group(4))
            test_data.append((uuid, edge_covered_count, edge_coverage, status_code_count))
    
    return test_data

# Plot graphs
def plot_graphs(test_data, output_prefix="test_analysis"):
    x = list(range(1, len(test_data) + 1))  # Treat test scenario as a unit
    edge_covered_counts = [data[1] for data in test_data]
    edge_coverages = [data[2] for data in test_data]
    status_code_counts = [data[3] for data in test_data]
    
    plt.figure(figsize=(8, 4))
    plt.plot(x, edge_covered_counts, marker='^', linestyle='-', color='g', label='Edge Covered Count')
    plt.xlabel("Test Process")
    plt.ylabel("Edge Covered Count")
    plt.title("Edge Covered Count Across Test Scenarios")
    plt.grid(True, linestyle='--', alpha=0.7)
    plt.legend()
    plt.savefig(f"{output_prefix}_edge_covered_count.svg", format="svg")
    plt.savefig(f"{output_prefix}_edge_covered_count.png", format="png")
    plt.close()
    
    plt.figure(figsize=(8, 4))
    plt.plot(x, edge_coverages, marker='o', linestyle='-', color='b', label='Edge Coverage')
    plt.xlabel("Test Process")
    plt.ylabel("Edge Coverage")
    plt.title("Edge Coverage Across Test Scenarios")
    plt.grid(True, linestyle='--', alpha=0.7)
    plt.legend()
    plt.savefig(f"{output_prefix}_edge_coverage.svg", format="svg")
    plt.savefig(f"{output_prefix}_edge_coverage.png", format="png")
    plt.close()
    
    plt.figure(figsize=(8, 4))
    plt.plot(x, status_code_counts, marker='s', linestyle='-', color='r', label='Status Code Count')
    plt.xlabel("Test Process")
    plt.ylabel("Covered Status Code Count")
    plt.title("Status Code Coverage Across Test Scenarios")
    plt.grid(True, linestyle='--', alpha=0.7)
    plt.legend()
    plt.savefig(f"{output_prefix}_status_codes.svg", format="svg")
    plt.savefig(f"{output_prefix}_status_codes.png", format="png")
    plt.close()

# Main function
def main(log_file):
    logs = read_logs(log_file)
    test_data = extract_test_scenarios(logs)
    if test_data:
        plot_graphs(test_data)
        print("Graphs saved successfully.")
    else:
        print("No relevant logs found.")

# Example usage
if __name__ == "__main__":
    import argparse
    parser = argparse.ArgumentParser(description="Visualize Test Process")
    parser.add_argument("--file_path", type=str, help="Path to the log file", required=True)
    args = parser.parse_args()
    main(args.file_path)
