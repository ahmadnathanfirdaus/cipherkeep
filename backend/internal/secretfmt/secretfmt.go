// Package secretfmt parses and encodes secret key/value sets in the supported
// interchange formats: dotenv (.env), JSON, and YAML.
//
// Secrets are stored flat as a string->string map. JSON and YAML are hierarchical:
// on encode, dotted keys (e.g. "database.master.host") are nested into a tree; on
// decode, nested structures are flattened back to dotted keys. Numeric path segments
// are treated as ordinary string keys (objects-only; no array reconstruction). The
// dotenv format is always flat — dotted keys are written literally.
package secretfmt

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Delimiter separates path segments in a flat (dotted) key.
const Delimiter = "."

// Format identifies a supported import/export format.
type Format string

const (
	FormatEnv  Format = "env"
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
)

// formatInfo is the metadata that defines a file type. Adding a new format is a
// matter of registering one entry here plus its parse/encode functions.
type formatInfo struct {
	Extension    string
	ContentType  string
	Hierarchical bool
	aliases      []string
}

var formats = map[Format]formatInfo{
	FormatEnv:  {Extension: "env", ContentType: "text/plain", Hierarchical: false, aliases: []string{"env", "dotenv", ".env"}},
	FormatJSON: {Extension: "json", ContentType: "application/json", Hierarchical: true, aliases: []string{"json"}},
	FormatYAML: {Extension: "yaml", ContentType: "application/yaml", Hierarchical: true, aliases: []string{"yaml", "yml"}},
}

// AllFormats returns the supported formats in a stable order.
func AllFormats() []Format { return []Format{FormatEnv, FormatJSON, FormatYAML} }

// Pair is a single ordered key/value entry, used when encoding.
type Pair struct {
	Key   string
	Value string
}

// ParseFormat normalizes a user-supplied format string.
func ParseFormat(s string) (Format, error) {
	want := strings.ToLower(strings.TrimSpace(s))
	for f, info := range formats {
		for _, alias := range info.aliases {
			if want == alias {
				return f, nil
			}
		}
	}
	return "", fmt.Errorf("unsupported format %q (expected env, json, or yaml)", s)
}

// Extension returns the conventional file extension for a format.
func (f Format) Extension() string {
	if info, ok := formats[f]; ok {
		return info.Extension
	}
	return "env"
}

// ContentType returns the MIME type for a format.
func (f Format) ContentType() string {
	if info, ok := formats[f]; ok {
		return info.ContentType
	}
	return "text/plain"
}

// Hierarchical reports whether the format nests dotted keys into a tree.
func (f Format) Hierarchical() bool {
	return formats[f].Hierarchical
}

// Parse decodes content in the given format into a key/value map.
func Parse(format Format, content string) (map[string]string, error) {
	switch format {
	case FormatEnv:
		return parseEnv(content)
	case FormatJSON:
		return parseJSON(content)
	case FormatYAML:
		return parseYAML(content)
	default:
		return nil, fmt.Errorf("unsupported format %q", format)
	}
}

// Encode serializes ordered pairs into the given format.
func Encode(format Format, pairs []Pair) (string, error) {
	switch format {
	case FormatEnv:
		return encodeEnv(pairs), nil
	case FormatJSON:
		return encodeJSON(pairs)
	case FormatYAML:
		return encodeYAML(pairs)
	default:
		return "", fmt.Errorf("unsupported format %q", format)
	}
}

// --- dotenv ---

func parseEnv(content string) (map[string]string, error) {
	out := make(map[string]string)
	for i, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(strings.TrimSuffix(raw, "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			return nil, fmt.Errorf("invalid line %d: expected KEY=value", i+1)
		}
		key := strings.TrimSpace(line[:eq])
		if key == "" {
			return nil, fmt.Errorf("invalid line %d: empty key", i+1)
		}
		out[key] = unquoteEnvValue(strings.TrimSpace(line[eq+1:]))
	}
	return out, nil
}

func unquoteEnvValue(v string) string {
	if len(v) >= 2 {
		if v[0] == '"' && v[len(v)-1] == '"' {
			inner := v[1 : len(v)-1]
			replacer := strings.NewReplacer(`\n`, "\n", `\r`, "\r", `\t`, "\t", `\"`, `"`, `\\`, `\`)
			return replacer.Replace(inner)
		}
		if v[0] == '\'' && v[len(v)-1] == '\'' {
			return v[1 : len(v)-1]
		}
	}
	return v
}

func encodeEnv(pairs []Pair) string {
	var b strings.Builder
	for _, p := range pairs {
		b.WriteString(p.Key)
		b.WriteByte('=')
		b.WriteString(quoteEnvValue(p.Value))
		b.WriteByte('\n')
	}
	return b.String()
}

func quoteEnvValue(v string) string {
	needsQuote := strings.ContainsAny(v, "\n\r\t \"'#") || v != strings.TrimSpace(v)
	if !needsQuote {
		return v
	}
	escaped := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`, "\t", `\t`).Replace(v)
	return `"` + escaped + `"`
}

// --- JSON ---

func parseJSON(content string) (map[string]string, error) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return map[string]string{}, nil
	}
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON: expected an object: %w", err)
	}
	out := make(map[string]string)
	if err := flatten("", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

func encodeJSON(pairs []Pair) (string, error) {
	tree, err := nest(pairs)
	if err != nil {
		return "", err
	}
	// json.MarshalIndent sorts map keys deterministically.
	b, err := json.MarshalIndent(tree, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b) + "\n", nil
}

// --- YAML ---

func parseYAML(content string) (map[string]string, error) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return map[string]string{}, nil
	}
	var raw map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &raw); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}
	out := make(map[string]string, len(raw))
	if err := flatten("", raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// flatten walks a decoded JSON/YAML value and writes leaf scalars into out using
// dot-joined paths. Maps recurse by key; arrays recurse by numeric index (which
// become ordinary string segments). Non-scalar leaves are rejected by scalarToString.
func flatten(prefix string, v interface{}, out map[string]string) error {
	switch t := v.(type) {
	case map[string]interface{}:
		for k, child := range t {
			if err := flatten(joinKey(prefix, k), child, out); err != nil {
				return err
			}
		}
	case []interface{}:
		for i, child := range t {
			if err := flatten(joinKey(prefix, strconv.Itoa(i)), child, out); err != nil {
				return err
			}
		}
	default:
		if prefix == "" {
			return fmt.Errorf("invalid input: expected an object at the top level")
		}
		s, err := scalarToString(prefix, v)
		if err != nil {
			return err
		}
		out[prefix] = s
	}
	return nil
}

func joinKey(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + Delimiter + key
}

// nest builds a tree of nested maps from flat dotted keys. A path that requires a
// segment to be both a value and a group is a conflict and returns an error.
func nest(pairs []Pair) (map[string]interface{}, error) {
	root := map[string]interface{}{}
	for _, p := range pairs {
		segments := strings.Split(p.Key, Delimiter)
		cursor := root
		for i, seg := range segments {
			if i == len(segments)-1 {
				if _, isMap := cursor[seg].(map[string]interface{}); isMap {
					return nil, fmt.Errorf("key conflict: %q is used as both a value and a group", p.Key)
				}
				cursor[seg] = p.Value
				break
			}
			child, ok := cursor[seg]
			if !ok {
				next := map[string]interface{}{}
				cursor[seg] = next
				cursor = next
				continue
			}
			next, isMap := child.(map[string]interface{})
			if !isMap {
				return nil, fmt.Errorf("key conflict: %q clashes with a value at %q", p.Key, seg)
			}
			cursor = next
		}
	}
	return root, nil
}

func scalarToString(key string, v interface{}) (string, error) {
	switch t := v.(type) {
	case nil:
		return "", nil
	case string:
		return t, nil
	case bool:
		return strconv.FormatBool(t), nil
	case int:
		return strconv.Itoa(t), nil
	case int64:
		return strconv.FormatInt(t, 10), nil
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64), nil
	default:
		return "", fmt.Errorf("key %q must be a scalar value, got %T", key, v)
	}
}

func encodeYAML(pairs []Pair) (string, error) {
	tree, err := nest(pairs)
	if err != nil {
		return "", err
	}
	doc := mapToYAMLNode(tree)
	b, err := yaml.Marshal(doc)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// mapToYAMLNode builds a sorted YAML mapping node from a nested string map. Leaf
// values are tagged !!str so values like "5432" or "true" round-trip as strings.
func mapToYAMLNode(m map[string]interface{}) *yaml.Node {
	node := &yaml.Node{Kind: yaml.MappingNode}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: k}
		var valNode *yaml.Node
		switch child := m[k].(type) {
		case map[string]interface{}:
			valNode = mapToYAMLNode(child)
		case string:
			valNode = &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: child}
		default:
			valNode = &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: fmt.Sprintf("%v", child)}
		}
		node.Content = append(node.Content, keyNode, valNode)
	}
	return node
}
