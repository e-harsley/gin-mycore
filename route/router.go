package coreRoute

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/e-harsley/gin-mycore/service"
	"github.com/e-harsley/gin-mycore/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetServiceField(i interface{}, fieldName string) (service.IBaseService, bool) {
	val := reflect.ValueOf(i)
	if val.Kind() == reflect.Struct {
		fieldVal := val.FieldByName(fieldName)

		if !fieldVal.IsValid() {
			return nil, true
		}

		return fieldVal.Interface().(service.IBaseService), false
	}

	return nil, true
}

func WrapRouter(router *gin.Engine, urlPath string, controller interface{}, limitter func(db *gorm.DB, ctx *gin.Context) *gorm.DB, permit func(ctx *gin.Context), preload func(db *gorm.DB, ctx *gin.Context) *gorm.DB) {

	router.GET(urlPath, func(ctx *gin.Context) {

		serviceMethod, is_error := GetServiceField(controller, "Service")
		if is_error {
			utils.RaiseException(ctx, "Controller has no service")
			return
		}
		limit := serviceMethod.LimitBy(ctx)
		limitter := limitter(limit, ctx)
		preload := preload(limitter, ctx)
		permit(ctx)
		serializers, is_error := utils.GetMapField(controller, "Serializer")
		if is_error {
			utils.RaiseException(ctx, "Controller has no serializers")
			return
		}
		results := reflect.ValueOf(controller).MethodByName("Find").Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(preload)})
		result := results[0].Interface()

		if bindTo, ok := serializers["Response"]; ok {
			service.ControllerResponse(ctx, result, bindTo)
			return
		}

	})
	router.POST(urlPath, func(ctx *gin.Context) {
		serviceMethod, is_error := GetServiceField(controller, "Service")
		if is_error {
			utils.RaiseException(ctx, "Controller has no service")
			return
		}
		limit := serviceMethod.LimitBy(ctx)

		permit(ctx)

		serializers, is_error := utils.GetMapField(controller, "Serializer")
		if is_error {
			utils.RaiseException(ctx, "Controller has no serializers")
			return
		}
		var data map[string]interface{}
		var err error
		if serializer, ok := serializers["Request"]; ok {
			data, err = utils.BindToMap(ctx, serializer)
			if err != nil {
				fmt.Println("hi")
				ctx.JSON(http.StatusUnprocessableEntity, utils.NewValidatorError(err))
				return
			}
		} else {
			utils.RaiseException(ctx, "Serializer has no schema request")
			return
		}
		handler := reflect.ValueOf(controller)
		_, ok := reflect.TypeOf(controller).MethodByName("Save")
		if ok {
			results := handler.MethodByName("Save").Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(data), reflect.ValueOf(limit)})
			result := results[0].Interface()

			if bindTo, ok := serializers["Response"]; ok {
				fmt.Println(result)
				service.ControllerResponse(ctx, result, bindTo)
				return
			}

		} else {
			utils.RaiseException(ctx, "Methon Save not found on controller")
			return
		}
	})

	routerGroup := router.Group(urlPath + "/:id")
	{
		routerGroup.GET("", func(ctx *gin.Context) {
			obj_id := ctx.Param("id")
			serviceMethod, is_error := GetServiceField(controller, "Service")
			if is_error {
				utils.RaiseException(ctx, "Controller has no service")
				return
			}
			limit := serviceMethod.LimitBy(ctx)
			limitter := limitter(limit, ctx)
			preload := preload(limitter, ctx)
			permit(ctx)
			serializers, is_error := utils.GetMapField(controller, "Serializer")
			if is_error {
				utils.RaiseException(ctx, "Controller has no serializers")
				return
			}

			results := reflect.ValueOf(controller).MethodByName("FindByID").Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(obj_id), reflect.ValueOf(preload)})

			result := results[0].Interface()

			if bindTo, ok := serializers["Response"]; ok {
				service.ControllerResponse(ctx, result, bindTo)
				return
			}
		})
		routerGroup.PUT("", func(ctx *gin.Context) {
			serializers, is_error := utils.GetMapField(controller, "Serializer")
			if is_error {
				utils.RaiseException(ctx, "Controller has no serializers")
				return
			}
			var data map[string]interface{}
			var err error
			if serializer, ok := serializers["Request"]; ok {
				data, err = utils.BindToMap(ctx, serializer)
				if err != nil {
					utils.RaiseException(ctx, err.Error())
					return
				}
			} else {
				utils.RaiseException(ctx, "Serializer has not schema request")
				return
			}
			obj_id := ctx.Param("id")
			results := reflect.ValueOf(controller).MethodByName("Update").Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(obj_id), reflect.ValueOf(data)})
			result := results[0].Interface()

			if bindTo, ok := serializers["Response"]; ok {
				fmt.Println(result)
				service.ControllerResponse(ctx, result, bindTo)
				return
			}
		})

		controllerGroup := routerGroup.Group("/:controllerName")
		{
			controllerGroup.Use(func(ctx *gin.Context) {
				controllerName := ctx.Param("controllerName")
				method := reflect.ValueOf(controller).MethodByName(strings.Title(controllerName))
				if method.IsValid() {
					serializers, is_error := utils.GetMapField(controller, "Serializer")
					if is_error {
						utils.RaiseException(ctx, "Controller has no serializers")
						return
					}
					var data map[string]interface{}
					var err error
					if serializer, ok := serializers[controllerName]; ok {
						data, err = utils.BindToMap(ctx, serializer)
						if err != nil {
							utils.RaiseException(ctx, err.Error())
							return
						}
					} else {
						utils.RaiseException(ctx, "Serializer has not schema request")
						return
					}
					handler := reflect.ValueOf(controller)
					_, ok := reflect.TypeOf(controller).MethodByName(strings.Title(controllerName))
					if ok {
						results := handler.MethodByName("Save").Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(data)})
						result := results[0].Interface()

						if bindTo, ok := serializers[controllerName+"Response"]; ok {
							fmt.Println(result)
							service.ControllerResponse(ctx, result, bindTo)
							return
						}

					} else {
						utils.RaiseException(ctx, "Methon Save not found on controller")
						return
					}
				} else {
					ctx.JSON(http.StatusNotFound, gin.H{
						"error": "Resource not found",
					})
					return
				}
			})
		}
	}
}
