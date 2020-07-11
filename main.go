package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/miekg/dns"
	"gopkg.in/yaml.v2"
)

type void struct{}

// Slots .
type Slots map[string]map[string]bool

// Config .
type Config struct {
	OutputFile string
	SlotsFile  string
	CPUs       int
	Workers    int
	Verbose    bool
	ResolveDNS bool
}

const googleDNS = "8.8.8.8:53"

var usage = `Usage: dns-slots [options...] < domains-file

Options:
  -o  File to output slot machine results. Default is stdout.
  -s  File tha contains the options for each slot. Default is slots/small.yml.
  -w  Number of parallelized workers. Default is 8.
  -c  Number of CPU cores. Machine default is %d.
  -v  Run in verbose mode.
  -d  Do a DNS lookup on each slot machine result and output only those with DNS records.
  -h  Print usage and exit.
`

func main() {
	outputFile := flag.String("o", "", "-o output-file")
	slotsFile := flag.String("s", "slots/small.yml", "-s slots-file")
	workers := flag.Int("w", 8, "")
	cpus := flag.Int("c", runtime.GOMAXPROCS(-1), "")
	verbose := flag.Bool("v", false, "")
	resolveDNS := flag.Bool("d", false, "")
	help := flag.Bool("h", false, "")
	flag.Parse()

	// show usage
	if *help {
		fmt.Fprintf(os.Stdout, usage, runtime.NumCPU())
		os.Exit(0)
	}

	conf := &Config{
		OutputFile: *outputFile,
		SlotsFile:  *slotsFile,
		Workers:    *workers,
		CPUs:       *cpus,
		Verbose:    *verbose,
		ResolveDNS: *resolveDNS,
	}

	runtime.GOMAXPROCS(conf.CPUs)

	err := run(conf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v", err)
		panic("failed to spin slots")
	}
}

func run(conf *Config) error {
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
		defer fp.Close()
		output = fp
	} else {
		// write results in standard output
		output = os.Stdout
	}

	// cretaing a logger for thread safety
	resultLogger := log.New(output, "", 0)

	domains := make(chan string, conf.Workers)
	tracker := make(chan void)
	gather := make(chan string)

	for i := 0; i < conf.Workers; i++ {
		go doWork(domains, tracker, gather, slots, conf.ResolveDNS)
	}

	// gather results, in a single background thread
	// TODO: look into doing this in conf.Workers threads
	go func() {
		for result := range gather {
			resultLogger.Println(result)
		}
		var v void
		tracker <- v
	}()

	// read input from stdin
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		domains <- strings.ToLower(sc.Text())
	}
	close(domains)

	// unblock signals by all (finished) workers
	for i := 0; i < conf.Workers; i++ {
		<-tracker
	}
	close(gather)
	<-tracker

	return nil
}

func doWork(domains <-chan string, tracker chan<- void, gather chan<- string, slots Slots, resolveDNS bool) {
	for domain := range domains {
		tokens := tokenize(domain)
		matches, indices, _ := matchSlots(tokens, slots)
		spin(gather, tokens, matches, indices, slots, map[string]bool{}, resolveDNS)
	}
	// signal this worker is done
	var v void
	tracker <- v
}

// spin to win!
func spin(gather chan<- string, tokens []string, matches []string, indices []int, slots Slots, seen map[string]bool, resolveDNS bool) {
	// memoize
	outcome := strings.Join(tokens, "")
	if _, seenBefore := seen[outcome]; seenBefore {
		return
	}
	seen[outcome] = true

	if resolveDNS {
		resolves, err := dnsResolves(outcome, googleDNS)
		if err == nil && resolves {
			// gather winning result
			gather <- outcome
		}
	} else {
		// gather result if domain has no dns record(s)
		gather <- outcome
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

			spin(gather, newTokens, matches[1:], indices[1:], slots, seen, resolveDNS)
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

func dnsResolves(fqdn, serverAddr string) (bool, error) {
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
