// Developer: zeelrupapara@gmail.com
// Description: Token management for GreenLync boilerplate
package v1

import (
	"time"
	model "greenlync-api-gateway/model/common/v1"

	"github.com/gofiber/fiber/v2"
)

// @Id				GetAllTokensHistroy
// @Description	Get All Tokens Histroy
// @Tags			System
// @Accept			json
// @Produce		json
// @Success		200	{array}		model.Token
// @Failure		500	{object}	http.HttpResponse
// @Security		BearerAuth
// @Router			/api/v1/system/tokens/history [get]
func (s *HttpServer) GetAllTokensHistroy(c *fiber.Ctx) error {
	tokens := &[]*model.Token{}

	time := time.Now().Add(-time.Hour).UnixNano()
	err := s.DB.Where("created_at < ?", time).Find(tokens).Error
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, tokens)
}

// @Id				DeleteAllTokensHistroy
// @Description	Delete All Token Histroy
// @Tags			System
// @Accept			json
// @Produce		json
// @Success		204
// @Failure		500	{object}	http.HttpResponse
// @Security		BearerAuth
// @Router			/api/v1/system/tokens/history [DELETE]
func (s *HttpServer) DeleteAllTokensHistroy(c *fiber.Ctx) error {
	time := time.Now().Add(-time.Hour).UnixNano()
	err := s.DB.Where("created_at < ?", time).Delete(&model.Token{}).Error
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseNoContent(c)
}
