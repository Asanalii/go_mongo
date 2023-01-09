Create table if not exists directors(
    ID bigserial primary key,
    name text not null,
    surname text not null,
    awards text[] not null
)