package interestsin

const deleteByFromToQry string = `FOR doc IN interestsIn
	FILTER doc._from == @from && doc._to == @to
	REMOVE doc._key IN interestsIn`

const insertDocByFacDepName string = `FOR doc IN interestGroups
	FILTER doc.name == @facDepName
	INSERT {
	"_from": @userNode,
	"_to": doc._id,
	"kind": @kind
	} INTO interestsIn`

const listInterestedInEdgeForUserQry string = `FOR v IN 1..1 OUTBOUND @userNode interestsIn
	RETURN v._key`
