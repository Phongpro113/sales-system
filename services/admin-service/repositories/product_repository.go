package repositories

import (
	"admin-service/models"
	"gorm.io/gorm"
)

type ProductRepository interface {
	GetAll(page, limit int) ([]models.Product, int64, error)
	Create(product *models.Product) error
	Update(id string, updates map[string]interface{}) (*models.Product, error)
	GetByID(id string) (*models.Product, error)
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db}
}

func (r *productRepository) GetAll(page, limit int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64
	offset := (page - 1) * limit

	query := r.db.Model(&models.Product{})
	query.Count(&total)
	err := query.Offset(offset).Limit(limit).Order("id DESC").Find(&products).Error
	return products, total, err
}

func (r *productRepository) Create(product *models.Product) error {
	return r.db.Create(product).Error
}

func (r *productRepository) GetByID(id string) (*models.Product, error) {
	var product models.Product
	err := r.db.First(&product, id).Error
	return &product, err
}

func (r *productRepository) Update(id string, updates map[string]interface{}) (*models.Product, error) {
	var product models.Product
	if err := r.db.First(&product, id).Error; err != nil {
		return nil, err
	}

	delete(updates, "id")
	delete(updates, "created_at")

	if err := r.db.Model(&product).Updates(updates).Error; err != nil {
		return nil, err
	}

	r.db.First(&product, id)
	return &product, nil
}
