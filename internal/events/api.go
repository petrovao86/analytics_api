package events

import (
	"errors"
	"net/http"

	"example.com/analytics_api/pkg/config"
	"example.com/analytics_api/pkg/handler"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

const (
	Handler = "events"
)

type ApiConfig struct {
	Path    string
	Storage string
}

func NewApiConfig() *ApiConfig {
	return &ApiConfig{
		Path:    "/events",
		Storage: "clickhouse://127.0.0.1:9000/default?sslmode=disable",
	}
}

func (c *ApiConfig) Read(cr config.IReader) error {
	var err error
	newC := *c

	path, pathErr := config.Get[string](cr, "path")
	if pathErr != nil && !errors.Is(pathErr, config.ErrNotFound) {
		err = errors.Join(err, pathErr)
	} else if pathErr == nil {
		newC.Path = path
	}

	storage, storageErr := config.Get[string](cr, "storage")
	if storageErr != nil && !errors.Is(pathErr, config.ErrNotFound) {
		err = errors.Join(err, storageErr)
	} else if storageErr == nil {
		newC.Storage = storage
	}

	if err != nil {
		return err
	}

	*c = newC
	return nil
}

type apiHandler struct {
	c    *ApiConfig
	repo IRepository
}

func NewHandler(opts ...handler.Opt) (handler.IHandler, error) {
	h := &apiHandler{
		c: NewApiConfig(),
	}
	for _, opt := range opts {
		if err := opt(h); err != nil {
			return nil, err
		}
	}
	repo, err := BuildRepo(h.c.Storage)
	if err != nil {
		return nil, err
	}
	h.repo = repo

	return h, nil
}

func (h *apiHandler) Configure(cr config.IReader) error {
	return h.c.Read(cr)
}

func (h *apiHandler) Path() string {
	return h.c.Path
}

func (h *apiHandler) AddRoutes(rg fiber.Router) {
	rg.Post("/", h.handler)
}

func (h *apiHandler) handler(ctx *fiber.Ctx) error {
	e := new(ApiEvent)
	if err := ctx.BodyParser(e); err != nil {
		logrus.WithError(err).Error("request parse error")
		return err
	}
	if err := validator.New().Struct(e); err != nil {
		logrus.WithError(err).Error("request validate error")
		return &fiber.Error{Code: http.StatusBadRequest, Message: err.Error()}
	}
	if err := h.repo.Insert(e); err != nil {
		logrus.WithError(err).Error("request insert error")
		return &fiber.Error{Code: http.StatusInternalServerError, Message: err.Error()}
	}
	return nil
}
