package db

import
(
	"database/sql"
	"fmt"
	"context"
	"log"
)
var server = "DLPSYNSQL01"
var port = 1433
var user = "sa"
var password = "Syn3rg1ze"

var db *sql.DB

func GetSqlDB (){
	// Connect to the DB
	var err error

	// Create connection string
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d",
	server, user, password, port)

	// Create connection pool
	db, err = sql.Open("sqlserver", connString)
	if err != nil {
	log.Fatal("Error creating connection pool: " + err.Error())
	}
	log.Printf("Connected!\n")

	// Close the database connection pool after program executes
	defer db.Close()
}
func SettlementNote(){

}

func AttachmentFiles()  {

//	-- DLPSYNSQL01 -  SQL server location
//	--DLPSYNSVR01 -- file location
//
//	select c.FBNumber, c.Bill_To_Code, c.In_DocFamilyID, m.In_DocLocation,m.In_DocID
//		from DELIVERYDOCS.dbo.Child C
//		inner join DELIVERYDOCS.dbo.Main M on M.In_DocID = C.In_DocFamilyID
//		where FBNumber = '71560627'
//and M.DeliveryDate is not null

}

// Gets and prints SQL Server version
func SelectVersion(){
	// Use background context
	ctx := context.Background()

	// Ping database to see if it's still alive.
	// Important for handling network issues and long queries.
	err := db.PingContext(ctx)
	if err != nil {
		log.Fatal("Error pinging database: " + err.Error())
	}

	var result string

	// Run query and scan for result
	err = db.QueryRowContext(ctx, "SELECT @@version").Scan(&result)
	if err != nil {
		log.Fatal("Scan failed:", err.Error())
	}
	fmt.Printf("%s\n", result)
}