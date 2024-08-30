-- +goose Up
CREATE TABLE
  jobs (id UUID PRIMARY KEY, url TEXT UNIQUE NOT NULL);

-- +goose Down
DROP TABLE jobs;