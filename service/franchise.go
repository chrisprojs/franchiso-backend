package service

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/chrisprojs/Franchiso/config"
	"github.com/chrisprojs/Franchiso/models"
	"github.com/chrisprojs/Franchiso/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/olivere/elastic/v7"
)

type UploadFranchiseRequest struct {
	CategoryID      string `form:"category_id" binding:"required"`
	Brand           string `form:"brand" binding:"required"`
	Description     string `form:"description" binding:"required"`
	Investment      string `form:"investment" binding:"required"`
	MonthlyRevenue  string `form:"monthly_revenue" binding:"required"`
	ROI             string `form:"roi" binding:"required"`
	BranchCount     string `form:"branch_count" binding:"required"`
	YearFounded     string `form:"year_founded" binding:"required"`
	Website         string `form:"website" binding:"required"`
	WhatsappContact string `form:"whatsapp_contact" binding:"required"`

	// Files
	Logo     *multipart.FileHeader   `form:"logo"`
	AdPhotos []*multipart.FileHeader `form:"ad_photos"`
	Stpw     *multipart.FileHeader   `form:"stpw"`
	Nib      *multipart.FileHeader   `form:"nib"`
	Npwp     *multipart.FileHeader   `form:"npwp"`
}

type UploadFranchiseResponse struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func UploadFranchise(c *gin.Context, app *config.App) {
	var req UploadFranchiseRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := c.GetString("role")
	if role != "Franchisor" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak memiliki akses"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}

	// Upload logo
	var logoUrl string
	if req.Logo != nil {
		croppedBuf, format, err := utils.CropImageToSquare(req.Logo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal crop logo"})
			return
		}
		logoFileHeader := utils.BufferToFileHeader(croppedBuf, req.Logo.Filename, format)
		logoUrl, err = utils.UploadToStorageProxy(logoFileHeader)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload logo"})
			return
		}
	}

	// Upload ad_photos (multiple)
	adPhotoUrls := []string{}
	for _, fileHeader := range req.AdPhotos {
		croppedBuf, format, err := utils.CropImageToSquare(fileHeader)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal crop ad_photos"})
			return
		}
		adPhotoFileHeader := utils.BufferToFileHeader(croppedBuf, fileHeader.Filename, format)
		url, err := utils.UploadToStorageProxy(adPhotoFileHeader)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload ad_photos"})
			return
		}
		adPhotoUrls = append(adPhotoUrls, url)
	}

	// Upload stpw
	var stpwUrl string
	if req.Stpw != nil {
		stpwUrl, _ = utils.UploadToStorageProxy(req.Stpw)
	}

	// Upload nib
	var nibUrl string
	if req.Nib != nil {
		nibUrl, _ = utils.UploadToStorageProxy(req.Nib)
	}

	// Upload npwp
	var npwpUrl string
	if req.Npwp != nil {
		npwpUrl, _ = utils.UploadToStorageProxy(req.Npwp)
	}

	investment, err := strconv.Atoi(req.Investment)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid investment value"})
		return
	}
	monthlyRevenue, err := strconv.Atoi(req.MonthlyRevenue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid monthly revenue value"})
		return
	}
	roi, err := strconv.Atoi(req.ROI)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ROI value"})
		return
	}
	branchCount, err := strconv.Atoi(req.BranchCount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch count value"})
		return
	}
	yearFounded, err := strconv.Atoi(req.YearFounded)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year founded value"})
		return
	}

	// Save franchise
	franchise := models.Franchise{
		ID:              uuid.New(),
		UserID:          uuid.MustParse(userID),
		CategoryID:      uuid.MustParse(req.CategoryID),
		Brand:           req.Brand,
		Logo:            logoUrl,
		AdPhotos:        adPhotoUrls,
		Description:     req.Description,
		Investment:      investment,
		MonthlyRevenue:  monthlyRevenue,
		ROI:             roi,
		BranchCount:     branchCount,
		YearFounded:     yearFounded,
		Website:         req.Website,
		WhatsappContact: req.WhatsappContact,
		IsBoosted:       false,
		Stpw:            stpwUrl,
		NIB:             nibUrl,
		NPWP:            npwpUrl,
		Status:          "Menunggu Verifikasi",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	_, err = app.DB.Model(&franchise).Insert()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Gagal menyimpan data franchise: %v", err)})
		return
	}

	resp := UploadFranchiseResponse{
		ID:      franchise.ID.String(),
		Status:  franchise.Status,
		Message: "Data franchise berhasil disimpan, menunggu verifikasi.",
	}
	c.JSON(http.StatusOK, resp)
}

type EditFranchiseRequest struct {
	CategoryID      *string `form:"category_id"`
	Brand           *string `form:"brand"`
	Description     *string `form:"description"`
	Investment      *string `form:"investment"`
	MonthlyRevenue  *string `form:"monthly_revenue"`
	ROI             *string `form:"roi"`
	BranchCount     *string `form:"branch_count"`
	YearFounded     *string `form:"year_founded"`
	Website         *string `form:"website"`
	WhatsappContact *string `form:"whatsapp_contact"`

	// Files
	Logo     *multipart.FileHeader   `form:"logo"`
	AdPhotos []*multipart.FileHeader `form:"ad_photos"`
	Stpw     *multipart.FileHeader   `form:"stpw"`
	Nib      *multipart.FileHeader   `form:"nib"`
	Npwp     *multipart.FileHeader   `form:"npwp"`
}

func EditFranchise(c *gin.Context, app *config.App) {
	franchiseID := c.Param("id")
	var req EditFranchiseRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := c.GetString("role")
	if role != "Franchisor" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak memiliki akses"})
		return
	}

	userID := c.GetString("user_id")
	franchise := &models.Franchise{}
	err := app.DB.Model(franchise).
		Where("id = ?", franchiseID).
		Where("user_id = ?", userID).
		Select()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Franchise tidak ditemukan"})
		return
	}

	columnsToUpdate := []string{}

	// Update fields if provided
	if req.CategoryID != nil && franchise.CategoryID.String() != *req.CategoryID {
		franchise.CategoryID = uuid.MustParse(*req.CategoryID)
		columnsToUpdate = append(columnsToUpdate, "category_id")
	}
	if req.Brand != nil && franchise.Brand != *req.Brand {
		franchise.Brand = *req.Brand
		columnsToUpdate = append(columnsToUpdate, "brand")
	}
	if req.Description != nil && franchise.Description != *req.Description {
		franchise.Description = *req.Description
		columnsToUpdate = append(columnsToUpdate, "description")
	}
	investment, err := strconv.Atoi(*req.Investment)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid investment value"})
		return
	}
	if req.Investment != nil && franchise.Investment != investment {
		franchise.Investment = investment
		columnsToUpdate = append(columnsToUpdate, "investment")
	}
	monthlyRevenue, err := strconv.Atoi(*req.MonthlyRevenue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid monthly revenue value"})
		return
	}
	if req.MonthlyRevenue != nil && franchise.MonthlyRevenue != monthlyRevenue {
		franchise.MonthlyRevenue = monthlyRevenue
		columnsToUpdate = append(columnsToUpdate, "monthly_revenue")
	}
	roi, err := strconv.Atoi(*req.ROI)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ROI value"})
		return
	}
	if req.ROI != nil && franchise.ROI != roi {
		franchise.ROI = roi
		columnsToUpdate = append(columnsToUpdate, "roi")
	}
	branchCount, err := strconv.Atoi(*req.BranchCount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch count value"})
		return
	}
	if req.BranchCount != nil && franchise.BranchCount != branchCount {
		franchise.BranchCount = branchCount
		columnsToUpdate = append(columnsToUpdate, "branch_count")
	}
	yearFounded, err := strconv.Atoi(*req.YearFounded)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year founded value"})
		return
	}
	if req.YearFounded != nil && franchise.YearFounded != yearFounded {
		franchise.YearFounded = yearFounded
		columnsToUpdate = append(columnsToUpdate, "year_founded")
	}
	if req.Website != nil && franchise.Website != *req.Website {
		franchise.Website = *req.Website
		columnsToUpdate = append(columnsToUpdate, "website")
	}
	if req.WhatsappContact != nil && franchise.WhatsappContact != *req.WhatsappContact {
		franchise.WhatsappContact = *req.WhatsappContact
		columnsToUpdate = append(columnsToUpdate, "whatsapp_contact")
	}

	// Logo
	if req.Logo != nil {
		croppedBuf, format, err := utils.CropImageToSquare(req.Logo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal crop logo"})
			return
		}
		logoFileHeader := utils.BufferToFileHeader(croppedBuf, req.Logo.Filename, format)
		logoUrl, err := utils.UploadToStorageProxy(logoFileHeader)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload logo"})
			return
		}
		franchise.Logo = logoUrl
		columnsToUpdate = append(columnsToUpdate, "logo")
	}

	// AdPhotos (multiple)
	if req.AdPhotos != nil {
		adPhotoUrls := []string{}
		for _, fileHeader := range req.AdPhotos {
			croppedBuf, format, err := utils.CropImageToSquare(fileHeader)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal crop ad_photos"})
				return
			}
			adPhotoFileHeader := utils.BufferToFileHeader(croppedBuf, fileHeader.Filename, format)
			url, err := utils.UploadToStorageProxy(adPhotoFileHeader)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload ad_photos"})
				return
			}
			adPhotoUrls = append(adPhotoUrls, url)
		}
		franchise.AdPhotos = adPhotoUrls
		columnsToUpdate = append(columnsToUpdate, "ad_photos")
	}

	// NPWP, NIB, SPTW can only be edited if status is Rejected/Waiting for Verification
	if franchise.Status != "Terverifikasi" {
		if req.Stpw != nil {
			stpwUrl, err := utils.UploadToStorageProxy(req.Stpw)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload stpw"})
				return
			}
			franchise.Stpw = stpwUrl
			columnsToUpdate = append(columnsToUpdate, "stpw")
		}
		if req.Nib != nil {
			nibUrl, err := utils.UploadToStorageProxy(req.Nib)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload nib"})
				return
			}
			franchise.NIB = nibUrl
			columnsToUpdate = append(columnsToUpdate, "nib")
		}
		if req.Npwp != nil {
			npwpUrl, err := utils.UploadToStorageProxy(req.Npwp)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload npwp"})
				return
			}
			franchise.NPWP = npwpUrl
			columnsToUpdate = append(columnsToUpdate, "npwp")
		}
	}

	if franchise.Status == "Ditolak" {
		franchise.Status = "Menunggu Verifikasi"
		columnsToUpdate = append(columnsToUpdate, "status")
	}
	franchise.UpdatedAt = time.Now()

	_, err = app.DB.Model(franchise).
		Column(columnsToUpdate...).
		WherePK().
		Where("user_id = ?", userID).
		Update()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Gagal mengupdate franchise: %v", err)})
		return
	}

	// Elasticsearch sync if verified
	if franchise.Status == "Terverifikasi" {
		var user models.User
		if franchise.User != nil {
			user = *franchise.User
		}
		var category models.Category
		if franchise.Category != nil {
			category = *franchise.Category
		}

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
			Refresh("true").
			Do(context.Background())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal sinkronisasi ke Elasticsearch"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Franchise berhasil diupdate"})
}

func DisplayFranchiseDetailByID(c *gin.Context, app *config.App) {
	franchiseID := c.Param("id")
	showPrivate := c.DefaultQuery("showPrivate", "false")

	if showPrivate == "true" {
		// Get from Postgres
		franchise := &models.Franchise{}
		err := app.DB.Model(franchise).
			Relation("User").
			Relation("Category").
			Where("franchise.id = ?", franchiseID).
			Select()
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Franchise tidak ditemukan"})
			return
		}

		role := c.GetString("role")
		userID := c.GetString("user_id")
		if role != "Admin" && !(role == "Franchisor" && userID == franchise.UserID.String()) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak"})
			return
		}

		c.JSON(http.StatusOK, franchise)
		return
	}

	// Get from Elasticsearch
	res, err := app.ES.Get().
		Index("franchises").
		Id(franchiseID).
		Do(context.Background())
	if err != nil || !res.Found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Franchise tidak ditemukan"})
		return
	}
	var franchise models.FranchiseES
	err = json.Unmarshal(res.Source, &franchise)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal decode data franchise dari Elasticsearch"})
		return
	}
	c.JSON(http.StatusOK, franchise)
}

type DisplayMyFranchisesResponse struct {
	Franchises []models.Franchise `json:"franchises"`
}

func DisplayMyFranchises(c *gin.Context, app *config.App) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}

	role := c.GetString("role")
	if role != "Franchisor" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak memiliki akses"})
		return
	}

	var franchises []models.Franchise
	err := app.DB.Model(&franchises).
		Where("user_id = ?", userID).
		Relation("Category").
		Select()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data franchise"})
		return
	}

	response := DisplayMyFranchisesResponse{
		Franchises: franchises,
	}

	c.JSON(http.StatusOK, response)
}

type SearchFranchiseRequest struct {
	SearchQuery       string
	Category          *string
	MinInvestment     *int
	MaxInvestment     *int
	MinMonthlyRevenue *int
	MinROI            *int
	MaxROI            *int
	MinBranchCount    *int
	MaxBranchCount    *int
	MinYearFounded    *int
	MaxYearFounded    *int
	Page              *int
	Limit             *int
}

type SearchFranchiseResponse struct {
	Total      int64                `json:"total"`
	Franchises []models.FranchiseES `json:"franchises"`
}

func SearchingFranchise(c *gin.Context, app *config.App) {
	// Bind request parameters from query string
	var req SearchFranchiseRequest

	// Get parameters from query string
	if searchQuery := c.Query("searchQuery"); searchQuery != "" {
		req.SearchQuery = searchQuery
	}

	if category := c.Query("category"); category != "" {
		req.Category = &category
	}

	if minInvestmentStr := c.Query("minInvestment"); minInvestmentStr != "" {
		if minInvestment, err := strconv.Atoi(minInvestmentStr); err == nil {
			req.MinInvestment = &minInvestment
		}
	}

	if maxInvestmentStr := c.Query("maxInvestment"); maxInvestmentStr != "" {
		if maxInvestment, err := strconv.Atoi(maxInvestmentStr); err == nil {
			req.MaxInvestment = &maxInvestment
		}
	}

	if minMonthlyRevenueStr := c.Query("minMonthlyRevenue"); minMonthlyRevenueStr != "" {
		if minMonthlyRevenue, err := strconv.Atoi(minMonthlyRevenueStr); err == nil {
			req.MinMonthlyRevenue = &minMonthlyRevenue
		}
	}

	if minROIStr := c.Query("minROI"); minROIStr != "" {
		if minROI, err := strconv.Atoi(minROIStr); err == nil {
			req.MinROI = &minROI
		}
	}

	if maxROIStr := c.Query("maxROI"); maxROIStr != "" {
		if maxROI, err := strconv.Atoi(maxROIStr); err == nil {
			req.MaxROI = &maxROI
		}
	}

	if minBranchCountStr := c.Query("minBranchCount"); minBranchCountStr != "" {
		if minBranchCount, err := strconv.Atoi(minBranchCountStr); err == nil {
			req.MinBranchCount = &minBranchCount
		}
	}

	if maxBranchCountStr := c.Query("maxBranchCount"); maxBranchCountStr != "" {
		if maxBranchCount, err := strconv.Atoi(maxBranchCountStr); err == nil {
			req.MaxBranchCount = &maxBranchCount
		}
	}

	if minYearFoundedStr := c.Query("minYearFounded"); minYearFoundedStr != "" {
		if minYearFounded, err := strconv.Atoi(minYearFoundedStr); err == nil {
			req.MinYearFounded = &minYearFounded
		}
	}

	if maxYearFoundedStr := c.Query("maxYearFounded"); maxYearFoundedStr != "" {
		if maxYearFounded, err := strconv.Atoi(maxYearFoundedStr); err == nil {
			req.MaxYearFounded = &maxYearFounded
		}
	}

	// Set default values for page and limit
	defaultPage := 1
	defaultLimit := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = &page
		} else {
			req.Page = &defaultPage
		}
	} else {
		req.Page = &defaultPage
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = &limit
		} else {
			req.Limit = &defaultLimit
		}
	} else {
		req.Limit = &defaultLimit
	}

	// Debug: Check if franchise data exists in ES
	if req.SearchQuery != "" {
		fmt.Printf("=== DEBUG: Mencari dengan query '%s' ===\n", req.SearchQuery)

		// Check all franchises in ES
		allRes, err := app.ES.Search().
			Index("franchises").
			Size(100).
			Do(context.Background())
		if err == nil {
			fmt.Printf("Total franchise di ES: %d\n", allRes.Hits.TotalHits.Value)
			for _, hit := range allRes.Hits.Hits {
				var f models.FranchiseES
				if err := json.Unmarshal(hit.Source, &f); err == nil {
					fmt.Printf("ES Franchise: ID=%s, Brand=%s, Description=%s\n", f.ID, f.Brand, f.Description)
				}
			}
		}
	}

	// Build ES query
	boolQuery := elastic.NewBoolQuery()

	// Filter by search query (fulltext)
	if req.SearchQuery != "" {
		// Use combined query for flexible search
		searchQuery := elastic.NewBoolQuery()

		// Query for brand: wildcard to search for substring in the middle of text (case-insensitive)
		// Example: "urg" will find "Burger King"
		brandWildcardQuery := elastic.NewWildcardQuery("brand", "*"+req.SearchQuery+"*")
		brandWildcardQuery.CaseInsensitive(true)

		// Alternative query: match query for search with token analysis (better for complete words)
		brandMatchQuery := elastic.NewMatchQuery("brand", req.SearchQuery)
		brandMatchQuery.Operator("or")
		brandMatchQuery.Fuzziness("AUTO")

		// Combination: one of the above queries must match
		searchQuery.Should(brandWildcardQuery)
		searchQuery.Should(brandMatchQuery)
		searchQuery.MinimumShouldMatch("1")
		boolQuery.Must(searchQuery)
	}

	// Filter by category
	if req.Category != nil {
		boolQuery.Filter(elastic.NewTermQuery("category.category_id", *req.Category))
	}

	// Filter by investment range
	if req.MinInvestment != nil || req.MaxInvestment != nil {
		rangeQuery := elastic.NewRangeQuery("investment")
		if req.MinInvestment != nil {
			rangeQuery.Gte(*req.MinInvestment)
		}
		if req.MaxInvestment != nil {
			rangeQuery.Lte(*req.MaxInvestment)
		}
		boolQuery.Filter(rangeQuery)
	}

	// Filter by monthly revenue range
	if req.MinMonthlyRevenue != nil {
		rangeQuery := elastic.NewRangeQuery("monthly_revenue")
		rangeQuery.Gte(*req.MinMonthlyRevenue)
		boolQuery.Filter(rangeQuery)
	}

	// Filter by ROI range
	if req.MinROI != nil || req.MaxROI != nil {
		rangeQuery := elastic.NewRangeQuery("roi")
		if req.MinROI != nil {
			rangeQuery.Gte(*req.MinROI)
		}
		if req.MaxROI != nil {
			rangeQuery.Lte(*req.MaxROI)
		}
		boolQuery.Filter(rangeQuery)
	}

	// Filter by branch count range
	if req.MinBranchCount != nil || req.MaxBranchCount != nil {
		rangeQuery := elastic.NewRangeQuery("branch_count")
		if req.MinBranchCount != nil {
			rangeQuery.Gte(*req.MinBranchCount)
		}
		if req.MaxBranchCount != nil {
			rangeQuery.Lte(*req.MaxBranchCount)
		}
		boolQuery.Filter(rangeQuery)
	}

	// Filter by year founded range
	if req.MinYearFounded != nil || req.MaxYearFounded != nil {
		rangeQuery := elastic.NewRangeQuery("year_founded")
		if req.MinYearFounded != nil {
			rangeQuery.Gte(*req.MinYearFounded)
		}
		if req.MaxYearFounded != nil {
			rangeQuery.Lte(*req.MaxYearFounded)
		}
		boolQuery.Filter(rangeQuery)
	}

	// Query ke ES
	searchService := app.ES.Search().
		Index("franchises").
		Query(boolQuery)

	// Sorting: is_boosted first, then by relevance if search query exists, then by created_at
	if req.SearchQuery != "" {
		searchService = searchService.Sort("is_boosted", false).Sort("_score", false)
	} else {
		searchService = searchService.Sort("is_boosted", false)
	}

	// Pagination
	from := (*req.Page - 1) * *req.Limit
	searchService = searchService.From(from).Size(*req.Limit)

	// Debug: Print query for troubleshooting
	if req.SearchQuery != "" {
		fmt.Printf("Search Query: %s\n", req.SearchQuery)
	}

	// Execute
	res, err := searchService.Do(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mencari franchise"})
		return
	}

	// Debug: Print total hits
	fmt.Printf("Total hits: %d\n", res.Hits.TotalHits.Value)

	// Get results
	franchises := []models.FranchiseES{}
	for _, hit := range res.Hits.Hits {
		var f models.FranchiseES
		if err := json.Unmarshal(hit.Source, &f); err == nil {
			// Debug: Print score for debugging
			if req.SearchQuery != "" && hit.Score != nil {
				fmt.Printf("Franchise: %s, Score: %f\n", f.Brand, *hit.Score)
			}
			franchises = append(franchises, f)
		}
	}

	// Create response
	response := SearchFranchiseResponse{
		Total:      res.Hits.TotalHits.Value,
		Franchises: franchises,
	}

	c.JSON(http.StatusOK, response)
}

type CategoryResponse struct {
	Categories []models.Category `json:"categories"`
}

func CategoryList(c *gin.Context, app *config.App) {
	// Get all categories from database
	var categories []models.Category
	err := app.DB.Model(&categories).Select()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data kategori",
		})
		return
	}

	var response CategoryResponse
	for _, category := range categories {
		response.Categories = append(response.Categories, models.Category{
			ID:       category.ID,
			Category: category.Category,
		})
	}

	c.JSON(http.StatusOK, response)
}

// DeleteFranchise deletes franchise data from Postgres, and if the status
// is "Verified" then also deletes its document from Elasticsearch.
func DeleteFranchise(c *gin.Context, app *config.App) {
	franchiseID := c.Param("id")

	role := c.GetString("role")
	if role != "Franchisor" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak memiliki akses"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}

	// Get franchise to verify ownership and check status
	franchise := &models.Franchise{}
	err := app.DB.Model(franchise).
		Where("id = ?", franchiseID).
		Where("user_id = ?", userID).
		Select()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Franchise tidak ditemukan"})
		return
	}

	// If verified, delete from Elasticsearch first
	if franchise.Status == "Terverifikasi" {
		_, err := app.ES.Delete().
			Index("franchises").
			Id(franchise.ID.String()).
			Refresh("true").
			Do(context.Background())
		if err != nil {
			if esErr, ok := err.(*elastic.Error); ok && esErr.Status == http.StatusNotFound {
				// Ignore if document not found in ES
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus franchise dari Elasticsearch"})
				return
			}
		}
	}

	// Delete from Postgres
	_, err = app.DB.Model((*models.Franchise)(nil)).
		Where("id = ?", franchiseID).
		Where("user_id = ?", userID).
		Delete()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Gagal menghapus franchise: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Franchise berhasil dihapus"})
}
