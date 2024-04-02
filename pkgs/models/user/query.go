package user

const listFacDepOfAllUsersQry string = `FOR doc IN users
	RETURN { key: doc._key, faculty: doc.faculty, field: doc.field }`
