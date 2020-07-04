package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/miekg/dns"
	"gopkg.in/yaml.v2"
)

// Slots .
type Slots map[string]map[string]bool

// Config .
type Config struct {
	OutputFile string
	SlotsFile  string
	Verbose    bool
	LookupDNS  bool
}

const googleDNS = "8.8.8.8:53"

var usage = `Usage: dns-slots [options...] < domains-file

Options:
  -o  File to output slot machine results. Default is stdout.
  -s  File tha contains the options for each slot. Default is slots-small.yml.
  -v  Run in verbose mode.
  -d  Do a DNS lookup on each slot machine result and output only those with DNS records.
  -h  Print usage and exit.
`

func main() {
	outputFile := flag.String("o", "", "-o output-file")
	slotsFile := flag.String("s", "slots-small.yml", "-s slots-file")
	verbose := flag.Bool("v", false, "")
	lookupDNS := flag.Bool("d", false, "")
	help := flag.Bool("h", false, "")
	flag.Parse()

	// show usage
	if *help {
		fmt.Fprint(os.Stdout, usage)
		os.Exit(0)
	}

	conf := &Config{
		OutputFile: *outputFile,
		SlotsFile:  *slotsFile,
		Verbose:    *verbose,
		LookupDNS:  *lookupDNS,
	}

	err := doWork(conf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v", err)
		panic("failed to spin slots")
	}
}

func doWork(conf *Config) error {
	slots, err := readSlotsFile(conf.SlotsFile)
	if err != nil {
		return err
	}

	var output io.Writer
	if len(conf.OutputFile) > 0 {
		// write results to file
		fp, err := os.OpenFile(conf.OutputFile, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			return err
		}
		output = fp
	} else {
		// write results in standard output
		output = os.Stdout
	}

	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		domain := strings.ToLower(sc.Text())
		tokens := tokenize(domain)
		matches, indices, _ := matchSlots(tokens, slots)
		spin(tokens, matches, indices, slots, map[string]bool{}, output, conf.LookupDNS)
	}

	return nil
}

// spin to win!
func spin(tokens []string, matches []string, indices []int, slots Slots, seen map[string]bool, output io.Writer, lookupDNS bool) {
	// memoize
	outcome := strings.Join(tokens, "")
	if _, seenBefore := seen[outcome]; seenBefore {
		return
	}
	seen[outcome] = true

	if lookupDNS {
		exists, err := dnsRecordsExist(outcome, googleDNS)
		if err == nil && exists {
			// output winning result
			fmt.Fprintln(output, outcome)
		}
	} else {
		fmt.Fprintln(output, outcome)
	}

	if len(matches) == 0 || len(indices) == 0 {
		return
	}

	for i := 0; i < len(indices); i++ {
		slot, ok := slots[matches[i]]
		if !ok {
			panic("oops, this should have never happened")
		}

		for v := range slot {
			newTokens := make([]string, len(tokens))
			copy(newTokens, tokens)

			newTokens[indices[i]] = v

			spin(newTokens, matches[1:], indices[1:], slots, seen, output, lookupDNS)
		}
	}
}

func matchSlots(tokens []string, slots Slots) (matches []string, indices []int, combinations int) {
	matches = []string{}
	indices = []int{}
	combinations = 1

	for tokenIdx, token := range tokens {
		for slotName, slot := range slots {
			if _, ok := slot[token]; ok {
				matches = append(matches, slotName)
				indices = append(indices, tokenIdx)
				combinations *= len(slot)
			}
		}
	}

	return matches, indices, combinations
}

// tokenize splits a DNS entry on "." and "-"
func tokenize(s string) []string {
	result := []string{}

	prevIdx := 0
	for idx, c := range s {
		if c == '.' || c == '-' {
			result = append(result, s[prevIdx:idx], string(c))
			prevIdx = idx + 1
		}
	}
	// append the final segment of the domain (e.g. example."com")
	result = append(result, s[prevIdx:])

	return result
}

func dnsRecordsExist(fqdn, serverAddr string) (bool, error) {
	var m dns.Msg
	m.SetQuestion(dns.Fqdn(fqdn), dns.TypeA)
	in, err := dns.Exchange(&m, serverAddr)
	if err != nil {
		return false, err
	}
	if len(in.Answer) == 0 {
		return false, nil
	}

	return true, nil
}

func readSlotsFile(filename string) (Slots, error) {
	yamlMap := map[string][]string{}

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(content, yamlMap)
	if err != nil {
		return nil, err
	}

	res := map[string]map[string]bool{}

	for k, v := range yamlMap {
		mm := map[string]bool{}
		for _, vv := range v {
			mm[vv] = true
		}
		res[k] = mm
	}

	return res, nil
}
