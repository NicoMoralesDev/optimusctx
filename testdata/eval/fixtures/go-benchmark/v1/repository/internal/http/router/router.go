package router

type Route struct {
	Method string
	Path   string
	Name   string
}

func Routes() []Route {
	return []Route{
		{Method: "GET", Path: "/healthz", Name: "health"},
		{Method: "GET", Path: "/rollouts", Name: "rollouts"},
	}
}
