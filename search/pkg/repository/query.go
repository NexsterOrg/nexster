package repository

const listAllUsersQry string = `FOR v IN users
	RETURN  { "key": v._key, "username": v.username, "image_url": v.image_url, "batch": v.batch, "indexNo": v.index_no,
		"field": v.field, "faculty": v.faculty }`
