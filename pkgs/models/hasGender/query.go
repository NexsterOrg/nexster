package hasgender

const deleteByFromToQry string = `FOR doc IN hasGender
	FILTER doc._from == @from && doc._to == @to
	REMOVE doc._key IN hasGender`
