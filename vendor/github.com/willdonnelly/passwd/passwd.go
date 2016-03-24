// Package passwd parsed content and files in form of /etc/passwd
package passwd

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"
)

// An Entry contains all the fields for a specific user
type Entry struct {
	Pass  string
	Uid   string
	Gid   string
	Gecos string
	Home  string
	Shell string
}

// Parse opens the '/etc/passwd' file and parses it into a map from usernames
// to Entries
func Parse() (map[string]Entry, error) {
	return ParseFile("/etc/passwd")
}

// ParseFile opens the file and parses it into a map from usernames to Entries
func ParseFile(path string) (map[string]Entry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	return ParseReader(file)
}

// ParseReader consumes the contents of r and parses it into a map from
// usernames to Entries
func ParseReader(r io.Reader) (map[string]Entry, error) {
	lines := bufio.NewReader(r)
	entries := make(map[string]Entry)
	for {
		line, _, err := lines.ReadLine()
		if err != nil {
			break
		}
		name, entry, err := parseLine(string(copyBytes(line)))
		if err != nil {
			return nil, err
		}
		entries[name] = entry
	}
	return entries, nil
}

func parseLine(line string) (string, Entry, error) {
	fs := strings.Split(line, ":")
	if len(fs) != 7 {
		return "", Entry{}, errors.New("Unexpected number of fields in /etc/passwd")
	}
	return fs[0], Entry{fs[1], fs[2], fs[3], fs[4], fs[5], fs[6]}, nil
}

func copyBytes(x []byte) []byte {
	y := make([]byte, len(x))
	copy(y, x)
	return y
}
