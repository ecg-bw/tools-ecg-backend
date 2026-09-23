CREATE TABLE IF NOT EXISTS editions (
    id SERIAL PRIMARY KEY,
    year INT NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS weekends (
    id SERIAL PRIMARY KEY,
    edition_id INT NOT NULL REFERENCES editions(id) ON DELETE CASCADE,
    weekend_number SMALLINT NOT NULL CHECK (weekend_number BETWEEN 1 AND 3),
    label VARCHAR(50),
    UNIQUE (edition_id, weekend_number)
);

CREATE TABLE IF NOT EXISTS market_days (
    id SERIAL PRIMARY KEY,
    weekend_id INT NOT NULL REFERENCES weekends(id) ON DELETE CASCADE,
    date DATE NOT NULL UNIQUE,
    day_of_week VARCHAR(10) NOT NULL CHECK (day_of_week IN ('Freitag', 'Samstag', 'Sonntag'))
);

CREATE TABLE IF NOT EXISTS shifts (
    id SERIAL PRIMARY KEY,
    day_id INT NOT NULL REFERENCES market_days(id) ON DELETE CASCADE,
    title VARCHAR(100) NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    required_slots SMALLINT NOT NULL CHECK (required_slots > 0),
    notes TEXT,
    CONSTRAINT check_shift_time CHECK (start_time != end_time)
);

CREATE TABLE IF NOT EXISTS volunteers (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    email VARCHAR(120) NOT NULL UNIQUE,
    phone VARCHAR(30),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT
);

CREATE TABLE IF NOT EXISTS shift_assignments (
    id SERIAL PRIMARY KEY,
    shift_id INT NOT NULL REFERENCES shifts(id) ON DELETE CASCADE,
    volunteer_id INT NOT NULL REFERENCES volunteers(id) ON DELETE CASCADE,
    role_id INT REFERENCES roles(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'confirmed' CHECK (status IN ('pending', 'confirmed', 'cancelled')),
    assigned_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (shift_id, volunteer_id)
);

-- Indexe für schnelle Lookups und Joins
CREATE INDEX IF NOT EXISTS idx_weekends_edition_id ON weekends(edition_id);
CREATE INDEX IF NOT EXISTS idx_market_days_weekend_id ON market_days(weekend_id);
CREATE INDEX IF NOT EXISTS idx_shifts_day_id ON shifts(day_id);
CREATE INDEX IF NOT EXISTS idx_assignments_shift_id ON shift_assignments(shift_id);
CREATE INDEX IF NOT EXISTS idx_assignments_volunteer_id ON shift_assignments(volunteer_id);