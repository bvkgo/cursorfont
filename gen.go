// Copyright (c) 2024 BVK Chaitanya

//go:build ignore

// This program generates Go constants for X11 cursor font ids
// from /usr/include/X11/cursorfont.h file.
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/format"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	type Entry struct {
		Name    string
		Value   uint32
		Comment string
	}
	var items []*Entry

	cursorfont, err := os.ReadFile("/usr/include/X11/cursorfont.h")
	if err != nil {
		log.Fatalf("reading cursorfont.h: %v", err)
	}
	re := `^#define \s+ (XC_[a-zA-Z_0-9]+) \s+ ([0-9]+) \s* (/\* .* \*/)? .* $`
	cursorfontRe := regexp.MustCompile(strings.Replace(re, " ", "", -1))

	s := bufio.NewScanner(bytes.NewBuffer(cursorfont))
	for s.Scan() {
		text := strings.TrimSpace(s.Text())
		m := cursorfontRe.FindStringSubmatch(text)
		if m == nil {
			continue
		}
		value, err := strconv.ParseUint(m[2], 10, 32)
		if err != nil {
			log.Fatalf("could not parse keysym value: %v", err)
		}
		item := &Entry{
			Name:    m[1],
			Value:   uint32(value),
			Comment: m[3],
		}
		items = append(items, item)
	}

	// Prepare a buffer with constant definitions.
	constsBuf := &bytes.Buffer{}
	for _, item := range items {
		if len(item.Comment) > 0 {
			fmt.Fprintf(constsBuf, "%s = %d %s\n", item.Name, item.Value, item.Comment)
		} else {
			fmt.Fprintf(constsBuf, "%s = %d\n", item.Name, item.Value)
		}
	}

	// Prepare a buffer with entries for fond-id string to value map.
	symsBuf := &bytes.Buffer{}
	for _, item := range items {
		fmt.Fprintf(symsBuf, "%q: 0x%x,\n", item.Name, item.Value)
	}

	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, `// generated by go generate; DO NOT EDIT.
package cursorfont

import "github.com/jezek/xgb/xproto"

const (
`+constsBuf.String()+`
)

var String2CursorMap = map[string]xproto.Cursor{
`+symsBuf.String()+`
}
`)

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		log.Printf("%s\n", buf.Bytes())
		log.Fatalf("formatting output: %v", err)
	}

	if err := os.WriteFile("cursorids.go", formatted, 0644); err != nil {
		log.Fatalf("writing cursorids.go: %v", err)
	}
}
