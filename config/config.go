package config

import (
	"context"
	"fmt"

	"github.com/go-pg/pg/v10"
	"github.com/olivere/elastic/v7"
	"github.com/redis/go-redis/v9"
)

type App struct {
	DB         *pg.DB
	ES         *elastic.Client
	Redis      *redis.Client
	Midtrans   *MidtransConfig
	GoogleMaps *GoogleMapsConfig
	Email      *EmailConfig
}

type dbLogger struct{}

func (d dbLogger) BeforeQuery(ctx context.Context, evt *pg.QueryEvent) (context.Context, error) {
	query, _ := evt.FormattedQuery()
	fmt.Println("QUERY:", string(query))
	return ctx, nil
}
func (d dbLogger) AfterQuery(ctx context.Context, evt *pg.QueryEvent) error {
	return nil
}
