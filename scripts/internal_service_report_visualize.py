import json
import networkx as nx
import matplotlib.pyplot as plt

def load_json(file_path):
    with open(file_path, 'r') as f:
        return json.load(f)

def build_graph(data):
    G = nx.DiGraph()
    
    for edge in data["finalRuntimeGraph"]["edges"]:
        source = edge["source"]["serviceName"]
        target = edge["target"]["serviceName"]
        method = edge["target"]["simpleAPIMethod"]["method"]
        
        G.add_edge(source, target, label=method)
    
    return G

def draw_graph(G, show_label=False):
    plt.figure(figsize=(10, 6))
    pos = nx.spring_layout(G)
    
    nx.draw(G, pos, with_labels=True, node_color='lightblue', edge_color='gray', node_size=3000, font_size=10, font_weight='bold')
    
    if show_label:
        edge_labels = {(u, v): d['label'] for u, v, d in G.edges(data=True)}
        nx.draw_networkx_edge_labels(G, pos, edge_labels=edge_labels, font_size=8, font_color='red')
    
    plt.title("Service Dependency Graph")
    plt.savefig("service_dependency_graph.svg")

def main(file_path, show_label=False):
    data = load_json(file_path)
    G = build_graph(data)
    draw_graph(G, show_label)

if __name__ == "__main__":
    import argparse
    parser = argparse.ArgumentParser(description="Visualize Service Graph")
    parser.add_argument("--file_path", type=str, help="Path to the JSON file", required=True)
    parser.add_argument("--show_label", action='store_true', help="Display edge labels", required=False)
    args = parser.parse_args()
    main(args.file_path, args.show_label)
