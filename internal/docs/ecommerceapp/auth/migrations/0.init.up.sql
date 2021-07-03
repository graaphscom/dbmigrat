create table "user"
(
    id       serial primary key,
    username varchar not null unique,
    password varchar not null
);

create table access_token
(
    user_id  integer references "user" (id) not null,
    token    varchar                        not null unique,
    valid_to timestamp                      not null
)