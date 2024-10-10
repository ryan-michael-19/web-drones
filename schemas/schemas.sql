DROP TABLE IF EXISTS public.bot_actions; -- legacy
DROP TABLE IF EXISTS public.bot_movement_ledger;
DROP TABLE IF EXISTS public.bots;
DROP TABLE IF EXISTS public.mines;
DROP TABLE IF EXISTS public.users;
CREATE TABLE public.bots (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    identifier text UNIQUE NOT NULL,
    inventory_count smallint NOT NULL,
    name text NOT NULL
);

CREATE TABLE public.bot_movement_ledger (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    bot_id bigint references bots(id) NOT NULL,
    time_action_started timestamp with time zone NOT NULL,
    new_x numeric NOT NULL,
    new_y numeric NOT NULL
);

CREATE TABLE public.mines (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    x numeric NOT NULL,
    y numeric NOT NULL
);

CREATE TABLE public.users (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    username text NOT NULL UNIQUE,
    password text NOT NULL
)