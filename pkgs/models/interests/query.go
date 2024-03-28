package interests

const listBasicInterestInfoQry string = `FOR doc IN interests
	FILTER DATE_ISO8601(doc.expireAt) <= DATE_ISO8601(DATE_NOW())
	LIMIT @limit
	RETURN { "_key": doc._key, "name": doc.name }`
