package db

import ("fmt"
	"database/sql"
	_ "github.com/alexbrainman/odbc"
	"log"
)

var db2 *sql.DB

func GetDb2(){
	// ********************************************************
	// Create the database handle, confirm driver is present
	// *********************************************************
	connectString := &configuration.db2user + ":" + configuration.db2pass + configuration.db2db
	log.Println(connectString)
	db2, err = sql.Open("odbc", "DSN=" + configuration.db2Server  )
	if err != nil {
		log.Fatalf("Error on initializing database connection: %s", err.Error())
	}
	fmt.Println("db opened at root:****@/test")
	db2.SetMaxIdleConns(100)
	defer db2.Close()
	// make sure connection is available
	err = db2.Ping()
	if err != nil {
		log.Fatalf("Error on opening database connection: %s", err.Error())
	}else {fmt.Println("verified db is open")}

}

var (
	CLAIM_ID int
	DETAIL_LINE_ID int
	BILL_NUMBER string
	CLAIM_STATUS string
	user3 string
	CONTACT_NAME string
	CLAIMANT_EMAIL string
	ASSIGNED_TO string
	ITEM_DESCR string
	ITEM_IS_REQUIRED string
)

//Check for email
func CheckSetlmentStatus(){

	var sqlStr = `select distinct C.CLAIM_ID,
		t.DETAIL_LINE_ID, t.BILL_NUMBER,
		C.CLAIM_STATUS, C.user3, c.CONTACT_NAME, c.CLAIMANT_EMAIL,
		LC.ASSIGNED_TO,
		LI.ITEM_DESCR, LI.ITEM_IS_REQUIRED
		from CLAIM C
		join LIST_CHECKIN LC on lc.LIST_CODE = c.CLAIM_ID
		join LIST_ITEM LI on lc.LIST_ID = li.ITEM_ID
		join TLORDER T on C.ORDER_ID = t.DETAIL_LINE_ID
		join CLIENT CL on T.CUSTOMER = CL.CLIENT_ID
		where C.CLAIM_ID = 2019001099
		and LC.LIST_ID = 100
		and C.CLAIM_STATUS in ('CLOSED', 'OPEN')
		and lc.IS_COMPLETE = 'True'
		`
	fmt.Print(sqlStr)
	rows, err := db.Query(sqlStr)
	if err != nil {
		log.Println(err)
	}

	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&CLAIM_ID, &DETAIL_LINE_ID, &BILL_NUMBER, &CLAIM_STATUS, &user3, &CONTACT_NAME, &CLAIMANT_EMAIL, &ASSIGNED_TO, &ITEM_DESCR, &ITEM_IS_REQUIRED)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(CLAIM_ID, DETAIL_LINE_ID, BILL_NUMBER, CLAIM_STATUS, user3, CONTACT_NAME, CLAIMANT_EMAIL, ASSIGNED_TO, ITEM_DESCR, ITEM_IS_REQUIRED)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	log.Println()
}
