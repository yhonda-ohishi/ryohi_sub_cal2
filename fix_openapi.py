import json

# Read the mixed file
with open('docs/openapi.json', 'r', encoding='utf-8') as f:
    data = json.load(f)

# Fix references and structure
if 'definitions' in data:
    if 'components' not in data:
        data['components'] = {}
    data['components']['schemas'] = data['definitions']
    del data['definitions']

# Fix all references
def fix_refs(obj):
    if isinstance(obj, dict):
        for key, value in list(obj.items()):
            if key == '$ref' and isinstance(value, str):
                obj[key] = value.replace('#/definitions/', '#/components/schemas/')
            else:
                fix_refs(value)
    elif isinstance(obj, list):
        for item in obj:
            fix_refs(item)
    return obj

data = fix_refs(data)

# Write back
with open('docs/openapi.json', 'w', encoding='utf-8') as f:
    json.dump(data, f, indent=2, ensure_ascii=False)

print('Fixed OpenAPI 3.0 file')