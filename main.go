package main

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// Required:
	endpoint = flag.String("endpoint", "", "container endpoint (required)")
	host     = flag.String("host", "", "HTTP host header, if different than endpoint")
	insecure = flag.Bool("insecure", false, "ignore server certificate")
	// Optional:
	path  = flag.String("path", "mediastorm/"+strings.Replace(time.Now().UTC().Format(time.RFC3339Nano), ":", "-", -1), "path to write")
	tps   = flag.Int("tps", 1, "PutObject TPS")
	size  = flag.Int("size", 512, "content size to write (random bytes)")
	count = flag.Int("n", 0, "number of requests to send, 0 means infinite")
	// debug    = flag.Bool("debug", false, "enable SDK debugging of HTTP requests")
	poolsize = flag.Int("poolsize", http.DefaultMaxIdleConnsPerHost, "connection pool size")
)

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	if *endpoint == "" {
		fmt.Fprintln(os.Stderr, "Non-empty endpoint is required.")
		flag.Usage()
		os.Exit(1)
	}

	tr := &http.Transport{
		MaxIdleConnsPerHost: *poolsize,
		IdleConnTimeout:     30 * time.Second,
	}
	if *insecure {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	client := &http.Client{Transport: tr}

	buf := make([]byte, *size)
	if _, err := rand.Read(buf); err != nil {
		log.Fatal(err)
	}

	log.Printf("MediaStorm: PutObject %s (%d B) @ %d TPS", *path, *size, *tps)

	interval := time.Duration(int(time.Second) / *tps)
	timer := time.NewTicker(interval)
	defer timer.Stop()

	start := time.Now()
	sent := new(int64)

	go func() {
		for {
			total := atomic.LoadInt64(sent)
			fmt.Printf("=> @ %.1f TPS\n", float64(total)/time.Since(start).Seconds())
			time.Sleep(5 * time.Second)
		}
	}()

	var wg sync.WaitGroup

	for i := 1; *count == 0 || i <= *count; i++ {
		wg.Add(1)
		go func() {
			start := time.Now().UTC()
			defer func() {
				end := time.Now().UTC()
				fmt.Println("METRICS", start.Unix(), float64(end.Sub(start).Nanoseconds())/float64(time.Millisecond), start, end)
			}()
			defer wg.Done()
			req, err := http.NewRequest("PUT", *endpoint+"/"+*path, bytes.NewReader(buf))
			if err != nil {
				log.Printf("[ERROR] NewRequest: %v", err)
				return
			}
			if *host != "" {
				req.Host = *host
			}
			req.Header.Set("User-Agent", "MediaStorm")
			req.Header.Set("X-Request-ID", fmt.Sprint(i))
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("[ERROR] PutObject: %v", err)
				return
			}
			defer resp.Body.Close()
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("[ERROR] Reading Response: %v", err)
				return
			}
			log.Printf("[INFO] %d :: %s", resp.StatusCode, b)
			atomic.AddInt64(sent, 1)
		}()
		<-timer.C
	}
	wg.Wait()
}
