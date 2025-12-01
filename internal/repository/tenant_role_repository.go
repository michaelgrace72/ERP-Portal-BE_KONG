package repository

import (
	"context"
	"go-gin-clean/internal/entity"

	"gorm.io/gorm"
)

type TenantRoleRepository struct {
	db       *gorm.DB
	baseRepo BaseRepository[entity.TenantRole]
}

func NewTenantRoleRepository(db *gorm.DB) *TenantRoleRepository {
	baseRepo := NewBaseRepository[entity.TenantRole](db)
	return &TenantRoleRepository{
		db:       db,
		baseRepo: *baseRepo,
	}
}

func (r *TenantRoleRepository) FindByID(ctx context.Context, id int64) (*entity.TenantRole, error) {
	return r.baseRepo.FindByID(ctx, id)
}

func (r *TenantRoleRepository) FindByTenantAndName(ctx context.Context, tenantID int64, name string) (*entity.TenantRole, error) {
	return r.baseRepo.FindFirst(ctx, "tenant_id = ? AND name = ? AND deleted_at IS NULL", tenantID, name)
}

func (r *TenantRoleRepository) Create(ctx context.Context, role *entity.TenantRole) (*entity.TenantRole, error) {
	if err := r.db.WithContext(ctx).Create(role).Error; err != nil {
		return nil, err
	}
	return role, nil
}

func (r *TenantRoleRepository) FindAllByTenantID(ctx context.Context, tenantID int64) ([]*entity.TenantRole, error) {
	var roles []*entity.TenantRole
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}
