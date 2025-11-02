CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE profiles (
    id UUID PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    age INTEGER NOT NULL CHECK (age >= 17 AND age <= 150),
    tg VARCHAR(100) DEFAULT NULL,
    resume_id UUID DEFAULT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    profession_category VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes
CREATE INDEX idx_profiles_email ON profiles(email);
CREATE INDEX idx_profiles_profession_category ON profiles(profession_category);
CREATE INDEX idx_profiles_deleted_at ON profiles(deleted_at);
CREATE INDEX idx_profiles_resume_id ON profiles(resume_id);
CREATE INDEX idx_profiles_created_at ON profiles(created_at);
CREATE INDEX idx_profiles_name ON profiles(first_name, last_name);
CREATE INDEX idx_profiles_age ON profiles(age);

-- Create function for automatic updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for automatic updated_at
CREATE TRIGGER update_profiles_updated_at
    BEFORE UPDATE ON profiles
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Insert sample data for testing
INSERT INTO profiles (id, first_name, last_name, age, tg, email, description, profession_category) VALUES
    ('fad6142e-e147-4c9c-b3d6-9950478d0557','Иван', 'Иванов', 25, '@ivanov', 'ivanov@example.com', 'Опытный разработчик с 5 годами опыта', 'IT'),
    ('fad6142e-e247-4c9c-b3d6-9950478d0557','Мария', 'Петрова', 28, '@maria', 'petrova@example.com', 'UI/UX дизайнер с креативным подходом', 'Design'),
    ('fad6142e-e147-4c9c-b3d6-9950408d0557','Алексей', 'Сидоров', 30, '@alex', 'sidorov@example.com', 'Маркетолог с опытом в digital', 'Marketing'),
    ('fad6242e-e147-4c9c-b3d6-9950478d0557','Елена', 'Козлова', 22, '@elena', 'kozlova@example.com', 'Начинающий разработчик', 'IT'),
    ('fad6142e-e1a7-4c9c-b3d6-9950478d0557','Дмитрий', 'Смирнов', 35, '@dmitry', 'smirnov@example.com', 'Менеджер проектов', 'Management');