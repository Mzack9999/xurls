/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package main

import (
	"log"
	"os"
	"sort"
	"text/template"

	"golang.org/x/net/idna"

	"github.com/mvdan/xurls"
)

var regexTmpl = template.Must(template.New("regex").Parse(`// Generated by regexgen

package xurls

const (
	gtld = ` + "`" + `{{.Gtld}}` + "`" + `
)
`))

func reverseJoin(a []string, sep string) string {
	if len(a) == 0 {
		return ""
	}
	if len(a) == 1 {
		return a[0]
	}
	n := len(sep) * (len(a) - 1)
	for i := 0; i < len(a); i++ {
		n += len(a[i])
	}

	b := make([]byte, n)
	bp := copy(b, a[len(a)-1])
	for i := len(a) - 2; i >= 0; i-- {
		s := a[i]
		bp += copy(b[bp:], sep)
		bp += copy(b[bp:], s)
	}
	return string(b)
}

func writeRegex(tlds []string) error {
	allTldsSet := make(map[string]struct{})
	for _, tldlist := range [...][]string{tlds, xurls.PseudoTLDs} {
		for _, tld := range tldlist {
			allTldsSet[tld] = struct{}{}
			asciiTld, err := idna.ToASCII(tld)
			if err != nil {
				return err
			}
			allTldsSet[asciiTld] = struct{}{}
		}
	}
	var allTlds []string
	for tld := range allTldsSet {
		allTlds = append(allTlds, tld)
	}
	sort.Strings(allTlds)
	f, err := os.Create("regex.go")
	if err != nil {
		return err
	}
	return regexTmpl.Execute(f, struct {
		Gtld string
	}{
		Gtld: `(?i)(` + reverseJoin(allTlds, `|`) + `)(?-i)`,
	})
}

func main() {
	if err := writeRegex(xurls.TLDs); err != nil {
		log.Fatalf("Could not write regex.go: %s", err)
	}
}