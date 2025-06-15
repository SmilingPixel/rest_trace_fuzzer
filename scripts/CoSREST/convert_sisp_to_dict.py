import json

"""
Script to extract valid samples from a structured API parameter validation schema.

Input JSON schema (each item in the top-level list):
{
  "op_name": "string",              # API path including parameter(s)
  "type": "string",                 # Data type (e.g., "string")
  "param_name": "string",           # Name of the path parameter
  "format": ["regex", ...],         # Allowed formats as regex strings
  "valid": [                        # Valid sample definitions
    {
      "param_name": "string",
      "category": "string",         # Category of the valid sample
      "description": "string",      # Description of the category
      "samples": ["string", ...]    # Sample valid values (take one per category)
    },
    ...
  ],
  "invalid": [                      # (Ignored here) Invalid sample definitions
    ...
  ]
}

Output JSON format:
[
  {
    "name": "param_name",           # e.g., "orderId"
    "value": "valid_sample"         # e.g., "abc-123-def"
  },
  ...
]
"""

def extract_valid_samples(input_file: str, output_file: str):
    with open(input_file, 'r') as f:
        data = json.load(f)

    output = []

    for entry in data:
        param_name = entry.get('param_name')
        valid_entries = entry.get('valid', [])
        seen_categories = set()

        for valid in valid_entries:
            category = valid.get('category')
            samples = valid.get('samples', [])

            if category not in seen_categories and samples:
                seen_categories.add(category)
                output.append({
                    "name": param_name,
                    "value": samples[0]
                })

    # Sort output by 'name'
    output_sorted = sorted(output, key=lambda x: x['name'])

    with open(output_file, 'w') as f:
        json.dump(output_sorted, f, indent=2)


if __name__ == '__main__':
    extract_valid_samples('sisp.json', 'tt_fuzz_dict.json')
