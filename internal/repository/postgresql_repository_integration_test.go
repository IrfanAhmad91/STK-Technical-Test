// +build integration

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stk/menu-tree-api/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite is a test suite for integration tests
type IntegrationTestSuite struct {
	suite.Suite
	db   *sql.DB
	repo *PostgreSQLRepository
}

// SetupSuite runs once before all tests
func (s *IntegrationTestSuite) SetupSuite() {
	// Read database connection from environment or use defaults
	dbHost := getEnvOrDefault("DB_HOST", "localhost")
	dbPort := getEnvOrDefault("DB_PORT", "5432")
	dbUser := getEnvOrDefault("DB_USER", "postgres")
	dbPassword := getEnvOrDefault("DB_PASSWORD", "postgres")
	dbName := getEnvOrDefault("DB_NAME", "menu_tree_test")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", dsn)
	require.NoError(s.T(), err, "Failed to connect to test database")

	err = db.Ping()
	require.NoError(s.T(), err, "Failed to ping test database")

	s.db = db
	s.repo = NewPostgreSQLRepository(db)
}

// TearDownSuite runs once after all tests
func (s *IntegrationTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
}

// SetupTest runs before each test
func (s *IntegrationTestSuite) SetupTest() {
	// Clean up the database before each test
	_, err := s.db.Exec("TRUNCATE TABLE menu_items RESTART IDENTITY CASCADE")
	require.NoError(s.T(), err, "Failed to truncate menu_items table")
}

// TestGetDescendantIDsIntegration tests GetDescendantIDs with real database
func (s *IntegrationTestSuite) TestGetDescendantIDsIntegration() {
	ctx := context.Background()

	// Create a hierarchical structure:
	//       1
	//      / \
	//     2   3
	//    / \
	//   4   5
	//      /
	//     6

	items := []domain.MenuItem{
		{Name: "Root", ParentID: nil, Position: 0},             // ID 1
		{Name: "Child1", ParentID: intPtr(1), Position: 0},      // ID 2
		{Name: "Child2", ParentID: intPtr(1), Position: 1},      // ID 3
		{Name: "GrandChild1", ParentID: intPtr(2), Position: 0}, // ID 4
		{Name: "GrandChild2", ParentID: intPtr(2), Position: 1}, // ID 5
		{Name: "GreatGrandChild", ParentID: intPtr(5), Position: 0}, // ID 6
	}

	// Insert all items
	for i := range items {
		err := s.repo.Create(ctx, &items[i])
		require.NoError(s.T(), err)
	}

	// Test: Get descendants of root (ID 1)
	descendants, err := s.repo.GetDescendantIDs(ctx, 1)
	assert.NoError(s.T(), err)
	assert.ElementsMatch(s.T(), []int{2, 3, 4, 5, 6}, descendants, "Root should have all other items as descendants")

	// Test: Get descendants of item 2
	descendants, err = s.repo.GetDescendantIDs(ctx, 2)
	assert.NoError(s.T(), err)
	assert.ElementsMatch(s.T(), []int{4, 5, 6}, descendants, "Item 2 should have items 4, 5, 6 as descendants")

	// Test: Get descendants of item 5
	descendants, err = s.repo.GetDescendantIDs(ctx, 5)
	assert.NoError(s.T(), err)
	assert.ElementsMatch(s.T(), []int{6}, descendants, "Item 5 should have item 6 as descendant")

	// Test: Get descendants of leaf node (ID 6)
	descendants, err = s.repo.GetDescendantIDs(ctx, 6)
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), descendants, "Leaf node should have no descendants")

	// Test: Get descendants of item 3 (no children)
	descendants, err = s.repo.GetDescendantIDs(ctx, 3)
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), descendants, "Item 3 should have no descendants")
}

// TestGetAncestorsIntegration tests GetAncestors with real database
func (s *IntegrationTestSuite) TestGetAncestorsIntegration() {
	ctx := context.Background()

	// Create a hierarchical structure:
	//       1
	//      / \
	//     2   3
	//    / \
	//   4   5
	//      /
	//     6

	items := []domain.MenuItem{
		{Name: "Root", ParentID: nil, Position: 0},             // ID 1
		{Name: "Child1", ParentID: intPtr(1), Position: 0},      // ID 2
		{Name: "Child2", ParentID: intPtr(1), Position: 1},      // ID 3
		{Name: "GrandChild1", ParentID: intPtr(2), Position: 0}, // ID 4
		{Name: "GrandChild2", ParentID: intPtr(2), Position: 1}, // ID 5
		{Name: "GreatGrandChild", ParentID: intPtr(5), Position: 0}, // ID 6
	}

	// Insert all items
	for i := range items {
		err := s.repo.Create(ctx, &items[i])
		require.NoError(s.T(), err)
	}

	// Test: Get ancestors of root (ID 1)
	ancestors, err := s.repo.GetAncestors(ctx, 1)
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), ancestors, "Root should have no ancestors")

	// Test: Get ancestors of item 2
	ancestors, err = s.repo.GetAncestors(ctx, 2)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), ancestors, 1, "Item 2 should have 1 ancestor")
	assert.Equal(s.T(), 1, ancestors[0].ID, "Item 2's ancestor should be item 1")

	// Test: Get ancestors of item 6 (deepest node)
	ancestors, err = s.repo.GetAncestors(ctx, 6)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), ancestors, 3, "Item 6 should have 3 ancestors")
	
	ancestorIDs := make([]int, len(ancestors))
	for i, a := range ancestors {
		ancestorIDs[i] = a.ID
	}
	assert.ElementsMatch(s.T(), []int{1, 2, 5}, ancestorIDs, "Item 6's ancestors should be items 1, 2, 5")

	// Test: Get ancestors of item 3
	ancestors, err = s.repo.GetAncestors(ctx, 3)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), ancestors, 1, "Item 3 should have 1 ancestor")
	assert.Equal(s.T(), 1, ancestors[0].ID, "Item 3's ancestor should be item 1")
}

// TestCircularReferenceDetectionIntegration tests circular reference detection
func (s *IntegrationTestSuite) TestCircularReferenceDetectionIntegration() {
	ctx := context.Background()

	// Create a simple hierarchy:
	//   1
	//   |
	//   2
	//   |
	//   3

	items := []domain.MenuItem{
		{Name: "Item1", ParentID: nil, Position: 0},        // ID 1
		{Name: "Item2", ParentID: intPtr(1), Position: 0},  // ID 2
		{Name: "Item3", ParentID: intPtr(2), Position: 0},  // ID 3
	}

	for i := range items {
		err := s.repo.Create(ctx, &items[i])
		require.NoError(s.T(), err)
	}

	// Scenario: Try to move item 1 to be a child of item 3
	// This would create a circular reference: 3 -> 1 -> 2 -> 3

	// Get ancestors of item 3
	ancestors, err := s.repo.GetAncestors(ctx, 3)
	assert.NoError(s.T(), err)

	// Check if item 1 is an ancestor of item 3
	isAncestor := false
	for _, ancestor := range ancestors {
		if ancestor.ID == 1 {
			isAncestor = true
			break
		}
	}

	assert.True(s.T(), isAncestor, "Item 1 should be an ancestor of item 3")
	// Therefore, moving item 1 to be a child of item 3 would create a circular reference
}

// TestCascadeDeleteIntegration tests cascade delete using GetDescendantIDs
func (s *IntegrationTestSuite) TestCascadeDeleteIntegration() {
	ctx := context.Background()

	// Create a hierarchical structure:
	//       1
	//      / \
	//     2   3
	//    / \
	//   4   5

	items := []domain.MenuItem{
		{Name: "Root", ParentID: nil, Position: 0},             // ID 1
		{Name: "Child1", ParentID: intPtr(1), Position: 0},      // ID 2
		{Name: "Child2", ParentID: intPtr(1), Position: 1},      // ID 3
		{Name: "GrandChild1", ParentID: intPtr(2), Position: 0}, // ID 4
		{Name: "GrandChild2", ParentID: intPtr(2), Position: 1}, // ID 5
	}

	for i := range items {
		err := s.repo.Create(ctx, &items[i])
		require.NoError(s.T(), err)
	}

	// Get descendants before delete
	descendants, err := s.repo.GetDescendantIDs(ctx, 2)
	assert.NoError(s.T(), err)
	assert.ElementsMatch(s.T(), []int{4, 5}, descendants, "Item 2 should have items 4, 5 as descendants")

	// Delete item 2 (should cascade to 4 and 5)
	deletedCount, err := s.repo.Delete(ctx, 2)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 3, deletedCount, "Should delete 3 items (2, 4, 5)")

	// Verify items 4 and 5 are deleted
	_, err = s.repo.FindByID(ctx, 4)
	assert.Error(s.T(), err, "Item 4 should be deleted")

	_, err = s.repo.FindByID(ctx, 5)
	assert.Error(s.T(), err, "Item 5 should be deleted")

	// Verify items 1 and 3 still exist
	item1, err := s.repo.FindByID(ctx, 1)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Root", item1.Name)

	item3, err := s.repo.FindByID(ctx, 3)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Child2", item3.Name)
}

// TestDeepHierarchyIntegration tests recursive CTEs with a deeply nested hierarchy
func (s *IntegrationTestSuite) TestDeepHierarchyIntegration() {
	ctx := context.Background()

	// Create a deep linear hierarchy: 1 -> 2 -> 3 -> 4 -> 5 -> 6 -> 7 -> 8 -> 9 -> 10
	items := make([]domain.MenuItem, 10)
	for i := 0; i < 10; i++ {
		items[i] = domain.MenuItem{
			Name:     fmt.Sprintf("Item%d", i+1),
			Position: 0,
		}
		if i > 0 {
			parentID := i // Parent is the previous item (after insertion, it will have ID = i)
			items[i].ParentID = &parentID
		}
	}

	// Insert items one by one to get proper parent IDs
	for i := range items {
		if i > 0 {
			items[i].ParentID = &items[i-1].ID
		}
		err := s.repo.Create(ctx, &items[i])
		require.NoError(s.T(), err)
	}

	// Test: Get all descendants of the root (should be 9 items)
	descendants, err := s.repo.GetDescendantIDs(ctx, items[0].ID)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), descendants, 9, "Root should have 9 descendants")

	// Test: Get all ancestors of the deepest item (should be 9 items)
	ancestors, err := s.repo.GetAncestors(ctx, items[9].ID)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), ancestors, 9, "Deepest item should have 9 ancestors")
}

// TestRecursiveCTEWithTransactionIntegration tests recursive CTEs within transactions
func (s *IntegrationTestSuite) TestRecursiveCTEWithTransactionIntegration() {
	ctx := context.Background()

	// Create a simple hierarchy
	items := []domain.MenuItem{
		{Name: "Root", ParentID: nil, Position: 0},
		{Name: "Child", ParentID: intPtr(1), Position: 0},
	}

	for i := range items {
		err := s.repo.Create(ctx, &items[i])
		require.NoError(s.T(), err)
	}

	// Test: Use GetDescendantIDs within a transaction
	err := s.repo.WithTransaction(ctx, func(txRepo MenuRepository) error {
		descendants, err := txRepo.GetDescendantIDs(ctx, 1)
		if err != nil {
			return err
		}
		assert.ElementsMatch(s.T(), []int{2}, descendants)
		return nil
	})
	assert.NoError(s.T(), err)

	// Test: Use GetAncestors within a transaction
	err = s.repo.WithTransaction(ctx, func(txRepo MenuRepository) error {
		ancestors, err := txRepo.GetAncestors(ctx, 2)
		if err != nil {
			return err
		}
		assert.Len(s.T(), ancestors, 1)
		assert.Equal(s.T(), 1, ancestors[0].ID)
		return nil
	})
	assert.NoError(s.T(), err)
}

// TestInIntegration runs the integration test suite
func TestIntegration(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests. Set RUN_INTEGRATION_TESTS=true to run.")
	}

	suite.Run(t, new(IntegrationTestSuite))
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func intPtr(i int) *int {
	return &i
}
