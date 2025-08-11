-- =====================
-- GLOBAL USERS
-- =====================
CREATE TABLE users (
    user_id BIGINT PRIMARY KEY,
    username TEXT,
    avatar TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- =====================
-- GUILDS (per-server settings)
-- =====================
CREATE TABLE guilds (
    guild_id BIGINT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    name TEXT,
    currency_name TEXT DEFAULT 'Coins',
    settings JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- =====================
-- GUILD MEMBERS (economy + cooldowns)
-- =====================
CREATE TABLE guild_members (
    guild_id BIGINT NOT NULL REFERENCES guilds(guild_id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    balance BIGINT DEFAULT 0,
    last_daily TIMESTAMPTZ,
    base_daily_hours INT DEFAULT 24,
    joined_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (guild_id, user_id)
);

-- =====================
-- ROLE-BASED DAILY MODIFIERS
-- =====================
CREATE TABLE role_daily_modifiers (
    guild_id BIGINT NOT NULL REFERENCES guilds(guild_id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL,
    min_hours INT NOT NULL,
    PRIMARY KEY (guild_id, role_id)
);

-- =====================
-- DISABLED COMMANDS (per guild)
-- =====================
CREATE TABLE disabled_commands (
    guild_id BIGINT NOT NULL REFERENCES guilds(guild_id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('command', 'category')),
    PRIMARY KEY (guild_id, name)
);

-- =====================
-- UNIFIED REMINDERS
-- =====================
CREATE TABLE reminders (
    reminder_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    guild_id BIGINT REFERENCES guilds(guild_id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    remind_at TIMESTAMPTZ NOT NULL,
    source TEXT NOT NULL,
    data JSONB,
    sent BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- =====================
-- SCHEDULED MESSAGES (guild/channel announcements)
-- =====================
CREATE TABLE scheduled_messages (
    message_id BIGSERIAL PRIMARY KEY,
    guild_id BIGINT NOT NULL REFERENCES guilds(guild_id) ON DELETE CASCADE,
    channel_id BIGINT NOT NULL,
    content TEXT NOT NULL,
    send_at TIMESTAMPTZ NOT NULL,
    sent BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- =====================
-- TRANSACTIONS (economy audit)
-- =====================
CREATE TABLE transactions (
    transaction_id BIGSERIAL PRIMARY KEY,
    guild_id BIGINT NOT NULL REFERENCES guilds(guild_id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    amount BIGINT NOT NULL,
    balance_before BIGINT,
    balance_after BIGINT,
    reason TEXT,
    meta JSONB,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- =====================
-- INDEXES
-- =====================
CREATE INDEX idx_gm_guild_balance_desc ON guild_members (guild_id, balance DESC);
CREATE INDEX idx_gm_user ON guild_members (user_id);
CREATE INDEX idx_tx_guild_time ON transactions (guild_id, created_at DESC);
CREATE INDEX idx_reminders_due ON reminders (sent, remind_at);
CREATE INDEX idx_scheduled_due ON scheduled_messages (sent, send_at);
