package ddev

import "encoding/json"

// DdevJSONOutput wraps the DDEV -j output JSON envelope.
// DDEV CLI output with -j flag produces lines like:
// {"level":"info","msg":"...","raw":{...},"time":"..."}
type DdevJSONOutput struct {
	Level string          `json:"level"`
	Msg   string          `json:"msg"`
	Raw   json.RawMessage `json:"raw"`
}

// DescribeResult holds parsed output from "ddev describe <name> -j".
type DescribeResult struct {
	Name            string `json:"name"`
	Status          string `json:"status"`
	AppRoot         string `json:"approot"`
	DatabaseType    string `json:"dbinfo_database_type"`
	DatabaseVersion string `json:"dbinfo_database_version"`
	MutagenEnabled  bool   `json:"mutagen_enabled"`
}

// ProjectInfo holds parsed output from "ddev list -j" (one entry).
type ProjectInfo struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	AppRoot string `json:"approot"`
}
