CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(1024) NOT NULL UNIQUE,
    password VARCHAR(1024) NOT NULL,
    picture VARCHAR(1024) NULL,
    email_verified BOOLEAN NOT NULL DEFAULT 'false',
    auth_providers TEXT[] NOT NULL DEFAULT ARRAY['normal'],
    token_hash VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT unique_email UNIQUE (email)
);