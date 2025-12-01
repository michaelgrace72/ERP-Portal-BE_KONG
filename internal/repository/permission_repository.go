package repository

import (
	"context"

	"go-gin-clean/internal/entity"

	"gorm.io/gorm"
)

type PermissionRepository struct {
	*BaseRepository[entity.Permission]
}

func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{
		BaseRepository: NewBaseRepository[entity.Permission](db),
	}
}

// FindByRoleID finds all permissions for a specific role
func (r *PermissionRepository) FindByRoleID(ctx context.Context, roleID int64) ([]entity.Permission, error) {
	var permissions []entity.Permission
	if err := r.db.WithContext(ctx).Where("role_id = ?", roleID).Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}
