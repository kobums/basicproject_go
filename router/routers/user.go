package routers

import (

	"strconv"

	"basicproject_go/global/log"


	"basicproject_go/controllers/rest"


	"basicproject_go/models"
	"github.com/gofiber/fiber/v2"
)

// SetupUserRoutes sets up routes for user domain
func SetupUserRoutes(group fiber.Router) {

	group.Get("/user", func(c *fiber.Ctx) error {
		page_, _ := strconv.Atoi(c.Query("page"))
		pagesize_, _ := strconv.Atoi(c.Query("pagesize"))
		var controller rest.UserController
		controller.Init(c)
		controller.Index(page_, pagesize_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Get("/user/:id", func(c *fiber.Ctx) error {
		id_, _ := strconv.ParseInt(c.Params("id"), 10, 64)
		var controller rest.UserController
		controller.Init(c)
		controller.Read(id_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Post("/user", func(c *fiber.Ctx) error {
		item_ := &models.UserUpdate{}
		err := c.BodyParser(item_)
		if err != nil {
		    log.Error().Msg(err.Error())
		}
		var controller rest.UserController
		controller.Init(c)
		controller.Insert(item_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Post("/user/batch", func(c *fiber.Ctx) error {
		var items_ *[]models.UserUpdate
		items__ref := &items_
		err := c.BodyParser(items__ref)
		if err != nil {
		    log.Error().Msg(err.Error())
		}
		var controller rest.UserController
		controller.Init(c)
		controller.Insertbatch(items_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Post("/user/count", func(c *fiber.Ctx) error {

		var controller rest.UserController
		controller.Init(c)
		controller.Count()
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Put("/user", func(c *fiber.Ctx) error {
		item_ := &models.UserUpdate{}
		err := c.BodyParser(item_)
		if err != nil {
		    log.Error().Msg(err.Error())
		}
		var controller rest.UserController
		controller.Init(c)
		controller.Update(item_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Delete("/user", func(c *fiber.Ctx) error {
		item_ := &models.User{}
		err := c.BodyParser(item_)
		if err != nil {
		    log.Error().Msg(err.Error())
		}
		var controller rest.UserController
		controller.Init(c)
		controller.Delete(item_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Delete("/user/batch", func(c *fiber.Ctx) error {
		item_ := &[]models.User{}
		err := c.BodyParser(item_)
		if err != nil {
		    log.Error().Msg(err.Error())
		}
		var controller rest.UserController
		controller.Init(c)
		controller.Deletebatch(item_)
		controller.Close()
		return c.JSON(controller.Result)
	})

}