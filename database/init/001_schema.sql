-- Fichar Database Schema
-- PostgreSQL initialization script

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Table: employees
CREATE TABLE IF NOT EXISTS employees (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    dni VARCHAR(20) NOT NULL UNIQUE,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('ADMIN', 'EMPLOYEE')),
    password_hash VARCHAR(255) NOT NULL,
    must_change_password BOOLEAN DEFAULT TRUE,
    daily_hours NUMERIC(4,2) DEFAULT 8.0,
    monthly_hours NUMERIC(5,2) DEFAULT 160.0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Table: shifts
CREATE TABLE IF NOT EXISTS shifts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    expected_hours NUMERIC(4, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Table: employee_shift_assignments
CREATE TABLE IF NOT EXISTS employee_shift_assignments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    shift_id UUID NOT NULL REFERENCES shifts(id) ON DELETE CASCADE,
    start_date DATE NOT NULL,
    end_date DATE
);

-- Table: attendances
CREATE TABLE IF NOT EXISTS attendances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    check_in TIMESTAMP,
    check_out TIMESTAMP,
    worked_hours NUMERIC(4, 2),
    late BOOLEAN DEFAULT FALSE,
    is_remote BOOLEAN DEFAULT FALSE,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    corrected BOOLEAN DEFAULT FALSE,
    correction_reason TEXT,
    corrected_by UUID REFERENCES employees(id),
    corrected_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Table: logs
CREATE TABLE IF NOT EXISTS logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_attendances_employee_date ON attendances(employee_id, date);
CREATE INDEX IF NOT EXISTS idx_attendances_date ON attendances(date);
CREATE INDEX IF NOT EXISTS idx_shifts_employee ON employee_shift_assignments(employee_id);
CREATE INDEX IF NOT EXISTS idx_logs_user ON logs(user_id);
CREATE INDEX IF NOT EXISTS idx_logs_created ON logs(created_at);
CREATE INDEX IF NOT EXISTS idx_employees_dni ON employees(dni);

-- Create initial admin user (password: 'admin', must be changed on first login)
-- Password hash for 'admin' using bcrypt cost 10
INSERT INTO employees (dni, first_name, last_name, role, password_hash, must_change_password)
VALUES (
    '00000000',
    'Admin',
    'System',
    'ADMIN',
    '$2b$10$4C7Q/fmhZAjoyz2HxADAjOhqqdG.vLLbGR9SP3Kbuccn1m6b8l49i',
    TRUE
) ON CONFLICT (dni) DO NOTHING;
