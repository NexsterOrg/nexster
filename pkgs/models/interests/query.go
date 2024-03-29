package interests

const listBasicInterestInfoQry string = `FOR doc IN interests
	FILTER DATE_ISO8601(doc.expireAt) <= DATE_ISO8601(DATE_NOW())
	LIMIT @limit
	RETURN { "_key": doc._key, "name": doc.name }`

const listYtVideosForInterestQry string = `FOR v IN 2..2 ANY @userNode interestsIn, interestBelongsTo
	FILTER v.ytVideos != null
	RETURN v.ytVideos`
