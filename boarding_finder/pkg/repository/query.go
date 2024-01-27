package repository

const occurrenceCountQry string = `FOR doc IN boardingOwners
	FILTER doc.mainContact == @mainContact
	return doc._key`

const getEdgeFromToQry string = `FOR doc IN boardingAdOwned
	FILTER doc._from == @from && doc._to == @to
	RETURN doc._key`

const delEdgeFromToQry string = `FOR edge IN boardingAdOwned
	FILTER edge._from == @from && edge._to == @to
	REMOVE edge IN boardingAdOwned
	RETURN OLD._key`
