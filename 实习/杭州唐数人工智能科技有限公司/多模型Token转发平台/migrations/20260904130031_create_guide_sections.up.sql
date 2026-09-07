-- Create the guide_sections table for the platform usage guide.
CREATE TABLE IF NOT EXISTS guide_sections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    title VARCHAR(255) NOT NULL,
    content_md TEXT NOT NULL,
    audience VARCHAR(20) NOT NULL DEFAULT 'user' CHECK (audience IN ('user', 'admin')),
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS guide_section_audience ON guide_sections(audience);
CREATE INDEX IF NOT EXISTS guide_section_is_enabled ON guide_sections(is_enabled);
CREATE INDEX IF NOT EXISTS guide_section_sort_order ON guide_sections(sort_order);
