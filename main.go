package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: ipx <METHOD> <CIDR> <HOST_HEADER>")
		os.Exit(1)
	}

	method := os.Args[1]
	cidr := os.Args[2]
	hostHeader := os.Args[3]

	ip, ipnet, err := net.ParseCIDR(cidr)

	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup

	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		wg.Add(1)
		//go sendHTTPRequest(ip.String(), "http", hostHeader, method, &wg)
		go sendHTTPRequest(ip.String(), "https", hostHeader, method, &wg)
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

func sendHTTPRequest(host string, scheme string, hostHeader string, method string, wg *sync.WaitGroup) {
	defer wg.Done()

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest(method, scheme+"://"+host, nil)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	req.Host = hostHeader
	resp, err := client.Do(req)

	if err != nil {
		// fmt.Println("Error: ", err)
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}

		contentLength := len(body)

		fmt.Printf(Yellow+"%s "+Reset+Cyan+"%s "+Reset+Green+"[%d] "+Reset+Magenta+"[%d] "+Reset+"%s\n", host, hostHeader, resp.StatusCode, contentLength, scheme)
		//fmt.Printf("Origin: %s Host: %s Status-Code: %d Content-Length: %d Protocol: %s\n", host, hostHeader, resp.StatusCode, contentLength, scheme)
	}
}
