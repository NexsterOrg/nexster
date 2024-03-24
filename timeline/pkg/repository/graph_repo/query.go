package graphrepo

const listInterestGroupsQry string = `FOR v IN 1..1 OUTBOUND @userNode interestsIn
RETURN {"key": v._key, "name": v.name, "type": v.type }`
