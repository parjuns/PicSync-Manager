package main

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createRandomProduct(t *testing.T) Product {
	arg := CreateProductParams{
		Name:        RandomString(5),
		Description: RandomString(10),
		Images:      []string{RandomString(5), RandomString(5)},
		Price:       strconv.Itoa(int(RandomInt(100, 1000))),
		UserID:      int(RandomInt(1, 100)),
	}

	productId, err := testPostgresStore.CreateProduct(arg)
	assert.NoError(t, err)
	assert.NotZero(t, productId)
	product, err := testPostgresStore.GetProduct(productId)
	assert.NoError(t, err)
	assert.NotEmpty(t, product)
	return product
}
func Test_DB_CreateProduct(t *testing.T) {
	createRandomProduct(t)
}

func Test_DB_AddProductCompressImages(t *testing.T) {
	product := createRandomProduct(t)
	arg1 := AddProductCompressImagesParams{
		ID:               int(product.ID),
		CompressedImages: []string{"./images/" + RandomString(5) + ".png", "./images/" + RandomString(5) + ".png"},
	}
	err := testPostgresStore.AddProductCompressImages(arg1)
	assert.NoError(t, err)
}

func Test_DB_GetProduct(t *testing.T) {
	product := createRandomProduct(t)

	newProduct, err := testPostgresStore.GetProduct(int(product.ID))
	assert.NoError(t, err)
	assert.NotEmpty(t, newProduct)

	assert.Equal(t, product.ID, newProduct.ID)
	assert.Equal(t, product.Name, newProduct.Name)
	assert.Equal(t, product.Description, newProduct.Description)
	assert.Equal(t, product.Images, newProduct.Images)
	assert.Equal(t, product.Price, newProduct.Price)
	assert.Equal(t, product.UserID, newProduct.UserID)

	newProduct, err = testPostgresStore.GetProduct(int(0))
	assert.Error(t, err)
	assert.Empty(t, newProduct)

}
func Test_DB_CheckUserID(t *testing.T) {
	product := createRandomProduct(t)
	err := testPostgresStore.CheckUserID(int(product.UserID))
	assert.NoError(t, err)
	err = testPostgresStore.CheckUserID(int(0))
	assert.Error(t, err)
}
