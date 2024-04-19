package user

const listFacDepOfAllUsersQry string = `FOR doc IN users
	RETURN { key: doc._key, faculty: doc.faculty, field: doc.field }`

const getUserSignupsQry string = `RETURN LENGTH (FOR user IN users
	FILTER user.createdAt >= @fromDate && user.createdAt <=@toDate
	RETURN user._key)`

//haritha mihimal all users
//const getAllUsersCountQry string = `RETURN LENGTH(FOR user IN users
//	RETURN user._key)`
//

const getAllUsersCountQry string = `
LET totalUsers = LENGTH(FOR user IN users RETURN user._key)
LET maleUsers = LENGTH(FOR user IN users FILTER user.gender == "male" RETURN user._key)
LET femaleUsers = LENGTH(FOR user IN users FILTER user.gender == "female" RETURN user._key)
RETURN { totalUsers, maleUsers, femaleUsers }`
