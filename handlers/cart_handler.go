package handlers

import (
	"errors"
	"log"
	"net/http"
	"store_backend/models"
	"store_backend/repositories"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type CartHandler struct {
	repos repositories.Repositories
}

func (h *CartHandler) RegisterRoutes(e *echo.Echo) error {
	carts := e.Group("/carts")

	carts.GET("", h.GetCarts)
	carts.GET("/:id", h.GetCart)
	carts.POST("", h.CreateCart)
	carts.DELETE("/:id", h.DeleteCart)

	carts.POST("/:id/checkout", h.DeleteCart)

	cartProducts := carts.Group("/:id/products")
	cartProducts.GET("", h.GetCartProducts)
	cartProducts.POST("/:productId", h.AddProductToCart)
	cartProducts.DELETE("/:productId", h.RemoveProductFromCart)
	cartProducts.DELETE("", h.ClearCart)

	return nil
}

func (h *CartHandler) GetCarts(c echo.Context) error {
	carts, err := h.repos.Carts.GetAll()

	if err != nil {
		log.Printf("error getting carts: %v", err)
		return c.NoContent(501)
	}

	return c.JSON(http.StatusOK, carts)
}

type GetCartRequest struct {
	ID uint `param:"id" validate:"required"`
}

func (h *CartHandler) GetCart(c echo.Context) error {
	req := GetCartRequest{}
	if err := c.Bind(&req); err != nil {
		return err
	}

	cart, err := h.checkCartExists(req.ID)
	if err != nil {
		return h.handleCartError(c, err, "Failed to get cart")
	}

	return c.JSON(http.StatusOK, cart)
}

func (h *CartHandler) CreateCart(c echo.Context) error {
	cart := &models.Cart{}

	newCart, err := h.repos.Carts.Create(cart)
	if err != nil {
		log.Printf("error creating cart: %v", err)
		return h.returnErrorJSON(c, http.StatusInternalServerError, "Failed to create cart")
	}

	return c.JSON(http.StatusCreated, newCart)
}

type DeleteCartRequest struct {
	ID uint `param:"id" validate:"required"`
}

func (h *CartHandler) DeleteCart(c echo.Context) error {
	data := DeleteCartRequest{}

	if err := bindAndValidate(c, &data); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	_, err := h.checkCartExists(data.ID)
	if err != nil {
		return h.handleCartError(c, err, "Failed to delete cart")
	}

	err = h.repos.Carts.Delete(data.ID)
	if err != nil {
		log.Printf("error deleting cart: %v", err)
		return h.returnErrorJSON(c, http.StatusInternalServerError, "Failed to delete cart")
	}

	return c.NoContent(http.StatusNoContent)
}

type AddProductToCartRequest struct {
	ID        uint `param:"id" validate:"required"`
	ProductID uint `param:"productId" validate:"required"`
}

func (h *CartHandler) AddProductToCart(c echo.Context) error {
	data := AddProductToCartRequest{}

	if err := bindAndValidate(c, &data); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	_, err := h.checkCartExists(data.ID)
	if err != nil {
		return h.handleCartError(c, err, FailedAddToCart)
	}

	err = h.checkProductExists(data.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return h.returnErrorJSON(c, http.StatusNotFound, "Product not found")
		}
		log.Printf("error getting product: %v", err)
		return h.returnErrorJSON(c, http.StatusInternalServerError, FailedAddToCart)
	}

	err = h.repos.Carts.AddProduct(data.ID, data.ProductID)
	if err != nil {
		log.Printf("error adding product to cart: %v", err)
		return h.returnErrorJSON(c, http.StatusInternalServerError, FailedAddToCart)
	}

	return h.returnUpdatedCart(c, data.ID, "Product added, but failed to retrieve updated cart")
}

type RemoveProductFromCartRequest struct {
	ID        uint `param:"id" validate:"required"`
	ProductID uint `param:"productId" validate:"required"`
}

func (h *CartHandler) RemoveProductFromCart(c echo.Context) error {
	data := RemoveProductFromCartRequest{}

	if err := bindAndValidate(c, &data); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	_, err := h.checkCartExists(data.ID)
	if err != nil {
		return h.handleCartError(c, err, "Failed to remove product from cart")
	}

	err = h.repos.Carts.RemoveProduct(data.ID, data.ProductID)
	if err != nil {
		log.Printf("error removing product from cart: %v", err)
		return h.returnErrorJSON(c, http.StatusInternalServerError, "Failed to remove product from cart")
	}

	return h.returnUpdatedCart(c, data.ID, "Product removed, but failed to retrieve updated cart")
}

type GetCartProductsRequest struct {
	ID uint `param:"id" validate:"required"`
}

func (h *CartHandler) GetCartProducts(c echo.Context) error {
	data := GetCartProductsRequest{}

	if err := bindAndValidate(c, &data); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	products, err := h.repos.Carts.GetProducts(data.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return h.returnErrorJSON(c, http.StatusNotFound, CartNotFound)
		}

		log.Printf("error getting cart products: %v", err)
		return h.returnErrorJSON(c, http.StatusInternalServerError, "Failed to get cart products")
	}

	return c.JSON(http.StatusOK, products)
}

type ClearCartRequest struct {
	ID uint `param:"id" validate:"required"`
}

func (h *CartHandler) ClearCart(c echo.Context) error {
	data := ClearCartRequest{}

	if err := bindAndValidate(c, &data); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	err := h.repos.Carts.ClearCart(data.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return h.returnErrorJSON(c, http.StatusNotFound, CartNotFound)
		}

		log.Printf("error clearing cart: %v", err)
		return h.returnErrorJSON(c, http.StatusInternalServerError, "Failed to clear cart")
	}

	return h.returnUpdatedCart(c, data.ID, "Cart cleared, but failed to retrieve updated cart")
}

func (h *CartHandler) checkCartExists(id uint) (*models.Cart, error) {
	return h.repos.Carts.GetByID(id)
}

func (h *CartHandler) checkProductExists(id uint) error {
	_, err := h.repos.Products.GetByID(id)
	return err
}

func (h *CartHandler) handleCartError(c echo.Context, err error, message string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return h.returnErrorJSON(c, http.StatusNotFound, CartNotFound)
	}
	log.Printf(ErrGettingCart, err)
	return h.returnErrorJSON(c, http.StatusInternalServerError, message)
}

func (h *CartHandler) returnErrorJSON(c echo.Context, status int, message string) error {
	return c.JSON(status, map[string]string{
		"error": message,
	})
}

func (h *CartHandler) returnUpdatedCart(c echo.Context, id uint, errorMessage string) error {
	cart, err := h.repos.Carts.GetByID(id)
	if err != nil {
		log.Printf(ErrUpdatedCart, err)
		return h.returnErrorJSON(c, http.StatusInternalServerError, errorMessage)
	}
	return c.JSON(http.StatusOK, cart)
}

const (
	ErrGettingCart = "error getting cart: %v"
	ErrUpdatedCart = "error getting updated cart: %v"

	CartNotFound    = "Cart not found"
	FailedAddToCart = "Failed to add product to cart"
)
