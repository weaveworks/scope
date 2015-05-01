package nameserver

import (
	"github.com/miekg/dns"
)

const (
	localTTL    uint32 = 30 // somewhat arbitrary; we don't expect anyone downstream to cache results
	negLocalTTL        = 30 // TTL for negative local resolutions
	minUDPSize         = 512
	maxUDPSize         = 65535
)

func makeHeader(r *dns.Msg, q *dns.Question) *dns.RR_Header {
	return &dns.RR_Header{
		Name: q.Name, Rrtype: q.Qtype,
		Class: dns.ClassINET, Ttl: localTTL}
}

func makeReply(r *dns.Msg, as []dns.RR) *dns.Msg {
	m := new(dns.Msg)
	m.SetReply(r)
	m.RecursionAvailable = true
	m.Answer = as
	return m
}

func makeTruncatedReply(r *dns.Msg) *dns.Msg {
	// for truncated response, we create a minimal reply with the Truncated bit set
	reply := new(dns.Msg)
	reply.SetReply(r)
	reply.Truncated = true
	return reply
}

func makeAddressReply(r *dns.Msg, q *dns.Question, addrs []ZoneRecord) *dns.Msg {
	answers := make([]dns.RR, len(addrs))
	header := makeHeader(r, q)
	count := 0
	for _, addr := range addrs {
		ip := addr.IP()
		ip4 := ip.To4()

		switch q.Qtype {
		case dns.TypeA:
			if ip4 != nil {
				answers[count] = &dns.A{Hdr: *header, A: ip}
				count++
			}
		case dns.TypeAAAA:
			if ip4 == nil {
				answers[count] = &dns.AAAA{Hdr: *header, AAAA: ip}
				count++
			}
		}
	}
	return makeReply(r, answers[:count])
}

func makePTRReply(r *dns.Msg, q *dns.Question, names []ZoneRecord) *dns.Msg {
	answers := make([]dns.RR, len(names))
	header := makeHeader(r, q)
	for i, name := range names {
		answers[i] = &dns.PTR{Hdr: *header, Ptr: name.Name()}
	}
	return makeReply(r, answers)
}

func makeDNSFailResponse(r *dns.Msg) *dns.Msg {
	m := new(dns.Msg)
	m.SetReply(r)
	m.RecursionAvailable = true
	m.Rcode = dns.RcodeNameError
	return m
}

func makeDNSNotImplResponse(r *dns.Msg) *dns.Msg {
	m := new(dns.Msg)
	m.SetReply(r)
	m.RecursionAvailable = true
	m.Rcode = dns.RcodeNotImplemented
	return m
}

// get the maximum UDP-reply length
func getMaxReplyLen(r *dns.Msg, proto dnsProtocol) int {
	maxLen := minUDPSize
	if proto == protTCP {
		maxLen = maxUDPSize
	} else if opt := r.IsEdns0(); opt != nil {
		maxLen = int(opt.UDPSize())
	}
	return maxLen
}
