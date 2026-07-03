CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL,
    slug VARCHAR(150) UNIQUE NOT NULL,
    parent_id UUID REFERENCES categories(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS equipment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    condition VARCHAR(20) NOT NULL DEFAULT 'used' CHECK (condition IN ('new', 'used')),
    region VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_equipment_category ON equipment(category_id);
CREATE INDEX IF NOT EXISTS idx_equipment_owner ON equipment(owner_id);

INSERT INTO categories (name, slug) VALUES
    ('Excavators', 'excavators'),
    ('Cranes', 'cranes'),
    ('Generators', 'generators'),
    ('Compressors', 'compressors'),
    ('Welding Equipment', 'welding-equipment'),
    ('Concrete Mixers', 'concrete-mixers'),
    ('Loaders', 'loaders'),
    ('Trucks & Transport', 'trucks-transport')
ON CONFLICT (slug) DO NOTHING;
