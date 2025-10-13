package users

type UserRole struct {
	Name string
}

type UserDetails struct {
	FullName   string
	Level      uint32
	Department string
}

type User struct {
	ID             uint32
	Email          string
	Details        UserDetails
	AccountBalance float64
}
