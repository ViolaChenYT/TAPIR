package tapir

import (
	"fmt"

	"github.com/ViolaChenYT/TAPIR/IR"
)

func RunApp() {
	// Create a new BankDB instance
	bankDB := NewBankDB(6) // Account IDs will have a length of 6 digits

	// Create two accounts
	accountID1, err := bankDB.CreateAccount("Alice")
	if err != nil {
		fmt.Println("Error creating account:", err)
		return
	}
	fmt.Println("Created account with ID:", accountID1)

	accountID2, err := bankDB.CreateAccount("Bob")
	if err != nil {
		fmt.Println("Error creating account:", err)
		return
	}
	fmt.Println("Created account with ID:", accountID2)

	// Deposit initial amounts
	err = bankDB.Deposit(accountID1, 1000)
	if err != nil {
		fmt.Println("Error depositing to account:", err)
		return
	}

	err = bankDB.Deposit(accountID2, 500)
	if err != nil {
		fmt.Println("Error depositing to account:", err)
		return
	}

	// Perform a transaction from account 1 to account 2
	err = bankDB.Transaction(accountID1, accountID2, 200)
	if err != nil {
		fmt.Println("Error performing transaction:", err)
		return
	}

	fmt.Println("Transaction completed successfully")

	// Retrieve and print updated account balances
	account1Balance, err := bankDB.QueryBalance(accountID1)
	if err != nil {
		fmt.Println("Error retrieving account:", err)
		return
	}
	fmt.Println("Account 1 balance:", account1.Balance)

	account2Balance, err := bankDB.QueryBalance(accountID2)
	if err != nil {
		fmt.Println("Error retrieving account:", err)
		return
	}
	fmt.Println("Account 2 balance:", account2.Balance)
}

func Main() {
	fmt.Println("Hello, Tapir!")
	server := IR.NewServer(1)
	server.Start()
}
