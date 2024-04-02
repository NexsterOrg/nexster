package mediaowner

const getOwnerForMediaQry string = `FOR v IN 1..1 OUTBOUND @mediaNode mediaOwnerEdges 
	RETURN v._id`
