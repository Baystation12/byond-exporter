package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/baystation12/byond-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	pb "github.com/prometheus/client_model/go"
)

var bind = flag.String("bind", "", "host to bind on")
var host = flag.String("host", "", "host to gather metrics from")
var key = flag.String("key", "", "key to use to authenticate to the game server")

func main() {
	flag.Parse()

	if *bind == "" {
		log.Fatal("required flag: bind")
	}

	if *host == "" {
		log.Fatal("required flag: host")
	}

	log.Println("starting")

	gatherer := NewBYONDGatherer(*host, *key)
	handler := promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{})
	http.Handle("/metrics", handler)
	log.Fatal(http.ListenAndServe(*bind, nil))
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
