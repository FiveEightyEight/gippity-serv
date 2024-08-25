DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS user_metadata CASCADE;
DROP TABLE IF EXISTS chats CASCADE;
DROP TABLE IF EXISTS messages CASCADE;
DROP TABLE IF EXISTS ai_models CASCADE;
DROP TABLE IF EXISTS chat_ai_models CASCADE;
DROP TABLE IF EXISTS user_preferences CASCADE;

-- Users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_login TIMESTAMPTZ,
    is_active BOOLEAN DEFAULT TRUE
);

-- User metadata table
CREATE TABLE user_metadata (
    user_id INTEGER PRIMARY KEY REFERENCES users(id),
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
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    title VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_updated TIMESTAMPTZ DEFAULT NOW(),
    is_archived BOOLEAN DEFAULT FALSE
);

-- Messages table
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    chat_id INTEGER REFERENCES chats(id),
    user_id INTEGER REFERENCES users(id),
    role VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    is_edited BOOLEAN DEFAULT FALSE
);

-- AI Models table
CREATE TABLE ai_models (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    version VARCHAR(20) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE
);

-- Chat-AI Model association table
CREATE TABLE chat_ai_models (
    chat_id INTEGER REFERENCES chats(id),
    ai_model_id INTEGER REFERENCES ai_models(id),
    PRIMARY KEY (chat_id, ai_model_id)
);

-- User preferences table
CREATE TABLE user_preferences (
    user_id INTEGER PRIMARY KEY REFERENCES users(id),
    default_ai_model INTEGER REFERENCES ai_models(id),
    theme VARCHAR(20) DEFAULT 'light',
    message_display_count INTEGER DEFAULT 50,
    notifications_enabled BOOLEAN DEFAULT TRUE
);

-- Add foreign key constraint for chats in users table
ALTER TABLE users ADD COLUMN last_chat_id INTEGER REFERENCES chats(id);
