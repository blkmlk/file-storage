CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

DROP TYPE IF EXISTS file_status;
CREATE TYPE file_status AS ENUM('created', 'uploaded');

CREATE TABLE files (
    id uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    name VARCHAR(200) NULL UNIQUE,
    content_type varchar(255) NOT NULL DEFAULT '',
    hash varchar(200) NOT NULL DEFAULT '',
    size int NOT NULL DEFAULT 0,
    status file_status NOT NULL DEFAULT 'created'::file_status,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE storages (
    id uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    host varchar(100) NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE file_parts (
    id uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    file_id uuid NOT NULL REFERENCES files(id) ON DELETE CASCADE ON UPDATE CASCADE,
    remote_id varchar(255) NOT NULL DEFAULT '',
    seq INTEGER NOT NULL DEFAULT 0,
    storage_id uuid NOT NULL REFERENCES storages(id) ON DELETE CASCADE ON UPDATE CASCADE,
    size INTEGER NOT NULL DEFAULT 0,
    hash VARCHAR(200) NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX file_parts_file_id_idx ON file_parts(file_id uuid_ops, seq);