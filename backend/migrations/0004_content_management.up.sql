-- Content Management Migration
-- Creates tables for managing static pages and FAQs

-- Create content_pages table for static pages (About, Policies, etc.)
CREATE TABLE IF NOT EXISTS content_pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    content TEXT NOT NULL,
    type TEXT NOT NULL, -- 'about', 'policy', etc.
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    meta_title TEXT,
    meta_description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create faqs table for frequently asked questions
CREATE TABLE IF NOT EXISTS faqs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'General',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for content_pages
CREATE INDEX IF NOT EXISTS idx_content_pages_type ON content_pages(type);
CREATE INDEX IF NOT EXISTS idx_content_pages_is_active ON content_pages(is_active);
CREATE INDEX IF NOT EXISTS idx_content_pages_slug ON content_pages(slug);

-- Create indexes for faqs
CREATE INDEX IF NOT EXISTS idx_faqs_category ON faqs(category);
CREATE INDEX IF NOT EXISTS idx_faqs_is_active ON faqs(is_active);
CREATE INDEX IF NOT EXISTS idx_faqs_sort_order ON faqs(sort_order);

-- Insert default content pages
INSERT INTO content_pages (title, slug, content, type, meta_title, meta_description) VALUES
(
    'Our Story',
    'our-story',
    '# Our Story

Ethnic Treasure was founded in 2015 with a simple mission: to preserve and celebrate India''s rich textile heritage while providing sustainable livelihoods to skilled artisans across the country.

## Our Journey

What began as a small passion project has grown into a movement that connects over 500+ artisans with appreciative customers worldwide. We work directly with weaving communities, ensuring fair wages and preserving traditional techniques that have been passed down through generations.

## Our Values

- **Authenticity**: Every piece in our collection is handcrafted using traditional techniques
- **Sustainability**: We use natural dyes and eco-friendly processes wherever possible
- **Fair Trade**: Artisans receive fair compensation for their craftsmanship
- **Preservation**: We actively work to preserve endangered craft forms

## The Artisans Behind Your Treasures

Each piece in our collection tells a story - of the artisan who created it, the tradition it represents, and the future it helps sustain. From Banarasi weavers to Kanchipuram silk experts, our partners are the true custodians of India''s textile heritage.

Join us in our mission to keep these traditions alive, one beautiful creation at a time.',
    'about',
    'Our Story - Ethnic Treasure | Handcrafted Traditional Indian Clothing',
    'Learn about Ethnic Treasure''s journey since 2015, our commitment to preserving India''s textile heritage, and the 500+ artisans who create our handcrafted collections.'
),
(
    'Our Artisans',
    'artisans',
    '# Our Artisans

At Ethnic Treasure, we believe that behind every beautiful creation is a skilled artisan whose hands bring life to threads. Our artisan community is the heart and soul of everything we do.

## Meet Our Weaving Communities

### Banarasi Weavers
- Location: Varanasi, Uttar Pradesh
- Specialty: intricate brocade work and gold/silver thread weaving
- Generations of expertise: 8+ on average

### Kanchipuram Silk Experts
- Location: Kanchipuram, Tamil Nadu
- Specialty: pure silk sarees with temple borders
- Known for: durability and vibrant colors

### Chikankari Artisans
- Location: Lucknow, Uttar Pradesh
- Specialty: delicate white-on-white embroidery
- Technique: 32 different stitch types

## Empowering Artisans

We work directly with artisan communities, eliminating middlemen and ensuring:

- Fair wages that reflect true craftsmanship
- Safe working conditions
- Skills development programs
- Healthcare and education support
- Preservation of traditional techniques

## Artisan Stories

Every month, we feature stories from our artisan community, sharing their dreams, challenges, and the meaning behind their craft. These aren''t just products - they''re legacies being woven into the fabric of modern India.',
    'about',
    'Our Artisans - Ethnic Treasure | Meet the Master Craftsmen',
    'Meet the skilled artisans behind Ethnic Treasure''s handcrafted collections. From Banarasi weavers to Chikankari experts, discover the stories and traditions.'
),
(
    'Crafts & Techniques',
    'crafts',
    '# Crafts & Techniques

India''s textile heritage spans thousands of years, with each region developing unique techniques that have been perfected over generations. At Ethnic Treasure, we celebrate and preserve these ancient crafts.

## Weaving Techniques

### Handloom Weaving
- **Process**: Traditional wooden looms operated entirely by hand
- **Time**: 2-4 weeks for a single saree
- **Regions**: Varanasi, Kanchipuram, Mysore, Pochampally

### Block Printing
- **Process**: Hand-carved wooden blocks dipped in natural dyes
- **Regions**: Jaipur, Bagru, Sanganer
- **Specialty**: Geometric patterns and floral motifs

### Tie & Dye (Bandhani)
- **Process**: Tiny dots tied before dyeing to create patterns
- **Regions**: Rajasthan, Gujarat
- **Meaning**: Each dot represents a blessing

## Embroidery Styles

### Chikankari
- **Origin**: Lucknow, 16th century
- **Technique**: 32 different stitch types
- **Character**: Delicate white-on-white work

### Zardozi
- **Technique**: Gold and silver thread embroidery
- **History**: Mughal era royal courts
- **Modern Use**: Bridal and festive wear

### Kantha
- **Origin**: Bengal
- **Technique**: Running stitch storytelling
- **Sustainability**: Upcycled fabric layers

## Dyeing Traditions

### Natural Dyes
- **Sources**: Plants, minerals, and insects
- **Benefits**: Eco-friendly and therapeutic
- **Colors**: Indigo, turmeric, madder, lac

### Bandhani Dyeing
- **Process**: Tie-resist dyeing
- **Patterns**: Dots, waves, and geometric designs
- **Cultural Significance**: Wedding and festival traditions

## Preserving Heritage

We actively work to preserve these techniques through:
- Documentation of traditional methods
- Training programs for younger generations
- Research and development of sustainable practices
- Fair trade partnerships with artisan communities

Each piece in our collection is a testament to these living traditions, carrying forward centuries of craftsmanship into the modern world.',
    'about',
    'Traditional Indian Crafts & Techniques | Ethnic Treasure Heritage',
    'Explore India''s rich textile heritage - handloom weaving, block printing, embroidery styles, and natural dyeing techniques preserved by Ethnic Treasure artisans.'
);

-- Insert default FAQs
INSERT INTO faqs (question, answer, category, sort_order) VALUES
('What is the return policy?', 'We offer a 30-day return policy for all unused items in original packaging. Please contact our customer service team to initiate a return. Refunds are processed within 5-7 business days after we receive the returned item.', 'Returns', 1),
('How long does shipping take?', 'Standard shipping takes 5-7 business days within India. Express shipping takes 2-3 business days. International shipping takes 10-15 business days. You can track your order using the tracking number provided after dispatch.', 'Shipping', 1),
('Do you ship internationally?', 'Yes, we ship to over 50 countries worldwide. International shipping rates and delivery times vary by destination. Please check our shipping policy for detailed information.', 'Shipping', 2),
('How do I know my size?', 'We provide detailed size charts for all our products. Each product page includes measurements in both inches and centimeters. If you''re unsure, our customer service team can help you find the perfect fit.', 'Products', 1),
('Are your products authentic?', 'Absolutely! All our products are handcrafted by skilled artisans using traditional techniques. We provide authenticity certificates with our premium products and work directly with artisan communities.', 'Products', 2),
('What payment methods do you accept?', 'We accept all major credit/debit cards, UPI, net banking, and cash on delivery (for select locations). All transactions are secured with industry-standard encryption.', 'Payments', 1),
('How do I care for my ethnic wear?', 'Each product comes with specific care instructions. Generally, we recommend dry cleaning for silk and embroidered items, gentle hand washing for cotton, and avoiding direct sunlight to preserve colors.', 'Products', 3),
('Do you offer customization?', 'Yes, we offer customization services for select products. Please contact our team at least 4 weeks before your required date to discuss your customization needs.', 'General', 1),
('How can I track my order?', 'Once your order is dispatched, you''ll receive a tracking number via email. You can use this number to track your order on our website or the courier''s tracking portal.', 'Orders', 1),
('What if I receive a damaged item?', 'We take utmost care in packaging, but if you receive a damaged item, please contact us within 48 hours with photos. We''ll arrange for a replacement or refund immediately.', 'Returns', 2);
