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
