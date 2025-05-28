-- Создание базы данных
CREATE DATABASE telemed;

-- Соединение с базой данных telemed_db (это не работает в скрипте, нужно подключиться через psql)
-- \c telemed_db
CREATE TYPE appointment_status AS ENUM ('Записан', 'Принят', 'Завершен');
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
CREATE INDEX IF NOT EXISTS idx_doctors_full_name ON doctors(full_name);

CREATE TABLE services (
    id SERIAL PRIMARY KEY,
    doctor_id INT REFERENCES doctors(id),
    name TEXT NOT NULL,
    CONSTRAINT services_doctor_id_name_unique UNIQUE (doctor_id, name)
);

CREATE TABLE schedules ( -- Renamed from schedule
    id SERIAL PRIMARY KEY,
    doctor_id INT NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    service_id INT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    start_time TIMESTAMPTZ NOT NULL,
    end_time   TIMESTAMPTZ NOT NULL,
    color VARCHAR(7) NOT NULL,
    visible BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX IF NOT EXISTS idx_schedules_start_time_end_time ON schedules(start_time, end_time);

-- Создание таблицы patients
CREATE TABLE patients (
    id SERIAL PRIMARY KEY,
    full_name TEXT NOT NULL,
    iin VARCHAR(12) UNIQUE NOT NULL,
    telegram_id BIGINT UNIQUE NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_patients_full_name ON patients(full_name);

CREATE TABLE timeslots (
    id           SERIAL PRIMARY KEY,
    schedule_id  INT NOT NULL REFERENCES schedules(id) ON DELETE CASCADE, -- Updated FK reference
    start_time TIMESTAMPTZ NOT NULL,
    end_time   TIMESTAMPTZ NOT NULL,
    is_booked    BOOLEAN   NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (schedule_id, start_time, end_time)
);

-- Создание таблицы appointments
CREATE TABLE appointments (
    id           SERIAL PRIMARY KEY,
    -- doctor_id INT NOT NULL REFERENCES doctors(id) ON DELETE CASCADE, -- Removed doctor_id
    service_id INT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    timeslot_id  INT NOT NULL REFERENCES timeslots(id) ON DELETE RESTRICT,
    patient_id   INT NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    status   appointment_status NOT NULL DEFAULT 'Записан',
    created_at   TIMESTAMP          NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_appointments_status ON appointments(status);
CREATE INDEX IF NOT EXISTS idx_appointments_created_at ON appointments(created_at);

-- Создание таблицы messages
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    appointment_id INT NOT NULL REFERENCES appointments(id) ON DELETE CASCADE,
    sender TEXT NOT NULL CHECK (sender IN ('patient', 'doctor', 'bot')),
    content TEXT NOT NULL,
    attachment_url TEXT,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_messages_sent_at ON messages(sent_at);

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
CREATE INDEX IF NOT EXISTS idx_video_sessions_room_name ON video_sessions(room_name);
CREATE INDEX IF NOT EXISTS idx_video_sessions_started_at ON video_sessions(started_at);
CREATE INDEX IF NOT EXISTS idx_video_sessions_ended_at ON video_sessions(ended_at);

CREATE TABLE medical_records (
    id SERIAL PRIMARY KEY,
    patient_id INT UNIQUE REFERENCES patients(id) ON DELETE CASCADE,
    gender VARCHAR(10),
    birth_date DATE,
    city VARCHAR(100),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE medical_histories ( -- Renamed from medical_history
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
CREATE INDEX IF NOT EXISTS idx_medical_histories_entry_type ON medical_histories(entry_type);
CREATE INDEX IF NOT EXISTS idx_medical_histories_event_date ON medical_histories(event_date);

CREATE TABLE diagnoses (
  id SERIAL PRIMARY KEY,
  code VARCHAR(10)    NOT NULL UNIQUE,
  name TEXT           NOT NULL,
  description TEXT,
  created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_diagnoses_name ON diagnoses(name);

CREATE TABLE appointment_details (
  appointment_id INT PRIMARY KEY
    REFERENCES appointments(id) ON DELETE CASCADE,
  complaints TEXT,
  diagnosis TEXT, -- Reverted from diagnosis_id INT REFERENCES diagnoses(id)
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
  schedule TEXT NOT NULL, -- Note: 'schedule' here is a column name, not the table
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


-- Специализации
INSERT INTO specializations (name) VALUES
('Терапевт'), ('Педиатр'), ('Кардиолог');

-- Врачи
INSERT INTO doctors (full_name, specialization_id) VALUES
('Марина Цветаева', 2), -- Pediatrician (ID: 1 after insert)
('Айгуль Омарова', 1), -- Therapist (ID: 2 after insert)
('Жанат Мусаев', 3);    -- Cardiologist (ID: 3 after insert)

-- Тестовые данные для таблицы services
-- IDs for doctors will be 1, 2, 3 respectively after the above inserts.
INSERT INTO services (doctor_id, name) VALUES
  (1, 'Консультация педиатра'),        -- Service ID: 1
  (1, 'Вакцинация и профилактика'),   -- Service ID: 2
  (1, 'Детский осмотр'),             -- Service ID: 3
  (2, 'Консультация терапевта'),     -- Service ID: 4
  (2, 'Снятие ЭКГ'),                 -- Service ID: 5
  (2, 'УЗИ сердца'),                  -- Service ID: 6
  (3, 'Кардиологическая консультация'),-- Service ID: 7
  (3, 'Эхокардиография'),             -- Service ID: 8
  (3, 'Нагрузочная ЭКГ');            -- Service ID: 9

-- Добавление тестовых расписаний
-- Assuming service IDs correspond to the order above
INSERT INTO schedules (doctor_id, service_id, start_time, end_time, color, visible) VALUES -- Renamed table
  (1, 1, '2025-05-15 09:00', '2025-05-15 17:00', '#4CAF50', TRUE), -- Schedule ID: 1
  (2, 4, '2025-05-16 09:00', '2025-05-16 17:00', '#2196F3', TRUE), -- Schedule ID: 2
  (3, 7, '2025-05-17 09:00', '2025-05-17 17:00', '#FF9800', TRUE); -- Schedule ID: 3

-- Временные слоты
-- Assuming schedule IDs correspond to the order above
INSERT INTO timeslots (schedule_id, start_time, end_time, is_booked) VALUES
(1, '2025-05-15 10:00', '2025-05-15 10:30', FALSE), -- Timeslot ID: 1
(1, '2025-05-15 11:00', '2025-05-15 11:30', FALSE), -- Timeslot ID: 2
(2, '2025-05-16 09:30', '2025-05-16 10:00', FALSE), -- Timeslot ID: 3
(3, '2025-05-17 14:00', '2025-05-17 14:30', FALSE), -- Timeslot ID: 4
(3, '2025-05-17 15:00', '2025-05-17 15:30', FALSE); -- Timeslot ID: 5

-- Test data for video call messaging
INSERT INTO patients (full_name, iin, telegram_id) VALUES
('Тестовый Пациент', '000000000001', 123456789); -- Patient ID: 1

-- Create a test appointment for video call
-- Patient ID 1, Service ID 4 ('Консультация терапевта' by Doctor 2), Timeslot ID 3 (belongs to Schedule 2, which is Doctor 2, Service 4)
-- This is now consistent as doctor_id is derived from service_id via schedules.
INSERT INTO appointments (patient_id, service_id, timeslot_id, created_at) VALUES
(1, 4, 3, CURRENT_TIMESTAMP); -- Appointment ID: 1 (assuming this is the first appointment overall)

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
  ('A09',   'Другие и неуточненные инфекции кишечника',                    ''), -- Diagnosis ID: 1
  ('E11.9', 'Сахарный диабет 2 типа без осложнений',                      ''), -- Diagnosis ID: 2
  ('E66.9', 'Ожирение неуточненное',                                      ''), -- Diagnosis ID: 3
  ('F41.1', 'Обостренное тревожное расстройство',                         ''), -- Diagnosis ID: 4
  ('G43.9', 'Мигрень неуточненная',                                      ''), -- Diagnosis ID: 5
  ('I10',   'Эссенциальная (первичная) гипертензия',                     ''), -- Diagnosis ID: 6
  ('I20.9', 'Стенокардия неуточненная',                                  ''), -- Diagnosis ID: 7
  ('I48.0', 'Пароксизмальная мерцательная аритмия',                       ''), -- Diagnosis ID: 8
  ('I50.9', 'Сердечная недостаточность неуточненная',                   ''), -- Diagnosis ID: 9
  ('J00',   'Острый назофарингит [простуда]',                            ''), -- Diagnosis ID: 10
  ('J01.0', 'Острый этмоидит',                                           ''), -- Diagnosis ID: 11
  ('J02.0', 'Острый катаральный фарингит',                               ''), -- Diagnosis ID: 12
  ('J03.0', 'Острый тонзиллит',                                          ''), -- Diagnosis ID: 13
  ('J06.9', 'Острая верхняя респираторная инфекция неуточненная',        ''), -- Diagnosis ID: 14
  ('J18.9', 'Пневмония неуточненная',                                    ''), -- Diagnosis ID: 15
  ('K21.0', 'ГЭРБ с эзофагитом',                                         ''), -- Diagnosis ID: 16
  ('K21.9', 'ГЭРБ без эзофагита',                                        ''), -- Diagnosis ID: 17
  ('M54.5', 'Боль в пояснице',                                           ''), -- Diagnosis ID: 18
  ('R07.9', 'Боль в грудной клетке неуточненная',                        ''), -- Diagnosis ID: 19
  ('R51',   'Головная боль',                                             ''); -- Diagnosis ID: 20

-- Create medical records for test patients
INSERT INTO medical_records (patient_id, gender, birth_date, city) VALUES
(1, 'Мужской', '1990-03-12', 'Алматы'); -- Medical Record ID: 1

-- Add chronic diseases to medical histories
INSERT INTO medical_histories (medical_record_id, entry_type, description, event_date) VALUES
(1, 'chronic_disease', 'Гипертония', '2018-05-20'),
(1, 'chronic_disease', 'Астма', '2015-11-10');

-- Add allergies to medical histories
INSERT INTO medical_histories (medical_record_id, entry_type, description, event_date) VALUES
(1, 'allergy', 'Пенициллин', '2012-07-15'),
(1, 'allergy', 'Пыльца', '2016-04-22');

-- Add vaccinations to medical histories
INSERT INTO medical_histories (medical_record_id, entry_type, description, event_date) VALUES
(1, 'vaccination', 'COVID-19', '2021-06-15'),
(1, 'vaccination', 'Грипп', '2022-11-03');

-- Add surgeries to medical histories
INSERT INTO medical_histories (medical_record_id, entry_type, description, event_date) VALUES
(1, 'surgery', 'Аппендэктомия', '2010-09-12'),
(1, 'surgery', 'Удаление кисты', '2017-04-05');

-- Add examinations to medical histories
INSERT INTO medical_histories (medical_record_id, entry_type, description, event_date) VALUES
(1, 'examination', 'Флюорография — без патологий', '2023-08-15'),
(1, 'examination', 'ЭКГ — синусовая аритмия', '2024-01-22'),
(1, 'examination', 'УЗИ щитовидной железы — узел 1 см', '2023-11-10');

-- Пример записи для appointment_details с использованием diagnosis TEXT
-- And appointment with id=1 exists.
-- Using 'Другие и неуточненные инфекции кишечника' as sample diagnosis text (was ID 1, code 'A09')
INSERT INTO appointment_details (appointment_id, complaints, diagnosis, assignments, created_at, updated_at) VALUES
(1, 'Частый жидкий стул, боли в животе', 'Другие и неуточненные инфекции кишечника', 'Смекта, обильное питье, диета', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

INSERT INTO doctors (full_name, specialization_id) VALUES ('Григорьев Максим Валерьевич', 1);
INSERT INTO doctors (full_name, specialization_id) VALUES ('Фёдорова Алина Романовна', 2);
INSERT INTO doctors (full_name, specialization_id) VALUES ('Степанов Арсений Кириллович', 3);
INSERT INTO doctors (full_name, specialization_id) VALUES ('Беляева София Львовна', 1);
INSERT INTO doctors (full_name, specialization_id) VALUES ('Андреев Даниил Артёмович', 2);
INSERT INTO doctors (full_name, specialization_id) VALUES ('Виноградова Полина Глебовна', 3);
INSERT INTO doctors (full_name, specialization_id) VALUES ('Богданов Марк Денисович', 1);
INSERT INTO doctors (full_name, specialization_id) VALUES ('Комарова Ева Ярославовна', 2);
INSERT INTO doctors (full_name, specialization_id) VALUES ('Киселёв Лев Игоревич', 1);
INSERT INTO doctors (full_name, specialization_id) VALUES ('Абрамова Милана Эмировна', 3);
INSERT INTO doctors (full_name, specialization_id) VALUES ('Тихонов Руслан Альбертович', 2);
INSERT INTO doctors (full_name, specialization_id) VALUES ('Мельникова Вероника Макаровна', 1);
INSERT INTO doctors (full_name, specialization_id) VALUES ('Щербаков Глеб Робертович', 3);
INSERT INTO doctors (full_name, specialization_id) VALUES ('Кузьмина Ульяна Давидовна', 2);

-- Additional mock patients
INSERT INTO patients (full_name, iin, telegram_id) VALUES ('Иванов Петр Сидорович', '000000000002', 100000000);
INSERT INTO patients (full_name, iin, telegram_id) VALUES ('Сергеева Анна Викторовна', '000000000003', 100000001);
INSERT INTO patients (full_name, iin, telegram_id) VALUES ('Кузнецова Ольга Игоревна', '000000000004', 100000002);
INSERT INTO patients (full_name, iin, telegram_id) VALUES ('Васильев Дмитрий Алексеевич', '000000000005', 100000003);
INSERT INTO patients (full_name, iin, telegram_id) VALUES ('Михайлова Екатерина Денисовна', '000000000006', 100000004);
INSERT INTO patients (full_name, iin, telegram_id) VALUES ('Новиков Артем Вячеславович', '000000000007', 100000005);
INSERT INTO patients (full_name, iin, telegram_id) VALUES ('Федорова Анастасия Юрьевна', '000000000008', 100000006);
INSERT INTO patients (full_name, iin, telegram_id) VALUES ('Волков Максим Романович', '000000000009', 100000007);
INSERT INTO patients (full_name, iin, telegram_id) VALUES ('Алексеева Полина Сергеевна', '000000000010', 100000008);
INSERT INTO patients (full_name, iin, telegram_id) VALUES ('Лебедев Иван Дмитриевич', '000000000011', 100000009);
INSERT INTO patients (full_name, iin, telegram_id) VALUES ('Смирнова Дарья Павловна', '000000000012', 100000010);
INSERT INTO patients (full_name, iin, telegram_id) VALUES ('Козлов Андрей Игоревич', '000000000013', 100000011);
INSERT INTO patients (full_name, iin, telegram_id) VALUES ('Морозова Алена Максимовна', '000000000014', 100000012);
INSERT INTO patients (full_name, iin, telegram_id) VALUES ('Егоров Кирилл Евгеньевич', '000000000015', 100000013);
INSERT INTO patients (full_name, iin, telegram_id) VALUES ('Павлова София Артемовна', '000000000016', 100000014);
INSERT INTO patients (full_name, iin, telegram_id) VALUES ('Орлов Никита Александрович', '000000000017', 100000015);