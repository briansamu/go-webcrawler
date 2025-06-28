package storage

import "webcrawler/internal/models"

type Storage interface {
	Connect()
	Disconnect()
	InsertPage(page models.Page)
}
