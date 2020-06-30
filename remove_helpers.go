package main

import (
	"strings"
)

func TrimBlanks(lines []string) []byte {
	return []byte(strings.TrimSpace(strings.Join(lines, "\n")) + "\n")
}

func FindGoCopyright(lines []string) (int, int) {
	startIdx := -1

	var copyright bool
	for idx, line := range lines {
		if startIdx == -1 && strings.TrimSpace(line) == "/*" {
			startIdx = idx
			continue
		}

		if startIdx > -1 && strings.TrimSpace(line) == "*/" {
			if copyright {
				return startIdx, idx
			}

			// reset
			startIdx = -1
			copyright = false
			continue
		}

		if startIdx > -1 && // inside comment block
			idx == startIdx+1 && // next line
			strings.HasPrefix(strings.TrimSpace(line), "Copyright ") {
			copyright = true
		}
	}
	return -1, -1
}

func FindBashCopyright(lines []string) (int, int) {
	startIdx := -1

	var copyright bool
	for idx, line := range lines {
		if startIdx == -1 && strings.HasPrefix(strings.TrimSpace(line), "# Copyright ") {
			startIdx = idx
			copyright = true
			continue
		}

		if startIdx > -1 && !strings.HasPrefix(strings.TrimSpace(line), "#") {
			if copyright {
				return startIdx, idx - 1
			}

			// reset
			startIdx = -1
			copyright = false
			continue
		}
	}
	return -1, -1
}
