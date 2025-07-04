-- ===========================================
-- PostgreSQL 数据库建表 SQL（支持完整业务 + RBAC 权限）
-- ===========================================

-- 枚举定义
CREATE TYPE txn_type AS ENUM ('IN', 'OUT', 'SALE'); -- 入库、出库、销售
CREATE TYPE order_status AS ENUM ('PENDING', 'SHIPPED', 'COMPLETED', 'CANCELLED'); -- 订单状态
CREATE TYPE permission_type AS ENUM ('page', 'button'); -- 权限类型：页面/按钮

-- 用户表
CREATE TABLE IF NOT EXISTS users (
                                     id              SERIAL PRIMARY KEY,
                                     username        VARCHAR(50)  NOT NULL UNIQUE,
    password_hash   VARCHAR(255) NOT NULL,
    full_name       VARCHAR(100),
    email           VARCHAR(100),
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW()
    );

-- 角色表
CREATE TABLE IF NOT EXISTS roles (
                                     id              SERIAL PRIMARY KEY,
                                     name            VARCHAR(50)  NOT NULL UNIQUE,  -- 系统标识（如 admin）
    display_name    VARCHAR(100) NOT NULL,         -- 显示名（如管理员）
    description     TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
    );

-- 用户角色关联表（多对多）
CREATE TABLE IF NOT EXISTS user_roles (
                                          user_id         INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id         INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
    );

-- 模块表（例如库存模块、销售模块等）
CREATE TABLE IF NOT EXISTS modules (
                                       id              SERIAL PRIMARY KEY,
                                       name            VARCHAR(50)  NOT NULL UNIQUE,
    description     TEXT
    );

-- 菜单表（页面菜单结构）
CREATE TABLE IF NOT EXISTS menus (
                                     id              SERIAL PRIMARY KEY,
                                     name            VARCHAR(100) NOT NULL,
    path            VARCHAR(255),
    icon            VARCHAR(100),
    parent_id       INT REFERENCES menus(id) ON DELETE SET NULL,
    order_index     INT DEFAULT 0,
    module_id       INT REFERENCES modules(id) ON DELETE CASCADE
    );

-- 权限表（按钮/接口等操作权限）
CREATE TABLE IF NOT EXISTS permissions (
                                           id              SERIAL PRIMARY KEY,
                                           name            VARCHAR(100) NOT NULL UNIQUE,  -- 例如 inventory.in
    description     TEXT,
    type            permission_type NOT NULL DEFAULT 'button',
    menu_id         INT REFERENCES menus(id) ON DELETE CASCADE
    );

-- 角色权限表
CREATE TABLE IF NOT EXISTS role_permissions (
                                                role_id         INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id   INT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
    );

-- 用户权限表（可用于临时权限）
CREATE TABLE IF NOT EXISTS user_permissions (
                                                user_id         INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission_id   INT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, permission_id)
    );

-- 产品信息
CREATE TABLE IF NOT EXISTS product (
                                       id               SERIAL PRIMARY KEY,
                                       model_name       VARCHAR(100) NOT NULL,
    brand            VARCHAR(100),
    category         VARCHAR(100),
    subcategory      VARCHAR(100),
    tertiary_category VARCHAR(100),
    manufacturer     VARCHAR(100),
    supplier         VARCHAR(100),
    barcode          VARCHAR(50) UNIQUE,
    barcode_url      TEXT,
    price            NUMERIC(12, 2),
    rrp_price        NUMERIC(12, 2),
    cost_price       NUMERIC(12, 2),
    unit             VARCHAR(20),
    spec             TEXT,
    origin           VARCHAR(100),
    standards        TEXT,
    weight_kg        NUMERIC(10, 2),
    lift_capacity_kg NUMERIC(10, 2),
    lift_height_mm   NUMERIC(10, 2),
    power_source     VARCHAR(50),
    warranty         VARCHAR(100),
    marketing_info   TEXT,
    remarks          TEXT,
    photo_url        TEXT,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW()
    );

-- 库存信息（按仓库）
CREATE TABLE IF NOT EXISTS inventory (
                                         id             SERIAL PRIMARY KEY,
                                         product_id     INT REFERENCES product(id) ON DELETE CASCADE,
    region_store   VARCHAR(100),
    actual_qty     INT NOT NULL DEFAULT 0,
    locked_qty     INT NOT NULL DEFAULT 0,
    available_qty  INT GENERATED ALWAYS AS (actual_qty - locked_qty) STORED
    );

-- 库存变动日志
CREATE TABLE IF NOT EXISTS inventory_transaction (
                                                     id           BIGSERIAL PRIMARY KEY,
                                                     product_id   INT NOT NULL REFERENCES product(id),
    type         txn_type NOT NULL,
    quantity     INT NOT NULL,
    region_store VARCHAR(100),
    user_id      INT REFERENCES users(id),
    reference    VARCHAR(100),
    note         TEXT,
    created_at   TIMESTAMPTZ DEFAULT NOW()
    );

-- 客户信息
CREATE TABLE IF NOT EXISTS customer (
                                        id           SERIAL PRIMARY KEY,
                                        name         VARCHAR(100) NOT NULL,
    contact      VARCHAR(100),
    address      VARCHAR(255),
    created_at   TIMESTAMPTZ DEFAULT NOW()
    );

-- 销售订单表
CREATE TABLE IF NOT EXISTS sales_order (
                                           id            SERIAL PRIMARY KEY,
                                           order_number  VARCHAR(50) NOT NULL UNIQUE,
    customer_id   INT REFERENCES customer(id),
    status        order_status NOT NULL DEFAULT 'PENDING',
    total_amount  NUMERIC(12,2) NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    created_by    INT REFERENCES users(id),
    shipped_by    INT REFERENCES users(id)
    );

-- 销售订单明细
CREATE TABLE IF NOT EXISTS sales_order_item (
                                                id           BIGSERIAL PRIMARY KEY,
                                                order_id     INT NOT NULL REFERENCES sales_order(id) ON DELETE CASCADE,
    product_id   INT NOT NULL REFERENCES product(id),
    quantity     INT NOT NULL,
    unit_price   NUMERIC(12,2) NOT NULL
    );

-- 产品附件（如说明书、证书等）
CREATE TABLE IF NOT EXISTS product_attachment (
                                                  id           SERIAL PRIMARY KEY,
                                                  product_id   INT NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    name         VARCHAR(255),
    file_type    VARCHAR(50),
    file_size    INT,
    uploaded_at  TIMESTAMPTZ DEFAULT NOW()
    );

-- 产品审批记录（可选）
CREATE TABLE IF NOT EXISTS product_approval (
                                                id           SERIAL PRIMARY KEY,
                                                product_id   INT NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    step         VARCHAR(100),
    approver     VARCHAR(100),
    status       VARCHAR(20),
    comment      TEXT,
    time         TIMESTAMPTZ
    );
