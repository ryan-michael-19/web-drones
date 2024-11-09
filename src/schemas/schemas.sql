DROP TABLE IF EXISTS bot_movement_ledger;
DROP TABLE IF EXISTS bots;
DROP TABLE IF EXISTS mines;
DROP TABLE IF EXISTS users;

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL
);
CREATE TABLE bots (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    user_id BIGINT REFERENCES users(id) NOT NULL, 
    identifier TEXT UNIQUE NOT NULL,
    inventory_count SMALLINT NOT NULL,
    name text NOT NULL
);

CREATE TABLE bot_movement_ledger (
    id bigserial PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    bot_id BIGINT REFERENCES bots(id) NOT NULL,
    user_id BIGINT REFERENCES users(id) NOT NULL, 
    time_action_started TIMESTAMP WITH TIME ZONE NOT NULL,
    new_x NUMERIC NOT NULL,
    new_y NUMERIC NOT NULL
);

CREATE TABLE mines (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    user_id BIGINT REFERENCES users(id) NOT NULL, 
    x NUMERIC NOT NULL,
    y NUMERIC NOT NULL
);

