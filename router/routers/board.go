package routers

import (

	"strconv"

	"basicproject_go/global/log"


	"basicproject_go/controllers/rest"


	"basicproject_go/models"
	"github.com/gofiber/fiber/v2"
)

// SetupBoardRoutes sets up routes for board domain
func SetupBoardRoutes(group fiber.Router) {

	group.Get("/board", func(c *fiber.Ctx) error {
		page_, _ := strconv.Atoi(c.Query("page"))
		pagesize_, _ := strconv.Atoi(c.Query("pagesize"))
		var controller rest.BoardController
		controller.Init(c)
		controller.Index(page_, pagesize_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Get("/board/:id", func(c *fiber.Ctx) error {
		id_, _ := strconv.ParseInt(c.Params("id"), 10, 64)
		var controller rest.BoardController
		controller.Init(c)
		controller.Read(id_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Post("/board", func(c *fiber.Ctx) error {
		item_ := &models.Board{}
		err := c.BodyParser(item_)
		if err != nil {
		    log.Error().Msg(err.Error())
		}
		var controller rest.BoardController
		controller.Init(c)
		controller.Insert(item_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Post("/board/batch", func(c *fiber.Ctx) error {
		var items_ *[]models.Board
		items__ref := &items_
		err := c.BodyParser(items__ref)
		if err != nil {
		    log.Error().Msg(err.Error())
		}
		var controller rest.BoardController
		controller.Init(c)
		controller.Insertbatch(items_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Post("/board/count", func(c *fiber.Ctx) error {

		var controller rest.BoardController
		controller.Init(c)
		controller.Count()
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Put("/board", func(c *fiber.Ctx) error {
		item_ := &models.Board{}
		err := c.BodyParser(item_)
		if err != nil {
		    log.Error().Msg(err.Error())
		}
		var controller rest.BoardController
		controller.Init(c)
		controller.Update(item_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Delete("/board", func(c *fiber.Ctx) error {
		item_ := &models.Board{}
		err := c.BodyParser(item_)
		if err != nil {
		    log.Error().Msg(err.Error())
		}
		var controller rest.BoardController
		controller.Init(c)
		controller.Delete(item_)
		controller.Close()
		return c.JSON(controller.Result)
	})

	group.Delete("/board/batch", func(c *fiber.Ctx) error {
		item_ := &[]models.Board{}
		err := c.BodyParser(item_)
		if err != nil {
		    log.Error().Msg(err.Error())
		}
		var controller rest.BoardController
		controller.Init(c)
		controller.Deletebatch(item_)
		controller.Close()
		return c.JSON(controller.Result)
	})

}