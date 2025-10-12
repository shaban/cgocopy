package users

import (
	"fmt"
	"unsafe"

	cgocopy "github.com/shaban/cgocopy/pkg/cgocopy"
)

func Example_cgocopy_users() {
	cgocopy.Reset()
	defer cgocopy.Reset()

	if err := registerUserStructs(); err != nil {
		panic(err)
	}

	cgocopy.Finalize()

	cUsersPtr, count := newSampleUsers()
	if cUsersPtr == nil {
		panic("createUsers returned nil")
	}
	defer freeSampleUsers(cUsersPtr, count)

	mapping, ok := cgocopy.GetMapping[User]()
	if !ok {
		panic("User mapping missing after registration")
	}

	goUsers := make([]User, count)
	for i := range goUsers {
		cPtr := unsafe.Add(cUsersPtr, uintptr(i)*mapping.CSize)
		if err := cgocopy.Copy(&goUsers[i], cPtr); err != nil {
			panic(err)
		}
	}

	for _, u := range goUsers {
		fmt.Printf("User %d: %s (level %d)\n", u.ID, u.Details.FullName, u.Details.Level)
		fmt.Printf("  Email: %s\n", u.Email)
		fmt.Printf("  Roles:\n")
		for _, role := range u.Details.Roles {
			fmt.Printf("    - %s\n", role.Name)
		}
		fmt.Printf("  Balance: $%.2f\n\n", u.AccountBalance)
	}

	// Output:
	// User 1000: Ada Lovelace (level 1)
	//   Email: ada@example.com
	//   Roles:
	//     - author
	//     - client
	//     - developer
	//     - admin
	//   Balance: $1234.56
	//
	// User 1001: Alan Turing (level 2)
	//   Email: alan@example.net
	//   Roles:
	//     - admin
	//     - moderator
	//     - support
	//     - analyst
	//   Balance: $1276.56
}

func registerUserStructs() error {
	if err := cgocopy.RegisterStruct[User](cgocopy.DefaultCStringConverter); err != nil {
		return fmt.Errorf("register User: %w", err)
	}
	return nil
}
