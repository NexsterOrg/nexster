package repository

const upcomingEventsQry string = `FOR doc IN events
	FILTER DATE_TIMESTAMP(doc.date) >= DATE_NOW()
	SORT DATE_TIMESTAMP(doc.date) ASC
	LIMIT @offset, @count
	LET res1 = (
		FOR v IN 1..1 OUTBOUND doc._id postedBy
		RETURN { "key": v._key, "username": v.username, "indexNo": v.index_no }
	)
	LET res2 = (
		FOR v, e IN 1..1 INBOUND doc._id eventReactedBy
		COLLECT love = e.love , going = e.going WITH COUNT INTO cnt
		SORT null
		RETURN {going , love, "count": cnt }
	)

	RETURN { "key": doc._key, "link": doc.link, "title": doc.title, "date": doc.date, "description": doc.description, 
	"venue": doc.venue, "mode": doc.mode, "eventLink": doc.eventLink, "createdAt": doc.createdAt, "postedBy": res1, "reactionStates": res2 }`

const getEventReactionKeyQry string = `FOR v, e IN 1..1 OUTBOUND @userNode eventReactedBy
	FILTER v._id == @eventNode
	RETURN e._key`

const getEventQry string = `FOR doc IN events
	FILTER doc._id == @eventNode
	LIMIT 1
	LET res1 = (
		FOR v IN 1..1 OUTBOUND @eventNode postedBy
		RETURN { "key": v._key, "username": v.username, "indexNo": v.index_no }
	)
	LET res2 = (
		FOR v, e IN 1..1 INBOUND @eventNode eventReactedBy
		COLLECT love = e.love , going = e.going WITH COUNT INTO cnt
		SORT null
		RETURN {going , love, "count": cnt }
	)
	RETURN { "key": doc._key, "link": doc.link, "title": doc.title, "date": doc.date, "description": doc.description, 
	"venue": doc.venue, "mode": doc.mode, "eventLink": doc.eventLink, "createdAt": doc.createdAt, "postedBy": res1, "reactionStates": res2 }`

const getEventLoveUserQry string = `FOR v, e IN 1..1 INBOUND @eventNode eventReactedBy
	FILTER e.love
	LIMIT @offset, @count
	RETURN { "key": v._key, "username": v.username, "imageUrl": v.image_url, "faculty": v.faculty, 
		"field": v.field, "batch": v.batch, "indexNo": v.index_no }`

const getOwnerUserKey string = `FOR v IN 1..1 OUTBOUND
	@eventNode postedBy
	RETURN v._key`
