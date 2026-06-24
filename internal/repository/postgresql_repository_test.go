package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stk/menu-tree-api/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetDescendantIDs tests the recursive CTE query for retrieving descendant IDs
func TestGetDescendantIDs(t *testing.T) {
	tests := []struct {
		name           string
		itemID         int
		setupMock      func(mock sqlmock.Sqlmock)
		expectedIDs    []int
		expectedError  bool
		errorContains  string
	}{
		{
			name:   "successfully retrieves descendant IDs",
			itemID: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id"}).
					AddRow(2).
					AddRow(3).
					AddRow(4)

				mock.ExpectQuery(`WITH RECURSIVE descendants`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedIDs:   []int{2, 3, 4},
			expectedError: false,
		},
		{
			name:   "returns empty slice when no descendants",
			itemID: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id"})

				mock.ExpectQuery(`WITH RECURSIVE descendants`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedIDs:   []int{},
			expectedError: false,
		},
		{
			name:   "handles database query error",
			itemID: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`WITH RECURSIVE descendants`).
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			expectedIDs:   nil,
			expectedError: true,
			errorContains: "failed to get descendants",
		},
		{
			name:   "handles row scan error",
			itemID: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id"}).
					AddRow("invalid") // Wrong type

				mock.ExpectQuery(`WITH RECURSIVE descendants`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedIDs:   nil,
			expectedError: true,
			errorContains: "failed to scan descendant id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.setupMock(mock)

			repo := NewPostgreSQLRepository(db)
			ctx := context.Background()

			ids, err := repo.GetDescendantIDs(ctx, tt.itemID)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if len(tt.expectedIDs) == 0 {
					assert.Empty(t, ids)
				} else {
					assert.Equal(t, tt.expectedIDs, ids)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestGetAncestors tests the recursive CTE query for retrieving ancestors
func TestGetAncestors(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		itemID         int
		setupMock      func(mock sqlmock.Sqlmock)
		expectedItems  []domain.MenuItem
		expectedError  bool
		errorContains  string
	}{
		{
			name:   "successfully retrieves ancestors",
			itemID: 5,
			setupMock: func(mock sqlmock.Sqlmock) {
				grandParentID := 1
				rows := sqlmock.NewRows([]string{"id", "name", "parent_id", "position", "created_at", "updated_at"}).
					AddRow(2, "Parent", &grandParentID, 0, now, now).
					AddRow(1, "GrandParent", nil, 0, now, now)

				mock.ExpectQuery(`WITH RECURSIVE ancestors`).
					WithArgs(5).
					WillReturnRows(rows)
			},
			expectedItems: []domain.MenuItem{
				{ID: 2, Name: "Parent", ParentID: intPtr(1), Position: 0, CreatedAt: now, UpdatedAt: now},
				{ID: 1, Name: "GrandParent", ParentID: nil, Position: 0, CreatedAt: now, UpdatedAt: now},
			},
			expectedError: false,
		},
		{
			name:   "returns empty slice when no ancestors (root level item)",
			itemID: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "parent_id", "position", "created_at", "updated_at"})

				mock.ExpectQuery(`WITH RECURSIVE ancestors`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedItems: []domain.MenuItem{},
			expectedError: false,
		},
		{
			name:   "handles database query error",
			itemID: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`WITH RECURSIVE ancestors`).
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			expectedItems: nil,
			expectedError: true,
			errorContains: "failed to get ancestors",
		},
		{
			name:   "handles row scan error",
			itemID: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "parent_id", "position", "created_at", "updated_at"}).
					AddRow("invalid", "Name", nil, 0, now, now) // Invalid ID type

				mock.ExpectQuery(`WITH RECURSIVE ancestors`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedItems: nil,
			expectedError: true,
			errorContains: "failed to scan ancestor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.setupMock(mock)

			repo := NewPostgreSQLRepository(db)
			ctx := context.Background()

			items, err := repo.GetAncestors(ctx, tt.itemID)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if len(tt.expectedItems) == 0 {
					assert.Empty(t, items)
				} else {
					assert.Equal(t, len(tt.expectedItems), len(items))
					for i, expected := range tt.expectedItems {
						assert.Equal(t, expected.ID, items[i].ID)
						assert.Equal(t, expected.Name, items[i].Name)
						assert.Equal(t, expected.ParentID, items[i].ParentID)
						assert.Equal(t, expected.Position, items[i].Position)
					}
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestRecursiveCTEWithTransaction tests that recursive queries work within transactions
func TestRecursiveCTEWithTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	ctx := context.Background()

	t.Run("GetDescendantIDs within transaction", func(t *testing.T) {
		mock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"id"}).
			AddRow(2).
			AddRow(3)

		mock.ExpectQuery(`WITH RECURSIVE descendants`).
			WithArgs(1).
			WillReturnRows(rows)

		mock.ExpectCommit()

		err := repo.WithTransaction(ctx, func(txRepo MenuRepository) error {
			ids, err := txRepo.GetDescendantIDs(ctx, 1)
			if err != nil {
				return err
			}
			assert.Equal(t, []int{2, 3}, ids)
			return nil
		})

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetAncestors within transaction", func(t *testing.T) {
		now := time.Now()
		mock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"id", "name", "parent_id", "position", "created_at", "updated_at"}).
			AddRow(1, "Parent", nil, 0, now, now)

		mock.ExpectQuery(`WITH RECURSIVE ancestors`).
			WithArgs(2).
			WillReturnRows(rows)

		mock.ExpectCommit()

		err := repo.WithTransaction(ctx, func(txRepo MenuRepository) error {
			items, err := txRepo.GetAncestors(ctx, 2)
			if err != nil {
				return err
			}
			assert.Equal(t, 1, len(items))
			assert.Equal(t, "Parent", items[0].Name)
			return nil
		})

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestCircularReferenceDetection tests using GetAncestors to detect circular references
func TestCircularReferenceDetection(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	ctx := context.Background()

	t.Run("can detect if new parent is a descendant", func(t *testing.T) {
		// Scenario: Moving item 1 to be child of item 5
		// Item 5's ancestors are [3, 1], so item 1 is an ancestor of 5
		// This would create a circular reference

		now := time.Now()
		rows := sqlmock.NewRows([]string{"id", "name", "parent_id", "position", "created_at", "updated_at"}).
			AddRow(3, "Item3", intPtr(1), 0, now, now).
			AddRow(1, "Item1", nil, 0, now, now)

		mock.ExpectQuery(`WITH RECURSIVE ancestors`).
			WithArgs(5).
			WillReturnRows(rows)

		ancestors, err := repo.GetAncestors(ctx, 5)
		assert.NoError(t, err)

		// Check if itemID 1 is in ancestors (would create circular reference)
		itemID := 1
		isCircular := false
		for _, ancestor := range ancestors {
			if ancestor.ID == itemID {
				isCircular = true
				break
			}
		}

		assert.True(t, isCircular, "Should detect that moving item 1 under item 5 would create circular reference")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestCascadeDeleteUsingDescendants tests that GetDescendantIDs can be used for cascade operations
func TestCascadeDeleteUsingDescendants(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	ctx := context.Background()

	t.Run("retrieve all descendants before cascade delete", func(t *testing.T) {
		// Scenario: Deleting item 1 should also delete items 2, 3, 4 (descendants)
		rows := sqlmock.NewRows([]string{"id"}).
			AddRow(2).
			AddRow(3).
			AddRow(4)

		mock.ExpectQuery(`WITH RECURSIVE descendants`).
			WithArgs(1).
			WillReturnRows(rows)

		descendants, err := repo.GetDescendantIDs(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, []int{2, 3, 4}, descendants)

		// This information can be used to:
		// 1. Warn the user about cascade delete
		// 2. Log which items will be deleted
		// 3. Perform additional cleanup operations
		totalToDelete := len(descendants) + 1 // descendants + the item itself
		assert.Equal(t, 4, totalToDelete)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestGetDescendantIDsDeepHierarchy tests recursive CTE with deeply nested hierarchy
func TestGetDescendantIDsDeepHierarchy(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	ctx := context.Background()

	t.Run("handles deeply nested hierarchy", func(t *testing.T) {
		// Simulate a deep tree: 1 -> 2 -> 3 -> 4 -> 5 -> 6 -> 7 -> 8 -> 9 -> 10
		rows := sqlmock.NewRows([]string{"id"})
		for i := 2; i <= 10; i++ {
			rows.AddRow(i)
		}

		mock.ExpectQuery(`WITH RECURSIVE descendants`).
			WithArgs(1).
			WillReturnRows(rows)

		descendants, err := repo.GetDescendantIDs(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, 9, len(descendants))

		// Verify all IDs from 2-10 are present
		expectedIDs := []int{2, 3, 4, 5, 6, 7, 8, 9, 10}
		assert.Equal(t, expectedIDs, descendants)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestContextCancellation tests that context cancellation is respected
func TestContextCancellation(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)

	t.Run("GetDescendantIDs respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// When context is cancelled before query, the query may or may not be executed
		// depending on timing. We just verify that an error is returned.
		_, err := repo.GetDescendantIDs(ctx, 1)
		assert.Error(t, err)
	})

	t.Run("GetAncestors respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// When context is cancelled before query, the query may or may not be executed
		// depending on timing. We just verify that an error is returned.
		_, err := repo.GetAncestors(ctx, 1)
		assert.Error(t, err)
	})
}

// Helper function to create int pointers
func intPtr(i int) *int {
	return &i
}
