"""
This script compares the covered edges between two JSON report files.
Each edge is identified by the combination of source and target service name, endpoint, and method.
The script outputs a JSON file listing:
    - Edges covered (hitCount > 0) in file1 but not in file2
    - Edges covered in file2 but not in file1

Usage:
    python compare_edges.py <file1.json> <file2.json> [output.json]

Arguments:
    file1.json      Path to the first JSON report file
    file2.json      Path to the second JSON report file
    output.json     (Optional) Path to the output JSON file (default: edge_diff.json)

Output:
    A JSON file with two lists: "only_in_file1" and "only_in_file2", each containing edge details.
"""

import json
import sys

def load_edges(filename):
    """
    Load covered edges from a JSON report file.

    Args:
        filename (str): Path to the JSON report file.

    Returns:
        dict: A dictionary mapping edge keys to edge details (source and target).
              The edge key is a tuple:
              (source_service, source_endpoint, source_method, target_service, target_endpoint, target_method)
              Only edges with hitCount > 0 are included.
    """
    with open(filename, 'r', encoding='utf-8') as f:
        data = json.load(f)
    edges = data['finalCallInfoGraph']['edges']
    covered_edges = {}
    for edge in edges:
        if edge.get('hitCount', 0) > 0:
            edge_key = (
                edge['source']['serviceName'],
                edge['source']['simpleAPIMethod']['endpoint'],
                edge['source']['simpleAPIMethod']['method'],
                edge['target']['serviceName'],
                edge['target']['simpleAPIMethod']['endpoint'],
                edge['target']['simpleAPIMethod']['method'],
            )
            covered_edges[edge_key] = {
                "source": edge['source'],
                "target": edge['target']
            }
    return covered_edges

def compare_coverage(file1, file2):
    """
    Compare covered edges between two JSON report files.

    Args:
        file1 (str): Path to the first JSON report file.
        file2 (str): Path to the second JSON report file.

    Returns:
        tuple: Two lists:
            - only_in_1: Edges covered in file1 but not in file2.
            - only_in_2: Edges covered in file2 but not in file1.
            Each edge is represented as a dictionary with "source" and "target".
    """
    covered1 = load_edges(file1)
    covered2 = load_edges(file2)

    only_in_1_keys = set(covered1.keys()) - set(covered2.keys())
    only_in_2_keys = set(covered2.keys()) - set(covered1.keys())

    only_in_1 = [covered1[k] for k in only_in_1_keys]
    only_in_2 = [covered2[k] for k in only_in_2_keys]

    return only_in_1, only_in_2

if __name__ == '__main__':
    if len(sys.argv) < 3:
        print("Usage: python compare_edges.py <file1.json> <file2.json> [output.json]")
        sys.exit(1)
    file1 = sys.argv[1]
    file2 = sys.argv[2]
    output_file = sys.argv[3] if len(sys.argv) > 3 else 'edge_diff.json'

    only_in_1, only_in_2 = compare_coverage(file1, file2)
    result = {
        "only_in_file1": only_in_1,
        "only_in_file2": only_in_2
    }

    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump(result, f, indent=4, ensure_ascii=False)
    print(f"Result written to {output_file}")
