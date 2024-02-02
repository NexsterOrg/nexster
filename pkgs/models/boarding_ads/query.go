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

const listAdsWithFilters string = `FOR doc IN boardingAds
	FILTER doc.status == @status
	FILTER doc.rent >= @minRent && doc.rent <= @maxRent
	FILTER doc.distance <= @maxDistance
	FILTER doc.beds >= @minBeds && doc.beds <= @maxBeds
	FILTER doc.baths >= @minBaths && doc.baths <= @maxBaths
	FILTER doc.gender IN @genders
	FILTER doc.bills IN @billTypes
	SORT %s 
	LIMIT @offset, @count
	LET ownerAddress = (
		FOR v IN 1..1 OUTBOUND doc boardingAdOwned
		LIMIT 1
		RETURN v.address
	 )
	RETURN { "key" : doc._key, "title": doc.title, "imageUrls": doc.imageUrls, "rent": doc.rent, "ownerAddr": ownerAddress,
	"beds": doc.beds, "baths": doc.baths, "gender": doc.gender, "distance": doc.distance, "createdAt": doc.createdAt,
	"locationSameAsOwner": doc.locationSameAsOwner, "address": doc.address }`

const countAdsWithFilter string = `LET totalCount = LENGTH(FOR doc IN boardingAds
	FILTER doc.status == @status
	FILTER doc.rent >= @minRent && doc.rent <= @maxRent
	FILTER doc.distance <= @maxDistance
	FILTER doc.beds >= @minBeds && doc.beds <= @maxBeds
	FILTER doc.baths >= @minBaths && doc.baths <= @maxBaths
	FILTER doc.gender IN @genders
	FILTER doc.bills IN @billTypes
	RETURN 1) RETURN totalCount`
