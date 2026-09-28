package main

import (
	"bytes"
	"strings"
	"testing"
)

const solvedState = `{
  "U": ["W","W","W","W","W","W","W","W","W"],
  "R": ["G","G","G","G","G","G","G","G","G"],
  "F": ["O","O","O","O","O","O","O","O","O"],
  "D": ["Y","Y","Y","Y","Y","Y","Y","Y","Y"],
  "L": ["B","B","B","B","B","B","B","B","B"],
  "B": ["X","X","X","X","X","X","X","X","X"]
}`

// edgeSwapCase exchanges the UR and UF edges while keeping both unflipped,
// producing a pure parity violation with all color counts intact.
func edgeSwapCase() string {
	s := solvedState
	// UR and UF share the same U-face color, so exchanging just the side
	// stickers R:1 and F:1 transposes the two pieces without flipping them.
	s = strings.Replace(s, `"R": ["G","G","G","G","G","G","G","G","G"]`,
		`"R": ["G","O","G","G","G","G","G","G","G"]`, 1)
	s = strings.Replace(s, `"F": ["O","O","O","O","O","O","O","O","O"]`,
		`"F": ["O","G","O","O","O","O","O","O","O"]`, 1)
	return s
}

func TestRunExitCodes(t *testing.T) {
	tests := []struct {
		name string
		json string
		code int
	}{
		{"legal", solvedState, 0},
		{"illegal edge transposition parity", edgeSwapCase(), 1},
		{
			"illegal color count imbalance",
			strings.Replace(solvedState,
				`["W","W","W","W","W","W","W","W","W"]`,
				`["W","W","W","W","W","G","W","W","W"]`, 1),
			1,
		},
		{"missing face", `{"U":["W","W","W","W","W","W","W","W","W"]}`, 2},
		{"short face", func() string {
			return strings.Replace(solvedState,
				`"R": ["G","G","G","G","G","G","G","G","G"]`,
				`"R": ["G","G","G","G","G","G","G","G"]`, 1)
		}(), 2},
		{"unknown face key", strings.Replace(solvedState, `"U":`, `"Q":`, 1), 2},
		{"malformed json", `{not json`, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(strings.NewReader(tt.json), &stdout, &stderr)
			if code != tt.code {
				t.Fatalf("exit code = %d, want %d (stdout=%s stderr=%s)",
					code, tt.code, stdout.String(), stderr.String())
			}
			if tt.code == 1 && !strings.Contains(stdout.String(), `"legal": false`) {
				t.Fatalf("illegal state did not emit a report: %s", stdout.String())
			}
			if tt.code == 2 && !strings.Contains(stderr.String(), `"error"`) {
				t.Fatalf("malformed input did not emit an error: %s", stderr.String())
			}
		})
	}
}
