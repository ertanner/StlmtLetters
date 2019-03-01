package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/gomail.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"database/sql"
	_ "github.com/alexbrainman/odbc"

	"./EmailPOC2/src/db/db.go"
)

//type MailConfig struct {
type IDNumber struct {
	ID         string   `json:"ID"`
	DocsNeeded []string `json:"DocsNeeded"`
	Trigger    string   `json:"Trigger"`
	From       string   `json:"From"`
	CC         string   `json:"CC"`
	To         string   `json:"To"`
	Subject    string   `json:"Subject"`
}
type MailConfig struct {
	ID_Numbers map[string]IDNumber `json:"ID_Numbers"`
}

type Claim struct {
	CLAIM_ID         int    `json:"CLAIM_ID"`
	DETAIL_LINE_ID   int    `json:"DETAIL_LINE_ID"`
	BILL_NUMBER      string `json:"BILL_NUMBER"`
	CLAIM_STATUS     string `json:"CLAIM_STATUS"`
	USER3            string `json:"USER3"`
	CONTACT_NAME     string `json:"CONTACT_NAME"`
	CLAIMANT_EMAIL   string `json:"CLAIMANT_EMAIL"`
	ASSIGNED_TO      string `json:"ASSIGNED_TO"`
	ITEM_DESCR       string `json:"ITEM_DESCR"`
	ITEM_IS_REQUIRED string `json:"ITEM_IS_REQUIRED"`
}

type Note struct {
	THE_NOTE string `json:"THE_NOTE"`
}

type Path struct {
	FBNumber    string `json:"FBNumber"`
	BillToCode  string `json:"Bill_To_Code"`
	DocName     string `json:"In_DocFamilyID"`
	DocLocation string `json:"In_DocLocation"`
	DocId       string `json:"In_DocID"`
	DocExt      string `json:"In_DocFileExt"`
}

type Config struct {
	Pwd string `json:"pwd"`
}

var err error

func main() {
	fmt.Println("Starting mail program.")
	Config := LoadConfiguration("acxpuwd")

	// open the db
	db, err := sql.Open("odbc",
		"etanner:judy123456789123@tcp(DLUTMDB08:50000)/DYLT_REP")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connecting to database...")

	log.Println("Database connection opened...")
	db.SetMaxOpenConns(10)
	//db.SetMaxIdleConns(3)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxIdleConns(0)

	err = db.Ping()
	if err != nil {
		fmt.Println("Ping error")
	}
	fmt.Println("DB opened")

	// fetch data
	var (
		id   int
		name string
	)
	rows, err := db.Query("select id, idname from TMWIN.DYLT_SETTLEMENT_LETTERS")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(id, name)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	b, err := ioutil.ReadFile("mail_config.json")
	if err != nil {
		panic(err)
	}

	m := make(map[string]MailConfig)
	err = json.Unmarshal(b, &m)
	fmt.Println(m, err)

	//MailConfig := LoadMailConfig("mail_config.json")
	//fmt.Println("gettign mail config")
	//fmt.Println(MailConfig.ID_Numbers)

	// fmt.Println("Pwd ="  + Config.Pwd)
	d := gomail.NewDialer("DLEXCH01.daylight.ads", 587, "etanner", Config.Pwd)

	s, err := d.Dial()
	if err != nil {
		panic(err)
	}

	//buffer, err := bimg.Read("image.jpg")
	//if err != nil {
	//	fmt.Fprintln(os.Stderr, err)
	//}
	//newImage, err := bimg.NewImage(buffer).Resize(800, 600)
	//if err != nil {
	//	fmt.Fprintln(os.Stderr, err)
	//}
	//
	//bimg.Write("new.jpg", newImage)

	// get a list of the claims to be processed
	claims := make([]Claim, 0)
	response, err := http.Get("http://localhost:1340/delDocStlmt")
	if err != nil {
		fmt.Printf("The http request failed with error %s\n", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		json.Unmarshal(data, &claims)
		fmt.Println(claims[1].CLAIM_ID)
	}

	// for each item in the list call the email function.
	for i := 0; i < len(claims); i++ {
		fmt.Println(strconv.Itoa(claims[i].CLAIM_ID) + " " + claims[i].BILL_NUMBER)
		sendEmail(s, claims[i].CLAIM_ID, claims[i].BILL_NUMBER, claims[i].CONTACT_NAME, claims[i].USER3, claims[i].DETAIL_LINE_ID)
	}
}

func sendEmail(s gomail.Sender, CLAIM_ID int, BILL string, CONTACT string, Assigned string, DLID int) {
	//send the email
	m := gomail.NewMessage()
	m.SetHeader("From", "etanner@dylt.com")
	m.SetHeader("To", "etanner@dylt.com") //, "claims@madegoods.com") //c.CLAIMANT_EMAIL
	m.SetAddressHeader("Cc", "etanner@dylt.com", "")

	// get the note for the DLID
	notes := make([]Note, 0)
	response, err := http.Get("http://localhost:1340/delDocNote?DLID=" + strconv.Itoa(DLID))
	if err != nil {
		fmt.Printf("The http request failed with error %s\n", err)
	} else {
		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("The ReadAll failed with error %s\n", err)
		} else {
			json.Unmarshal(data, &notes)
			//fmt.Println(notes[0].THE_NOTE)
		}
	}

	//get the delivery doc's from Synergize
	var pathtofile string
	var fileName string
	paths := make([]Path, 0)
	pthResponse, err := http.Get("http://localhost:1340/delDocs?billNo=" + BILL)
	if err != nil {
		fmt.Printf("The http request failed with error %s\n", err)
	} else {
		data, err := ioutil.ReadAll(pthResponse.Body)
		if err != nil {
			fmt.Printf("The ReadAll failed with error %s\n", err)
		} else {
			if len(data) < 1 {
				fmt.Println("Error no data was retrieved.")
			} else {
				json.Unmarshal(data, &paths)
				fmt.Println("Got to paths = " + strconv.Itoa(len(paths)))

				for i := 0; i < len(paths); i++ {
					fileName = paths[i].DocId + "." + paths[i].DocExt
					fmt.Println("File path and name: " + fileName)
					if len(pathtofile) != 0 {
						fmt.Println("Got to doc location" + paths[i].DocLocation)
						pathtofile = paths[i].DocLocation[15:len(paths[i].DocLocation)]
					} else {
						pathtofile = "0"
					}
				}
			}
		}
	}
	fmt.Println("Z:\\" + pathtofile)
	fmt.Println(fileName)

	// set up the email with the data from above
	var str = "Settlement Letter, Claim ID " + strconv.Itoa(CLAIM_ID) + ", for Pro " + BILL
	m.SetHeader("Subject", str)
	m.SetBody("text/html", "Dear <b>"+CONTACT+",</b>"+
		"<br><br>"+
		notes[0].THE_NOTE+
		"<br><br>"+
		"Closing: Sincerely,<br>"+
		"         "+Assigned+" <br>"+
		"         Claims Analyst <br><br>"+
		"Attachment: Delivery Receipt for the same Freight Bill #"+BILL+" from Synergize")

	m.Attach("Z:\\" + pathtofile + "\\" + fileName)

	// send the email.
	if err := gomail.Send(s, m); err != nil {
		log.Printf("Could not send email to %q: %v", CONTACT, err)
	}

	// wait 15 sec for the mailserver.  There is a limit to the # of email per min going out.
	timer1 := time.NewTimer(15 * time.Second)
	<-timer1.C

	m.Reset()
}
func LoadConfiguration(file string) Config {
	var config Config
	pwd, _ := os.Getwd()
	configFile, err := os.Open(pwd + "\\" + file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

func LoadMailConfig(file string) MailConfig {
	var mailConfig MailConfig
	pwd, _ := os.Getwd()

	configFile, err := os.Open(pwd + "\\" + file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&mailConfig)

	fmt.Println("mc")
	fmt.Println(mailConfig.ID_Numbers["0"].ID)
	fmt.Println("done mc")
	return mailConfig
}
