-- Add the new 500-row option for installations still using the original defaults.
-- Custom page-size lists are left unchanged.
UPDATE settings
SET value = '[10,20,50,100,500]',
    updated_at = NOW()
WHERE key = 'table_page_size_options'
  AND value = '[10,20,50,100]';
