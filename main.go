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

	"github.com/spf13/cobra"
)

var (
	cidr               string
	domain             string
	delta              int
	proxyContentLength int
)

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "ipx",
	Short: "Origin IP Discovery Tool",
	Run: func(cmd *cobra.Command, args []string) {
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

		var wg sync.WaitGroup

		for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
			wg.Add(1)
			go worker(ip.String(), "http", domain, &wg)
		}

		wg.Wait()
	},
}

func init() {
	rootCmd.Flags().StringVar(&cidr, "cidr", "", "CIDR range to scan for IPs (e.g., 192.168.1.0/24)")
	rootCmd.Flags().StringVar(&domain, "domain", "", "Domain name to use in the Host header")
	rootCmd.Flags().IntVar(&delta, "delta", 0, "Allowed range for content length difference (default 0)")

	rootCmd.MarkFlagRequired("cidr")
	rootCmd.MarkFlagRequired("domain")
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
