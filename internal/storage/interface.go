package storage

import "webcrawler/internal/models"

type Storage interface {
	Connect()
	Disconnect()
	InsertPage(page models.Page)
	SearchPages(query string, page, limit int) ([]models.Page, int, error)
	GetPages(page, limit int) ([]models.Page, int, error)
	GetTotalPages() (int, error)
}
