-- ===========================================
-- 切换客户端编码
-- ===========================================
SET client_encoding = 'UTF8';  -- 客户端使用 UTF8 编码

-- ===========================================
-- 0. 通用枚举类型（ENUM）
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
  'draft',             -- 草稿
  'pending_tech',      -- 技术审
  'pending_purchase',  -- 采购审
  'pending_finance',   -- 财务审
  'ready_published',   -- 准备发布
  'published',         -- 已发布
  'rejected',          -- 驳回
  'closed'             -- 关闭
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
-- 1. 新增：供应商表
-- ===========================================
CREATE TABLE suppliers (
                           id          SERIAL PRIMARY KEY,                  -- 供应商ID，自增
                           name        VARCHAR(100) NOT NULL,               -- 供应商名称
                           contact     VARCHAR(100),                        -- 联系人
                           phone       VARCHAR(50),                         -- 电话
                           email       VARCHAR(100),                        -- 邮箱
                           address     VARCHAR(255),                        -- 地址
                           created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),  -- 创建时间
                           updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()   -- 更新时间
);

-- ===========================================
-- 1. 用户与权限（RBAC）
-- ===========================================

-- 用户表
CREATE TABLE users (
                       id            SERIAL       PRIMARY KEY,              -- 用户ID，自增
                       username      VARCHAR(50)  UNIQUE NOT NULL,          -- 登录用户名
                       password_hash VARCHAR(256) NOT NULL,                 -- 密码哈希
                       version       BIGINT       NOT NULL DEFAULT 1,       -- 乐观锁
                       created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),    -- 创建时间
                       updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),    -- 更新时间
                       is_deleted    BOOLEAN      NOT NULL DEFAULT FALSE     -- 软删除标记
);

-- 角色表
CREATE TABLE roles (
                       id   SERIAL       PRIMARY KEY,           -- 角色ID
                       name VARCHAR(50)  UNIQUE NOT NULL        -- 角色名
);

-- 权限表
CREATE TABLE permissions (
                             id   SERIAL         PRIMARY KEY,         -- 权限ID
                             name VARCHAR(100)   UNIQUE NOT NULL      -- 权限名
);

-- 用户与角色关联表
CREATE TABLE user_roles (
                            user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,  -- 用户ID
                            role_id INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,  -- 角色ID
                            PRIMARY KEY(user_id, role_id)                                  -- 联合主键
);

-- 角色与权限关联表
CREATE TABLE role_permissions (
                                  role_id       INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,       -- 角色ID
                                  permission_id INT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE, -- 权限ID
                                  PRIMARY KEY(role_id, permission_id)                                       -- 联合主键
);

-- ===========================================
-- 2. 修改：附件表，支持多类型单据关联
-- ===========================================
CREATE TABLE attachments (
                             id           SERIAL       PRIMARY KEY,           -- 附件ID
                             file_name    VARCHAR(255) NOT NULL,              -- 文件名
                             file_type    VARCHAR(100) NOT NULL,              -- 文件类型
                             file_size    INT,                                -- 文件大小（字节）
                             url          TEXT NOT NULL,                     -- 存储路径或链接
                             uploaded_by  INT REFERENCES users(id),           -- 上传者用户ID
                             uploaded_at  TIMESTAMPTZ NOT NULL DEFAULT now(), -- 上传时间
                             ref_type     VARCHAR(20) NOT NULL,               -- 关联类型：'quote','order','inbound','outbound',...
                             ref_id       INT  NOT NULL                       -- 关联记录ID
);

-- ===========================================
-- 3. 地区与仓库关联（多对多）
-- ===========================================

-- 地区表
CREATE TABLE regions (
                         id          SERIAL       PRIMARY KEY,             -- 地区ID
                         name        VARCHAR(100) NOT NULL UNIQUE,         -- 地区名称
                         created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),   -- 创建时间
                         updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()    -- 更新时间
);

-- 仓库表
CREATE TABLE warehouses (
                            id          SERIAL       PRIMARY KEY,             -- 仓库ID
                            name        VARCHAR(100) NOT NULL UNIQUE,         -- 仓库名称
                            location    VARCHAR(255),                         -- 地址
                            version     BIGINT       NOT NULL DEFAULT 1,      -- 乐观锁
                            created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),   -- 创建时间
                            updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),   -- 更新时间
                            is_deleted  BOOLEAN      NOT NULL DEFAULT FALSE    -- 软删除
);

-- 地区与仓库关联表
CREATE TABLE region_warehouses (
                                   region_id    INT NOT NULL REFERENCES regions(id)    ON DELETE CASCADE,  -- 地区ID
                                   warehouse_id INT NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,  -- 仓库ID
                                   PRIMARY KEY(region_id, warehouse_id)                                  -- 联合主键
);

CREATE INDEX idx_region_wh_region    ON region_warehouses(region_id);
CREATE INDEX idx_region_wh_warehouse ON region_warehouses(warehouse_id);

-- ===========================================
-- 4. 门店与客户
-- ===========================================

-- 门店表
CREATE TABLE stores (
                        id           SERIAL        PRIMARY KEY,          -- 门店ID
                        code         VARCHAR(50)   NOT NULL UNIQUE,      -- 门店编码
                        name         VARCHAR(100)  NOT NULL,             -- 门店名称
                        region_id    INT           NOT NULL REFERENCES regions(id), -- 所属地区
                        address      VARCHAR(255),                       -- 地址
                        manager_id   INT           REFERENCES users(id),  -- 负责人ID
                        version      BIGINT        NOT NULL DEFAULT 1,   -- 乐观锁
                        created_at   TIMESTAMPTZ   NOT NULL DEFAULT now(),-- 创建时间
                        updated_at   TIMESTAMPTZ   NOT NULL DEFAULT now(),-- 更新时间
                        is_deleted   BOOLEAN      NOT NULL DEFAULT FALSE  -- 软删除
);
CREATE INDEX idx_stores_region  ON stores(region_id);
CREATE INDEX idx_stores_manager ON stores(manager_id);

-- 客户表
CREATE TABLE customers (
                           id               SERIAL               PRIMARY KEY,        -- 客户ID，自增
                           store_id         INT      REFERENCES stores(id),           -- 所属门店
                           type             customer_type_enum NOT NULL DEFAULT 'retail',
    -- 客户类型：retail/wholesale/online
                           name             VARCHAR(100) NOT NULL,                   -- 客户公司或个人名称

    -- 默认账单与送货地址快照
                           billing_address  VARCHAR(255) NOT NULL,                   -- 默认账单地址
                           shipping_address VARCHAR(255) NOT NULL,                   -- 默认送货地址

    -- 默认联系人快照
                           contact_name     VARCHAR(100) NOT NULL,                   -- 联系人姓名
                           contact_phone    VARCHAR(50)  NOT NULL,                   -- 联系人电话
                           contact_email    VARCHAR(100),                            -- 联系人邮箱

                           version          BIGINT    NOT NULL DEFAULT 1,            -- 乐观锁版本
                           created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),      -- 创建时间
                           updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),      -- 更新时间
                           is_deleted       BOOLEAN   NOT NULL DEFAULT FALSE         -- 软删除标记
);

CREATE INDEX idx_customers_store  ON customers(store_id);
CREATE INDEX idx_customers_type   ON customers(type);

-- ===========================================
-- 5. 币种与汇率
-- ===========================================

-- 当前汇率表
CREATE TABLE currency_rates (
                                code        currency_code_enum PRIMARY KEY,       -- 币种
                                rate_to_aud NUMERIC(14,6)     NOT NULL,           -- 相对AUD汇率
                                version     BIGINT            NOT NULL DEFAULT 1, -- 乐观锁
                                updated_at  TIMESTAMPTZ        NOT NULL DEFAULT now() -- 更新时间
);

-- 汇率历史表
CREATE TABLE currency_rate_history (
                                       id             SERIAL              PRIMARY KEY,   -- 历史记录ID
                                       code           currency_code_enum  NOT NULL,      -- 币种
                                       rate_to_aud    NUMERIC(14,6)       NOT NULL,      -- 生效汇率
                                       effective_date DATE                NOT NULL       -- 生效日期
);
CREATE INDEX idx_crh_code_date ON currency_rate_history(code,effective_date);

-- ===========================================
-- 6. 产品分类（多级）
-- ===========================================
CREATE TABLE product_categories (
                                    id         SERIAL       PRIMARY KEY,             -- 分类ID
                                    name       VARCHAR(100) UNIQUE NOT NULL,         -- 分类名称
                                    parent_id  INT          REFERENCES product_categories(id) ON DELETE SET NULL, -- 父级ID
                                    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),   -- 创建时间
                                    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now()    -- 更新时间
);

-- ===========================================
-- 7. 产品表
-- ===========================================
CREATE TABLE products (
                          id                   SERIAL               PRIMARY KEY,        -- 产品ID
                          djj_code             VARCHAR(50)          UNIQUE NOT NULL,    -- SKU
                          name_cn              VARCHAR(100)         NOT NULL,           -- 中文名
                          name_en              VARCHAR(100),                             -- 英文名
                          manufacturer         VARCHAR(100),                             -- 厂商
                          manufacturer_code    VARCHAR(100),                             -- 厂家编码
                          supplier             VARCHAR(100),                             -- 供应商
                          model                VARCHAR(100),                             -- 型号
                          category_id          INT REFERENCES product_categories(id),     -- 一级分类ID
                          subcategory_id       INT REFERENCES product_categories(id),     -- 二级分类ID
                          tertiary_category_id INT REFERENCES product_categories(id),     -- 三级分类ID
                          technical_specs      JSONB    DEFAULT '{}'::JSONB,               -- 技术参数
                          specs                TEXT,                                      -- 规格
                          price                NUMERIC(12,2) NOT NULL CHECK(price>=0),   -- 进价
                          rrp_price            NUMERIC(12,2) CHECK(rrp_price>=0),         -- 建议零售价
                          currency             currency_code_enum NOT NULL DEFAULT 'AUD',  -- 币种
                          status               product_status_enum NOT NULL DEFAULT 'draft',  -- 产品状态
                          application_status   application_status_enum NOT NULL DEFAULT 'open',-- 审批状态
                          product_type         product_type_enum   NOT NULL DEFAULT 'others',-- 主分类
                          standard_warranty    VARCHAR(100),                             -- 标准保修期
                          remarks              TEXT,                                      -- 备注
                          marketing_info       TEXT,                                      -- 营销信息
                          training_docs        TEXT,                                      -- 培训资料
                          weight_kg            NUMERIC(10,2) CHECK(weight_kg>=0),         -- 重量(kg)
                          lift_capacity_kg     NUMERIC(10,2) CHECK(lift_capacity_kg>=0),  -- 起重量(kg)
                          lift_height_mm       NUMERIC(10,2) CHECK(lift_height_mm>=0),    -- 起升高度(mm)
                          power_source         VARCHAR(100),                             -- 动力来源
                          other_specs          JSONB    DEFAULT '{}'::JSONB,              -- 扩展字段1
                          extra_info           JSONB    DEFAULT '{}'::JSONB,              -- 扩展字段2
                          metadata             JSONB    DEFAULT '{}'::JSONB,              -- 元数据
                          version              BIGINT   NOT NULL DEFAULT 1,               -- 乐观锁
                          created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),        -- 创建时间
                          updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),        -- 更新时间
                          is_deleted           BOOLEAN  NOT NULL DEFAULT FALSE           -- 软删除
);
CREATE INDEX idx_products_supplier  ON products(supplier);
CREATE INDEX idx_products_category  ON products(category_id);
CREATE INDEX idx_products_rrp_price ON products(rrp_price);

-- ===========================================
-- 新增：公司表（companies）
-- ===========================================
CREATE TABLE companies (
                           id            SERIAL       PRIMARY KEY,                  -- 自增主键
                           code          VARCHAR(50)  UNIQUE NOT NULL,              -- 公司代码（例如 DJJ_PERTH）
                           name          VARCHAR(100) NOT NULL,                     -- 公司名称
                           email         VARCHAR(100),                              -- 联系邮箱
                           phone         VARCHAR(50),                               -- 联系电话
                           website       VARCHAR(255),                              -- 公司官网
                           abn           VARCHAR(20),                               -- Australian Business Number
                           address       VARCHAR(255),                              -- 公司地址
                           logo_base64   TEXT,                                      -- Logo 的 Base64 编码
                           bank_name     VARCHAR(100),                              -- 银行户名
                           bsb           VARCHAR(20),                               -- BSB 号
                           account_no    VARCHAR(50),                               -- 银行账号
                           created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),       -- 记录创建时间
                           updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now()        -- 记录更新时间
);


-- ===========================================
-- 8. 新品上线审核记录表
-- ===========================================
CREATE TABLE product_launch_reviews (
                                        id           SERIAL                PRIMARY KEY,            -- 审核ID
                                        product_id   INT     NOT NULL       REFERENCES products(id) ON DELETE CASCADE, -- 产品ID
                                        stage        review_stage_enum     NOT NULL,               -- 审核阶段
                                        status       approval_status_enum  NOT NULL DEFAULT 'pending',-- 审批结果
                                        comments     TEXT,                                       -- 审批意见
                                        reviewer_id  INT                   REFERENCES users(id),  -- 审核人ID
                                        reviewed_at  TIMESTAMPTZ,                              -- 审核时间
                                        created_at   TIMESTAMPTZ NOT NULL DEFAULT now()        -- 创建时间
);
CREATE INDEX idx_plr_product ON product_launch_reviews(product_id);

-- ===========================================
-- 9. 产品专用附件表
-- ===========================================
CREATE TABLE product_attachments (
                                     id           SERIAL       PRIMARY KEY,               -- 附件ID
                                     product_id   INT          NOT NULL REFERENCES products(id) ON DELETE CASCADE, -- 产品ID
                                     file_name    VARCHAR(255) NOT NULL,                  -- 文件名
                                     file_type    VARCHAR(100) NOT NULL,                  -- 文件类型
                                     file_size    INT,                                    -- 文件大小
                                     url          TEXT         NOT NULL,                  -- 链接
                                     uploaded_by  INT          REFERENCES users(id),      -- 上传者
                                     uploaded_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),    -- 上传时间
                                     metadata     JSONB        DEFAULT '{}'::JSONB        -- 元数据
);
CREATE INDEX idx_prod_attach_product ON product_attachments(product_id);

-- ===========================================
-- 10. 库存与日志表
-- ===========================================
-- ===========================================
-- 入库明细（inventory 入库动作的明细，一般称 inbound_items）
-- ===========================================
CREATE TABLE inbound_items (
                               id                SERIAL        PRIMARY KEY,              -- 明细ID
                               inbound_id        INT   NOT NULL REFERENCES inventory_inbounds(id) ON DELETE CASCADE, -- 入库单ID
                               product_id        INT   NOT NULL REFERENCES products(id),  -- 产品ID
                               received_qty      INT   NOT NULL,                         -- 实收数量
                               unit_price        NUMERIC(12,2) NOT NULL,                 -- 本次入库参考单价
                               remark            TEXT,                                   -- 额外备注
                               created_at        TIMESTAMPTZ NOT NULL DEFAULT now()      -- 记录时间
);

-- ===========================================
-- 出库明细（inventory 出库动作的明细，一般称 outbound_items）
-- ===========================================
CREATE TABLE outbound_items (
                                id                SERIAL        PRIMARY KEY,              -- 明细ID
                                outbound_id       INT   NOT NULL REFERENCES inventory_outbounds(id) ON DELETE CASCADE, -- 出库单ID
                                product_id        INT   NOT NULL REFERENCES products(id),  -- 产品ID
                                shipped_qty       INT   NOT NULL,                         -- 发货数量
                                unit_price        NUMERIC(12,2) NOT NULL,                 -- 本次出库参考单价
                                remark            TEXT,                                   -- 额外备注
                                created_at        TIMESTAMPTZ NOT NULL DEFAULT now()      -- 记录时间
);

-- ===========================================
-- 库存表：仍旧关联 products + warehouse
-- ===========================================
CREATE TABLE inventory (
                           id             SERIAL       PRIMARY KEY,                 -- 库存ID
                           product_id     INT  NOT NULL REFERENCES products(id),    -- 产品
                           warehouse_id   INT  NOT NULL REFERENCES warehouses(id), -- 仓库
                           on_hand        INT  NOT NULL DEFAULT 0,                  -- 在库数量
                           reserved       INT  NOT NULL DEFAULT 0,                  -- 预留数量
                           updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),       -- 更新时间
--                            UNIQUE(product_id, warehouse_id)                        -- 单产品+仓库唯一
);

CREATE INDEX idx_inventory_product   ON inventory(product_id);
CREATE INDEX idx_inventory_warehouse ON inventory(warehouse_id);

CREATE TABLE inventory_logs (
                                id           SERIAL       PRIMARY KEY,                    -- 日志ID
                                inventory_id INT          NOT NULL REFERENCES inventory(id), -- 库存ID
                                change_type  VARCHAR(20)  NOT NULL,                       -- 变动类型
                                quantity     INT          NOT NULL,                       -- 变动数量
                                operator_id  INT NOT NULL REFERENCES users(id),           -- 操作人ID
                                remark       TEXT,                                         -- 备注
                                created_at   TIMESTAMPTZ  NOT NULL DEFAULT now()           -- 记录时间
);
CREATE INDEX idx_invlog_inventory ON inventory_logs(inventory_id);

-- ===========================================
-- 订单主表（orders）
--    —— 客户确认报价后生成的正式订单
-- ===========================================
CREATE TABLE orders (
                        id               SERIAL               PRIMARY KEY,      -- 订单ID
                        quote_id         INT      REFERENCES quotes(id),        -- 来源报价单
                        order_number     VARCHAR(50) UNIQUE NOT NULL,           -- 订单编号
                        store_id         INT      REFERENCES stores(id),        -- 门店ID
                        customer_id      INT      NOT NULL REFERENCES customers(id),-- 客户ID
                        order_date       DATE    NOT NULL,                      -- 下单日期
                        currency         currency_code_enum NOT NULL DEFAULT 'AUD',-- 币种
                        shipping_address VARCHAR(255) NOT NULL,                 -- 最终发货地址
                        total_amount     NUMERIC(14,2),                         -- 合计金额（可冗余）
                        status           order_status_enum NOT NULL DEFAULT 'draft', -- 订单状态
                        created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),    -- 创建时间
                        updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()     -- 更新时间
);

CREATE INDEX idx_orders_quote    ON orders(quote_id);
CREATE INDEX idx_orders_store    ON orders(store_id);
CREATE INDEX idx_orders_customer ON orders(customer_id);
CREATE INDEX idx_orders_status   ON orders(status);

-- ===========================================
-- 订单明细表（order_items）
--    —— 每条订单的产品/服务行项
-- ===========================================
CREATE TABLE order_items (
                             id           SERIAL        PRIMARY KEY,                   -- 明细ID
                             order_id     INT   NOT NULL REFERENCES orders(id) ON DELETE CASCADE,-- 关联订单
                             product_id   INT   NOT NULL REFERENCES products(id),  -- 产品ID
                             quantity     INT   NOT NULL,                          -- 数量
                             unit_price   NUMERIC(12,2) NOT NULL                   -- 锁定时单价
);

CREATE INDEX idx_order_items_order   ON order_items(order_id);
CREATE INDEX idx_order_items_product ON order_items(product_id);

-- ===========================================
-- 拣货单主表（picking_lists）
--    —— 根据订单生成，用于仓库拣货
-- ===========================================
CREATE TABLE picking_lists (
                               id               SERIAL               PRIMARY KEY,      -- 拣货单ID
                               picking_number   VARCHAR(50) UNIQUE NOT NULL,           -- 拣货单编号
                               order_id         INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,-- 来源订单
                               delivery_address VARCHAR(255) NOT NULL,                 -- 复制自 order.shipping_address
                               status           VARCHAR(20) NOT NULL DEFAULT 'draft',  -- 拣货状态（draft/picked/done 等）
                               created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),    -- 创建时间
                               updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()     -- 更新时间
);
CREATE INDEX idx_picking_lists_order  ON picking_lists(order_id);
CREATE INDEX idx_picking_lists_status ON picking_lists(status);

-- ===========================================
-- 拣货单明细表（picking_list_items）
--    —— 每条拣货单的行项目
-- ===========================================
CREATE TABLE picking_list_items (
                                    id               SERIAL PRIMARY KEY,                    -- 明细ID
                                    picking_list_id  INT    NOT NULL REFERENCES picking_lists(id) ON DELETE CASCADE,-- 关联拣货单
                                    product_id       INT    NOT NULL REFERENCES products(id),-- 产品ID
                                    quantity         INT    NOT NULL,                        -- 数量
                                    location         VARCHAR(100),                           -- 库位（可选）
                                    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()      -- 创建时间
);

CREATE INDEX idx_picking_items_list   ON picking_list_items(picking_list_id);
CREATE INDEX idx_picking_items_product ON picking_list_items(product_id);


-- ===========================================
-- 12. 发票与付款表
-- ===========================================
CREATE TABLE invoices (
                          id             SERIAL       PRIMARY KEY,               -- 发票ID
                          order_id       INT          NOT NULL REFERENCES orders(id), -- 订单ID
                          invoice_number VARCHAR(50)  UNIQUE NOT NULL,            -- 发票编号
                          issue_date     DATE         NOT NULL,                   -- 开票日期
                          total_amount   NUMERIC(14,2)                             -- 总金额
);
CREATE INDEX idx_invoices_order ON invoices(order_id);

CREATE TABLE payments (
                          id       SERIAL       PRIMARY KEY,                     -- 付款ID
                          order_id INT          NOT NULL REFERENCES orders(id),   -- 订单ID
                          amount   NUMERIC(12,2) NOT NULL,                        -- 付款金额
                          paid_at  TIMESTAMPTZ  NOT NULL DEFAULT now()            -- 付款时间
);
CREATE INDEX idx_payments_order ON payments(order_id);

-- ===========================================
-- 13. PD 检查 & 发货表
-- ===========================================
CREATE TABLE pd_inspections (
                                id         SERIAL       PRIMARY KEY,                   -- 检查ID
                                order_id   INT          NOT NULL REFERENCES orders(id) ON DELETE CASCADE, -- 订单ID
                                passed     BOOLEAN      NOT NULL,                      -- 是否通过
                                created_at TIMESTAMPTZ  NOT NULL DEFAULT now()          -- 检查时间
);

CREATE TABLE shipments (
                           id         SERIAL       PRIMARY KEY,                   -- 发货ID
                           order_id   INT          NOT NULL REFERENCES orders(id) ON DELETE CASCADE, -- 订单ID
                           shipped_at TIMESTAMPTZ  NOT NULL DEFAULT now()          -- 发货时间
);

-- ===========================================
-- 14. 客户活动、提醒、审批日志表
-- ===========================================
CREATE TABLE customer_activities (
                                     id            SERIAL       PRIMARY KEY,                -- 活动ID
                                     customer_id   INT          NOT NULL REFERENCES customers(id), -- 客户ID
                                     activity_type VARCHAR(50),                              -- 活动类型
                                     created_at    TIMESTAMPTZ  NOT NULL DEFAULT now()        -- 活动时间
);

CREATE TABLE reminders (
                           id         SERIAL       PRIMARY KEY,                   -- 提醒ID
                           ref_type   VARCHAR(20),                                 -- 关联类型
                           ref_id     INT,                                         -- 关联ID
                           remind_at  TIMESTAMPTZ  NOT NULL,                       -- 提醒时间
                           message    TEXT                                          -- 提醒消息
);

CREATE TABLE approval_logs (
                               id         SERIAL       PRIMARY KEY,                   -- 审批日志ID
                               ref_type   VARCHAR(20),                                 -- 关联类型
                               ref_id     INT,                                         -- 关联ID
                               result     VARCHAR(20),                                 -- 审批结果
                               created_at TIMESTAMPTZ NOT NULL DEFAULT now()            -- 记录时间
);

-- ===========================================
-- 15. 报价申请与报价单表
-- ===========================================
-- ===========================================
-- 报价单主表（quotes）
--    —— 客户下单前，销售填写的报价
-- ===========================================
CREATE TABLE quotes (
                        id             SERIAL               PRIMARY KEY,           -- 报价单ID
                        store_id       INT      REFERENCES stores(id),            -- 门店ID
                        customer_id    INT      NOT NULL REFERENCES customers(id), -- 客户ID
                        quote_number   VARCHAR(50) NOT NULL UNIQUE,               -- 报价编号
                        sales_rep      VARCHAR(100) NOT NULL,                     -- 销售代表
                        quote_date     DATE      NOT NULL,                        -- 报价日期
                        currency       currency_code_enum NOT NULL DEFAULT 'AUD', -- 币种
                        sub_total      NUMERIC(14,2)       NOT NULL,               -- 小计
                        gst_total      NUMERIC(14,2)       NOT NULL,               -- GST 金额
                        total_amount   NUMERIC(14,2)       NOT NULL,               -- 总金额
                                                                                   -- 快照客户地址，免得后续变更影响历史
                        billing_address  VARCHAR(255) NOT NULL,                   -- 账单地址
                        shipping_address VARCHAR(255) NOT NULL,                   -- 发货地址
                        remarks        TEXT,                                      -- 备注
                        warranty_notes TEXT,                                      -- 保修及特殊备注
                        status         approval_status_enum DEFAULT 'pending',    -- 报价审批状态
                        created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),        -- 创建时间
                        updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()         -- 更新时间
);

CREATE INDEX idx_quotes_store    ON quotes(store_id);
CREATE INDEX idx_quotes_customer ON quotes(customer_id);
CREATE INDEX idx_quotes_status   ON quotes(status);

-- ===========================================
-- 报价单明细表（quote_items）
--    —— 每条报价的产品/服务行项
-- ===========================================
CREATE TABLE quote_items (
                             id           SERIAL       PRIMARY KEY,                  -- 明细ID
                             quote_id     INT    NOT NULL REFERENCES quotes(id) ON DELETE CASCADE,
    -- 关联报价单
                             product_id   INT    REFERENCES products(id),            -- 产品ID（可选）
                             description  TEXT         NOT NULL,                     -- 描述
                             quantity     INT          NOT NULL,                     -- 数量
                             unit         VARCHAR(20)  NOT NULL,                     -- 单位
                             unit_price   NUMERIC(12,2) NOT NULL,                    -- 单价
                             discount     NUMERIC(12,2) DEFAULT 0,                   -- 折扣
                             total_price  NUMERIC(14,2) NOT NULL,                    -- 金额（含折扣后）
                             goods_nature goods_nature_enum DEFAULT 'contract',      -- 货物性质
                             created_at   TIMESTAMPTZ NOT NULL DEFAULT now()         -- 创建时间
);

CREATE INDEX idx_quote_items_quote ON quote_items(quote_id);
CREATE INDEX idx_quote_items_product ON quote_items(product_id);

-- ===========================================
-- 16. 统一审计历史表 & 枚举更新
-- ===========================================
CREATE TYPE audited_table_enum AS ENUM (
  'inventory', 'orders', 'quote_requests', 'quote_lists',
  'products', 'product_launch_reviews'
);

CREATE TABLE audited_history (
                                 history_id BIGSERIAL            PRIMARY KEY,            -- 历史ID
                                 table_name audited_table_enum   NOT NULL,               -- 表名
                                 record_id  INT                  NOT NULL,               -- 记录ID
                                 store_id   INT,                                         -- 门店ID
                                 changed_by INT                  NOT NULL REFERENCES users(id), -- 操作人ID
                                 operation  VARCHAR(10)          NOT NULL,               -- 操作类型（INSERT/UPDATE/DELETE）
                                 payload    JSONB                NOT NULL,               -- 变更前数据
                                 changed_at TIMESTAMPTZ          NOT NULL DEFAULT now()  -- 变更时间
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

-- 在各表上注册审计触发器
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
