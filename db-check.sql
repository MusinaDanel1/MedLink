-- Check if the services table exists and has the right constraints
DO $$
BEGIN
    -- Check if the services table exists
    IF NOT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'services') THEN
        -- Create the services table
        CREATE TABLE services (
            id SERIAL PRIMARY KEY,
            doctor_id INTEGER NOT NULL,
            name VARCHAR(255) NOT NULL
        );
        
        -- Add a unique constraint
        ALTER TABLE services ADD CONSTRAINT services_doctor_id_name_unique UNIQUE (doctor_id, name);
    ELSE
        -- Check if the unique constraint exists
        IF NOT EXISTS (
            SELECT 1 FROM pg_constraint 
            WHERE conname = 'services_doctor_id_name_unique' 
            AND conrelid = 'services'::regclass
        ) THEN
            -- Add the unique constraint
            ALTER TABLE services ADD CONSTRAINT services_doctor_id_name_unique UNIQUE (doctor_id, name);
        END IF;
    END IF;
END;
$$;

-- Check if specialized services exist for the doctor
WITH doctor_data AS (
    SELECT id, full_name FROM doctors WHERE full_name = 'Марина Цветаева'
)
SELECT 
    doctor_data.id, 
    doctor_data.full_name,
    COUNT(s.id) AS service_count,
    ARRAY_AGG(s.name) AS service_names
FROM doctor_data
LEFT JOIN services s ON s.doctor_id = doctor_data.id
GROUP BY doctor_data.id, doctor_data.full_name; 