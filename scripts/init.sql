-- Создание базы данных
CREATE DATABASE telemed;

-- Соединение с базой данных telemed_db (это не работает в скрипте, нужно подключиться через psql)
-- \c telemed_db
CREATE TYPE appointment_status AS  ENUM ('Записан', 'Принят');
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
CREATE TABLE services (
    id SERIAL PRIMARY KEY,
    doctor_id INT REFERENCES doctors(id),
    name TEXT NOT NULL
);

CREATE TABLE schedule (
    id SERIAL PRIMARY KEY,
    doctor_id INT NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    service_id INT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    start_time TIMESTAMPTZ NOT NULL,
    end_time   TIMESTAMPTZ NOT NULL,
    color VARCHAR(7) NOT NULL,
    visible BOOLEAN NOT NULL DEFAULT TRUE
);
-- Создание таблицы patients
CREATE TABLE patients (
    id SERIAL PRIMARY KEY,
    full_name TEXT NOT NULL,
    iin VARCHAR(12) UNIQUE NOT NULL,
    telegram_id BIGINT UNIQUE NOT NULL
);

CREATE TABLE timeslots (
    id           SERIAL PRIMARY KEY,
    schedule_id  INT NOT NULL REFERENCES schedule(id) ON DELETE CASCADE,
    start_time TIMESTAMPTZ NOT NULL,
    end_time   TIMESTAMPTZ NOT NULL,
    is_booked    BOOLEAN   NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (schedule_id, start_time, end_time)
);

-- Создание таблицы appointments
CREATE TABLE appointments (
    id           SERIAL PRIMARY KEY,
    doctor_id INT NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    service_id INT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    timeslot_id  INT NOT NULL REFERENCES timeslots(id) ON DELETE RESTRICT,
    patient_id   INT NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    status   appointment_status NOT NULL DEFAULT 'Записан',
    created_at   TIMESTAMP          NOT NULL DEFAULT NOW()
);

-- Создание таблицы messages
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    appointment_id INT NOT NULL REFERENCES appointments(id) ON DELETE CASCADE,
    sender TEXT NOT NULL CHECK (sender IN ('patient', 'doctor', 'bot')),
    content TEXT NOT NULL,
    attachment_url TEXT,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
    appointment_id INT NOT NULL REFERENCES appointments(id) ON DELETE CASCADE,
    room_name TEXT NOT NULL,
    video_url TEXT NOT NULL,
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE medical_records (
    id SERIAL PRIMARY KEY,
    patient_id INT UNIQUE REFERENCES patients(id) ON DELETE CASCADE,
    gender VARCHAR(10),
    birth_date DATE,
    city VARCHAR(100),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE medical_history (
  id SERIAL PRIMARY KEY,
  medical_record_id INT
    REFERENCES medical_records(id) ON DELETE CASCADE,
  entry_type VARCHAR(50) NOT NULL
    CHECK (entry_type IN (
      'chronic_disease','allergy','vaccination',
      'surgery','examination'
    )),
  description TEXT NOT NULL,
  event_date DATE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE diagnoses (
  id SERIAL PRIMARY KEY,
  code VARCHAR(10)    NOT NULL UNIQUE,   
  name TEXT           NOT NULL,          
  description TEXT,                     
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE appointment_details (
  appointment_id INT PRIMARY KEY
    REFERENCES appointments(id) ON DELETE CASCADE,
  complaints TEXT,
  diagnosis TEXT ,
  assignments TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE prescriptions (
  id SERIAL PRIMARY KEY,
  appointment_id INT
    REFERENCES appointments(id) ON DELETE CASCADE,
  medication TEXT NOT NULL,
  dosage TEXT NOT NULL,
  schedule TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);



-- Специализации
INSERT INTO specializations (name) VALUES 
('Терапевт'), ('Педиатр'), ('Кардиолог');

-- Врачи
INSERT INTO doctors (full_name, specialization_id) VALUES 
('Марина Цветаева', 2),
('Айгуль Омарова', 1),
('Жанат Мусаев', 3);

-- Тестовые данные для таблицы services
INSERT INTO services (doctor_id, name) VALUES
  (1, 'Консультация педиатра'),
  (1, 'Вакцинация и профилактика'),
  (1, 'Детский осмотр'),
  (2, 'Консультация терапевта'),
  (2, 'Снятие ЭКГ'),
  (2, 'УЗИ сердца'),
  (3, 'Кардиологическая консультация'),
  (3, 'Эхокардиография'),
  (3, 'Нагрузочная ЭКГ');

-- Добавление тестовых расписаний
INSERT INTO schedule (doctor_id, service_id, start_time, end_time, color, visible) VALUES
  (1, 1, '2025-05-15 09:00', '2025-05-15 17:00', '#4CAF50', true),
  (2, 4, '2025-05-16 09:00', '2025-05-16 17:00', '#2196F3', true),
  (3, 7, '2025-05-17 09:00', '2025-05-17 17:00', '#FF9800', true);

-- Временные слоты
INSERT INTO timeslots (schedule_id, start_time, end_time, is_booked) VALUES 
(1, '2025-05-15 10:00', '2025-05-15 10:30', false),
(1, '2025-05-15 11:00', '2025-05-15 11:30', false),
(2, '2025-05-16 09:30', '2025-05-16 10:00', false),
(3, '2025-05-17 14:00', '2025-05-17 14:30', false),
(3, '2025-05-17 15:00', '2025-05-17 15:30', false);

-- Тестовые данные для таблицы services
INSERT INTO services (doctor_id, name) VALUES
  (1, 'Консультация педиатра'),
  (1, 'Вакцинация и профилактика'),
  (1, 'Детский осмотр'),
  (2, 'Консультация терапевта'),
  (2, 'Снятие ЭКГ'),
  (2, 'УЗИ сердца'),
  (3, 'Кардиологическая консультация'),
  (3, 'Эхокардиография'),
  (3, 'Нагрузочная ЭКГ');

-- Test data for video call messaging
INSERT INTO patients (full_name, iin, telegram_id) VALUES 
('Тестовый Пациент', '000000000001', 123456789);

-- Create a test appointment for video call (using user ID 2 which is Марина Цветаева)
INSERT INTO appointments (patient_id, doctor_id, service_id, timeslot_id, created_at) VALUES 
(1, 1, 4, 3, CURRENT_TIMESTAMP);

-- Create a test video session
INSERT INTO video_sessions (appointment_id, room_name, video_url) VALUES 
(1, 'test-room-1', 'https://meet.jit.si/test-room-1');

-- Add some test messages
INSERT INTO messages (appointment_id, sender, content, sent_at) VALUES 
(1, 'patient', 'Здравствуйте, доктор!', CURRENT_TIMESTAMP - interval '5 minutes'),
(1, 'doctor', 'Здравствуйте! Как я могу вам помочь?', CURRENT_TIMESTAMP - interval '4 minutes'),
(1, 'patient', 'У меня болит голова последние 2 дня', CURRENT_TIMESTAMP - interval '3 minutes'),
(1, 'doctor', 'Давайте обсудим ваши симптомы подробнее', CURRENT_TIMESTAMP - interval '2 minutes'),
(1, 'bot', 'Видеозвонок начался', CURRENT_TIMESTAMP - interval '1 minute');

-- Пример заполнения справочника диагнозов (ICD-10)
INSERT INTO diagnoses (code, name, description) VALUES
  ('A09',   'Другие и неуточненные инфекции кишечника',                    ''),
  ('E11.9', 'Сахарный диабет 2 типа без осложнений',                      ''),
  ('E66.9', 'Ожирение неуточненное',                                      ''),
  ('F41.1', 'Обостренное тревожное расстройство',                         ''),
  ('G43.9', 'Мигрень неуточненная',                                      ''),
  ('I10',   'Эссенциальная (первичная) гипертензия',                     ''),
  ('I20.9', 'Стенокардия неуточненная',                                  ''),
  ('I48.0', 'Пароксизмальная мерцательная аритмия',                       ''),
  ('I50.9', 'Сердечная недостаточность неуточненная',                   ''),
  ('J00',   'Острый назофарингит [простуда]',                            ''),
  ('J01.0', 'Острый этмоидит',                                           ''),
  ('J02.0', 'Острый катаральный фарингит',                               ''),
  ('J03.0', 'Острый тонзиллит',                                          ''),
  ('J06.9', 'Острая верхняя респираторная инфекция неуточненная',        ''),
  ('J18.9', 'Пневмония неуточненная',                                    ''),
  ('K21.0', 'ГЭРБ с эзофагитом',                                         ''),
  ('K21.9', 'ГЭРБ без эзофагита',                                        ''),
  ('M54.5', 'Боль в пояснице',                                           ''),
  ('R07.9', 'Боль в грудной клетке неуточненная',                        ''),
  ('R51',   'Головная боль',                                             '');

-- Create medical records for test patients
INSERT INTO medical_records (patient_id, gender, birth_date, city) VALUES
(1, 'Мужской', '1990-03-12', 'Алматы');

-- Add chronic diseases to medical history
INSERT INTO medical_history (medical_record_id, entry_type, description, event_date) VALUES
(1, 'chronic_disease', 'Гипертония', '2018-05-20'),
(1, 'chronic_disease', 'Астма', '2015-11-10');

-- Add allergies to medical history
INSERT INTO medical_history (medical_record_id, entry_type, description, event_date) VALUES
(1, 'allergy', 'Пенициллин', '2012-07-15'),
(1, 'allergy', 'Пыльца', '2016-04-22');

-- Add vaccinations to medical history
INSERT INTO medical_history (medical_record_id, entry_type, description, event_date) VALUES
(1, 'vaccination', 'COVID-19', '2021-06-15'),
(1, 'vaccination', 'Грипп', '2022-11-03');

-- Add surgeries to medical history
INSERT INTO medical_history (medical_record_id, entry_type, description, event_date) VALUES
(1, 'surgery', 'Аппендэктомия', '2010-09-12'),
(1, 'surgery', 'Удаление кисты', '2017-04-05');

-- Add examinations to medical history
INSERT INTO medical_history (medical_record_id, entry_type, description, event_date) VALUES
(1, 'examination', 'Флюорография — без патологий', '2023-08-15'),
(1, 'examination', 'ЭКГ — синусовая аритмия', '2024-01-22'),
(1, 'examination', 'УЗИ щитовидной железы — узел 1 см', '2023-11-10');
