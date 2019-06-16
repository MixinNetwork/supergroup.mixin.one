ALTER TABLE messages ADD COLUMN quote_message_id VARCHAR(36) NOT NULL DEFAULT '';
ALTER TABLE distributed_messages ADD COLUMN quote_message_id VARCHAR(36) NOT NULL DEFAULT '';
