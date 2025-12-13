package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/chrisprojs/Franchiso/config"
	"github.com/chrisprojs/Franchiso/models"
	"github.com/chrisprojs/Franchiso/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	midtrans "github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	snap "github.com/midtrans/midtrans-go/snap"
)

type BoostFranchiseRequest struct {
	Package string `json:"package" binding:"required"` // ex: "7", "14", "30"
}

type BoostFranchiseResponse struct {
	SnapToken   string `json:"snap_token"`
	RedirectUrl string `json:"redirect_url"`
}

func BoostFranchise(c *gin.Context, app *config.App) {
	franchiseID := c.Param("id")
	var req BoostFranchiseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := c.GetString("role")
	if role != "Franchisor" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak memiliki akses"})
		return
	}

	// Validasi franchise milik user
	// Hitung harga paket
	var price int
	switch req.Package {
	case "7":
		price = 100000
	case "14":
		price = 180000
	case "30":
		price = 350000
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Paket tidak valid"})
		return
	}

	// Create Boost data (is_active: false)
	packageDay, err := strconv.Atoi(req.Package)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s", err)})
		return
	}

	boost := models.Boost{
		ID:          uuid.New(),
		FranchiseID: uuid.MustParse(franchiseID),
		StartDate:   time.Now(),
		EndDate:     time.Now().AddDate(0, 0, packageDay),
		IsActive:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	_, err = app.DB.Model(&boost).Insert()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat boost"})
		return
	}

	// Buat Snap Token Midtrans
	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  boost.ID.String(),
			GrossAmt: int64(price),
		},
	}
	snapClient := app.Midtrans.SnapClient
	snapResp, err := snapClient.CreateTransaction(snapReq)
	if snapResp.Token == "" || snapResp.RedirectURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Gagal membuat pembayaran: %v", err)})
		return
	}

	res := BoostFranchiseResponse{
		SnapToken:   snapResp.Token,
		RedirectUrl: snapResp.RedirectURL,
	}

	c.JSON(http.StatusOK, res)
}

func BoostPurchaseCallback(c *gin.Context, app *config.App) {
	var notif coreapi.TransactionStatusResponse
	if err := c.ShouldBindJSON(&notif); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if notif.TransactionStatus == "" || 
	notif.TransactionStatus != "settlement" && 
	notif.TransactionStatus != "capture" {
		c.JSON(http.StatusContinue, gin.H{"message": fmt.Sprintf("Status is %s", notif.TransactionStatus)})
		return
	}

	// Ambil boost berdasarkan order_id
	boostID := notif.OrderID
	boost := &models.Boost{}
	err := app.DB.Model(boost).Where("id = ?", boostID).Select()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Boost tidak ditemukan: %v", err)})
		return
	}
	if boost.IsActive {
		c.JSON(http.StatusContinue, gin.H{"message": "Boost already active"})
		return
	}

	// Catat payment
	transactionTime, err := utils.ParseStringToTime(notif.TransactionTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	grossAmount, err := strconv.ParseFloat(notif.GrossAmount, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	payment := models.Payment{
		ID:                uuid.New(),
		BoostID:           boost.ID,
		TransactionID:     uuid.MustParse(notif.TransactionID),
		GrossAmount:       float64(grossAmount),
		PaymentType:       notif.PaymentType,
		TransactionTime:   transactionTime,
		TransactionStatus: notif.TransactionStatus,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	_, err = app.DB.Model(&payment).Insert()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Gagal mencatat payment: %s", err)})
		return
	}

	// If payment is successful, activate boost & franchise
	if notif.TransactionStatus == "settlement" || notif.TransactionStatus == "capture" {
		_, err = app.DB.Model((*models.Boost)(nil)).
			Set("is_active = ?", true).
			Set("updated_at = ?", time.Now()).
			Where("id = ?", boost.ID).
			Update()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Gagal update Boost: %v", err)})
			return
		}

		// Update franchise directly without select (avoid N+1)
		_, err = app.DB.Model((*models.Franchise)(nil)).
			Set("is_boosted = ?", true).
			Set("updated_at = ?", time.Now()).
			Where("id = ?", boost.FranchiseID).
			Update()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Gagal update Franchise PG: %v", err)})
			return
		}

		// Sinkronisasi ke Elasticsearch
		_, err = app.ES.Update().
			Index("franchises").
			Id(boost.FranchiseID.String()).
			Doc(map[string]interface{}{
				"is_boosted": true,
				"updated_at": time.Now(),
			}).
			Do(context.Background())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Gagal update Franchise ES: %v", err)})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}
