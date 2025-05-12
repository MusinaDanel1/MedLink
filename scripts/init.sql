-- Создание базы данных
CREATE DATABASE telemed;

-- Соединение с базой данных telemed (это не работает в скрипте, нужно подключиться через psql)
-- \c telemed

-- Создание таблицы users
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    iin VARCHAR(12) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    full_name TEXT NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    is_blocked BOOLEAN NOT NULL DEFAULT FALSE,
    blocked_reason TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Создание таблицы patients
CREATE TABLE patients (
    id SERIAL PRIMARY KEY,
    full_name TEXT NOT NULL,
    phone TEXT NOT NULL,
    telegram_id TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы appointments
CREATE TABLE appointments (
    id SERIAL PRIMARY KEY,
    patient_id INT REFERENCES patients(id) ON DELETE CASCADE,
    doctor_id INT REFERENCES users(id) ON DELETE CASCADE,
    date_time TIMESTAMP NOT NULL,
    status TEXT CHECK (status IN ('scheduled', 'completed', 'canceled')) DEFAULT 'scheduled',
    symptoms TEXT,
    diagnosis TEXT,
    treatment TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы pdf_files
CREATE TABLE pdf_files (
    id SERIAL PRIMARY KEY,
    appointment_id INT REFERENCES appointments(id) ON DELETE CASCADE,
    file_url TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы video_sessions
CREATE TABLE video_sessions (
    id SERIAL PRIMARY KEY,
    appointment_id INT REFERENCES appointments(id) ON DELETE CASCADE,
    video_url TEXT NOT NULL,
    started_at TIMESTAMP,
    ended_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы messages
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    patient_id INT REFERENCES patients(id) ON DELETE CASCADE,
    sender TEXT CHECK (sender IN ('patient', 'bot')) NOT NULL,
    content TEXT,
    attachment_url TEXT,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);