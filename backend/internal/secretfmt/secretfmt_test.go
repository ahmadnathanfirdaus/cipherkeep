package secretfmt

import "testing"

func TestParseFormat(t *testing.T) {
	cases := map[string]Format{
		"env": FormatEnv, "dotenv": FormatEnv, ".env": FormatEnv,
		"json": FormatJSON, "yaml": FormatYAML, "yml": FormatYAML, "YAML": FormatYAML,
	}
	for in, want := range cases {
		got, err := ParseFormat(in)
		if err != nil || got != want {
			t.Errorf("ParseFormat(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	if _, err := ParseFormat("xml"); err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestParseEnv(t *testing.T) {
	content := "# comment\nexport FOO=bar\nBAZ=\"hello world\"\nQUOTED='single'\n\nEMPTY=\n"
	got, err := Parse(FormatEnv, content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]string{"FOO": "bar", "BAZ": "hello world", "QUOTED": "single", "EMPTY": ""}
	assertMapEqual(t, got, want)
}

func TestParseEnvInvalid(t *testing.T) {
	if _, err := Parse(FormatEnv, "no-equals-sign"); err == nil {
		t.Error("expected error for line without '='")
	}
}

func TestParseJSON(t *testing.T) {
	got, err := Parse(FormatJSON, `{"A":"1","B":"two"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertMapEqual(t, got, map[string]string{"A": "1", "B": "two"})

	// Numbers and booleans are accepted and stringified.
	got2, err := Parse(FormatJSON, `{"PORT": 5432, "TLS": true}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertMapEqual(t, got2, map[string]string{"PORT": "5432", "TLS": "true"})
}

func TestNestedFlatten(t *testing.T) {
	yamlIn := "database:\n  master:\n    host: localhost\n    port: 5432\n  replica:\n    host: localhost\n    port: 5433\n"
	got, err := Parse(FormatYAML, yamlIn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]string{
		"database.master.host":  "localhost",
		"database.master.port":  "5432",
		"database.replica.host": "localhost",
		"database.replica.port": "5433",
	}
	assertMapEqual(t, got, want)

	// JSON nesting flattens the same way.
	jsonGot, err := Parse(FormatJSON, `{"database":{"master":{"host":"localhost","port":5432}}}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertMapEqual(t, jsonGot, map[string]string{
		"database.master.host": "localhost",
		"database.master.port": "5432",
	})
}

func TestNestedRoundTrip(t *testing.T) {
	flat := map[string]string{
		"database.master.host":  "localhost",
		"database.master.port":  "5432",
		"database.replica.host": "10.0.0.1",
		"app.name":              "cipherkeep",
		"flat_key":              "v",
	}
	for _, format := range []Format{FormatJSON, FormatYAML} {
		pairs := make([]Pair, 0, len(flat))
		for k, v := range flat {
			pairs = append(pairs, Pair{Key: k, Value: v})
		}
		encoded, err := Encode(format, pairs)
		if err != nil {
			t.Fatalf("Encode(%s) error: %v", format, err)
		}
		decoded, err := Parse(format, encoded)
		if err != nil {
			t.Fatalf("Parse(%s) error: %v\n%s", format, err, encoded)
		}
		assertMapEqual(t, decoded, flat)
	}
}

func TestNestConflict(t *testing.T) {
	// "database" is both a value and a parent of "database.host".
	_, err := Encode(FormatYAML, []Pair{
		{Key: "database", Value: "x"},
		{Key: "database.host", Value: "y"},
	})
	if err == nil {
		t.Error("expected a key-conflict error")
	}
}

func TestParseYAML(t *testing.T) {
	got, err := Parse(FormatYAML, "A: hello\nPORT: 8080\nENABLED: true\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertMapEqual(t, got, map[string]string{"A": "hello", "PORT": "8080", "ENABLED": "true"})
}

func TestRoundTrip(t *testing.T) {
	original := map[string]string{
		"SIMPLE":  "value",
		"SPACES":  "has spaces",
		"SPECIAL": `quote" and #hash`,
		"NEWLINE": "line1\nline2",
		"EMPTY":   "",
		"URLISH":  "postgres://u:p@h:5432/db?sslmode=disable",
	}
	for _, format := range []Format{FormatEnv, FormatJSON, FormatYAML} {
		pairs := make([]Pair, 0, len(original))
		for k, v := range original {
			pairs = append(pairs, Pair{Key: k, Value: v})
		}
		encoded, err := Encode(format, pairs)
		if err != nil {
			t.Fatalf("Encode(%s) error: %v", format, err)
		}
		decoded, err := Parse(format, encoded)
		if err != nil {
			t.Fatalf("Parse(%s) error: %v\nencoded:\n%s", format, err, encoded)
		}
		assertMapEqual(t, decoded, original)
	}
}

func assertMapEqual(t *testing.T, got, want map[string]string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("map size mismatch: got %d, want %d (%v)", len(got), len(want), got)
	}
	for k, w := range want {
		if g, ok := got[k]; !ok || g != w {
			t.Errorf("key %q: got %q (present=%v), want %q", k, g, ok, w)
		}
	}
}
