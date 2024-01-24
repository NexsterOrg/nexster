package repository

const occurrenceCountQry string = `FOR doc IN boardingOwner
	FILTER doc.mainContact == @mainContact
	return doc._key`
