package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"trackpocket/internal/model/domain"
	"trackpocket/internal/model/web"
	"trackpocket/internal/service"
)

func newTestCategoryService() (service.CategoryService, *CategoryRepositoryMock) {
	categoryRepoMock := new(CategoryRepositoryMock)
	svc := service.NewCategoryService(categoryRepoMock)
	return svc, categoryRepoMock
}

// ============ CREATE ============

func TestCategoryCreate_Success(t *testing.T) {
	svc, categoryRepoMock := newTestCategoryService()

	userID := uuid.New()
	req := web.CreateCategoryRequest{
		Name: "Freelance",
		Type: "INCOME",
	}

	expected := domain.Category{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      req.Name,
		Type:      req.Type,
		IsDefault: false,
		IsActive:  true,
	}

	categoryRepoMock.On("Create", mock.Anything, mock.Anything).Return(expected, nil)

	result, err := svc.Create(context.Background(), userID, req)

	assert.NoError(t, err)
	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, "Freelance", result.Name)
	assert.False(t, result.IsDefault)

	categoryRepoMock.AssertExpectations(t)
}

func TestCategoryCreate_Failed_RepositoryError(t *testing.T) {
	svc, categoryRepoMock := newTestCategoryService()

	userID := uuid.New()
	req := web.CreateCategoryRequest{Name: "Freelance", Type: "INCOME"}

	categoryRepoMock.On("Create", mock.Anything, mock.Anything).Return(domain.Category{}, errors.New("db error"))

	_, err := svc.Create(context.Background(), userID, req)

	assert.Error(t, err)
	categoryRepoMock.AssertExpectations(t)
}

// ============ FIND ALL ============

func TestCategoryFindAll_Success(t *testing.T) {
	svc, categoryRepoMock := newTestCategoryService()

	userID := uuid.New()
	expected := []domain.Category{
		{ID: uuid.New(), UserID: userID, Name: "Gaji", Type: "INCOME", IsDefault: true},
		{ID: uuid.New(), UserID: userID, Name: "Makanan", Type: "EXPENSE", IsDefault: true},
	}

	categoryRepoMock.On("FindAllByUserID", mock.Anything, userID).Return(expected, nil)

	result, err := svc.FindAll(context.Background(), userID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Gaji", result[0].Name)

	categoryRepoMock.AssertExpectations(t)
}

func TestCategoryFindAll_EmptyResult(t *testing.T) {
	svc, categoryRepoMock := newTestCategoryService()

	userID := uuid.New()
	categoryRepoMock.On("FindAllByUserID", mock.Anything, userID).Return([]domain.Category{}, nil)

	result, err := svc.FindAll(context.Background(), userID)

	assert.NoError(t, err)
	assert.Len(t, result, 0)
	categoryRepoMock.AssertExpectations(t)
}

// ============ UPDATE ============

func TestCategoryUpdate_Success(t *testing.T) {
	svc, categoryRepoMock := newTestCategoryService()

	userID := uuid.New()
	categoryID := uuid.New()

	existing := domain.Category{ID: categoryID, UserID: userID, Name: "Old Name"}
	updated := domain.Category{ID: categoryID, UserID: userID, Name: "New Name"}

	categoryRepoMock.On("FindByID", mock.Anything, categoryID).Return(existing, nil)
	categoryRepoMock.On("Update", mock.Anything, categoryID, "New Name").Return(updated, nil)

	req := web.UpdateCategoryRequest{Name: "New Name"}
	result, err := svc.Update(context.Background(), userID, categoryID, req)

	assert.NoError(t, err)
	assert.Equal(t, "New Name", result.Name)
	categoryRepoMock.AssertExpectations(t)
}

func TestCategoryUpdate_Failed_NotFound(t *testing.T) {
	svc, categoryRepoMock := newTestCategoryService()

	userID := uuid.New()
	categoryID := uuid.New()

	categoryRepoMock.On("FindByID", mock.Anything, categoryID).Return(domain.Category{}, errors.New("not found"))

	req := web.UpdateCategoryRequest{Name: "New Name"}
	_, err := svc.Update(context.Background(), userID, categoryID, req)

	assert.Error(t, err)
	categoryRepoMock.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
}

func TestCategoryUpdate_Failed_NotOwner(t *testing.T) {
	svc, categoryRepoMock := newTestCategoryService()

	userID := uuid.New()
	otherUserID := uuid.New()
	categoryID := uuid.New()

	existing := domain.Category{ID: categoryID, UserID: otherUserID, Name: "Old Name"}
	categoryRepoMock.On("FindByID", mock.Anything, categoryID).Return(existing, nil)

	req := web.UpdateCategoryRequest{Name: "New Name"}
	_, err := svc.Update(context.Background(), userID, categoryID, req)

	assert.Error(t, err)
	categoryRepoMock.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
}

// ============ DELETE ============

func TestCategoryDelete_Success(t *testing.T) {
	svc, categoryRepoMock := newTestCategoryService()

	userID := uuid.New()
	categoryID := uuid.New()

	existing := domain.Category{ID: categoryID, UserID: userID, IsDefault: false}

	categoryRepoMock.On("FindByID", mock.Anything, categoryID).Return(existing, nil)
	categoryRepoMock.On("HasTransactions", mock.Anything, categoryID).Return(false, nil)
	categoryRepoMock.On("SoftDelete", mock.Anything, categoryID).Return(nil)

	err := svc.Delete(context.Background(), userID, categoryID)

	assert.NoError(t, err)
	categoryRepoMock.AssertExpectations(t)
}

func TestCategoryDelete_Failed_NotFound(t *testing.T) {
	svc, categoryRepoMock := newTestCategoryService()

	userID := uuid.New()
	categoryID := uuid.New()

	categoryRepoMock.On("FindByID", mock.Anything, categoryID).Return(domain.Category{}, errors.New("not found"))

	err := svc.Delete(context.Background(), userID, categoryID)

	assert.Error(t, err)
	categoryRepoMock.AssertNotCalled(t, "SoftDelete", mock.Anything, mock.Anything)
}

func TestCategoryDelete_Failed_NotOwner(t *testing.T) {
	svc, categoryRepoMock := newTestCategoryService()

	userID := uuid.New()
	otherUserID := uuid.New()
	categoryID := uuid.New()

	existing := domain.Category{ID: categoryID, UserID: otherUserID}
	categoryRepoMock.On("FindByID", mock.Anything, categoryID).Return(existing, nil)

	err := svc.Delete(context.Background(), userID, categoryID)

	assert.Error(t, err)
	categoryRepoMock.AssertNotCalled(t, "SoftDelete", mock.Anything, mock.Anything)
}

func TestCategoryDelete_Failed_IsDefault(t *testing.T) {
	svc, categoryRepoMock := newTestCategoryService()

	userID := uuid.New()
	categoryID := uuid.New()

	existing := domain.Category{ID: categoryID, UserID: userID, IsDefault: true}
	categoryRepoMock.On("FindByID", mock.Anything, categoryID).Return(existing, nil)

	err := svc.Delete(context.Background(), userID, categoryID)

	assert.Error(t, err)
	categoryRepoMock.AssertNotCalled(t, "SoftDelete", mock.Anything, mock.Anything)
	categoryRepoMock.AssertNotCalled(t, "HasTransactions", mock.Anything, mock.Anything)
}

func TestCategoryDelete_WithTransactions_StillSoftDeletes(t *testing.T) {
	svc, categoryRepoMock := newTestCategoryService()

	userID := uuid.New()
	categoryID := uuid.New()

	existing := domain.Category{ID: categoryID, UserID: userID, IsDefault: false}

	categoryRepoMock.On("FindByID", mock.Anything, categoryID).Return(existing, nil)
	categoryRepoMock.On("HasTransactions", mock.Anything, categoryID).Return(true, nil) // punya transaksi
	categoryRepoMock.On("SoftDelete", mock.Anything, categoryID).Return(nil)

	err := svc.Delete(context.Background(), userID, categoryID)

	assert.NoError(t, err) // tetep sukses, karena soft-delete jalan baik ada/gak ada transaksi
	categoryRepoMock.AssertExpectations(t)
}

// ============ SEED DEFAULT CATEGORIES ============

func TestSeedDefaultCategories_Success(t *testing.T) {
	svc, categoryRepoMock := newTestCategoryService()

	userID := uuid.New()

	// Create dipanggil berkali-kali (satu per default category) - pakai mock.Anything biar gak perlu list manual 8 kali
	categoryRepoMock.On("Create", mock.Anything, mock.Anything).Return(domain.Category{ID: uuid.New(), UserID: userID}, nil)

	err := svc.SeedDefaultCategories(context.Background(), userID)

	assert.NoError(t, err)
	categoryRepoMock.AssertNumberOfCalls(t, "Create", 8) // sesuai jumlah default category di service
}

func TestSeedDefaultCategories_Failed_StopsOnFirstError(t *testing.T) {
	svc, categoryRepoMock := newTestCategoryService()

	userID := uuid.New()

	categoryRepoMock.On("Create", mock.Anything, mock.Anything).Return(domain.Category{}, errors.New("db error")).Once()

	err := svc.SeedDefaultCategories(context.Background(), userID)

	assert.Error(t, err)
	categoryRepoMock.AssertNumberOfCalls(t, "Create", 1) // berhenti di percobaan pertama yang gagal
}