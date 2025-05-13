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

CREATE TABLE specializations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL
);
CREATE TABLE doctors (
    id SERIAL PRIMARY KEY,
    full_name TEXT NOT NULL,
    specialization_id INT REFERENCES specializations(id) 
);
-- Создание таблицы patients
CREATE TABLE patients (
    id SERIAL PRIMARY KEY,
    full_name TEXT NOT NULL,
    iin VARCHAR(12) UNIQUE NOT NULL,
    telegram_id BIGINT UNIQUE NOT NULL
);

CREATE TABLE timeslots (
    id SERIAL PRIMARY KEY,
    doctor_id INT REFERENCES doctors(id),
    appointment_time TIMESTAMP NOT NULL,
    is_booked BOOLEAN DEFAULT FALSE
);

-- Создание таблицы appointments
CREATE TABLE appointments (
    id SERIAL PRIMARY KEY,
    patient_id INT REFERENCES patients(id) ON DELETE CASCADE,
    doctor_id INT REFERENCES users(id) ON DELETE CASCADE,
    timeslots_id INT REFERENCES timeslots(id),
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


-- Специализации
INSERT INTO specializations (name) VALUES 
('Терапевт'), ('Педиатр'), ('Кардиолог');

-- Врачи
INSERT INTO doctors (full_name, specialization_id) VALUES 
('Марина Цветаева', 2),
('Айгуль Омарова', 1),
('Жанат Мусаев', 3);

-- Временные слоты
INSERT INTO timeslots (doctor_id, appointment_time) VALUES 
(1, '2025-05-15 10:00'),
(1, '2025-05-15 11:00'),
(2, '2025-05-16 09:30'),
(3, '2025-05-17 14:00'),
(3, '2025-05-17 15:00');
