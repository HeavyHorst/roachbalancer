package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/HeavyHorst/roachbalancer/balancer"
	_ "github.com/lib/pq"
)

func main() {
	b := balancer.New("root", "/certs", false, "xxx.xxx.xxx.xxx:26257")
	go b.Listen(0) // 0 means random high port
	b.WaitReady()

	// Connect to the "bank" database.
	db, err := sql.Open("postgres",
		fmt.Sprintf("postgresql://maxroach@%s/bank?ssl=true&sslmode=require&sslrootcert=/certs/ca.crt&sslkey=/certs/client.maxroach.key&sslcert=/certs/client.maxroach.crt", b.GetAddr()))
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}
	defer db.Close()

	// Create the "accounts" table.
	if _, err := db.Exec(
		"CREATE TABLE IF NOT EXISTS accounts (id INT PRIMARY KEY, balance INT)"); err != nil {
		log.Fatal(err)
	}

	// Insert two rows into the "accounts" table.
	if _, err := db.Exec(
		"INSERT INTO accounts (id, balance) VALUES (3, 2000), (4, 250)"); err != nil {
		log.Fatal(err)
	}

	// Print out the balances.
	rows, err := db.Query("SELECT id, balance FROM accounts")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	fmt.Println("Initial balances:")
	for rows.Next() {
		var id, balance int
		if err := rows.Scan(&id, &balance); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%d %d\n", id, balance)
	}
}
