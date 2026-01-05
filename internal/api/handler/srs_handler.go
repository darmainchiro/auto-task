package handler

import (
	"srs-automation/internal/core/service"

	"github.com/gofiber/fiber/v2"
)

type SRSHandler struct {
	service *service.SRSService
}

func NewSRSHandler(service *service.SRSService) *SRSHandler {
	return &SRSHandler{service: service}
}

type GenerateSRSRequest struct {
	DocumentID uint   `json:"document_id"`
	Title      string `json:"title"`
}

func (h *SRSHandler) Generate(c *fiber.Ctx) error {
	var req GenerateSRSRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	srs, err := h.service.GenerateSRS(req.DocumentID, req.Title)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "SRS generated successfully",
		"data":    srs,
	})
}

func (h *SRSHandler) GetByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid SRS ID",
		})
	}

	srs, err := h.service.GetSRS(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "SRS not found",
		})
	}

	return c.JSON(fiber.Map{
		"data": srs,
	})
}

func (h *SRSHandler) GetByDocument(c *fiber.Ctx) error {
	docID, err := c.ParamsInt("documentId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid document ID",
		})
	}

	srsList, err := h.service.GetSRSByDocument(uint(docID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": srsList,
	})
}

func (h *SRSHandler) GetAll(c *fiber.Ctx) error {
	srsList, err := h.service.GetAllSRS()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": srsList,
	})
}

type UpdateSRSRequest struct {
	Content string `json:"content"`
	Status  string `json:"status"`
}

func (h *SRSHandler) Update(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid SRS ID",
		})
	}

	var req UpdateSRSRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.service.UpdateSRS(uint(id), req.Content, req.Status); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "SRS updated successfully",
	})
}

func (h *SRSHandler) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid SRS ID",
		})
	}

	if err := h.service.DeleteSRS(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "SRS deleted successfully",
	})
}
