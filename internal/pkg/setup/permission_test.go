// internal/pkg/setup/permission_test.go
package setup

import (
	"djj-inventory-system/internal/handler"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListPermissionModules(t *testing.T) {
	// 1. 设置测试环境 (DB, Gin Engine)
	_, router, err := SetupTest(t)
	if err != nil {
		t.Fatalf("Failed to set up test environment: %v", err)
	}

	// 2. 运行迁移脚本，确保权限已初始化
	// (SetupTest 已经隐式调用了，这里可以加日志确认)
	t.Log("Database migration completed, permissions should be initialized.")

	// 3. 构造并发送 API 请求
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/permissions/modules", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 4. 断言结果
	assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200")

	// 5. 解析并验证 JSON body
	var returnedModules []handler.PermissionModuleDTO
	err = json.Unmarshal(w.Body.Bytes(), &returnedModules)
	assert.NoError(t, err, "Failed to unmarshal response body")

	// 验证模块数量是否和配置一致
	expectedModuleCount := len(handler.PermissionModules)
	assert.Equal(t, expectedModuleCount, len(returnedModules), "Expected number of modules to match config")
	t.Logf("Successfully fetched %d permission modules.", len(returnedModules))

	// 计算并验证总权限数量
	totalPermissionsInConfig := 0
	for _, mod := range handler.PermissionModules {
		totalPermissionsInConfig += len(mod.Permissions)
	}
	totalPermissionsInResponse := 0
	for _, mod := range returnedModules {
		totalPermissionsInResponse += len(mod.Permissions)
	}
	assert.Equal(t, totalPermissionsInConfig, totalPermissionsInResponse, "Expected total number of permissions to match config")
	t.Logf("Total permissions in response: %d", totalPermissionsInResponse)

	// 抽样检查第一个模块的第一个权限
	assert.NotEmpty(t, returnedModules, "Returned modules should not be empty")
	firstModule := returnedModules[0]
	expectedFirstModule := handler.PermissionModules[0]
	assert.Equal(t, expectedFirstModule.Module, firstModule.Module, "First module name should match")
	assert.Equal(t, expectedFirstModule.Description, firstModule.Description, "First module description should match")

	assert.NotEmpty(t, firstModule.Permissions, "Permissions in the first module should not be empty")
	firstPermission := firstModule.Permissions[0]
	expectedFirstPermission := expectedFirstModule.Permissions[0]
	assert.Equal(t, expectedFirstPermission.ID, firstPermission.ID, "First permission ID should match")
	assert.Equal(t, expectedFirstPermission.Name, firstPermission.Name, "First permission name should match")
	assert.Equal(t, expectedFirstPermission.Label, firstPermission.Label, "First permission label should match")
	assert.Equal(t, expectedFirstPermission.Description, firstPermission.Description, "First permission description should match")

	t.Logf("Sample check for permission '%s' passed.", firstPermission.Label)
}
