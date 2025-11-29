package main

import (
	"context"
	"log"
	"time"

	"github.com/chrisprojs/Franchiso/config"
	"github.com/chrisprojs/Franchiso/models"
	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/olivere/elastic/v7"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	// Initialize database connections
	db := config.NewPostgres()
	es := config.NewElastic()

	// Check and remove expired boosts
	if err := checkAndRemoveExpiredBoosts(db, es); err != nil {
		log.Fatal("Error checking expired boosts:", err)
	}

	log.Println("Successfully checked and removed expired boosts")
}

func checkAndRemoveExpiredBoosts(db *pg.DB, es *elastic.Client) error {
	now := time.Now()

	// Get all active boosts that have expired
	var expiredBoosts []models.Boost
	err := db.Model(&expiredBoosts).
		Where("is_active = ? AND end_date < ?", true, now).
		Select()

	if err != nil {
		return err
	}

	log.Printf("Found %d expired boosts", len(expiredBoosts))

	for _, boost := range expiredBoosts {
		// Get franchise details
		var franchise models.Franchise
		err := db.Model(&franchise).
			Where("id = ?", boost.FranchiseID).
			Select()

		if err != nil {
			log.Printf("Error getting franchise %s: %v", boost.FranchiseID, err)
			continue
		}

		// Update franchise is_boosted to false in PostgreSQL
		err = updateFranchiseBoostStatus(db, franchise.ID, false)
		if err != nil {
			log.Printf("Error updating franchise boost status in PostgreSQL for %s: %v", franchise.ID, err)
			continue
		}

		// Update franchise is_boosted to false in Elasticsearch
		err = updateFranchiseBoostStatusES(es, franchise.ID.String(), false)
		if err != nil {
			log.Printf("Error updating franchise boost status in Elasticsearch for %s: %v", franchise.ID, err)
			continue
		}

		// Delete the expired boost
		err = deleteBoost(db, boost.ID)
		if err != nil {
			log.Printf("Error deleting boost %s: %v", boost.ID, err)
			continue
		}

		log.Printf("Successfully processed expired boost %s for franchise %s", boost.ID, franchise.ID)
	}

	return nil
}

func updateFranchiseBoostStatus(db *pg.DB, franchiseID uuid.UUID, isBoosted bool) error {
	_, err := db.Model((*models.Franchise)(nil)).
		Set("is_boosted = ?", isBoosted).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", franchiseID).
		Update()

	return err
}

func updateFranchiseBoostStatusES(es *elastic.Client, franchiseID string, isBoosted bool) error {
	ctx := context.Background()

	// Update the document in Elasticsearch
	_, err := es.Update().
		Index("franchises").
		Id(franchiseID).
		Doc(map[string]interface{}{
			"is_boosted": isBoosted,
			"updated_at": time.Now().Format(time.RFC3339),
		}).
		Do(ctx)

	return err
}

func deleteBoost(db *pg.DB, boostID uuid.UUID) error {
	_, err := db.Model((*models.Boost)(nil)).
		Where("id = ?", boostID).
		Delete()

	return err
}
