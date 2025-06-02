CREATE DATABASE telemed;

CREATE TYPE appointment_status AS ENUM ('Записан', 'Принят', 'Завершен');

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

CREATE TABLE schedules ( 
    id SERIAL PRIMARY KEY,
    doctor_id INT NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    service_id INT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    start_time TIMESTAMPTZ NOT NULL,
    end_time   TIMESTAMPTZ NOT NULL,
    color VARCHAR(7) NOT NULL,
    visible BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX IF NOT EXISTS idx_schedules_start_time_end_time ON schedules(start_time, end_time);

CREATE TABLE patients (
    id SERIAL PRIMARY KEY,
    full_name TEXT NOT NULL,
    iin VARCHAR(12) UNIQUE NOT NULL,
    telegram_id BIGINT UNIQUE NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_patients_full_name ON patients(full_name);

CREATE TABLE timeslots (
    id           SERIAL PRIMARY KEY,
    schedule_id  INT NOT NULL REFERENCES schedules(id) ON DELETE CASCADE, 
    start_time TIMESTAMPTZ NOT NULL,
    end_time   TIMESTAMPTZ NOT NULL,
    is_booked    BOOLEAN   NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (schedule_id, start_time, end_time)
);

CREATE TABLE appointments (
    id           SERIAL PRIMARY KEY,
    service_id INT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    timeslot_id  INT NOT NULL REFERENCES timeslots(id) ON DELETE RESTRICT,
    patient_id   INT NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    status   appointment_status NOT NULL DEFAULT 'Записан',
    created_at   TIMESTAMP          NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_appointments_status ON appointments(status);
CREATE INDEX IF NOT EXISTS idx_appointments_created_at ON appointments(created_at);

CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    appointment_id INT NOT NULL REFERENCES appointments(id) ON DELETE CASCADE,
    sender TEXT NOT NULL CHECK (sender IN ('patient', 'doctor', 'bot')),
    content TEXT NOT NULL,
    attachment_url TEXT,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_messages_sent_at ON messages(sent_at);

CREATE TABLE pdf_files (
    id SERIAL PRIMARY KEY,
    appointment_id INT REFERENCES appointments(id) ON DELETE CASCADE,
    file_url TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

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

CREATE TABLE medical_histories ( 
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
  diagnosis TEXT, 
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

INSERT INTO specializations (name) VALUES
  ('Терапевт'),         -- 1
  ('Педиатр'),          -- 2
  ('Кардиолог'),        -- 3
  ('Хирург'),           -- 4
  ('Невролог'),         -- 5
  ('Офтальмолог'),      -- 6
  ('Дерматолог'),       -- 7
  ('Эндокринолог'),     -- 8
  ('Гастроэнтеролог'),  -- 9
  ('Психиатр'),         --10
  ('Аллерголог'),       --11
  ('Уролог');           --12

-- Врачи
INSERT INTO doctors (full_name, specialization_id) VALUES
  ('Марина Цветаева', 2),       -- Педиатр
  ('Айгуль Омарова', 1),        -- Терапевт
  ('Жанат Мусаев', 3),          -- Кардиолог
  ('Алексей Смирнов', 4),       -- Хирург
  ('Гаухар Сагынова', 5),       -- Невролог
  ('Игорь Брагин', 6),          -- Офтальмолог
  ('Наталья Ким', 7),           -- Дерматолог
  ('Бахытжан Ермеков', 8),      -- Эндокринолог
  ('Ольга Соколова', 9),        -- Гастроэнтеролог
  ('Мурат Бейсеков', 10),       -- Психиатр
  ('Елена Жумагалиева', 11),    -- Аллерголог
  ('Руслан Тлеулин', 12);       -- Уролог   

INSERT INTO services (doctor_id, name) VALUES
  -- Педиатр (id = 1, Марина Цветаева)
  (1, 'Консультация педиатра'),
  (1, 'Вакцинация и профилактика'),
  (1, 'Детский осмотр'),

  -- Терапевт (id = 2, Айгуль Омарова)
  (2, 'Консультация терапевта'),
  (2, 'Измерение давления'),
  (2, 'Общий осмотр'),

  -- Кардиолог (id = 3, Жанат Мусаев)
  (3, 'Кардиологическая консультация'),
  (3, 'Эхокардиография'),
  (3, 'Нагрузочная ЭКГ'),

  -- Хирург (id = 4, Алексей Смирнов)
  (4, 'Хирургическая консультация'),
  (4, 'Удаление новообразований'),
  (4, 'Перевязка после операции'),

  -- Невролог (id = 5, Гаухар Сагынова)
  (5, 'Консультация невролога'),
  (5, 'Диагностика головных болей'),
  (5, 'Электроэнцефалография (ЭЭГ)'),

  -- Офтальмолог (id = 6, Игорь Брагин)
  (6, 'Проверка зрения'),
  (6, 'Подбор очков'),
  (6, 'Осмотр глазного дна'),

  -- Дерматолог (id = 7, Наталья Ким)
  (7, 'Консультация дерматолога'),
  (7, 'Диагностика кожных заболеваний'),
  (7, 'Удаление бородавок'),

  -- Эндокринолог (id = 8, Бахытжан Ермеков)
  (8, 'Приём эндокринолога'),
  (8, 'Консультация по диабету'),
  (8, 'УЗИ щитовидной железы'),

  -- Гастроэнтеролог (id = 9, Ольга Соколова)
  (9, 'Консультация гастроэнтеролога'),
  (9, 'ФГДС (гастроскопия)'),
  (9, 'Лечение ЖКТ-заболеваний'),

  -- Психиатр (id = 10, Мурат Бейсеков)
  (10, 'Психиатрическая консультация'),
  (10, 'Психотерапевтическая беседа'),
  (10, 'Назначение медикаментозного лечения'),

  -- Аллерголог (id = 11, Елена Жумагалиева)
  (11, 'Консультация аллерголога'),
  (11, 'Аллергопробы'),
  (11, 'Иммунотерапия'),

  -- Уролог (id = 12, Руслан Тлеулин)
  (12, 'Приём уролога'),
  (12, 'УЗИ почек и мочевого пузыря'),
  (12, 'Лечение инфекций мочеполовой системы');            

INSERT INTO schedules (doctor_id, service_id, start_time, end_time, color, visible) VALUES 
  (1, 1, '2025-05-15 09:00', '2025-05-15 17:00', '#4CAF50', TRUE), 
  (2, 4, '2025-05-16 09:00', '2025-05-16 17:00', '#2196F3', TRUE), 
  (3, 7, '2025-05-17 09:00', '2025-05-17 17:00', '#FF9800', TRUE); 

INSERT INTO timeslots (schedule_id, start_time, end_time, is_booked) VALUES
  (1, '2025-05-15 10:00', '2025-05-15 10:30', FALSE), 
  (1, '2025-05-15 11:00', '2025-05-15 11:30', FALSE), 
  (2, '2025-05-16 09:30', '2025-05-16 10:00', FALSE), 
  (3, '2025-05-17 14:00', '2025-05-17 14:30', FALSE), 
  (3, '2025-05-17 15:00', '2025-05-17 15:30', FALSE); 

INSERT INTO patients (full_name, iin, telegram_id) VALUES
  ('Ахматова Айгерим', '000000000001', 123456789),
  ('Иванов Петр Сидорович', '000000000002', 100000000),
  ('Сергеева Анна Викторовна', '000000000003', 100000001),
  ('Кузнецова Ольга Игоревна', '000000000004', 100000002),
  ('Васильев Дмитрий Алексеевич', '000000000005', 100000003),
  ('Михайлова Екатерина Денисовна', '000000000006', 100000004),
  ('Новиков Артем Вячеславович', '000000000007', 100000005),
  ('Федорова Анастасия Юрьевна', '000000000008', 100000006),
  ('Волков Максим Романович', '000000000009', 100000007),
  ('Алексеева Полина Сергеевна', '000000000010', 100000008),
  ('Лебедев Иван Дмитриевич', '000000000011', 100000009),
  ('Смирнова Дарья Павловна', '000000000012', 100000010),
  ('Козлов Андрей Игоревич', '000000000013', 100000011),
  ('Морозова Алена Максимовна', '000000000014', 100000012),
  ('Егоров Кирилл Евгеньевич', '000000000015', 100000013),
  ('Павлова София Артемовна', '000000000016', 100000014),
  ('Орлов Никита Александрович', '000000000017', 100000015);

INSERT INTO appointments (patient_id, service_id, timeslot_id, created_at) VALUES
  (1, 4, 3, CURRENT_TIMESTAMP); 

INSERT INTO messages (appointment_id, sender, content, sent_at) VALUES
  (1, 'patient', 'Здравствуйте, доктор!', CURRENT_TIMESTAMP - interval '5 minutes'),
  (1, 'doctor', 'Здравствуйте! Как я могу вам помочь?', CURRENT_TIMESTAMP - interval '4 minutes'),
  (1, 'patient', 'У меня болит голова последние 2 дня', CURRENT_TIMESTAMP - interval '3 minutes'),
  (1, 'doctor', 'Давайте обсудим ваши симптомы подробнее', CURRENT_TIMESTAMP - interval '2 minutes'),
  (1, 'bot', 'Видеозвонок начался', CURRENT_TIMESTAMP - interval '1 minute');

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

INSERT INTO medical_records (patient_id, gender, birth_date, city) VALUES
(1, 'Мужской', '1990-03-12', 'Алматы'); 

INSERT INTO medical_histories (medical_record_id, entry_type, description, event_date) VALUES
(1, 'chronic_disease', 'Гипертония', '2018-05-20'),
(1, 'chronic_disease', 'Астма', '2015-11-10');

INSERT INTO medical_histories (medical_record_id, entry_type, description, event_date) VALUES
(1, 'allergy', 'Пенициллин', '2012-07-15'),
(1, 'allergy', 'Пыльца', '2016-04-22');

INSERT INTO medical_histories (medical_record_id, entry_type, description, event_date) VALUES
(1, 'vaccination', 'COVID-19', '2021-06-15'),
(1, 'vaccination', 'Грипп', '2022-11-03');

INSERT INTO medical_histories (medical_record_id, entry_type, description, event_date) VALUES
(1, 'surgery', 'Аппендэктомия', '2010-09-12'),
(1, 'surgery', 'Удаление кисты', '2017-04-05');

INSERT INTO medical_histories (medical_record_id, entry_type, description, event_date) VALUES
(1, 'examination', 'Флюорография — без патологий', '2023-08-15'),
(1, 'examination', 'ЭКГ — синусовая аритмия', '2024-01-22'),
(1, 'examination', 'УЗИ щитовидной железы — узел 1 см', '2023-11-10');

INSERT INTO appointment_details (appointment_id, complaints, diagnosis, assignments, created_at, updated_at) VALUES
(1, 'Частый жидкий стул, боли в животе', 'Другие и неуточненные инфекции кишечника', 'Смекта, обильное питье, диета', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
