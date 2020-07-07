package main

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func renderValue(value []byte) string {
	var buf bytes.Buffer
	err := json.Indent(&buf, value, "", "  ")
	if err != nil {
		return fmt.Sprintf("\n```\n%s\n```", string(value))
	}

	return fmt.Sprintf("\n```json\n%s\n```", string(buf.Bytes()))
}
