package repository

const occurrenceCountQry string = `FOR doc IN boardingOwners
	FILTER doc.mainContact == @mainContact
	return doc._key`
