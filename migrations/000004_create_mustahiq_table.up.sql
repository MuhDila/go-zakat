CREATE TABLE IF NOT EXISTS mustahiq (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    phoneNumber VARCHAR(20) NOT NULL UNIQUE,
    address TEXT NOT NULL,
    asnafID UUID NOT NULL REFERENCES asnaf(id) ON DELETE RESTRICT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('active', 'inactive', 'pending')),
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mustahiq_name ON mustahiq(name);
CREATE INDEX IF NOT EXISTS idx_mustahiq_status ON mustahiq(status);
CREATE INDEX IF NOT EXISTS idx_mustahiq_asnaf_id ON mustahiq(asnafID);
