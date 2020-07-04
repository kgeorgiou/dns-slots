# 🎰 dns-slots

```
              .-------.
              |  DNS  |
  ____________|_______|____________
  |  __    __    ___  _____   __    |
  | / _\  / /   /___\/__   \ / _\   |
  | \ \  / /   //  //  / /\ \\ \    |
  | _\ \/ /___/ \_//  / /  \/_\ \ []|
  | \__/\____/\___/   \/     \__/ []|
  |===_______===_______===_______===|
  ||*|       |*|       |*|       |*|| __
  ||*|  dev  |*|  us   |*| west  |*||(__)
  ||*|_______|*|_______|*|_______|*|| ||
  |===_______===_______===_______===| ||
  ||*|       |*|       |*|       |*|| ||
  ||*|  prd  |*|  eu   |*| east  |*|| ||
  ||*|_______|*|_______|*|_______|*||_//
  |===_______===_______===_______===|_/
  ||*|       |*|       |*|       |*||
  ||*|  uat  |*|  af   |*| north |*||
  ||*|_______|*|_______|*|_______|*||
  |===___________________________===|
  |  /___________________________\  |
  |   |                         |   |
 _|    \_______________________/    |_
(_____________________________________)
```

Yet another tool for DNS enumeration.

In contrast to subdomain guessing tools, dns-slots is fed with subdomains collected *a priori* in order to make educated guesses to discover more.

For example if we know `dev-auth-1.example.com` is a valid subdomain, there's a good chance the following subdomains also exist: 
- `dev-auth-2.example.com`
- `prd-auth-1.example.com`
- `prd-auth-2.example.com`

## Give it a Spin!

**Build**  
```
go build
```

**Usage**
```
Usage: dns-slots [options...] < domains-file

Options:
  -o  File to output slot machine results. Default is stdout.
  -s  File tha contains the options for each slot. Default is slots-small.yml.
  -v  Run in verbose mode.
  -d  Do a DNS lookup on each slot machine result and output only those with DNS records.
  -h  Print usage and exit.
```

**Run**
```
$ dns-slots -o output.txt -d < known-domains.txt
```

## TO-DO
- [ ] Spin on non clearly delimitted slots (e.g. auth**1**)
- [ ] Smart-spin mode on sequence slots (e.g. 0, 1,2,3 and a,b,c)
  - If "1" doesn't exist, then it's very likely subsequent numbers won't exist either
- [ ] Add UTF-8 (rune) support
- [ ] Create
  - [x] slots-small.yml
  - [x] slots-medium.yml
  - [ ] slots-large.yml 
- [ ] Add a `-max-depth` flag to limit permutations
- [ ] Write benchmarks
