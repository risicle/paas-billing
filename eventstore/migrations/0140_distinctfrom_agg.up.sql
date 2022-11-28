-- **do not alter - add new migrations instead**

BEGIN;


CREATE TYPE distinctfrom_agg_stype_uuid (prior boolean, prev uuid);

CREATE FUNCTION distinctfrom_agg_sfunc (state anyelement, newval anyelement) RETURNS anyelement AS $$
	-- prior being null indicates this being the first element, which is never distinct
	SELECT
		state.prior IS NOT NULL AND (state.prior OR state.prev IS DISTINCT FROM newval) AS prior,
		newval AS prev;
$$ LANGUAGE SQL IMMUTABLE PARALLEL SAFE;

CREATE FUNCTION distinctfrom_agg_finalfunc (state anyelement) RETURNS boolean AS $$
	SELECT state.prior;
$$ LANGUAGE SQL IMMUTABLE PARALLEL SAFE;

CREATE AGGREGATE distinctfrom_agg (uuid) (
	SFUNC = distinctfrom_agg_sfunc
	, STYPE = distinctfrom_agg_stype_uuid
	, FINALFUNC = distinctfrom_agg_finalfunc
);


COMMIT;
