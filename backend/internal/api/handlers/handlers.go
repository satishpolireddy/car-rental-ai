package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/satishpolireddy/car-rental-ai/internal/models"
	"github.com/satishpolireddy/car-rental-ai/internal/repository"
	"github.com/satishpolireddy/car-rental-ai/internal/services"
)

type Handler struct {
	carRepo        *repository.CarRepository
	bookingService *services.BookingService
	aiService      *services.AIService
	customerRepo   *repository.CustomerRepository
	bookingRepo    *repository.BookingRepository
}

func NewHandler(
	carRepo *repository.CarRepository,
	bookingService *services.BookingService,
	aiService *services.AIService,
	customerRepo *repository.CustomerRepository,
	bookingRepo *repository.BookingRepository,
) *Handler {
	return &Handler{
		carRepo:        carRepo,
		bookingService: bookingService,
		aiService:      aiService,
		customerRepo:   customerRepo,
		bookingRepo:    bookingRepo,
	}
}

// SearchCars handles GET /api/v1/cars/search
func (h *Handler) SearchCars(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cars, err := h.carRepo.Search(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cars": cars, "total": len(cars)})
}

// GetCar handles GET /api/v1/cars/:id
func (h *Handler) GetCar(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid car id"})
		return
	}
	car, err := h.carRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "car not found"})
		return
	}
	c.JSON(http.StatusOK, car)
}

// ListLocations handles GET /api/v1/locations
func (h *Handler) ListLocations(c *gin.Context) {
	locations, err := h.carRepo.ListLocations(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load locations"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"locations": locations})
}

// CreateBooking handles POST /api/v1/bookings
func (h *Handler) CreateBooking(c *gin.Context) {
	var req models.BookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	booking, err := h.bookingService.CreateBooking(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, booking)
}

// GetBooking handles GET /api/v1/bookings/:id
func (h *Handler) GetBooking(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}
	booking, err := h.bookingService.GetBooking(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		return
	}
	c.JSON(http.StatusOK, booking)
}

// CancelBooking handles DELETE /api/v1/bookings/:id
func (h *Handler) CancelBooking(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}
	if err := h.bookingService.CancelBooking(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "booking cancelled"})
}

// GetCustomerBookings handles GET /api/v1/customers/:id/bookings
func (h *Handler) GetCustomerBookings(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	bookings, err := h.bookingService.GetCustomerBookings(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load bookings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bookings": bookings, "total": len(bookings)})
}

// AIRecommend handles POST /api/v1/ai/recommend
func (h *Handler) AIRecommend(c *gin.Context) {
	var req models.AIQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get available cars to provide context to the AI
	cars, err := h.carRepo.Search(c.Request.Context(), models.SearchRequest{
		Location: c.Query("location"),
	})
	if err != nil || len(cars) == 0 {
		cars = []models.Car{} // graceful fallback
	}

	resp, err := h.aiService.Recommend(c.Request.Context(), req, cars)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI service unavailable"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetCustomerDashboard handles GET /api/v1/customers/:id/dashboard
// Returns the full home page payload: profile, stats, current bookings, past bookings.
func (h *Handler) GetCustomerDashboard(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	ctx := c.Request.Context()

	customer, err := h.customerRepo.GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
		return
	}

	stats, err := h.bookingRepo.GetCustomerStats(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load stats"})
		return
	}

	current, err := h.bookingRepo.GetCurrentByCustomer(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load current bookings"})
		return
	}

	past, err := h.bookingRepo.GetPastByCustomer(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load past bookings"})
		return
	}

	c.JSON(http.StatusOK, models.CustomerDashboard{
		Customer:        customer,
		Stats:           stats,
		CurrentBookings: current,
		PastBookings:    past,
	})
}

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func parseUintParam(c *gin.Context, param string) (uint, error) {
	v, err := strconv.ParseUint(c.Param(param), 10, 64)
	return uint(v), err
}
