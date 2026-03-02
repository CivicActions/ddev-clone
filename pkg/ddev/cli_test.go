package ddev

import (
	"testing"
)

func TestExtractRaw_ValidInfoLine(t *testing.T) {
	input := []byte(`{"level":"info","msg":"describe","raw":{"name":"mysite","status":"running","approot":"/home/user/mysite"}}`)

	raw, err := extractRaw(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `{"name":"mysite","status":"running","approot":"/home/user/mysite"}`
	if string(raw) != expected {
		t.Errorf("expected %s, got %s", expected, string(raw))
	}
}

func TestExtractRaw_MultiLine(t *testing.T) {
	input := []byte(`{"level":"debug","msg":"loading config"}
{"level":"info","msg":"describe","raw":{"name":"testsite","status":"stopped"}}
{"level":"debug","msg":"done"}`)

	raw, err := extractRaw(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `{"name":"testsite","status":"stopped"}`
	if string(raw) != expected {
		t.Errorf("expected %s, got %s", expected, string(raw))
	}
}

func TestExtractRaw_NullRaw(t *testing.T) {
	input := []byte(`{"level":"info","msg":"no data","raw":null}`)

	_, err := extractRaw(input)
	if err == nil {
		t.Fatal("expected error for null raw, got nil")
	}
}

func TestExtractRaw_NoRawField(t *testing.T) {
	input := []byte(`{"level":"info","msg":"just a message"}`)

	_, err := extractRaw(input)
	if err == nil {
		t.Fatal("expected error for missing raw field, got nil")
	}
}

func TestExtractRaw_EmptyInput(t *testing.T) {
	input := []byte(``)

	_, err := extractRaw(input)
	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
}

func TestExtractRaw_NonJSONLines(t *testing.T) {
	input := []byte(`not json at all
{"level":"info","msg":"ok","raw":{"key":"value"}}
more non-json`)

	raw, err := extractRaw(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `{"key":"value"}`
	if string(raw) != expected {
		t.Errorf("expected %s, got %s", expected, string(raw))
	}
}

func TestExtractRaw_ListOutput(t *testing.T) {
	input := []byte(`{"level":"info","msg":"list","raw":[{"name":"site1","status":"running","approot":"/p1"},{"name":"site2","status":"stopped","approot":"/p2"}]}`)

	raw, err := extractRaw(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Raw should be an array
	if raw[0] != '[' {
		t.Errorf("expected array in raw, got: %s", string(raw))
	}
}
