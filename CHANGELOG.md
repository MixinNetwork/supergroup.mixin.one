# 2019-06-19

支持全部禁言, 只允许管理员发言

CREATE TABLE IF NOT EXISTS properties (
  name               VARCHAR(512) PRIMARY KEY,
  value              VARCHAR(1024) NOT NULL,
  created_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

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
