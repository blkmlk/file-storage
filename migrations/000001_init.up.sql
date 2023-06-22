CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE uploaded_files (
    id uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    name VARCHAR(200) NOT NULL UNIQUE,
    hash varchar(200) NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE storages (
    id uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4()
);

CREATE TABLE file_parts (
    id uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    file_id uuid NOT NULL REFERENCES uploaded_files(id) ON DELETE CASCADE ON UPDATE CASCADE,
    seq INTEGER NOT NULL DEFAULT 0,
    storage_id uuid NOT NULL REFERENCES storages(id) ON DELETE CASCADE ON UPDATE CASCADE,
    hash VARCHAR(200) NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);