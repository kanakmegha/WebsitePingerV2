package checker

import (
	"context"
	"net"
	"strings"
	"time"
)

type DNSResult struct {
	HasARecord    bool     `json:"has_a_record"`
	HasAAAARecord bool     `json:"has_aaaa_record"`
	ARecords      []string `json:"a_records"`
	AAAARecords   []string `json:"aaaa_records"`
	MXRecords     []string `json:"mx_records"`
	TXTRecords    []string `json:"txt_records"`
	SPFValid      bool     `json:"spf_valid"`
	DMARCValid    bool     `json:"dmarc_valid"`
	Error         string   `json:"error,omitempty"`
}

func CheckDNS(ctx context.Context, domain string, timeoutSec int) (*DNSResult, error) {
	if timeoutSec <= 0 {
		timeoutSec = 5
	}

	resolver := &net.Resolver{}
	ctxTimeout, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	result := &DNSResult{
		ARecords:    []string{},
		AAAARecords: []string{},
		MXRecords:   []string{},
		TXTRecords:  []string{},
	}

	// 1. IP Lookup (A and AAAA records)
	ips, err := resolver.LookupIPAddr(ctxTimeout, domain)
	if err == nil {
		for _, ip := range ips {
			if ip.IP.To4() != nil {
				result.ARecords = append(result.ARecords, ip.IP.String())
			} else if ip.IP.To16() != nil {
				result.AAAARecords = append(result.AAAARecords, ip.IP.String())
			}
		}
	}
	result.HasARecord = len(result.ARecords) > 0
	result.HasAAAARecord = len(result.AAAARecords) > 0

	// 2. MX Records Lookup
	mxs, err := resolver.LookupMX(ctxTimeout, domain)
	if err == nil {
		for _, mx := range mxs {
			result.MXRecords = append(result.MXRecords, mx.Host)
		}
	}

	// 3. TXT Records & SPF Validation
	txts, err := resolver.LookupTXT(ctxTimeout, domain)
	if err == nil {
		for _, txt := range txts {
			result.TXTRecords = append(result.TXTRecords, txt)
			if strings.HasPrefix(strings.TrimSpace(txt), "v=spf1") {
				result.SPFValid = true
			}
		}
	}

	// 4. DMARC Record Lookup (_dmarc.domain)
	dmarcDomain := "_dmarc." + domain
	dmarcTxts, err := resolver.LookupTXT(ctxTimeout, dmarcDomain)
	if err == nil {
		for _, txt := range dmarcTxts {
			if strings.HasPrefix(strings.TrimSpace(txt), "v=DMARC1") {
				result.DMARCValid = true
				break
			}
		}
	}

	return result, nil
}
