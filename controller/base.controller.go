package controller

import (
	"github.com/e-harsley/gin-mycore/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BaseController struct {
	Service    service.IBaseService
	Serializer map[string]interface{}
}

type IBaseController interface {
	Save(ctx *gin.Context, data map[string]interface{})
}

func (cls BaseController) Save(ctx *gin.Context, data map[string]interface{}, db *gorm.DB) service.ServiceResponse {
	return cls.Service.Save(ctx, data, db)

}

func (cls BaseController) Find(ctx *gin.Context, db *gorm.DB) service.ServiceResponse {
	return cls.Service.GetAll(ctx, db)

}

func (cls BaseController) FindByID(ctx *gin.Context, id string, db *gorm.DB) service.ServiceResponse {
	return cls.Service.GetById(ctx, id, db)
}

func (cls BaseController) Update(ctx *gin.Context, id string, data map[string]interface{}, db *gorm.DB) service.ServiceResponse {
	return cls.Service.Update(ctx, id, data, db)
}
