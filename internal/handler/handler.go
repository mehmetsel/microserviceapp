package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"microservice/internal/cache"
)

type Item struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
}

type Handler struct {
	cache *cache.Redis
}

func New(c *cache.Redis) *Handler {
	return &Handler{cache: c}
}

func (h *Handler) Health(c *gin.Context) {
	ctx := c.Request.Context()
	if err := h.cache.Client().Ping(ctx).Err(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "redis": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) ListItems(c *gin.Context) {
	ctx := c.Request.Context()

	var items []Item
	if err := h.cache.Get(ctx, "cache:items:all", &items); err == nil {
		c.Header("X-Cache", "HIT")
		c.JSON(http.StatusOK, items)
		return
	}

	keys, err := h.cache.Scan(ctx, "store:item:*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items = make([]Item, 0, len(keys))
	for _, key := range keys {
		var item Item
		if err := h.cache.Get(ctx, key, &item); err == nil {
			items = append(items, item)
		}
	}

	_ = h.cache.Set(ctx, "cache:items:all", items)
	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, items)
}

func (h *Handler) GetItem(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	cacheKey := fmt.Sprintf("cache:items:%s", id)

	var item Item
	if err := h.cache.Get(ctx, cacheKey, &item); err == nil {
		c.Header("X-Cache", "HIT")
		c.JSON(http.StatusOK, item)
		return
	}

	if err := h.cache.Get(ctx, fmt.Sprintf("store:item:%s", id), &item); err != nil {
		if err == redis.Nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	_ = h.cache.Set(ctx, cacheKey, item)
	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, item)
}

func (h *Handler) CreateItem(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		Name  string `json:"name"  binding:"required"`
		Value string `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item := Item{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Name:      req.Name,
		Value:     req.Value,
		CreatedAt: time.Now(),
	}

	if err := h.cache.SetPermanent(ctx, fmt.Sprintf("store:item:%s", item.ID), item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = h.cache.Delete(ctx, "cache:items:all")
	c.JSON(http.StatusCreated, item)
}

func (h *Handler) DeleteItem(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if err := h.cache.Delete(ctx,
		fmt.Sprintf("store:item:%s", id),
		fmt.Sprintf("cache:items:%s", id),
		"cache:items:all",
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
