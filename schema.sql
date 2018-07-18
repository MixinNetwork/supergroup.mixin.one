CREATE TABLE users (
    user_id           STRING(36) NOT NULL,
    identity_number   INT64 NOT NULL,
    full_name         STRING(512) NOT NULL,
    access_token      STRING(512) NOT NULL,
    avatar_url        STRING(1024) NOT NULL,
    trace_id          STRING(36) NOT NULL,
    state             STRING(128) NOT NULL,
    subscribed_at     TIMESTAMP NOT NULL,
) PRIMARY KEY(user_id);

CREATE INDEX users_by_subscribed ON users(subscribed_at) STORING(full_name);


CREATE TABLE messages (
    message_id            STRING(36) NOT NULL,
    user_id               STRING(36) NOT NULL,
    category              STRING(512) NOT NULL,
    data                  BYTES(MAX) NOT NULL,
    created_at            TIMESTAMP NOT NULL,
    updated_at            TIMESTAMP NOT NULL,
    state                 STRING(128) NOT NULL,
    last_distribute_at    TIMESTAMP NOT NULL,
) PRIMARY KEY(message_id);

CREATE INDEX messages_by_state_updated ON messages(state, updated_at);


CREATE TABLE distributed_messages (
    message_id            STRING(36) NOT NULL,
    conversation_id       STRING(36) NOT NULL,
    recipient_id          STRING(36) NOT NULL,
    user_id               STRING(36) NOT NULL,
    category              STRING(512) NOT NULL,
    data                  BYTES(MAX) NOT NULL,
    created_at            TIMESTAMP NOT NULL,
    updated_at            TIMESTAMP NOT NULL,
) PRIMARY KEY(message_id);

CREATE INDEX distributed_messages_by_updated ON distributed_messages(updated_at);


CREATE TABLE properties (
    key         STRING(512) NOT NULL,
    value       STRING(8192) NOT NULL,
    updated_at  TIMESTAMP NOT NULL,
) PRIMARY KEY(key);
