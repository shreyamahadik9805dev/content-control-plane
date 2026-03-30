package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/shreyafeo/content-control-plane/internal/service"
)

type Handler struct {
	podcasts *service.Podcasts
}

func New(podcasts *service.Podcasts) *Handler {
	return &Handler{podcasts: podcasts}
}

// Register mounts all HTTP routes on the given engine.
func (h *Handler) Register(r *gin.Engine) {
	r.GET("/health", h.health)

	r.POST("/sync/podcasts", h.syncPodcasts)
	r.GET("/podcasts", h.listPodcasts)
	r.GET("/podcasts/:id", h.getPodcast)
	r.POST("/podcasts/:id/pin", h.pinPodcast)
	r.GET("/audit-logs", h.auditLogs)
}

func (h *Handler) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) syncPodcasts(c *gin.Context) {
	q := c.Query("query")
	run, err := h.podcasts.Sync(c.Request.Context(), q)
	if err != nil {
		if errors.Is(err, service.ErrBadQuery) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, run)
}

func (h *Handler) listPodcasts(c *gin.Context) {
	pods, err := h.podcasts.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pods)
}

func (h *Handler) getPodcast(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	p, err := h.podcasts.Get(c.Request.Context(), id)
	if err != nil {
		if service.ErrIsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "podcast not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handler) pinPodcast(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var body service.PinRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "expected JSON body: {\"pinned\": true|false}"})
		return
	}
	if err := h.podcasts.SetPinned(c.Request.Context(), id, body.Pinned); err != nil {
		if service.ErrIsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "podcast not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	p, err := h.podcasts.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handler) auditLogs(c *gin.Context) {
	limit := 100
	if q := c.Query("limit"); q != "" {
		if n, err := strconv.Atoi(q); err == nil {
			limit = n
		}
	}
	logs, err := h.podcasts.AuditLogs(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}
