package boardingads

const getAdWithOwnerQry string = `FOR v, e, p IN 1..1 OUTBOUND @adId boardingAdOwned
	LET toVertex = p.vertices[1]
	RETURN {
		"from": p.vertices[0],
		"to": {
			"_key": toVertex._key,
			"createdAt": toVertex.createdAt,
			"name": toVertex.name,
			"mainContact": toVertex.mainContact,
			"otherContacts": toVertex.otherContacts,
			"address": toVertex.address,
			"location": toVertex.location,
			"status": toVertex.status,
		}
	}`
