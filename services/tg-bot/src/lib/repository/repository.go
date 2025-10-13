package repository

import (
	models "github.com/saintson-network-seller/additions/models"
	"errors"
	logger "tg-bot/src/lib/logger"
)

type Client struct {
	tableName string
	// В будущем здесь будут поля для подключения db *sql.DB, connectionString string
}

func NewClient(tableName string) *Client {
	return &Client{
		tableName: tableName,
	}
}

func (c *Client) CloseConnection() error {
	logger.Log.Infof("Closing connection to table: %s\n", c.tableName)
	return nil
}

func (c *Client) GetById(id int) (models.Product, error) {
	switch id {
	case 1:
		return models.Product{
			OfficialName:   "Premium Subscription",
			ShortName:      "premium",
			Description:    "Full access to all premium features",
			AmountCurrency: "RUB",
			AmountPrice:    100,
		}, nil
	case 2:
		return models.Product{
			OfficialName:   "Basic Subscription", 
			ShortName:      "basic",
			Description:    "Access to basic features",
			AmountCurrency: "RUB",
			AmountPrice:    150,
		}, nil
	case 3:
		return models.Product{
			OfficialName:   "Trial Subscription",
			ShortName:      "trial", 
			Description:    "7-day trial period",
			AmountCurrency: "RUB",
			AmountPrice:    200,
		}, nil
	default:
		return models.Product{}, errors.New("product not found")
	}
}

func (c *Client) Create(product *models.Product) error {
	
	logger.Log.Infof("Creating product in table %s: %s\n", c.tableName, product.OfficialName)
	return nil
	// заглушка insert
}

func (c *Client) Update(id string, product *models.Product) error {  
	logger.Log.Infof("Updating product %s in table %s\n", id, c.tableName)
	return nil
	// заглушка update  
}