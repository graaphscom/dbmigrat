create table "order"
(
    id    serial primary key,
    buyer integer references "user" (id) not null
);

create table order_item
(
    order_id   integer references "order" (id),
    product_id integer references product (id),
    quantity   integer not null,
    unit_price decimal not null,
    primary key (order_id, product_id)
)