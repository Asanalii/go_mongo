Create table if not exists directors(
    ID bigserial primary key,
    name text not null,
    surname text not null,
    DOB date not null
)