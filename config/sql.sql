-- ===========================================
-- 切换客户端编码
-- ===========================================
SET client_encoding = 'UTF8';  -- 客户端使用 UTF8 编码

-- ===========================================
-- 0. 通用枚举类型（ENUM）定义
-- ===========================================
-- 货物性质枚举
CREATE TYPE goods_nature_enum AS ENUM (
  'contract',           -- 合同货
  'multi_contract',     -- 多合同货
  'partial_contract',   -- 部分合同货
  'warranty',           -- 保修货
  'gift',               -- 赠品
  'self_purchased',     -- 自购件
  'consignment'         -- 寄售件
);
-- 审批结果枚举
CREATE TYPE approval_status_enum AS ENUM (
  'pending',            -- 待审批
  'approved',           -- 已通过
  'rejected'            -- 已驳回
);
-- 库存状态枚举（库存表使用）
CREATE TYPE stock_status_enum AS ENUM (
  'pending',            -- 待入库
  'in_stock',           -- 已入库
  'not_applicable'      -- 不适用
);
-- 订单状态枚举（订单表使用）
CREATE TYPE order_status_enum AS ENUM (
  'draft',                      -- 草稿
  'ordered',                    -- 已下单
  'deposit_received',           -- 已付定金
  'final_payment_received',     -- 已付尾款
  'pre_delivery_inspection',    -- 发货前检验
  'shipped',                    -- 已发货
  'delivered',                  -- 已完成
  'order_closed',               -- 订单关闭
  'cancelled'                   -- 已取消
);
-- 币种代码枚举
CREATE TYPE currency_code_enum AS ENUM (
  'AUD',                        -- 澳元
  'USD',                        -- 美元
  'CNY',                        -- 人民币
  'EUR',                        -- 欧元
  'GBP'                         -- 英镑
);
-- 客户类型枚举
CREATE TYPE customer_type_enum AS ENUM (
  'retail',                     -- 零售
  'wholesale',                  -- 批发
  'online'                      -- 电商
);
-- 订单类型枚举
CREATE TYPE order_type_enum AS ENUM (
  'purchase',                   -- 采购单
  'sales'                       -- 销售单
);
-- 产品主分类枚举
CREATE TYPE product_type_enum AS ENUM (
  'machine',    -- 主机
  'parts',      -- 配件
  'attachment', -- 属具
  'tools',      -- 工具
  'others'      -- 其他
);
-- 产品状态枚举
CREATE TYPE product_status_enum AS ENUM (
  'draft',
  'pending_tech',
  'pending_purchase',
  'pending_finance',
  'ready_published',
  'published',
  'rejected',
  'closed'
);
-- 审批流程状态枚举
CREATE TYPE application_status_enum AS ENUM (
  'open',    -- 审批中
  'closed'   -- 已结束
);
-- 审核阶段枚举
CREATE TYPE review_stage_enum AS ENUM (
  'technical',  -- 技术
  'purchasing', -- 采购
  'finance'     -- 财务
);

-- ===========================================
-- 1. 用户与权限（RBAC）基础表
-- ===========================================
-- 用户表，存储账号及软删除状态
CREATE TABLE users (
                       id            SERIAL       PRIMARY KEY,
                       username      VARCHAR(50)  UNIQUE NOT NULL,     -- 登录用户名
                       password_hash VARCHAR(256) NOT NULL,            -- 密码哈希
                       version       BIGINT       NOT NULL DEFAULT 1,  -- 乐观锁
                       created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
                       updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
                       is_deleted    BOOLEAN      NOT NULL DEFAULT FALSE -- 软删除
);
-- 角色表
CREATE TABLE roles (
                       id   SERIAL PRIMARY KEY,
                       name VARCHAR(50) UNIQUE NOT NULL                -- 角色名
);
-- 权限表
CREATE TABLE permissions (
                             id   SERIAL PRIMARY KEY,
                             name VARCHAR(100) UNIQUE NOT NULL               -- 权限名
);
-- 用户与角色关联表
CREATE TABLE user_roles (
                            user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                            role_id INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
                            PRIMARY KEY(user_id, role_id)
);
-- 角色与权限关联表
CREATE TABLE role_permissions (
                                  role_id       INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
                                  permission_id INT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
                                  PRIMARY KEY(role_id, permission_id)
);

-- ===========================================
-- 2. 通用附件表
-- ===========================================
CREATE TABLE attachments (
                             id          SERIAL       PRIMARY KEY,
                             file_name   VARCHAR(255) NOT NULL,           -- 文件名
                             file_type   VARCHAR(100) NOT NULL,           -- 文件类型
                             file_size   INT,                             -- 大小（字节）
                             url         TEXT         NOT NULL,           -- 存储路径或链接
                             uploaded_by INT REFERENCES users(id),        -- 上传者ID
                             uploaded_at TIMESTAMPTZ NOT NULL DEFAULT now()-- 上传时间
);

-- ===========================================
-- 3. 地区与仓库关联（多对多）
-- ===========================================
-- 地区表
CREATE TABLE regions (
                         id          SERIAL        PRIMARY KEY,
                         name        VARCHAR(100)  NOT NULL UNIQUE,    -- 地区名称
                         created_at  TIMESTAMPTZ   NOT NULL DEFAULT now(),
                         updated_at  TIMESTAMPTZ   NOT NULL DEFAULT now()
);
-- 仓库表
CREATE TABLE warehouses (
                            id          SERIAL       PRIMARY KEY,
                            name        VARCHAR(100) NOT NULL UNIQUE,     -- 仓库名称
                            location    VARCHAR(255),                      -- 地址
                            version     BIGINT       NOT NULL DEFAULT 1,   -- 乐观锁
                            created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
                            updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
                            is_deleted  BOOLEAN      NOT NULL DEFAULT FALSE -- 软删除
);
-- 地区与仓库关联表
CREATE TABLE region_warehouses (
                                   region_id    INT NOT NULL REFERENCES regions(id)    ON DELETE CASCADE, -- 地区ID
                                   warehouse_id INT NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE, -- 仓库ID
                                   PRIMARY KEY(region_id, warehouse_id)
);
CREATE INDEX idx_region_wh_region    ON region_warehouses(region_id);
CREATE INDEX idx_region_wh_warehouse ON region_warehouses(warehouse_id);

-- ===========================================
-- 4. 门店与客户
-- ===========================================
-- 门店表
CREATE TABLE stores (
                        id           SERIAL        PRIMARY KEY,
                        code         VARCHAR(50)   NOT NULL UNIQUE,     -- 门店编码
                        name         VARCHAR(100)  NOT NULL,             -- 门店名称
                        region_id    INT           NOT NULL REFERENCES regions(id), -- 所属地区
                        address      VARCHAR(255),                        -- 地址
                        manager_id   INT           REFERENCES users(id),  -- 负责人ID
                        version      BIGINT        NOT NULL DEFAULT 1,   -- 乐观锁
                        created_at   TIMESTAMPTZ   NOT NULL DEFAULT now(),
                        updated_at   TIMESTAMPTZ   NOT NULL DEFAULT now(),
                        is_deleted   BOOLEAN       NOT NULL DEFAULT FALSE -- 软删除
);
CREATE INDEX idx_stores_region  ON stores(region_id);
CREATE INDEX idx_stores_manager ON stores(manager_id);
-- 客户表
CREATE TABLE customers (
                           id         SERIAL              PRIMARY KEY,
                           store_id   INT     REFERENCES stores(id),    -- 门店ID
                           type       customer_type_enum NOT NULL DEFAULT 'retail', -- 客户类型
                           name       VARCHAR(100)       NOT NULL,     -- 名称
                           phone      VARCHAR(20),                         -- 电话
                           email      VARCHAR(100),                        -- 邮箱
                           address    VARCHAR(255),                        -- 地址
                           version    BIGINT             NOT NULL DEFAULT 1, -- 乐观锁
                           created_at TIMESTAMPTZ        NOT NULL DEFAULT now(),
                           updated_at TIMESTAMPTZ        NOT NULL DEFAULT now(),
                           is_deleted BOOLEAN            NOT NULL DEFAULT FALSE -- 软删除
);
CREATE INDEX idx_customers_store ON customers(store_id);

-- ===========================================
-- 5. 币种与汇率
-- ===========================================
CREATE TABLE currency_rates (
                                code        currency_code_enum PRIMARY KEY,  -- 币种
                                rate_to_aud NUMERIC(14,6)     NOT NULL,      -- 相对AUD汇率
                                version     BIGINT            NOT NULL DEFAULT 1,
                                updated_at  TIMESTAMPTZ        NOT NULL DEFAULT now()
);
CREATE TABLE currency_rate_history (
                                       id             SERIAL              PRIMARY KEY,
                                       code           currency_code_enum  NOT NULL,
                                       rate_to_aud    NUMERIC(14,6)       NOT NULL,
                                       effective_date DATE                NOT NULL     -- 生效日期
);
CREATE INDEX idx_crh_code_date ON currency_rate_history(code,effective_date);

-- ===========================================
-- 6. 产品分类字典表（多级分类）
-- ===========================================
CREATE TABLE product_categories (
                                    id         SERIAL       PRIMARY KEY,
                                    name       VARCHAR(100) UNIQUE NOT NULL,     -- 分类名称
                                    parent_id  INT          REFERENCES product_categories(id) ON DELETE SET NULL, -- 父分类
                                    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
                                    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- ===========================================
-- 7. 产品表
-- ===========================================
CREATE TABLE products (
                          id                   SERIAL               PRIMARY KEY,
                          djj_code             VARCHAR(50)          UNIQUE NOT NULL, -- SKU
                          name_cn              VARCHAR(100)         NOT NULL,        -- 中文名
                          name_en              VARCHAR(100),                         -- 英文名
                          manufacturer         VARCHAR(100),                         -- 厂商
                          manufacturer_code    VARCHAR(100),                         -- 厂家编码
                          supplier             VARCHAR(100),                         -- 供应商
                          model                VARCHAR(100),                         -- 型号
                          category_id          INT REFERENCES product_categories(id), -- 一级分类ID
                          subcategory_id       INT REFERENCES product_categories(id), -- 二级分类ID
                          tertiary_category_id INT REFERENCES product_categories(id), -- 三级分类ID
                          technical_specs      JSONB,                                    -- 技术参数
                          specs                TEXT,                                     -- 规格
                          price                NUMERIC(12,2)        NOT NULL CHECK(price>=0), -- 进价
                          rrp_price            NUMERIC(12,2)        CHECK(rrp_price>=0),    -- 建议零售价
                          currency             currency_code_enum   NOT NULL DEFAULT 'AUD',-- 币种
                          status               product_status_enum  NOT NULL DEFAULT 'draft',-- 产品状态
                          application_status   application_status_enum NOT NULL DEFAULT 'open',-- 审批流程状态
                          product_type         product_type_enum    NOT NULL DEFAULT 'others',-- 主分类
                          standard_warranty    VARCHAR(100),                         -- 保修期
                          remarks              TEXT,                                     -- 备注
                          marketing_info       TEXT,                                     -- 营销信息
                          training_docs        TEXT,                                     -- 培训资料
                          weight_kg            NUMERIC(10,2)        CHECK(weight_kg>=0),     -- 重量
                          lift_capacity_kg     NUMERIC(10,2)        CHECK(lift_capacity_kg>=0),-- 起重量
                          lift_height_mm       NUMERIC(10,2)        CHECK(lift_height_mm>=0),  -- 起升高度
                          power_source         VARCHAR(100),                         -- 动力源
                          other_specs          JSONB    DEFAULT '{}'::JSONB,            -- 扩展1
                          extra_info           JSONB    DEFAULT '{}'::JSONB,            -- 扩展2
                          metadata             JSONB    DEFAULT '{}'::JSONB,            -- 元数据
                          version              BIGINT   NOT NULL DEFAULT 1,             -- 乐观锁
                          created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),      -- 创建时间
                          updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),      -- 更新时间
                          is_deleted           BOOLEAN  NOT NULL DEFAULT FALSE         -- 软删除
);
CREATE INDEX idx_products_supplier  ON products(supplier);
CREATE INDEX idx_products_category  ON products(category_id);
CREATE INDEX idx_products_rrp_price ON products(rrp_price);

-- ===========================================
-- 8. 新品上线审核记录表
-- ===========================================
CREATE TABLE product_launch_reviews (
                                        id           SERIAL                PRIMARY KEY,
                                        product_id   INT     NOT NULL       REFERENCES products(id) ON DELETE CASCADE,-- 产品ID
                                        stage        review_stage_enum     NOT NULL,     -- 审核阶段
                                        status       approval_status_enum  NOT NULL DEFAULT 'pending',-- 审批结果
                                        comments     TEXT,                                  -- 审批意见
                                        reviewer_id  INT                   REFERENCES users(id),           -- 审核人
                                        reviewed_at  TIMESTAMPTZ,                                  -- 审核时间
                                        created_at   TIMESTAMPTZ NOT NULL DEFAULT now()           -- 创建时间
);
CREATE INDEX idx_plr_product ON product_launch_reviews(product_id);

-- ===========================================
-- 9. 产品专用附件表
-- ===========================================
CREATE TABLE product_attachments (
                                     id           SERIAL       PRIMARY KEY,
                                     product_id   INT          NOT NULL REFERENCES products(id) ON DELETE CASCADE,-- 产品ID
                                     file_name    VARCHAR(255) NOT NULL,          -- 文件名
                                     file_type    VARCHAR(100) NOT NULL,          -- 文件类型
                                     file_size    INT,                            -- 文件大小
                                     url          TEXT         NOT NULL,          -- 链接
                                     uploaded_by  INT          REFERENCES users(id),   -- 上传者
                                     uploaded_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),   -- 上传时间
                                     metadata     JSONB        DEFAULT '{}'::JSONB       -- 元数据
);
CREATE INDEX idx_prod_attach_product ON product_attachments(product_id);

-- ===========================================
-- 10. 库存与日志表
-- ===========================================
CREATE TABLE inventory (
                           id                 SERIAL       PRIMARY KEY,
                           product_id         INT          NOT NULL REFERENCES products(id),  -- 产品ID
                           warehouse_id       INT          NOT NULL REFERENCES warehouses(id),-- 仓库ID
                           on_hand            INT          NOT NULL DEFAULT 0,                -- 在库数量
                           reserved_for_order INT          NOT NULL DEFAULT 0,                -- 预留数量
                           version            BIGINT       NOT NULL DEFAULT 1,                -- 乐观锁
                           updated_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),             -- 更新时间
                           UNIQUE(product_id, warehouse_id)                                -- 复合唯一
);
CREATE INDEX idx_inventory_product   ON inventory(product_id);
CREATE INDEX idx_inventory_warehouse ON inventory(warehouse_id);

CREATE TABLE inventory_logs (
                                id           SERIAL       PRIMARY KEY,
                                inventory_id INT          NOT NULL REFERENCES inventory(id),       -- 库存ID
                                change_type  VARCHAR(20)  NOT NULL,                                -- 变动类型
                                quantity     INT          NOT NULL,                                -- 变动数量
                                operator_id  INT NOT NULL REFERENCES users(id),                     -- 操作人
                                remark       TEXT,                                                  -- 备注
                                created_at   TIMESTAMPTZ  NOT NULL DEFAULT now()                    -- 记录时间
);
CREATE INDEX idx_invlog_inventory ON inventory_logs(inventory_id);

-- ===========================================
-- 11. 订单与明细表
-- ===========================================
CREATE TABLE orders (
                        id          _SERIAL             PRIMARY KEY,
                        order_type   order_type_enum    NOT NULL,      -- 订单类型
                        order_number VARCHAR(50)        UNIQUE NOT NULL,-- 订单号
                        store_id     INT     REFERENCES stores(id),     -- 门店ID
                        partner_id   INT     NOT NULL,                   -- 合作方ID
                        order_date   DATE    NOT NULL,                  -- 订单日期
                        status       order_status_enum  NOT NULL DEFAULT 'draft',-- 订单状态
                        currency     currency_code_enum NOT NULL DEFAULT 'AUD', -- 币种
                        total_amount NUMERIC(14,2),                       -- 总金额
                        version      BIGINT             NOT NULL DEFAULT 1,      -- 乐观锁
                        created_at   TIMESTAMPTZ        NOT NULL DEFAULT now(),   -- 创建时间
                        updated_at   TIMESTAMPTZ        NOT NULL DEFAULT now()    -- 更新时间
);
CREATE INDEX idx_orders_store ON orders(store_id);

CREATE TABLE order_items (
                             id         SERIAL       PRIMARY KEY,
                             order_id   INT          NOT NULL REFERENCES orders(id) ON DELETE CASCADE,-- 订单ID
                             product_id INT          NOT NULL REFERENCES products(id),               -- 产品ID
                             quantity   INT          NOT NULL,                                       -- 数量
                             unit_price NUMERIC(12,2)                                             -- 单价
);
CREATE INDEX idx_items_order   ON order_items(order_id);
CREATE INDEX idx_items_product ON order_items(product_id);

-- ===========================================
-- 12. 发票与付款表
-- ===========================================
CREATE TABLE invoices (
                          id             SERIAL       PRIMARY KEY,
                          order_id       INT          NOT NULL REFERENCES orders(id),   -- 订单ID
                          invoice_number VARCHAR(50)  UNIQUE NOT NULL,                   -- 发票号
                          issue_date     DATE         NOT NULL,                           -- 开票日期
                          total_amount   NUMERIC(14,2)                                   -- 总金额
);
CREATE INDEX idx_invoices_order ON invoices(order_id);

CREATE TABLE payments (
                          id       SERIAL       PRIMARY KEY,
                          order_id INT          NOT NULL REFERENCES orders(id),          -- 订单ID
                          amount   NUMERIC(12,2) NOT NULL,                               -- 付款金额
                          paid_at  TIMESTAMPTZ  NOT NULL DEFAULT now()                   -- 支付时间
);
CREATE INDEX idx_payments_order ON payments(order_id);

-- ===========================================
-- 13. PD 检查 & 发货表
-- ===========================================
CREATE TABLE pd_inspections (
                                id         SERIAL       PRIMARY KEY,
                                order_id   INT          NOT NULL REFERENCES orders(id) ON DELETE CASCADE, -- 订单ID
                                passed     BOOLEAN      NOT NULL,                                        -- 是否通过
                                created_at TIMESTAMPTZ  NOT NULL DEFAULT now()                           -- 检查时间
);

CREATE TABLE shipments (
                           id         SERIAL       PRIMARY KEY,
                           order_id   INT          NOT NULL REFERENCES orders(id) ON DELETE CASCADE, -- 订单ID
                           shipped_at TIMESTAMPTZ  NOT NULL DEFAULT now()                           -- 发货时间
);

-- ===========================================
-- 14. 客户活动、提醒、审批日志表
-- ===========================================
CREATE TABLE customer_activities (
                                     id            SERIAL       PRIMARY KEY,
                                     customer_id   INT          NOT NULL REFERENCES customers(id), -- 客户ID
                                     activity_type VARCHAR(50),                                    -- 活动类型
                                     created_at    TIMESTAMPTZ  NOT NULL DEFAULT now()             -- 活动时间
);

CREATE TABLE reminders (
                           id         SERIAL       PRIMARY KEY,
                           ref_type   VARCHAR(20),                                         -- 关联类型
                           ref_id     INT,                                                 -- 关联ID
                           remind_at  TIMESTAMPTZ  NOT NULL,                               -- 提醒时间
                           message    TEXT                                                  -- 提醒消息
);

CREATE TABLE approval_logs (
                               id         SERIAL       PRIMARY KEY,
                               ref_type   VARCHAR(20),                                        -- 关联类型
                               ref_id     INT,                                                -- 关联ID
                               result     VARCHAR(20),                                        -- 审批结果
                               created_at TIMESTAMPTZ  NOT NULL DEFAULT now()                 -- 日志时间
);

-- ===========================================
-- 15. 报价申请与报价单表
-- ===========================================
CREATE TABLE quote_requests (
                                id           SERIAL       PRIMARY KEY,
                                store_id     INT           REFERENCES stores(id), -- 门店ID
                                quote_date   DATE          NOT NULL,              -- 报价日期
                                total_amount NUMERIC(14,2)                         -- 总金额
);

CREATE TABLE quote_lists (
                             id               SERIAL       PRIMARY KEY,
                             quote_request_id INT          REFERENCES quote_requests(id), -- 申请ID
                             list_date        DATE,                                      -- 列表日期
                             total_amount     NUMERIC(14,2)                              -- 总金额
);

-- ===========================================
-- 16. 统一审计历史表 & 枚举更新
-- ===========================================
CREATE TYPE audited_table_enum AS ENUM (
  'inventory',
  'orders',
  'quote_requests',
  'quote_lists'
);
ALTER TYPE audited_table_enum ADD VALUE 'products';
ALTER TYPE audited_table_enum ADD VALUE 'product_launch_reviews';

CREATE TABLE audited_history (
                                 history_id BIGSERIAL            PRIMARY KEY,
                                 table_name audited_table_enum   NOT NULL,        -- 表名
                                 record_id  INT                  NOT NULL,        -- 记录ID
                                 store_id   INT,                                    -- 门店ID
                                 changed_by INT                  NOT NULL REFERENCES users(id), -- 操作人ID
                                 operation  VARCHAR(10)          NOT NULL,        -- 操作类型
                                 payload    JSONB                NOT NULL,        -- 变更前数据
                                 changed_at TIMESTAMPTZ          NOT NULL DEFAULT now() -- 变更时间
);
CREATE INDEX idx_audhist_tbl   ON audited_history(table_name);
CREATE INDEX idx_audhist_store ON audited_history(store_id);
CREATE INDEX idx_audhist_user  ON audited_history(changed_by);
CREATE INDEX idx_audhist_time  ON audited_history(changed_at);

-- ===========================================
-- 17. 审计触发函数 & 触发器
-- ===========================================
CREATE OR REPLACE FUNCTION fn_audit_generic()
RETURNS TRIGGER AS $$
DECLARE
rec_json JSONB;
  sid      INT;
  uid      INT := current_setting('app.current_user_id')::INT;
BEGIN
  rec_json := to_jsonb(OLD);
BEGIN
    sid := OLD.store_id;
EXCEPTION WHEN undefined_column THEN
    sid := NULL;
END;
INSERT INTO audited_history(
    table_name, record_id, store_id,
    changed_by, operation, payload, changed_at
) VALUES (
             TG_TABLE_NAME::audited_table_enum,
             OLD.id,
             sid,
             uid,
             TG_OP,
             rec_json,
             now()
         );
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 注册触发器
DROP TRIGGER IF EXISTS trg_inventory_audit ON inventory;
CREATE TRIGGER trg_inventory_audit
    BEFORE UPDATE OR DELETE ON inventory
  FOR EACH ROW EXECUTE FUNCTION fn_audit_generic();

DROP TRIGGER IF EXISTS trg_orders_audit ON orders;
CREATE TRIGGER trg_orders_audit
    BEFORE UPDATE OR DELETE ON orders
  FOR EACH ROW EXECUTE FUNCTION fn_audit_generic();

DROP TRIGGER IF EXISTS trg_qr_audit ON quote_requests;
CREATE TRIGGER trg_qr_audit
    BEFORE UPDATE OR DELETE ON quote_requests
  FOR EACH ROW EXECUTE FUNCTION fn_audit_generic();

DROP TRIGGER IF EXISTS trg_ql_audit ON quote_lists;
CREATE TRIGGER trg_ql_audit
    BEFORE UPDATE OR DELETE ON quote_lists
  FOR EACH ROW EXECUTE FUNCTION fn_audit_generic();

DROP TRIGGER IF EXISTS trg_products_audit ON products;
CREATE TRIGGER trg_products_audit
    BEFORE UPDATE OR DELETE ON products
  FOR EACH ROW EXECUTE FUNCTION fn_audit_generic();

DROP TRIGGER IF EXISTS trg_plr_audit ON product_launch_reviews;
CREATE TRIGGER trg_plr_audit
    BEFORE UPDATE OR DELETE ON product_launch_reviews
  FOR EACH ROW EXECUTE FUNCTION fn_audit_generic();

-- ===========================================
-- 18. 回滚示例函数
-- ===========================================
CREATE OR REPLACE FUNCTION fn_rollback_audit(
  p_table audited_table_enum,
  p_store INT,
  p_user  INT,
  p_to_ts TIMESTAMPTZ
) RETURNS VOID AS $$
DECLARE
rec  audited_history%ROWTYPE;
  cols TEXT;
BEGIN
SELECT * INTO rec
FROM audited_history
WHERE table_name = p_table
  AND (p_store IS NULL OR store_id = p_store)
  AND changed_by = p_user
  AND changed_at <= p_to_ts
ORDER BY changed_at DESC
    LIMIT 1;

IF NOT FOUND THEN
    RAISE NOTICE 'No audit record found';
    RETURN;
END IF;

SELECT string_agg(format('%I = %L', key, value), ', ')
INTO cols
FROM jsonb_each_text(rec.payload)
WHERE key <> 'id';

EXECUTE format('UPDATE %I SET %s WHERE id = %L',
               rec.table_name, cols, rec.record_id);

RAISE NOTICE 'Rolled back % to % at %', rec.table_name, rec.record_id, rec.changed_at;
END;
$$ LANGUAGE plpgsql;
