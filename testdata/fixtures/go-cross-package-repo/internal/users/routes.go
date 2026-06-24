package users

func Routes() Router {
	r := NewRouter()
	r.Get("/users", listUsers)
	r.Post("/users", createUser)
	return r
}
