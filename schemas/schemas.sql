DROP SCHEMA public IF EXISTS CASCADE;  -- TODO: This will break anyone setting their username to "public"
CREATE SCHEMA IF NOT EXISTS %s AUTHORIZATION gorm;
SET search_path TO %s;
DROP TABLE IF EXISTS bot_movement_ledger;
DROP TABLE IF EXISTS bots;
DROP TABLE IF EXISTS mines;
DROP TABLE IF EXISTS users;
CREATE TABLE bots (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    identifier text UNIQUE NOT NULL,
    inventory_count smallint NOT NULL,
    name text NOT NULL
);

CREATE TABLE bot_movement_ledger (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    bot_id bigint references bots(id) NOT NULL,
    time_action_started timestamp with time zone NOT NULL,
    new_x numeric NOT NULL,
    new_y numeric NOT NULL
);

CREATE TABLE mines (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    x numeric NOT NULL,
    y numeric NOT NULL
);

CREATE TABLE users (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    username text NOT NULL UNIQUE,
    password text NOT NULL
);
SET search_path TO public;