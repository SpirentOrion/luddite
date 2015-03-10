package main

import "luddite"

func main() {
	s := luddite.NewService()
	s.AddCollectionResource("/users", newUserResource())
	s.Run(":8000")
}
