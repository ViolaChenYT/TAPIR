//go:build exclude
package tapir

import (
	"encoding/json"
	"fmt"
)

type Account struct {
	account_id   string `json:"account_id"`    // must be unique
	user_name    string `json:"user_name"`
	balance      int    `json:"balance"`       // must be positive
}

// BankDB represents the database for storing bank accounts.
type BankDB struct {
  storage            *Client
	new_account_id     int
	account_id_length  int
}

// NewTwitterDB creates a new TwitterDB instance.
func NewBankDB(account_id_length int) *BankDB {
	return &BankDB{
		storage: ,// TODO: initialize storage,
		num_accounts: 0,
		account_id_length: account_id_length,
	}
}

func (db *BankDB) getAccount(account_id string) (*Account, error) {
	// Retrieve account from storage
	account_string, err := db.storage.Get(accountID)
	if err != nil {
		return nil, err
	}

	// Deserialize account from string
	var account Account
	err = json.Unmarshal([]byte(account_string), &account)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (db *BankDB) updateAccount(account_id string) error {
	// Serialize the account to string
	account_JSON, err := json.Marshal(account)
	if err != nil {
		return err
	}

	account_string := string(accountJSON)

	// Store the account_string with key account_id
	err := db.storage.Put(account_id, account_string)
	return err
}

// Create a new account for the given user, return the generate account id
func (db *BankDB) CreateAccount(user_name string) (string, error) {
	db.new_account_id++

	// Generate account ID
	account_id := fmt.Sprintf("%0*d", db.account_id_length, db.new_account_id)

	// Create a new account
	account := Account{
		account_id: account_id,
		user_name:  user_name,
		balance:   0,
	}

	// Put into storage
	db.storage.Begin()
	err := db.updateAccount(account_id, account)
	
	if err != nil {
		// TODO: Print debug message
		db.storage.Abort()
		return "", err
	}
	db.storage.Commit()

	return account_id, nil
}

// Return the balance of the account with the given account_id
func (db *BankDB) QueryBalance(account_id string) (int, error) {
	db.storage.Begin()
	account, err := db.getAccount(account_id)

	if err != nil {
		// TODO: Print debug message
		db.storage.Abort()
		return -1, err
	}

	db.storage.Commit()

	return account.balance, nil
}

// Deposit adds the specified amount to the balance of the account with the given account_id
func (db *BankDB) Deposit(account_id string, amount int) error {
	db.storage.Begin()
	account, err := db.getAccount(account_id)

	if err != nil {
		// TODO: Print debug message
		db.storage.Abort()
		return err
	}

	// Update balance
	account.balance += amount

	err := db.updateAccount(account_id, account)
	if err != nil {
		// TODO: Print debug message
		db.storage.Abort()
		return err
	}
	db.storage.Commit()

	return nil
}

// Withdraw deducts the specified amount from the balance of the account with the given accountID
func (db *BankDB) Withdraw(accountID string, amount int) error {
	db.storage.Begin()
	account, err := db.getAccount(account_id)

	if err != nil {
		// TODO: Print debug message
		db.storage.Abort()
		return err
	}

	// Check if the balance is sufficient for withdrawal
	if account.balance < amount {
		db.storage.Abort()
		return fmt.Errorf("insufficient balance for withdrawal")
	}

	// Update balance
	account.balance -= amount

	err := db.updateAccount(account_id, account)
	if err != nil {
		// TODO: Print debug message
		db.storage.Abort()
		return err
	}
	db.storage.Commit()

	return nil
}

// Rename changes the username of the account with the given accountID
func (db *BankDB) Rename(account_id string, new_user_name string) error {
	db.storage.Begin()
	account, err := db.getAccount(account_id)

	if err != nil {
		// TODO: Print debug message
		db.storage.Abort()
		return err
	}

	// Update user name
	account.user_name += new_user_name

	err := db.updateAccount(account_id, account)
	if err != nil {
		// TODO: Print debug message
		db.storage.Abort()
		return err
	}
	db.storage.Commit()

	return nil
}

// Transaction transfers the specified amount from senderAccountID to receiverAccountID
func (db *BankDB) Transaction(sender_account_id string, receiver_account_id string, amount int) error {
	db.storage.Begin()
	sender_account, err := db.getAccount(sender_account_id)

	if err != nil {
		// TODO: Print debug message
		db.storage.Abort()
		return err
	}

	receiver_account, err := db.getAccount(receiver_account_id)

	if err != nil {
		// TODO: Print debug message
		db.storage.Abort()
		return err
	}

	// Check if the balance is sufficient for withdrawal
	if sender_account.balance < amount {
		db.storage.Abort()
		return fmt.Errorf("insufficient balance for transaction")
	}

	// Update balance
	sender_account.balance -= amount
	receiver_account.balance += amount

	// Update storage for sender
	err := db.updateAccount(sender_account_id, sender_account)
	if err != nil {
		// TODO: Print debug message
		db.storage.Abort()
		return err
	}
	// Update storage for receiver
	err := db.updateAccount(receiver_account_id, receiver_account)
	if err != nil {
		// TODO: Print debug message
		db.storage.Abort()
		return err
	}
	db.storage.Commit()

	return nil
}



