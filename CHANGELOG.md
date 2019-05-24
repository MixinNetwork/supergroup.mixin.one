# 2019-05-24

我们修改了数据库的表结构，如果是 2019-05-24 号之前部署的，需要更新一个数据库。

改进：优先给活跃用户发送消息

```
alter table users add column active_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
alter table distributed_messages add column active_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
CREATE INDEX IF NOT EXISTS message_shard_status_activex ON distributed_messages(shard, status, active_at, created_at);
```
