package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

type stack struct {
	id    string
	lines []string
}

type dedupedStack struct {
	text  string
	ids   []string
	count int
}

func main() {
	re := regexp.MustCompile(`^goroutine (\d+)`)

	scanner := bufio.NewScanner(bufio.NewReader(os.Stdin))
	lineNum := 0
	counter := -1
	var s *stack
	var stacks []*stack
	for scanner.Scan() {
		lineNum++

		line := scanner.Text()
		matched := re.FindStringSubmatch(line)
		if len(matched) == 2 {
			if s != nil {
				log.Fatalf("line %d: found start of new stack while parsing stack", lineNum)
			}

			s = &stack{
				id:    matched[1],
				lines: make([]string, 0, 10),
			}
			counter = 0
			continue
		}
		if counter == 0 && strings.TrimSpace(line) != "" {
			if strings.HasPrefix(line, "created by") {
				s.lines = append(s.lines, line)
			} else {
				i := strings.LastIndex(line, "(")
				if i <= 0 {
					log.Fatalf("line %d: found last '(' at index %d: %q", lineNum, i, line)
				}
				s.lines = append(s.lines, line[:i])
			}
			counter++
			continue
		}
		if counter == 1 {
			counter = 0
			s.lines = append(s.lines, line)
			continue
		}

		counter = -1
		if s != nil {
			stacks = append(stacks, s)
			s = nil
		}
	}
	if s != nil {
		stacks = append(stacks, s)
		s = nil
	}

	deduped := map[string]dedupedStack{}
	for _, s := range stacks {
		var b strings.Builder
		for _, line := range s.lines {
			b.WriteString(line)
			b.WriteString("\n")
		}
		text := b.String()
		hash := md5sum(text)

		entry := deduped[hash]
		entry.text = text
		entry.ids = append(entry.ids, s.id)
		entry.count++
		deduped[hash] = entry
	}

	for hash, entry := range deduped {
		fmt.Printf("hash: %s\n", hash)
		fmt.Printf("goroutines: %s\n", strings.Join(entry.ids, ","))
		fmt.Println(entry.text)
	}
}

func md5sum(v string) string {
	hash := md5.Sum([]byte(v))
	return hex.EncodeToString(hash[:])
}
