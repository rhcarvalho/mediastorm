package main

import (
	"bytes"
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/mediastoredata"
)

var (
	// Required:
	endpoint = flag.String("endpoint", "", "container endpoint (required)")
	// Optional:
	path  = flag.String("path", "mediastorm/"+strings.Replace(time.Now().UTC().Format(time.RFC3339Nano), ":", "-", -1), "path to write")
	tps   = flag.Int("tps", 1, "PutObject TPS")
	size  = flag.Int("size", 512, "content size to write (random bytes)")
	count = flag.Int("n", 0, "number of requests to send, 0 means infinite")
	debug = flag.Bool("debug", false, "enable SDK debugging of HTTP requests")
)

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	if *endpoint == "" {
		fmt.Fprintln(os.Stderr, "Non-empty endpoint is required.")
		flag.Usage()
		os.Exit(1)
	}

	cfg := aws.NewConfig().
		WithEndpoint(*endpoint).
		WithCredentialsChainVerboseErrors(true)
	if *debug {
		// Enable relevant debugging flags. See aws.LogDebug* for more.
		cfg.WithLogLevel(aws.LogDebugWithRequestRetries | aws.LogDebugWithRequestErrors)
	}
	s := session.Must(session.NewSession(cfg))

	_, err := s.Config.Credentials.Get()
	if err != nil {
		log.Fatalf("AWS credentials: %v", err)
	}
	if aws.StringValue(s.Config.Region) == "" {
		log.Fatal("AWS region is not defined. Set the AWS_REGION environment variable or relevant config files.")
	}

	msd := mediastoredata.New(s)

	buf := make([]byte, *size)
	if _, err := rand.Read(buf); err != nil {
		log.Fatal(err)
	}
	input := &mediastoredata.PutObjectInput{
		Body: bytes.NewReader(buf),
		Path: path,
	}

	log.Printf("MediaStorm: PutObject %s (%d B) @ %d TPS", *path, *size, *tps)

	interval := time.Duration(int(time.Second) / *tps)
	timer := time.NewTicker(interval)
	defer timer.Stop()
	for i := 1; *count == 0 || i <= *count; i++ {
		req, output := msd.PutObjectRequest(input)
		req.HTTPRequest.Header.Set("User-Agent", "MediaStorm")
		req.HTTPRequest.Header.Set("X-Request-ID", fmt.Sprint(i))
		err := req.Send()
		if err != nil {
			log.Printf("[ERROR] PutObject: %v", err)
		}
		log.Printf("[INFO] %v", output)
		<-timer.C
	}
}
