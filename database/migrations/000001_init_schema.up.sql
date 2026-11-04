-- 1. Tabelle: days
CREATE TABLE days (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    weekend_id BIGINT NULL,
    date DATE NOT NULL
);

-- 2. Tabelle: shifts
CREATE TABLE shifts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    day_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    required_slots INT NOT NULL DEFAULT 1,
    description TEXT NULL,
    CONSTRAINT fk_shifts_day 
        FOREIGN KEY (day_id) 
        REFERENCES days(id) 
        ON DELETE CASCADE
);

-- 3. Tabelle: shift_assignments
CREATE TABLE shift_assignments (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    shift_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    phone_number VARCHAR(50) NULL,
    CONSTRAINT fk_assignments_shift 
        FOREIGN KEY (shift_id) 
        REFERENCES shifts(id) 
        ON DELETE CASCADE
);

-- 4. Tabelle: waffeln
CREATE TABLE waffeln (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    shift_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT fk_waffeln_shift 
        FOREIGN KEY (shift_id) 
        REFERENCES shifts(id) 
        ON DELETE CASCADE
);