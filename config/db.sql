-- 0. 定义枚举类型
CREATE TYPE txn_type AS ENUM ('IN','OUT','SALE');
CREATE TYPE order_status AS ENUM ('PENDING','SHIPPED','COMPLETED');

-- 1. 系统用户表（操作人员）
CREATE TABLE IF NOT EXISTS warehouse_user (
                                              id             SERIAL PRIMARY KEY,
                                              username       VARCHAR(50)  NOT NULL UNIQUE,
    password_hash  VARCHAR(255) NOT NULL,        -- 存储 bcrypt/Scrypt/Argon2 哈希
    role           VARCHAR(20)  NOT NULL,        -- e.g. 'admin','manager','staff'
    full_name      VARCHAR(100),
    created_at     TIMESTAMPTZ DEFAULT NOW()
    );

-- 2. 产品表
CREATE TABLE IF NOT EXISTS product (
                                       id               SERIAL PRIMARY KEY,
                                       model_name       VARCHAR(100) NOT NULL,
    photo_url        TEXT,
    cost_price       NUMERIC(12,2) NOT NULL,
    sale_price       NUMERIC(12,2) NOT NULL,
    target_customer  VARCHAR(100),
    barcode          VARCHAR(50) UNIQUE,       -- “PROD-000123” 样式
    barcode_url      TEXT,                     -- 条码图片访问地址
    created_at       TIMESTAMPTZ DEFAULT NOW()
    );

-- 3. 当前库存
CREATE TABLE IF NOT EXISTS inventory (
                                         product_id   INT PRIMARY KEY
                                         REFERENCES product(id)
    ON DELETE CASCADE,
    actual_qty   INT NOT NULL DEFAULT 0,
    locked_qty   INT NOT NULL DEFAULT 0
    );

-- 4. 库存事务日志（入库/出库/销售）
CREATE TABLE IF NOT EXISTS inventory_transaction (
                                                     id           BIGSERIAL PRIMARY KEY,
                                                     product_id   INT         NOT NULL
                                                     REFERENCES product(id),
    type         txn_type    NOT NULL,         -- IN, OUT, SALE
    quantity     INT         NOT NULL,
    user_id      INT,                         -- 操作人员
    note         TEXT,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_invtxn_user FOREIGN KEY (user_id) REFERENCES warehouse_user(id)
    );

-- 5. 客户表（业务客户）
CREATE TABLE IF NOT EXISTS customer (
                                        id           SERIAL PRIMARY KEY,
                                        name         VARCHAR(100) NOT NULL,
    contact      VARCHAR(100),
    address      VARCHAR(255),
    created_at   TIMESTAMPTZ DEFAULT NOW()
    );

-- 6. 销售订单
CREATE TABLE IF NOT EXISTS sales_order (
                                           id            SERIAL PRIMARY KEY,
                                           order_number  VARCHAR(50)   NOT NULL UNIQUE,
    customer_id   INT           NOT NULL
    REFERENCES customer(id),
    status        order_status  NOT NULL DEFAULT 'PENDING',
    total_amount  NUMERIC(12,2) NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ   DEFAULT NOW(),
    created_by    INT,                            -- 下单人
    shipped_by    INT,                            -- 发货人
    CONSTRAINT fk_order_created_by FOREIGN KEY (created_by) REFERENCES warehouse_user(id),
    CONSTRAINT fk_order_shipped_by FOREIGN KEY (shipped_by) REFERENCES warehouse_user(id)
    );

-- 7. 订单明细
CREATE TABLE IF NOT EXISTS sales_order_item (
                                                id           BIGSERIAL PRIMARY KEY,
                                                order_id     INT         NOT NULL
                                                REFERENCES sales_order(id)
    ON DELETE CASCADE,
    product_id   INT         NOT NULL
    REFERENCES product(id),
    quantity     INT         NOT NULL,
    unit_price   NUMERIC(12,2) NOT NULL
    );
