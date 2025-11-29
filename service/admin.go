package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/chrisprojs/Franchiso/config"
	"github.com/chrisprojs/Franchiso/models"
	"github.com/gin-gonic/gin"
)

type DisplayAllRequestForVerificationFranchiseResponse struct {
	Franchises []models.Franchise `json:"franchises"`
}

// DisplayAllRequestForVerificationFranchise menampilkan seluruh franchise yang statusnya 'Menunggu Verifikasi'
func DisplayAllRequestForVerificationFranchise(c *gin.Context, app *config.App) {
	role := c.GetString("role")
	if role != "Admin" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak memiliki akses"})
		return
	}

	var franchises []models.Franchise
	err := app.DB.Model(&franchises).
		Where("status = ?", "Menunggu Verifikasi").
		Select()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data franchise: " + err.Error()})
		return
	}
	resp := DisplayAllRequestForVerificationFranchiseResponse{
		Franchises: franchises,
	}
	c.JSON(http.StatusOK, resp)
}

type VerifyFranchiseRequest struct {
	Status string `json:"status" binding:"required"`
}

func VerifyFranchise(c *gin.Context, app *config.App) {
	id := c.Param("id")
	var req VerifyFranchiseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := c.GetString("role")
	if role != "Admin" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak memiliki akses"})
		return
	}

	// Update franchise status
	var franchise models.Franchise
	_, err := app.DB.Model(&franchise).
		Where("id = ?", id).
		Set("status = ?", req.Status).
		Update()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Franchise tidak berhasil diupdate: %v", err)})
		return
	}

	// Query ulang franchise beserta relasinya
	err = app.DB.Model(&franchise).
		Relation("User").
		Relation("Category").
		Where("franchise.id = ?", id).
		Select()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Gagal mengambil data franchise: %v", err)})
		return
	}

	// Jika terverifikasi, sinkron ke ES
	if req.Status == "Terverifikasi" {
		var user models.User
		if franchise.User != nil {
			user = *franchise.User
		}

		var category models.Category
		if franchise.Category != nil {
			category = *franchise.Category
		}

		// Siapkan data sesuai mapping
		doc := map[string]interface{}{
			"id": franchise.ID.String(),
			"user": map[string]interface{}{
				"user_id": franchise.UserID.String(),
				"name":    user.Name,
			},
			"category": map[string]interface{}{
				"category_id": franchise.CategoryID.String(),
				"category":    category.Category,
			},
			"brand":            franchise.Brand,
			"logo":             franchise.Logo,
			"ad_photos":        franchise.AdPhotos,
			"description":      franchise.Description,
			"investment":       franchise.Investment,
			"monthly_revenue":  franchise.MonthlyRevenue,
			"roi":              franchise.ROI,
			"branch_count":     franchise.BranchCount,
			"year_founded":     franchise.YearFounded,
			"website":          franchise.Website,
			"whatsapp_contact": franchise.WhatsappContact,
			"is_boosted":       franchise.IsBoosted,
			"created_at":       franchise.CreatedAt,
			"updated_at":       franchise.UpdatedAt,
		}
		_, err := app.ES.Index().
			Index("franchises").
			Id(franchise.ID.String()).
			BodyJson(doc).
			Do(context.Background())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal sinkronisasi ke Elasticsearch"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status franchise berhasil diupdate"})
}
