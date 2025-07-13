CREATE TABLE order_request_log (
    id BiGSERIAL PRIMARY KEY,
    idempotency_token text unique not null,
    create_time timestamp default current_timestamp
)