package matcher

import (
	"net"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Matcher struct {
	IPPrefixes []string
	Domains    []string
	Alert      *Alert
}

func NewMatcher(ipPrefixes []string, domains []string) *Matcher {
	return &Matcher{
		IPPrefixes: ipPrefixes,
		Domains:    domains,
		Alert:      NewAlert(),
	}
}

func (m *Matcher) IPMatch(ip string) {
	for _, prefix := range m.IPPrefixes {
		_, ipnet, err := net.ParseCIDR(prefix)
		if err != nil {
			log.Errorf("Error parsing CIDR %s: %v", prefix, err)
			continue
		}
		parsedIP := net.ParseIP(ip)
		if parsedIP == nil {
			log.Errorf("Error parsing IP %s", ip)
		}
		if ipnet.Contains(parsedIP) {
			m.Alert.SendAlert("IP " + ip + " matches prefix " + prefix)
		}
	}
}

func (m *Matcher) DomainMatch(domain string) {
	for _, d := range m.Domains {
		if strings.EqualFold(domain, d) {
			m.Alert.SendAlert("Domain " + domain + " matches domain " + d)
		}
	}
}
