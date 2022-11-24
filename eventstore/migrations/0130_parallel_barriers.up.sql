-- **do not alter - add new migrations instead**

BEGIN;


CREATE OR REPLACE FUNCTION uuid_or_placeholder(str text)
RETURNS uuid AS $$
BEGIN
  RETURN str::uuid;
EXCEPTION WHEN invalid_text_representation THEN
  RETURN 'd5091c33-2f9d-4b15-82dc-4ad69717fc03'::uuid;
END;
$$ LANGUAGE plpgsql IMMUTABLE PARALLEL SAFE;


COMMIT;
