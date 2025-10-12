package users

type UserRole struct {
	Name string
}

type UserDetails struct {
	FullName string
	Roles    []UserRole
	Level    uint32
}

type User struct {
	ID             uint32
	Email          string
	Details        UserDetails
	AccountBalance float64
}
