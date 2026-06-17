package model

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID               uuid.UUID
	Title            string
	ShortDescription string
	Description      string
	SKU              string
	Slug             string
	Status           string
	Price            int64
	Currency         string
	DeletedAt        time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewProduct(
	title,
	shortDescription,
	description,
	sku,
	slug,
	status string,
	price int64,
	currency string,
) *Product {
	return &Product{
		Title:            title,
		ShortDescription: shortDescription,
		Description:      description,
		SKU:              sku,
		Slug:             slug,
		Status:           status,
		Price:            price,
		Currency:         currency,
	}
}
