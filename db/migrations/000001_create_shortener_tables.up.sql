CREATE TABLE IF NOT EXISTS users
(
    id bigint PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY
);

CREATE TABLE IF NOT EXISTS urls_id
(
    url_id character varying PRIMARY KEY,
    url character varying NOT NULL,
    user_id bigint NOT NULL,
    deleted boolean DEFAULT false,
    CONSTRAINT user_id FOREIGN KEY (user_id)
        REFERENCES users (id) MATCH SIMPLE
);
CREATE UNIQUE INDEX IF NOT EXISTS original_url ON urls_id (url);