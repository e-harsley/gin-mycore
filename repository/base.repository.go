package repository

import (
	"context"
	"fmt"

	"github.com/e-harsley/go-mycore/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BaseRepository[E any] struct {
	Db *gorm.DB
}

type IBaseRepository[E any] interface {
	Create(ctx context.Context, data map[string]interface{}, db *gorm.DB) (E, error)
	FindByID(ctx context.Context, id uint, db *gorm.DB) (E, error)
	UpdateRepo(ctx context.Context, id uint, data map[string]interface{}, db *gorm.DB) (E, error)
	Find(ctx context.Context, db *gorm.DB) ([]E, error)
	FindBy(ctx context.Context, specification Specification) ([]E, error)
	FindOne(ctx context.Context, specification Specification, preload ...string) (E, error)
	Limitter(ctx *gin.Context) *gorm.DB
}

func NewRepository[E any](db *gorm.DB) *BaseRepository[E] {
	return &BaseRepository[E]{
		Db: db,
	}
}

func (r BaseRepository[E]) Limitter(ctx *gin.Context) *gorm.DB {
	return r.Db.WithContext(ctx)
}

func (r *BaseRepository[E]) Create(ctx context.Context, data map[string]interface{}, db *gorm.DB) (E, error) {
	var model E

	err := db.Transaction(func(tx *gorm.DB) error {
		// Perform the create operation inside the transaction
		err := tx.Model(&model).Create(data).Error
		if err != nil {
			return err
		}

		// Get the last inserted record
		err = tx.Last(&model).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return *new(E), err
	}

	fmt.Println("created value:", model)
	return model, nil
}

func (r *BaseRepository[E]) FindByID(ctx context.Context, id uint, db *gorm.DB) (E, error) {
	var entity E

	err := db.First(&entity, id).Error
	if err != nil {
		return *new(E), err
	}

	return entity, nil
}

func (r *BaseRepository[E]) UpdateRepo(ctx context.Context, id uint, data map[string]interface{}, db *gorm.DB) (E, error) {
	fmt.Println("id", id)
	single, err := r.FindByID(ctx, id, db)
	fmt.Println(single, data)

	if err != nil {
		utils.PanicException(err, "could not find item by id")
	}

	err = db.Model(&single).Updates(data).Error

	if err != nil {
		return *new(E), err
	}

	return single, nil

}

func (r *BaseRepository[E]) Find(ctx context.Context, db *gorm.DB) ([]E, error) {

	var err error

	var entity []E

	err = db.Find(&entity).Error

	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (r *BaseRepository[E]) FindBy(ctx context.Context, specification Specification) ([]E, error) {

	var entities []E

	err := r.Db.WithContext(ctx).Where(specification.GetQuery(), specification.GetValues()...).Find(&entities).Error
	if err != nil {
		return nil, err
	}

	return entities, nil
}

func (r *BaseRepository[E]) FindOne(ctx context.Context, specification Specification, preload ...string) (E, error) {

	var err error

	var entity E

	err = r.Db.WithContext(ctx).Where(specification.GetQuery(), specification.GetValues()...).First(&entity).Error

	if err != nil {
		return *new(E), err
	}

	return entity, nil
}
