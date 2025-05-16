package dbutils_test

import (
	"reflect"
	"testing"

	"github.com/gurch101/gowebutils/pkg/dbutils"
)

func TestBuildSearchSelectFields(t *testing.T) {
	tests := []struct {
		name       string
		tableName  string
		fields     []string
		customMaps map[string]string
		expected   []string
	}{
		{
			name:       "basic fields without custom mappings",
			tableName:  "roles",
			fields:     []string{"id", "roleName", "createdAt"},
			customMaps: nil,
			expected: []string{
				"count(*) over()",
				"roles.id",
				"roles.role_name",
				"roles.created_at",
			},
		},
		{
			name:      "with custom mappings",
			tableName: "roles",
			fields:    []string{"id", "numUsers", "activeStatus"},
			customMaps: map[string]string{
				"numUsers":     "count(users.id) as num_users",
				"activeStatus": "CASE WHEN status = 1 THEN 'active' ELSE 'inactive' END as status",
			},
			expected: []string{
				"count(*) over()",
				"roles.id",
				"count(users.id) as num_users",
				"CASE WHEN status = 1 THEN 'active' ELSE 'inactive' END as status",
			},
		},
		{
			name:      "mixed custom and regular fields",
			tableName: "products",
			fields:    []string{"productId", "price", "inventoryCount"},
			customMaps: map[string]string{
				"inventoryCount": "(SELECT COUNT(*) FROM inventory WHERE product_id = products.id) as inventory_count",
			},
			expected: []string{
				"count(*) over()",
				"products.product_id",
				"products.price",
				"(SELECT COUNT(*) FROM inventory WHERE product_id = products.id) as inventory_count",
			},
		},
		{
			name:       "empty fields slice",
			tableName:  "users",
			fields:     []string{},
			customMaps: nil,
			expected: []string{
				"count(*) over()",
			},
		},
		{
			name:       "table name with schema prefix",
			tableName:  "public.accounts",
			fields:     []string{"accountId", "balance"},
			customMaps: nil,
			expected: []string{
				"count(*) over()",
				"public.accounts.account_id",
				"public.accounts.balance",
			},
		},
		{
			name:      "custom mapping overrides all fields",
			tableName: "orders",
			fields:    []string{"id", "total"},
			customMaps: map[string]string{
				"id":    "orders.order_id",
				"total": "(SELECT SUM(price * quantity) FROM order_items WHERE order_id = orders.id) as total",
			},
			expected: []string{
				"count(*) over()",
				"orders.order_id",
				"(SELECT SUM(price * quantity) FROM order_items WHERE order_id = orders.id) as total",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dbutils.BuildSearchSelectFields(tt.tableName, tt.fields, tt.customMaps)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("BuildSearchSelectFields() = %v, want %v", result, tt.expected)
			}
		})
	}
}
