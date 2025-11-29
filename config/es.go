package config

import (
	"os"
	"github.com/olivere/elastic/v7"
)

var ES *elastic.Client

func NewElastic() *elastic.Client {
	url := os.Getenv("ELASTIC_URL")
	if url == "" {
		url = "http://localhost:9200"
	}
	client, err := elastic.NewClient(
		elastic.SetURL(url),
		elastic.SetSniff(false),
	)
	if err != nil {
		panic("Unable to connect to Elasticsearch" + err.Error())
	}
	return client
} 