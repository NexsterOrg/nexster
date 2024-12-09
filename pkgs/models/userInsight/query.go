package userInsight

const getActiveUserCountForGivenTimeRangeQry string = `RETURN LENGTH(
    FOR doc IN userInsights
        FILTER doc.type == "activeUser" && TO_NUMBER(doc.year) >= DATE_YEAR(@from) && TO_NUMBER(doc.year) <= DATE_YEAR(@to)
        LET validTimestamps = (
            FOR ts IN doc.loginTimestamps
                FILTER ts >= @from && ts <= @to
                LIMIT 1
                RETURN ts
        )
        FILTER LENGTH(validTimestamps) > 0
        COLLECT userId = doc.userId INTO uniqueUserIds
        RETURN userId
)`
