-- **do not alter - add new migrations instead**

BEGIN;


CREATE FUNCTION last_agg_sfunc (state anyelement, newval anyelement) RETURNS anyelement AS $$
	SELECT newval;
$$ LANGUAGE SQL IMMUTABLE PARALLEL SAFE;

CREATE AGGREGATE last_agg (anyelement) (
	SFUNC = last_agg_sfunc
	, STYPE = anyelement
);


COMMIT;
