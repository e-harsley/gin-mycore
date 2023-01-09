package service

import (
	"fmt"
	"net/http"

	constant "github.com/e-harsley/gin-mycore/constants"
	"github.com/e-harsley/gin-mycore/utils"
	"gorm.io/gorm"

	"github.com/e-harsley/gin-mycore/repository"
	"github.com/gin-gonic/gin"
)

type BaseService[E any] struct {
	repository.IBaseRepository[E]
}

func GenerateService[E any](db *gorm.DB) *BaseService[E] {
	return &BaseService[E]{
		repository.NewRepository[E](db),
	}
}

type ServiceResponse struct {
	Status  string
	Message string
	Data    interface{}
}

type IBaseService interface {
	Save(ctx *gin.Context, data map[string]interface{}, db *gorm.DB) ServiceResponse
	GetById(ctx *gin.Context, obj_id string, db *gorm.DB) ServiceResponse
	Update(ctx *gin.Context, obj_id string, data map[string]interface{}, db *gorm.DB) ServiceResponse
	GetAll(ctx *gin.Context, db *gorm.DB) ServiceResponse
	LimitBy(ctx *gin.Context) *gorm.DB
}

func (cls *BaseService[E]) LimitBy(ctx *gin.Context) *gorm.DB {
	return cls.Limitter(ctx)
}

func (cls *BaseService[E]) Save(ctx *gin.Context, data map[string]interface{}, db *gorm.DB) ServiceResponse {
	service, err := cls.Create(ctx, data, db)
	fmt.Println("service create", service)

	if err != nil {
		return ServiceResponse{
			Message: err.Error(),
			Status:  constant.ERROR_STATUS,
		}
	}

	return ServiceResponse{
		Message: "Created Successfully",
		Status:  constant.SUCCESS_STATUS,
		Data:    service,
	}

}

func (cls *BaseService[E]) Update(ctx *gin.Context, obj_id string, data map[string]interface{}, db *gorm.DB) ServiceResponse {

	id := utils.StringToUint(obj_id)

	service, err := cls.UpdateRepo(ctx, id, data, db)

	if err != nil {
		return ServiceResponse{
			Message: err.Error(),
			Status:  constant.ERROR_STATUS,
		}
	}

	return ServiceResponse{
		Message: "Created Successfully",
		Status:  constant.SUCCESS_STATUS,
		Data:    service,
	}

}

func (cls *BaseService[E]) GetAll(ctx *gin.Context, db *gorm.DB) ServiceResponse {
	service, err := cls.Find(ctx, db)

	if err != nil {
		return ServiceResponse{
			Message: err.Error(),
			Status:  constant.ERROR_STATUS,
		}
	}
	return ServiceResponse{
		Message: "List of Items",
		Status:  constant.SUCCESS_STATUS,
		Data:    service,
	}
}

func (cls *BaseService[E]) GetById(ctx *gin.Context, obj_id string, db *gorm.DB) ServiceResponse {

	id := utils.StringToUint(obj_id)

	service, err := cls.FindByID(ctx, id, db)

	if err != nil {
		return ServiceResponse{
			Message: err.Error(),
			Status:  constant.ERROR_STATUS,
		}
	}
	return ServiceResponse{
		Message: "List of Items",
		Status:  constant.SUCCESS_STATUS,
		Data:    service,
	}
}

func ControllerResponse(ctx *gin.Context, data interface{}, bindTo interface{}) {
	dataMap := utils.ToJsonMap(data)
	if dataMap["Status"] == constant.ERROR_STATUS {
		utils.RaiseException(ctx, dataMap["Message"].(string))
		return
	}
	dataOp := utils.BindDataOperationStruct(dataMap["Data"], bindTo)
	fmt.Println(dataOp)
	if dataOp.IsError {
		utils.RaiseException(ctx, dataOp.Message)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status": constant.SUCCESS_STATUS,
		"data":   dataOp.Data,
	})

}
