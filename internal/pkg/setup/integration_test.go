// file: internal/pkg/setup/api_integration_test.go
package setup_test

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"

	"djj-inventory-system/internal/database"
	"djj-inventory-system/internal/logger"
	"djj-inventory-system/internal/model"
	"djj-inventory-system/internal/pkg/setup"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// registerPayload 用于 /api/register 的请求体
type registerPayload struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
	RoleIDs  []uint `json:"role_ids"`
}

// newTestDB 返回一个内存 SQLite 并自动 migrate 所有 model
func newTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开内存 sqlite 失败: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{}, &model.Role{}, &model.Permission{},
		&model.UserRole{}, &model.RolePermission{},
	); err != nil {
		t.Fatalf("自动 migrate 失败: %v", err)
	}
	return db
}

// 在每个测试前调用，初始化 logger
func init() {
	if err := logger.Init("./logs/test.log", zapcore.DebugLevel); err != nil {
		panic(err)
	}
	defer logger.Sync()
}

// setupTestServer 完成：
// 1) 新建内存 DB, migrate & seed RBAC
// 2) 启动 httptest.Server + cookiejar client
// 3) 注册并登录一个测试用户 fooexample/Secret123!
// 返回 httpexpect.Expect，用于后续认证接口测试
func setupTestServer(t *testing.T) *httpexpect.Expect {
	db := newTestDB(t)
	database.InitRBACSeed(db)

	router := setup.NewRouter(db)
	srv := httptest.NewServer(router)
	t.Cleanup(func() { srv.Close() })

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("初始化 CookieJar 失败: %v", err)
	}
	client := &http.Client{Jar: jar}

	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  srv.URL,
		Client:   client,
		Reporter: httpexpect.NewRequireReporter(t),
	})

	// 先注册
	payload := registerPayload{
		Username: "fooexample",
		Email:    "foo@example.com",
		Password: "Secret123!",
		RoleIDs:  []uint{1}, // 假设 InitRBACSeed 后有 id=1 的角色
	}
	e.POST("/api/register").
		WithJSON(payload).
		Expect().
		Status(http.StatusCreated)

	// 再登录
	e.POST("/api/login").
		WithJSON(map[string]string{
			"username": payload.Username,
			"password": payload.Password,
		}).
		Expect().
		Status(http.StatusOK)

	return e
}

// TestRegister 测试 /api/register
func TestRegister(t *testing.T) {
	db := newTestDB(t)
	database.InitRBACSeed(db)
	router := setup.NewRouter(db)
	srv := httptest.NewServer(router)
	t.Cleanup(func() { srv.Close() })

	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  srv.URL,
		Client:   &http.Client{},
		Reporter: httpexpect.NewRequireReporter(t),
	})

	payload := registerPayload{
		Username: gofakeit.Username(),
		Email:    gofakeit.Email(),
		Password: "TestPass123!",
		RoleIDs:  []uint{1},
	}
	obj := e.POST("/api/register").
		WithJSON(payload).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	// 应该返回一个 id 字段
	obj.Value("id").Number().Gt(0)
}

// TestLogin 测试 /api/login
func TestLogin(t *testing.T) {
	e := setupTestServer(t)

	// 这里登录已经在 setupTestServer 中跑过一遍了，
	// 这次再调用一次，应该仍然返回 200
	e.POST("/api/login").
		WithJSON(map[string]string{
			"username": "fooexample",
			"password": "Secret123!",
		}).
		Expect().
		Status(http.StatusOK)
}

// TestLogout 测试 /api/logout
func TestLogout(t *testing.T) {
	e := setupTestServer(t)

	// 注销应该返回 204 No Content
	e.POST("/api/logout").
		Expect().
		Status(http.StatusNoContent)
}

// TestRolesCRUD 测试 /api/roles 的增删改查
func TestRolesCRUD(t *testing.T) {
	e := setupTestServer(t)

	// 1) Create
	roleName := gofakeit.JobTitle()
	obj := e.POST("/api/roles").
		WithJSON(map[string]string{"name": roleName}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	id := int(obj.Value("id").Number().Raw())

	// 2) List
	arr := e.GET("/api/roles").
		Expect().
		Status(http.StatusOK).
		JSON().Array()
	arr.Length().Gt(0)

	// 3) Get
	e.GET(fmt.Sprintf("/api/roles/%d", id)).
		Expect().
		Status(http.StatusOK).
		JSON().Object().
		ValueEqual("name", roleName)

	// 4) Update
	newName := gofakeit.JobTitle()
	e.PUT(fmt.Sprintf("/api/roles/%d", id)).
		WithJSON(map[string]string{"name": newName}).
		Expect().
		Status(http.StatusOK).
		JSON().Object().
		ValueEqual("name", newName)

	// 5) Delete
	e.DELETE(fmt.Sprintf("/api/roles/%d", id)).
		Expect().
		Status(http.StatusNoContent)
}

// TestPermissionsCRUD 测试 /api/permissions 的增删改查
func TestPermissionsCRUD(t *testing.T) {
	e := setupTestServer(t)

	// 1) Create
	action := gofakeit.HackerVerb()
	object := gofakeit.HackerNoun()
	obj := e.POST("/api/permissions").
		WithJSON(map[string]string{
			"action": action,
			"object": object,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	id := int(obj.Value("id").Number().Raw())

	// 2) List
	list := e.GET("/api/permissions").
		Expect().
		Status(http.StatusOK).
		JSON().Array()
	list.Length().Gt(0)

	// 3) Get
	e.GET(fmt.Sprintf("/api/permissions/%d", id)).
		Expect().
		Status(http.StatusOK).
		JSON().Object().
		ValueEqual("action", action).
		ValueEqual("object", object)

	// 4) Update
	newAction := gofakeit.HackerVerb()
	newObject := gofakeit.HackerNoun()
	e.PUT(fmt.Sprintf("/api/permissions/%d", id)).
		WithJSON(map[string]string{
			"action": newAction,
			"object": newObject,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().Object().
		ValueEqual("action", newAction).
		ValueEqual("object", newObject)

	// 5) Delete
	e.DELETE(fmt.Sprintf("/api/permissions/%d", id)).
		Expect().
		Status(http.StatusNoContent)
}

// TestUsersCRUD 测试 /api/users 的增删改查
func TestUsersCRUD(t *testing.T) {
	e := setupTestServer(t)

	// 1) Create
	username := gofakeit.Username()
	email := gofakeit.Email()
	pass := "User12345!"
	obj := e.POST("/api/users").
		WithJSON(map[string]interface{}{
			"username": username,
			"email":    email,
			"password": pass,
			"role_ids": []uint{1},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()
	uid := int(obj.Value("id").Number().Raw())

	// 2) List
	e.GET("/api/users").
		Expect().
		Status(http.StatusOK).
		JSON().Array().
		Length().Gt(0)

	// 3) Get
	e.GET(fmt.Sprintf("/api/users/%d", uid)).
		Expect().
		Status(http.StatusOK).
		JSON().Object().
		ValueEqual("username", username).
		ValueEqual("email", email)

	// 4) Update
	newUsername := gofakeit.Username()
	e.PUT(fmt.Sprintf("/api/users/%d", uid)).
		WithJSON(map[string]string{"username": newUsername}).
		Expect().
		Status(http.StatusOK).
		JSON().Object().
		ValueEqual("username", newUsername)

	// 5) Delete
	e.DELETE(fmt.Sprintf("/api/users/%d", uid)).
		Expect().
		Status(http.StatusNoContent)
}

// TestUserRoleAssignments 测试 /api/users/:id/roles 的分配、列出和移除
func TestUserRoleAssignments(t *testing.T) {
	e := setupTestServer(t)

	// 先创建一个新用户
	username := gofakeit.Username()
	email := gofakeit.Email()
	pass := "Assign123!"
	obj := e.POST("/api/users").
		WithJSON(map[string]interface{}{
			"username": username,
			"email":    email,
			"password": pass,
			"role_ids": []uint{}, // 先不分配
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()
	uid := int(obj.Value("id").Number().Raw())

	// 分配角色 id=1
	e.POST(fmt.Sprintf("/api/users/%d/roles/1", uid)).
		Expect().
		Status(http.StatusNoContent)

	// 列出
	arr := e.GET(fmt.Sprintf("/api/users/%d/roles", uid)).
		Expect().
		Status(http.StatusOK).
		JSON().Array()
	arr.Length().Equal(1)
	arr.Element(0).Object().ValueEqual("id", 1)

	// 移除
	e.DELETE(fmt.Sprintf("/api/users/%d/roles/1", uid)).
		Expect().
		Status(http.StatusNoContent)

	// 再列，应该是空
	e.GET(fmt.Sprintf("/api/users/%d/roles", uid)).
		Expect().
		Status(http.StatusOK).
		JSON().Array().
		Empty()
}
