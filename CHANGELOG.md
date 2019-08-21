# 2019-08-21

coupons 跟 orders 两个表不再需要了

```
DROP TABLE IF EXISTS coupons;
DROP TABLE IF EXISTS orders;
```

# 2019-07-03

添加了更多支付方式，包括微信支付, 但是需要相关的证书等, config.tpl.yaml 也作了相应的修改。

```
AlTER TABLE users ADD COLUMN pay_method VARCHAR(512) NOT NULL DEFAULT ''

CREATE TABLE IF NOT EXISTS orders (
  order_id         VARCHAR(36) PRIMARY KEY CHECK (order_id ~* '^[0-9a-f-]{36,36}$'),
  trace_id         BIGSERIAL,
  user_id          VARCHAR(36) NOT NULL CHECK (user_id ~* '^[0-9a-f-]{36,36}$'),
  prepay_id        VARCHAR(36) DEFAULT '',
  state            VARCHAR(32) NOT NULL,
  amount           VARCHAR(128) NOT NULL,
  channel          VARCHAR(32) NOT NULL,
  transaction_id   VARCHAR(32) DEFAULT '',
  created_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  paid_at          TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS order_created_paidx ON orders(user_id,state,created_at);


CREATE TABLE IF NOT EXISTS coupons (
	coupon_id         VARCHAR(36) PRIMARY KEY CHECK (coupon_id ~* '^[0-9a-f-]{36,36}$'),
	code              VARCHAR(512) NOT NULL,
	user_id	          VARCHAR(36) NOT NULL CHECK (user_id ~* '^[0-9a-f-]{36,36}$'),
	occupied_by       VARCHAR(36),
	occupied_at       TIMESTAMP WITH TIME ZONE,
	created_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS coupons_codex ON coupons(code);
CREATE INDEX IF NOT EXISTS coupons_occupiedx ON coupons(occupied_by);
CREATE INDEX IF NOT EXISTS coupons_userx ON coupons(user_id);
```

Nginx 修改

```
location ~* \.(js|css|png|jpg|jpeg|gif|ico)$ {
  expires max;
  try_files $uri =404;
}
```

# 2019-06-19

支持全部禁言, 只允许管理员发言

```
CREATE TABLE IF NOT EXISTS properties (
  name               VARCHAR(512) PRIMARY KEY,
  value              VARCHAR(1024) NOT NULL,
  created_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

# 2019-05-29

大群支持引用一小时内的消息

```
ALTER TABLE messages ADD COLUMN quote_message_id VARCHAR(36) NOT NULL DEFAULT '';
ALTER TABLE distributed_messages ADD COLUMN quote_message_id VARCHAR(36) NOT NULL DEFAULT '';
```


# 2019-05-24

我们修改了数据库的表结构，如果是 2019-05-24 号之前部署的，需要更新一个数据库。

改进：优先给活跃用户发送消息

```
ALTER TABLE users ADD COLUMN active_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
CREATE INDEX IF NOT EXISTS users_activex ON users(active_at);
CREATE INDEX IF NOT EXISTS message_shard_status_recipientx ON distributed_messages(shard, status, recipient_id, created_at);
CREATE INDEX IF NOT EXISTS message_status_createdx ON distributed_messages(status, created_at);
CREATE INDEX IF NOT EXISTS message_createdx ON distributed_messages(created_at);
DROP INDEX IF EXISTS message_status;
```
