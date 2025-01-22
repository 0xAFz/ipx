package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	delta = 100
)

var proxyContentLength int

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: ipx <CIDR> <DOMAIN>")
		fmt.Println("Exmaple: ipx 192.168.1.0/24 example.com")
		os.Exit(1)
	}

	cidr := os.Args[1]
	domain := os.Args[2]

	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		fmt.Println("Failed to parse CIDR: ", err)
		os.Exit(1)
	}

	body, err := sendHTTPRequest(domain, "https", domain)
	if err != nil {
		fmt.Printf("Failed to resolve origin: %v\n", err)
		os.Exit(1)
	}

	proxyContentLength = len(body)

	// fmt.Printf("Reverse Proxy Content-Length: %d\n\n", proxyContentLength)

	var wg sync.WaitGroup

	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		wg.Add(1)
		go worker(ip.String(), "http", domain, &wg)
	}

	wg.Wait()
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func worker(host, scheme, domain string, wg *sync.WaitGroup) {
	defer wg.Done()

	body, err := sendHTTPRequest(host, scheme, domain)
	if err != nil {
		return
	}

	if !isWithinRange(len(body), proxyContentLength) {
		return
	}

	fmt.Printf("%s %d\n", host, len(body))
}

func sendHTTPRequest(host, scheme, domain string) ([]byte, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Skip SSL verification
		},
	}

	url := fmt.Sprintf("%s://%s", scheme, host)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Host = domain

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func isWithinRange(directLength, proxyLength int) bool {
	return math.Abs(float64(directLength)-float64(proxyLength)) <= float64(delta)
}
