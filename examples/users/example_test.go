package users

import (
	"fmt"
	"unsafe"

	cgocopy "github.com/shaban/cgocopy/pkg/cgocopy"
)

func Example_cgocopy_users() {
	// Create C user data
	cUsersPtr, count := createSampleUsers()
	if cUsersPtr == nil {
		panic("createUsers returned nil")
	}
	defer freeSampleUsers(cUsersPtr, count)

	// Get C struct size from metadata (registered in init())
	meta := cgocopy.GetMetadata[User]()
	if meta == nil {
		panic("User not registered")
	}

	// Copy C users to Go - v2 API is super simple!
	goUsers := make([]User, count)
	for i := range goUsers {
		// Calculate pointer to each C struct in the array
		cPtr := unsafe.Add(cUsersPtr, uintptr(i)*meta.Size)

		// Copy using v2 generic API - registration happens in init()!
		user, err := cgocopy.Copy[User](cPtr)
		if err != nil {
			panic(err)
		}
		goUsers[i] = user
	}

	for _, u := range goUsers {
		fmt.Printf("User %d: %s\n", u.ID, u.Details.FullName)
		fmt.Printf("  Email: %s\n", u.Email)
		fmt.Printf("  Department: %s (Level %d)\n", u.Details.Department, u.Details.Level)
		fmt.Printf("  Balance: $%.2f\n\n", u.AccountBalance)
	}

	// Output:
	// User 1000: Ada Lovelace
	//   Email: ada@example.com
	//   Department: Mathematics (Level 1)
	//   Balance: $1234.56
	//
	// User 1001: Alan Turing
	//   Email: alan@example.net
	//   Department: Computer Science (Level 2)
	//   Balance: $1276.56
}
