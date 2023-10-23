package main

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

type Storage interface {
	CreateProduct(CreateProductParams) (int, error)
	CheckUserID(int) error
	GetProduct(int) (Product, error)
	AddProductCompressImages(AddProductCompressImagesParams) error
}

type PostgresStore struct {
	db *sql.DB
}
type Config struct {
	Username string
	Password string
	Dbname   string
	Sslmode  string
}

func NewPostgresStore(config *Config) (*PostgresStore, error) {
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", config.Username, config.Password, config.Dbname, config.Sslmode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil
}

type CreateProductParams struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Images      []string `json:"images"`
	Price       string   `json:"price"`
	UserID      int      `json:"user_id"`
}

type AddProductCompressImagesParams struct {
	ID               int      `json:"id"`
	CompressedImages []string `json:"compressed_images"`
}

const (
	createProductQuery = `
	INSERT INTO products (
	name, description,images,price,user_id
	) VALUES (
	$1, $2, $3, $4, $5
	)
	RETURNING id
	`

	getProductQuery = `
	SELECT id,name,description,images,price,user_id,compressed_images,created_at FROM products WHERE
	id = $1
	`

	addProductCompressImagesQuery = `
	UPDATE products
	SET compressed_images = $2 ,updated_at = (SELECT NOW())
	WHERE products.id = $1
	`

	checkUserIdQuery = `
	SELECT id from users 
	Where users.id = $1
	`
)

func (s *PostgresStore) CreateProduct(arg CreateProductParams) (int, error) {

	var productId int
	err := s.db.QueryRow(createProductQuery,
		arg.Name,
		arg.Description,
		pq.Array(arg.Images),
		arg.Price,
		arg.UserID).Scan(&productId)

	if err != nil {
		return -1, err
	}

	return productId, nil
}

func (s *PostgresStore) AddProductCompressImages(arg AddProductCompressImagesParams) error {

	_, err := s.db.Exec(addProductCompressImagesQuery,
		arg.ID,
		pq.Array(arg.CompressedImages))

	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) GetProduct(id int) (Product, error) {

	row := s.db.QueryRow(getProductQuery, id)
	var i Product
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		pq.Array(&i.Images),
		&i.Price,
		&i.UserID,
		pq.Array(&i.CompressedImages),
		&i.CreatedAt,
	)
	if err != nil {
		return Product{}, err
	}

	return i, nil
}

func (s *PostgresStore) CheckUserID(id int) error {
	var userId int
	err := s.db.QueryRow(checkUserIdQuery, id).Scan(&userId)
	if err != nil {
		return err
	}
	return nil
}
