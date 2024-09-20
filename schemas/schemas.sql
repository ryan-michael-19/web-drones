DROP TABLE IF EXISTS public.bot_actions;
DROP TABLE IF EXISTS public.bots;
DROP TABLE IF EXISTS public.mines;
CREATE TABLE public.bots (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    identifier text UNIQUE,
    name text
    -- status text,
    -- x numeric,
    -- y numeric
);

CREATE TABLE public.bot_actions (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    bot_id bigint references bots(id),
    time_action_started timestamp with time zone,
    new_x numeric,
    new_y numeric
);

CREATE TABLE public.mines (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    x numeric,
    y numeric
);