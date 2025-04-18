package database

import (
	"math/rand"
	"store_backend/models"

	"github.com/go-faker/faker/v4"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func createCategories(db *gorm.DB) ([]models.Category, error) {
	categories := []models.Category{
		{Name: "Electronics"},
		{Name: "Clothing"},
		{Name: "Books"},
		{Name: "Home & Kitchen"},
		{Name: "Sports & Outdoors"},
	}

	for i := range categories {
		if err := db.Create(&categories[i]).Error; err != nil {
			return nil, err
		}
	}
	return categories, nil
}

func createProducts(db *gorm.DB, categories []models.Category) ([]models.Product, error) {
	products := make([]models.Product, 0)
	for _, category := range categories {
		numProducts := rand.Intn(6) + 5 // 5-10 products per category
		for i := 0; i < numProducts; i++ {
			price := decimal.NewFromFloat(float64(rand.Intn(99400)+599) / 100)
			product := models.Product{
				Name:       faker.Word() + " " + faker.Word(),
				Price:      price,
				CategoryID: &category.ID,
			}
			if err := db.Create(&product).Error; err != nil {
				return nil, err
			}
			products = append(products, product)
		}
	}
	return products, nil
}

func createCarts(db *gorm.DB, products []models.Product, numCarts int) error {
	for i := 0; i < numCarts; i++ {
		cart := models.Cart{}
		if err := db.Create(&cart).Error; err != nil {
			return err
		}

		// Add random products to cart (1-5 products)
		numProductsInCart := rand.Intn(5) + 1
		cartProducts := make([]models.Product, 0, numProductsInCart)

		for j := 0; j < numProductsInCart; j++ {
			randomProduct := products[rand.Intn(len(products))]
			cartProducts = append(cartProducts, randomProduct)
		}

		if err := db.Model(&cart).Association("Products").Append(cartProducts); err != nil {
			return err
		}
	}
	return nil
}

func Seed(db *gorm.DB) error {
	db.Exec("DELETE FROM cart_products")
	db.Exec("DELETE FROM products")
	db.Exec("DELETE FROM categories")
	db.Exec("DELETE FROM carts")

	// Create categories
	categories, err := createCategories(db)
	if err != nil {
		return err
	}

	// Create products
	products, err := createProducts(db, categories)
	if err != nil {
		return err
	}

	// Create carts
	if err := createCarts(db, products, 5); err != nil {
		return err
	}

	return nil
}

func GetRandomProducts(db *gorm.DB, count int) ([]models.Product, error) {
	var products []models.Product
	err := db.Order("RANDOM()").Limit(count).Find(&products).Error
	return products, err
}
