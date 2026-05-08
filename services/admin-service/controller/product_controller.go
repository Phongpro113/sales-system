package controller

import (
	"admin-service/models"
	"admin-service/repositories"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ProductController struct {
	repo repositories.ProductRepository
}

func NewProductController(repo repositories.ProductRepository) *ProductController {
	return &ProductController{repo}
}

func (ctrl *ProductController) formatImageURL(c *gin.Context, imageURL string) string {
	if imageURL == "" || strings.HasPrefix(imageURL, "http") {
		return imageURL
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://" + c.Request.Host
		if c.Request.TLS != nil {
			baseURL = "https://" + c.Request.Host
		}
	}

	if !strings.HasPrefix(imageURL, "/uploads/") {
		imageURL = "/uploads/" + imageURL
	}

	return baseURL + imageURL
}

func (ctrl *ProductController) GetProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	products, total, err := ctrl.repo.GetAll(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	for i := range products {
		products[i].ImageURL = ctrl.formatImageURL(c, products[i].ImageURL)
	}

	c.JSON(http.StatusOK, gin.H{
		"products":    products,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	})
}

func (ctrl *ProductController) CreateProduct(c *gin.Context) {
	name := c.PostForm("name")
	description := c.PostForm("description")
	priceStr := c.PostForm("price")
	stockStr := c.PostForm("stock")
	sku := c.PostForm("sku")
	category := c.PostForm("category")

	price, _ := strconv.ParseFloat(priceStr, 64)
	stock, _ := strconv.Atoi(stockStr)

	// Handle file upload
	imageURL := ""
	file, err := c.FormFile("image")
	if err == nil {
		uploadDir := "./uploads"
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			os.MkdirAll(uploadDir, os.ModePerm)
		}

		filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(file.Filename))
		filepath := filepath.Join(uploadDir, filename)

		if err := c.SaveUploadedFile(file, filepath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
			return
		}
		imageURL = fmt.Sprintf("/uploads/%s", filename)
	}

	product := models.Product{
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
		SKU:         sku,
		Category:    category,
		ImageURL:    imageURL,
	}

	if product.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product name is required"})
		return
	}

	if err := ctrl.repo.Create(&product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	// Format image URL for response
	product.ImageURL = ctrl.formatImageURL(c, product.ImageURL)

	c.JSON(http.StatusCreated, product)
}

func (ctrl *ProductController) UploadImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image uploaded"})
		return
	}

	// Create uploads directory if it doesn't exist
	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, os.ModePerm)
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(file.Filename))
	filepath := filepath.Join(uploadDir, filename)

	// Save file
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	// Return the full URL
	imageURL := ctrl.formatImageURL(c, fmt.Sprintf("/uploads/%s", filename))
	c.JSON(http.StatusOK, gin.H{"image_url": imageURL})
}

func (ctrl *ProductController) UpdateProduct(c *gin.Context) {
	id := c.Param("id")

	updates := make(map[string]interface{})

	// Get form fields
	if name := c.PostForm("name"); name != "" {
		updates["name"] = name
	}
	if desc := c.PostForm("description"); desc != "" {
		updates["description"] = desc
	}
	if priceStr := c.PostForm("price"); priceStr != "" {
		price, _ := strconv.ParseFloat(priceStr, 64)
		updates["price"] = price
	}
	if stockStr := c.PostForm("stock"); stockStr != "" {
		stock, _ := strconv.Atoi(stockStr)
		updates["stock"] = stock
	}
	if sku := c.PostForm("sku"); sku != "" {
		updates["sku"] = sku
	}
	if category := c.PostForm("category"); category != "" {
		updates["category"] = category
	}

	// Handle file upload if present
	file, err := c.FormFile("image")
	if err == nil {
		uploadDir := "./uploads"
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			os.MkdirAll(uploadDir, os.ModePerm)
		}

		filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(file.Filename))
		filepath := filepath.Join(uploadDir, filename)

		if err := c.SaveUploadedFile(file, filepath); err == nil {
			updates["image_url"] = fmt.Sprintf("/uploads/%s", filename)
		}
	}

	product, err := ctrl.repo.Update(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	// Format image URL for response
	product.ImageURL = ctrl.formatImageURL(c, product.ImageURL)

	c.JSON(http.StatusOK, product)
}

func (ctrl *ProductController) GetProduct(c *gin.Context) {
	id := c.Param("id")
	product, err := ctrl.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Format image URL
	product.ImageURL = ctrl.formatImageURL(c, product.ImageURL)

	c.JSON(http.StatusOK, product)
}

func (ctrl *ProductController) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "admin-service",
	})
}
