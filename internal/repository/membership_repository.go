package repository

import (
	"context"
	"go-gin-clean/internal/entity"

	"gorm.io/gorm"
)

type MembershipRepository struct {
	db       *gorm.DB
	baseRepo BaseRepository[entity.Membership]
}

func NewMembershipRepository(db *gorm.DB) *MembershipRepository {
	baseRepo := NewBaseRepository[entity.Membership](db)
	return &MembershipRepository{
		db:       db,
		baseRepo: *baseRepo,
	}
}

func (r *MembershipRepository) Create(ctx context.Context, membership *entity.Membership) (*entity.Membership, error) {
	if err := r.db.WithContext(ctx).Create(membership).Error; err != nil {
		return nil, err
	}
	return membership, nil
}

func (r *MembershipRepository) FindByID(ctx context.Context, id int64) (*entity.Membership, error) {
	return r.baseRepo.FindByID(ctx, id)
}

func (r *MembershipRepository) FindByUserAndTenant(ctx context.Context, userID, tenantID int64) (*entity.Membership, error) {
	return r.baseRepo.FindFirst(ctx, "user_id = ? AND tenant_id = ? AND deleted_at IS NULL", userID, tenantID)
}

// FindByUserID returns all memberships for a user (without preloading relations)
func (r *MembershipRepository) FindByUserID(ctx context.Context, userID int64) ([]entity.Membership, error) {
	var memberships []entity.Membership
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Find(&memberships).Error; err != nil {
		return nil, err
	}
	return memberships, nil
}

func (r *MembershipRepository) FindAllByUserID(ctx context.Context, userID int64) ([]*entity.Membership, error) {
	var memberships []*entity.Membership
	if err := r.db.WithContext(ctx).
		Preload("Tenant").
		Preload("Role").
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Find(&memberships).Error; err != nil {
		return nil, err
	}
	return memberships, nil
}

func (r *MembershipRepository) FindAllByTenantID(ctx context.Context, tenantID int64) ([]*entity.Membership, error) {
	var memberships []*entity.Membership
	if err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Role").
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Find(&memberships).Error; err != nil {
		return nil, err
	}
	return memberships, nil
}

func (r *MembershipRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&entity.Membership{}, id).Error
}
