package media

const ListMediaAfterGivenDateQry string = `FOR md IN media
	FILTER md.visibility == @visibility && md.created_date > DATE_ISO8601(@fromDate)
	RETURN md._key`
