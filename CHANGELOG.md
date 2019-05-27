# 2019-05-24

我们修改了数据库的表结构，如果是 2019-05-24 号之前部署的，需要更新一个数据库。

改进：优先给活跃用户发送消息

```
alter table users add column active_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
CREATE INDEX IF NOT EXISTS users_activex ON users(active_at);
CREATE INDEX IF NOT EXISTS message_shard_status_recipientx ON distributed_messages(shar+d, status, recipient_id, created_at);
CREATE INDEX IF NOT EXISTS message_status_createdx ON distributed_messages(status, crea+ted_at);
CREATE INDEX IF NOT EXISTS message_createdx ON distributed_messages(created_at);
DROP INDEX IF EXISTS message_status;
```
