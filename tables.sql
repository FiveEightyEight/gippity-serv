DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS user_metadata CASCADE;
DROP TABLE IF EXISTS chats CASCADE;
DROP TABLE IF EXISTS messages CASCADE;
DROP TABLE IF EXISTS ai_models CASCADE;
DROP TABLE IF EXISTS chat_ai_models CASCADE;
DROP TABLE IF EXISTS user_preferences CASCADE;

-- Enable the uuid-ossp extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    pk SERIAL PRIMARY KEY,
    id UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_login TIMESTAMPTZ,
    is_active BOOLEAN DEFAULT TRUE
);

CREATE INDEX idx_users_id ON users(id);

-- User metadata table
CREATE TABLE user_metadata (
    user_id UUID PRIMARY KEY REFERENCES users(id),
    preferred_language VARCHAR(10),
    timezone VARCHAR(50),
    interests TEXT[],
    profession VARCHAR(100),
    education_level VARCHAR(50),
    birth_year INTEGER,
    country VARCHAR(50),
    last_updated TIMESTAMPTZ DEFAULT NOW()
);

-- Chats table
CREATE TABLE chats (
    pk SERIAL PRIMARY KEY,
    id UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    title VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_updated TIMESTAMPTZ DEFAULT NOW(),
    is_archived BOOLEAN DEFAULT FALSE,
    ai_model_version VARCHAR(20)
);

CREATE INDEX idx_chats_id ON chats(id);

-- Messages table
CREATE TABLE messages (
    pk SERIAL PRIMARY KEY,
    id UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
    chat_id UUID REFERENCES chats(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    role VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    is_edited BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_messages_id ON messages(id);

-- AI Models table
CREATE TABLE ai_models (
    pk SERIAL PRIMARY KEY,
    id UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
    name VARCHAR(50) UNIQUE NOT NULL,
    version VARCHAR(20) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE
);

CREATE INDEX idx_ai_models_id ON ai_models(id);

-- Chat-AI Model association table
CREATE TABLE chat_ai_models (
    chat_id UUID REFERENCES chats(id),
    ai_model_id UUID REFERENCES ai_models(id),
    PRIMARY KEY (chat_id, ai_model_id)
);

-- User preferences table
CREATE TABLE user_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id),
    default_ai_model UUID REFERENCES ai_models(id),
    theme VARCHAR(20) DEFAULT 'light',
    message_display_count INTEGER DEFAULT 50,
    notifications_enabled BOOLEAN DEFAULT TRUE
);

-- Add foreign key constraint for chats in users table
ALTER TABLE users ADD COLUMN last_chat_id UUID REFERENCES chats(id);