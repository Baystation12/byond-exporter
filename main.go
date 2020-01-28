package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/Baystation12/byond-go/byond"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	pb "github.com/prometheus/client_model/go"
)

var bind = flag.String("bind", "", "host to bind on")
var host = flag.String("host", "", "host to gather metrics from")
var configFile = flag.String("config_file", "", "path to a bs12-style config.txt, used to set key if set")
var key = flag.String("key", "", "key to use to authenticate to the game server")

func main() {
	flag.Parse()

	if *bind == "" {
		log.Fatal("required flag: bind")
	}

	if *host == "" {
		log.Fatal("required flag: host")
	}

	setKey := *key
	if *configFile != "" {
		configKey, err := extractKey(*configFile)
		if err != nil {
			log.Fatalf("failed to extract key from config file: %s", err)
		}

		if configKey == "" {
			log.Printf("could not find key in config file %s, falling back to flag", *configFile)
		} else {
			setKey = configKey
		}
	}

	log.Println("starting")

	gatherer := NewBYONDGatherer(*host, setKey)
	handler := promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{})
	http.Handle("/metrics", handler)
	log.Fatal(http.ListenAndServe(*bind, nil))
}

func extractKey(fp string) (string, error) {
	re, err := regexp.Compile(`^\s*COMMS_PASSWORD\s+([^\s]+)\s*$`)
	if err != nil {
		return "", fmt.Errorf("failed to compile config regex: %s", err)
	}

	f, err := os.Open(fp)
	if err != nil {
		return "", fmt.Errorf("failed to open config file %s: %s", fp, err)
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	key := ""
	for s.Scan() {
		kg := re.FindStringSubmatch(s.Text())
		if len(kg) == 0 {
			continue
		}
		if kg[1] != "" {
			key = kg[1]
			break
		}
	}

	if err := s.Err(); err != nil {
		return "", fmt.Errorf("error while scanning config file %s: %s", fp, err)
	}

	if key == "" {
		return "", fmt.Errorf("could not find key config line in config file %s", fp)
	}

	return key, nil
}

type BYONDGatherer struct {
	client *byond.QueryClient
	key    string
}

func (g *BYONDGatherer) Gather() ([]*pb.MetricFamily, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := "prometheus_metrics"
	if g.key != "" {
		query += fmt.Sprintf(";key=%s", g.key)
	}

	resp, err := g.client.Query(ctx, []byte(query), true)
	if err != nil {
		return nil, err
	}

	var out []*pb.MetricFamily
	if err := json.Unmarshal(resp, &out); err != nil {
		return nil, err
	}

	return out, nil
}

func NewBYONDGatherer(host, key string) *BYONDGatherer {
	client := byond.NewQueryClient(host)
	return &BYONDGatherer{
		client: client,
		key:    key,
	}
}
