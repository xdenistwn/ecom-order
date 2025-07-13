CREATE TABLE orders (
    id BiGSERIAL PRIMARY KEY,
    user_id bigint not null,
    amount numeric not null,
    total_qty integer not null,
    payment_method varchar(50),
    shipping_address text,
    status integer not null,
    order_detail_id bigint references order_detail(id),
    create_time timestamp default current_timestamp,
    update_time timestamp default current_timestamp
)