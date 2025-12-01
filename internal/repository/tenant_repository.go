package repository

import (
	"context"
	"go-gin-clean/internal/entity"

	"gorm.io/gorm"
)

type TenantRepository struct {
	db       *gorm.DB
	baseRepo BaseRepository[entity.Tenant]
}

func NewTenantRepository(db *gorm.DB) *TenantRepository {
	baseRepo := NewBaseRepository[entity.Tenant](db)
	return &TenantRepository{
		db:       db,
		baseRepo: *baseRepo,
	}
}

func (r *TenantRepository) FindByID(ctx context.Context, id int64) (*entity.Tenant, error) {
	return r.baseRepo.FindByID(ctx, id)
}

func (r *TenantRepository) FindBySlug(ctx context.Context, slug string) (*entity.Tenant, error) {
	return r.baseRepo.FindFirst(ctx, "slug = ? AND deleted_at IS NULL", slug)
}

func (r *TenantRepository) ExistsBySlug(ctx context.Context, slug string) bool {
	exists, _ := r.baseRepo.WhereExisting(ctx, "slug = ? AND deleted_at IS NULL", slug)
	return exists
}

func (r *TenantRepository) Create(ctx context.Context, tenant *entity.Tenant) (*entity.Tenant, error) {
	if err := r.db.WithContext(ctx).Create(tenant).Error; err != nil {
		return nil, err
	}
	return tenant, nil
}

func (r *TenantRepository) Update(ctx context.Context, tenant *entity.Tenant) error {
	return r.db.WithContext(ctx).Save(tenant).Error
}

func (r *TenantRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&entity.Tenant{}, id).Error
}
