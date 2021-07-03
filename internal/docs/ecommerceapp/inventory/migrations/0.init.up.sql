create table product
(
    id   serial primary key,
    name varchar not null
);

create table delivery
(
    id                    serial primary key,
    product_id            integer references product (id) not null,
    delivered_units_count bigint                          not null
)