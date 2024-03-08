package clamav

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Summary struct {
	Infected int
	Scanned  int
}

func Scan(dir string) (*Summary, error) {
	s := &Summary{}
	cmd := exec.Command("clamscan", "-r", dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Infected files:") {
			parts := strings.Split(line, ":")
			if len(parts) != 2 {
				return nil, fmt.Errorf("unexpected output from clamscan: %s", line)
			}
			infected, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return nil, err
			}
			s.Infected = infected
		}
		if strings.Contains(line, "Scanned files:") {
			parts := strings.Split(line, ":")
			if len(parts) != 2 {
				return nil, fmt.Errorf("unexpected output from clamscan: %s", line)
			}
			scanned, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return nil, err
			}
			s.Scanned = scanned
		}
	}
	return s, nil
}
