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


CREATE TABLE packets (
	packet_id         STRING(36) NOT NULL,
	user_id	          STRING(36) NOT NULL,
	asset_id          STRING(36) NOT NULL,
	amount            STRING(128) NOT NULL,
	greeting          STRING(36) NOT NULL,
	total_count       INT64 NOT NULL,
	remaining_count   INT64 NOT NULL,
	remaining_amount  STRING(128) NOT NULL,
	state             STRING(36) NOT NULL,
	created_at        TIMESTAMP NOT NULL,
) PRIMARY KEY(packet_id);

CREATE INDEX packets_by_state_created ON packets(state, created_at);


CREATE TABLE participants (
	packet_id         STRING(36) NOT NULL,
	user_id	          STRING(36) NOT NULL,
	amount            STRING(128) NOT NULL,
	created_at        TIMESTAMP NOT NULL,
	paid_at           TIMESTAMP,
) PRIMARY KEY(packet_id, user_id),
INTERLEAVE IN PARENT packets ON DELETE CASCADE;

CREATE INDEX participants_by_created_paid ON participants(created_at, paid_at) STORING(amount);


CREATE TABLE assets (
	asset_id         STRING(36) NOT NULL,
	symbol           STRING(512) NOT NULL,
	name             STRING(512) NOT NULL,
	icon_url         STRING(1024) NOT NULL,
	price_btc        STRING(128) NOT NULL,
	price_usd        STRING(128) NOT NULL,
) PRIMARY KEY(asset_id);
