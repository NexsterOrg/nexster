package media

const getVisibilitForLink string = `FOR doc in media
	FILTER doc.link == @link
	LIMIT 1
	RETURN doc.visibility
	`
