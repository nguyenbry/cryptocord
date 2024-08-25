-- +goose Up
CREATE TABLE
  test (id SERIAL PRIMARY KEY, name TEXT NOT NULL);

-- +goose Down
DROP TABLE test;