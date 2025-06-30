-- ===========================================
-- 客户表（customers）
--    —— 存储系统中的客户基础信息及默认地址/联系人
-- ===========================================
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
-- 订单主表（orders）
--    —— 客户确认报价后生成的正式订单
-- ===========================================
CREATE TABLE orders (
                        id               SERIAL               PRIMARY KEY,      -- 订单ID
                        quote_id         INT      REFERENCES quotes(id),        -- 来源报价单
                        order_number     VARCHAR(50) UNIQUE NOT NULL,           -- 订单编号
                        store_id         INT      REFERENCES stores(id),        -- 门店ID
                        customer_id      INT      NOT NULL REFERENCES customers(id),
    -- 客户ID
                        order_date       DATE    NOT NULL,                      -- 下单日期
                        currency         currency_code_enum NOT NULL DEFAULT 'AUD',
    -- 币种
                        shipping_address VARCHAR(255) NOT NULL,                 -- 最终发货地址
                        total_amount     NUMERIC(14,2),                         -- 合计金额（可冗余）
                        status           order_status_enum NOT NULL DEFAULT 'draft',
    -- 订单状态
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
                             order_id     INT   NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    -- 关联订单
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
                               order_id         INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    -- 来源订单
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
                                    picking_list_id  INT    NOT NULL REFERENCES picking_lists(id) ON DELETE CASCADE,
    -- 关联拣货单
                                    product_id       INT    NOT NULL REFERENCES products(id),-- 产品ID
                                    quantity         INT    NOT NULL,                        -- 数量
                                    location         VARCHAR(100),                           -- 库位（可选）
                                    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()      -- 创建时间
);

CREATE INDEX idx_picking_items_list   ON picking_list_items(picking_list_id);
CREATE INDEX idx_picking_items_product ON picking_list_items(product_id);
