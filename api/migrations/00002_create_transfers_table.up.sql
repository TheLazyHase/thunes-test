-- for the MVP orders are executed immediatly every time, but we may want to have a creation_date and execution_date to allow people to book a
-- amount is BIGINT because you credibly could have an order of more than 2 billions - maybe not in the near future, but also it's not a widly unrealistic amount.

CREATE TABLE IF NOT EXISTS transfers (
    id BIGSERIAL PRIMARY KEY,
    source_id BIGSERIAL NOT NULL,
    destination_id BIGSERIAL NOT NULL,
    amount BIGINT NOT NULL DEFAULT 0,
    execution_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

ALTER TABLE users
    ADD balance BIGINT NOT NULL DEFAULT 0;
