# 2021-09-15
users 表添加 authorization_id, scope 两个字段, 用来支持 ed25519 签名

# 2021-03-18
利用 go1.6 的 embed 把 production 配置一起打包, 最新 config.tpl.yaml 

# 2020-08-12
配置文件: config.tpl.yaml 
增加：api_root, blaze_root 可以自动切换域名


# 2020-06-02

配置文件: config.tpl.yaml 
增加: message_tips_suspended 6 天不活跃的用户发送系统提示消息，并不再推送消息
users 表索引修改: CREATE INDEX IF NOT EXISTS users_subscribed_activex ON users(subscribed_at, active_at);

# 2019-11-05

配置文件: config.tpl.yaml 
删除: prohibited_message
增加: limit_message_number, 跟原有的 limit_message_duration 可以控制发消息频率

# 2019-09-06

大群发消息配置修改 `limit_message_frequency` 改成 `limit_message_duration`, 值为秒

# 2019-08-27

大群权限修改，需要添加 MESSAGES:REPRESENT, 添加了新的消息模板 `message_reward_memo`

# 2019-08-26

config.tpl.yaml 添加了新消息模版 `message_reward_label`

支持打赏，添加了两个表
```
CREATE TABLE IF NOT EXISTS broadcasters (
	user_id	          VARCHAR(36) PRIMARY KEY CHECK (user_id ~* '^[0-9a-f-]{36,36}$'),
	created_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	updated_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS broadcasters_updatedx ON broadcasters(updated_at);


CREATE TABLE IF NOT EXISTS rewards (
	reward_id           VARCHAR(36) PRIMARY KEY CHECK (reward_id ~* '^[0-9a-f-]{36,36}$'),
	user_id	            VARCHAR(36) NOT NULL CHECK (user_id ~* '^[0-9a-f-]{36,36}$'),
	recipient_id        VARCHAR(36) NOT NULL CHECK (recipient_id ~* '^[0-9a-f-]{36,36}$'),
	asset_id            VARCHAR(36) NOT NULL CHECK (asset_id ~* '^[0-9a-f-]{36,36}$'),
	amount              VARCHAR(128) NOT NULL,
	paid_at             TIMESTAMP WITH TIME ZONE NOT NULL,
	created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS rewards_paidx ON rewards(paid_at);
```


# 2019-08-21

coupons 跟 orders 两个表不再需要了

```
DROP TABLE IF EXISTS coupons;
DROP TABLE IF EXISTS orders;
```

# 2019-07-03

添加了更多支付方式，包括微信支付, 但是需要相关的证书等, config.tpl.yaml 也作了相应的修改。

```
AlTER TABLE users ADD COLUMN pay_method VARCHAR(512) NOT NULL DEFAULT '';

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
