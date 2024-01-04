package dispatcher

func MatchRoutes(dtx DispatchContext) *ResponseWriter {
	return DMap[dtx.Domain](dtx)
}
