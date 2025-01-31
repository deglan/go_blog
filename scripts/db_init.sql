CREATE EXTENSION IF NOT EXISTS citext;

DO
$$
BEGIN
    IF NOT EXISTS (
        SELECT FROM pg_database WHERE datname = 'social'
    ) THEN
        CREATE DATABASE social;
    END IF;
END
$$;