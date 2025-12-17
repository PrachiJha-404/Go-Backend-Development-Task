package handler

import (
	"net/http"
	"strconv"
	"time"
	"user-api/internal/models"
	"user-api/internal/service"
	"user-api/internal/validator"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type UserHandler struct {
	service   service.UserService
	logger    *zap.Logger
	validator *validator.Validator
}

func NewUserHandler(service service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		service:   service,
		logger:    logger,
		validator: validator.NewValidator(),
	}
}

func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	dbUsers, err := h.service.ListUsers(c.Context())
	if err != nil {
		h.logger.Error("failed to list users", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch users"})
	}
	return c.Status(http.StatusOK).JSON(dbUsers)
}

func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}
	dbUser, err := h.service.GetUser(c.Context(), int32(id))
	if err != nil {
		h.logger.Error("failed to get user", zap.Error(err))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}
	return c.Status(http.StatusOK).JSON(dbUser)
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Validate the request
	if err := h.validator.ValidateStruct(req); err != nil {
		h.logger.Warn("validation failed for create user", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	dob, err := time.Parse("2006-01-02", req.DOB)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid date format (Use YYYY-MM-DD)"})
	}
	dbUser, err := h.service.CreateUser(c.Context(), req.Name, dob)
	if err != nil {
		h.logger.Error("failed to create user", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create user"})
	}
	return c.Status(http.StatusOK).JSON(dbUser)
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}
	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Validate the request
	if err := h.validator.ValidateStruct(req); err != nil {
		h.logger.Warn("validation failed for update user", zap.Error(err))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	dob, err := time.Parse("2006-01-02", req.DOB)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid data format (use YYYY-MM-DD)"})
	}
	user, err := h.service.UpdateUser(c.Context(), int32(id), req.Name, dob)
	if err != nil {
		h.logger.Error("failed to update user", zap.Error(err))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update user"})
	}
	return c.Status(http.StatusOK).JSON(user)
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}
	err = h.service.DeleteUser(c.Context(), int32(id))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete user"})
	}
	return c.Status(http.StatusOK).SendStatus(http.StatusNoContent)
}
