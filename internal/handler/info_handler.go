package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type InfoHandler struct{}

func NewInfoHandler() *InfoHandler {
	return &InfoHandler{}
}

// InfoHandler возвращает количество дней до Нового года
func (h *InfoHandler) InfoHandler(c *gin.Context) {
	now := time.Now()

	nextYear := now.Year() + 1
	newYear := time.Date(nextYear, time.January, 1, 0, 0, 0, 0, now.Location())

	daysUntilNewYear := int(newYear.Sub(now).Hours() / 24)
	if daysUntilNewYear < 0 {
		daysUntilNewYear = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"days_until_new_year": daysUntilNewYear,
	})
}
