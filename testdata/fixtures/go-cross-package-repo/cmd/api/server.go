package api

import "example.com/repomind/internal/users"

func register(r Router) {
	r.Mount("/api", users.Routes())
}
