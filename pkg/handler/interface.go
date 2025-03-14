package handler

import (
	"github.com/gofiber/fiber/v2"
)

type IHandler interface {
	Path() string
	AddRoutes(rg fiber.Router)
}

type Opt func(IHandler) error

type Constructor func(opts ...Opt) (IHandler, error)

type IFactory interface {
	Register(name string, h Constructor) error
	Build(name string, opts ...Opt) (IHandler, error)
}
