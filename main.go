package main
import (
	"fmt"
	"log"
	"os"
	"errors"
	"path"
	"time"
	"sync"
	"net/http"
	"encoding/json"
	"strconv"
	"github.com/gorilla/mux"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
)
var lock sync.Mutex

type db_env struct {
	db *sql.DB
}

type company struct{
	Id int `json:"id"`
	Name string `json:"name"`
	Ceo_admin_name string `json:"ceo_admin_name"`
	Year_founded int `json:"year_founded"`
	Country_origin string `json:"country_origin"`
}

type rocket struct{
	Id int `json:"id"`
	Name string `json:"name"`
	Tons_to_leo float32 `json:"tons_to_leo"`
	No_of_stages int `json:"no_of_stages"`
	No_of_boosters int `json:"no_of_boosters"`
	Producer int `json:"producer_id"`
}

type launch struct{
	Id int `json:"id"`
	Mission_name string `json:"mission_name"`
	Launch_date_timestamp int `json:"launch_date_timestamp"`
	Launch_date string `json:"launch_date"`
	Vehicle int `json:"vehicle"`
	Tons_launched float32 `json:"tons_launched"`
	Outcome string `json:"outcome"`
}

type rocket_form_template struct{
	Companies []company
}

type launch_form_template struct{
	Rockets []rocket
}

//func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request){
//	w.Header().Set("Content-Type", "application/json")
//	w.WriteHeader(http.StatusOK)
//	w.Write([]byte(`{"message": "hello world"}`))
//}

func (env *db_env) check_table_exists(table_name string) bool{
	_, table_check := env.db.Query("SELECT * FROM "+ table_name +";")
	if table_check == nil {
		return true
	}
	return false
}

func (env *db_env) create_table(query string, table_name string){
	log.Println("Creating "+ table_name +" table...")
	prep_query, err := env.db.Prepare(query)
	if err != nil{
		log.Fatal(err.Error())
	}
	prep_query.Exec() // here query is executed and table created
	log.Println(table_name+" table created")
}

func (env *db_env) init_tables(){
	lock.Lock()
	defer lock.Unlock()
	
	company_table := `CREATE TABLE companies (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"name" TEXT,
		"ceo_admin_name" TEXT,
		"year_founded" INTEGER,
		"country_origin" TEXT
	);`

	rockets_table := `CREATE TABLE rockets (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"name" TEXT,
		"no_of_stages" INTEGER,
		"no_of_boosters" INTEGER DEFAULT 0,
		"producer" INTEGER NOT NULL,
		"tons_to_leo" REAL,
		FOREIGN KEY(producer) REFERENCES companies(id)
	);`


	launch_table := `CREATE TABLE launches (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"mission_name" TEXT,
		"launch_date" INTEGER,
		"vehicle" INTEGER,
		"tons_launched" REAL,
		"outcome" TEXT,
		FOREIGN KEY(vehicle) REFERENCES rockets(id)
	);`

	table_exists := env.check_table_exists("companies")
	if !table_exists{
		env.create_table(company_table, "companies")
	}
	table_exists = env.check_table_exists("rockets")
	if !table_exists{
		env.create_table(rockets_table, "rockets")
	}
	table_exists = env.check_table_exists("launches")
	if !table_exists{
		env.create_table(launch_table, "launches")
	}
}

func (env *db_env) insert_row(table_name string, query string){
	log.Println("Inserting record into "+table_name+" table")
	statement, err := env.db.Prepare(query)
	if err != nil{
		log.Fatalln(err.Error())
	}
	_, err  = statement.Exec()
	if err != nil{
		log.Fatalln(err.Error())
	}
}

func (env *db_env) populate_companies_table(){
	log.Println("Populating companies table...")
	rocket_lab := `INSERT INTO companies(name, ceo_admin_name, year_founded, country_origin) VALUES('Rocket Lab', 'Peter Beck', 2006, 'NZ');`
	spacex := `INSERT INTO companies(name, ceo_admin_name, year_founded, country_origin) VALUES ('SpaceX', 'Elon Musk', 2002, 'US');`
	ula := `INSERT INTO companies(name, ceo_admin_name, year_founded, country_origin) VALUES ('United Launch Alliance', 'Tory Bruno', 2006, 'US');`
	ariane_group := `INSERT INTO companies(name, ceo_admin_name, year_founded, country_origin) VALUES ('ArianeGroup', '--', 2015, 'FR');`
	env.insert_row("companies", rocket_lab)
	env.insert_row("companies", spacex)
	env.insert_row("companies", ula)
	env.insert_row("companies", ariane_group)
	log.Println("companies table populated")
}


func (env *db_env) populate_rockets_table(){
	var spacex_id int
	var rklb_id int
	spacex_row := env.db.QueryRow("SELECT id FROM companies WHERE name LIKE 'SpaceX'")
	rklb_row := env.db.QueryRow("SELECT id FROM companies WHERE name LIKE 'Rocket Lab'")
	
	switch err := spacex_row.Scan(&spacex_id); err{
		case sql.ErrNoRows:
			log.Println("Row does not exist!")
		case nil:
			log.Println(spacex_id)
		default:
			panic(err)
	}
	
	switch err := rklb_row.Scan(&rklb_id); err{
		case sql.ErrNoRows:
			log.Println("Row does not exist!")
		case nil:
			log.Println(rklb_id)
		default:
			panic(err)
	}
	
	log.Println("Populating rockets table...")
	
	falcon_9 := fmt.Sprintf(`INSERT INTO rockets(name, no_of_stages, producer, tons_to_leo) VALUES('Falcon 9', 2, %d, 22.8);`, spacex_id)
	falcon_heavy := fmt.Sprintf(`INSERT INTO rockets(name, no_of_stages, no_of_boosters, producer, tons_to_leo) VALUES('Falcon Heavy', 2, 2, %d, 63.8);`, spacex_id)
	electron := fmt.Sprintf(`INSERT INTO rockets(name, no_of_stages, producer, tons_to_leo) VALUES('Electron', 2, %d, 0.3);`, rklb_id)
	env.insert_row("rockets", falcon_9)
	env.insert_row("rockets", falcon_heavy)
	env.insert_row("rockets", electron)
	log.Println("rockets table populated")
}

func (env *db_env) populate_tables(){
	var count_rows int
	rows, err := env.db.Query("SELECT COUNT(*) FROM companies")
	if err != nil{
		log.Fatal(err)
	}
	for rows.Next(){
		if err := rows.Scan(&count_rows); err != nil{
			log.Fatal(err)
		}
	}
	if count_rows == 0{
		env.populate_companies_table()
	}
	rows.Close()
	rows, err = env.db.Query("SELECT COUNT(*) FROM rockets")
	if err != nil{
		log.Fatal(err)
	}
	for rows.Next(){
		if err := rows.Scan(&count_rows); err != nil{
			log.Fatal(err)
		}
	}
	if count_rows == 0{
		env.populate_rockets_table()
	}
	rows.Close()	
}

func list_rockets(db *sql.DB) []rocket{
	row, err := db.Query("SELECT * FROM rockets;")
	if err != nil{
		log.Fatal(err)
	}
	records := []rocket{}
	defer row.Close()
	for row.Next(){
		var vehicle rocket
		err = row.Scan(&vehicle.Id, &vehicle.Name, &vehicle.No_of_stages, &vehicle.No_of_boosters, &vehicle.Producer, &vehicle.Tons_to_leo)
		if err != nil {
			log.Fatal(err)
		}
		records = append(records, vehicle)
		log.Println(vehicle.Name)
	}
	return records
}

func (env *db_env) list_companies() []company{
	rows, err := env.db.Query("SELECT * FROM companies;")
	if err != nil{
		log.Fatal(err)
	}
	records := []company{}
	defer rows.Close()
	for rows.Next(){
		var record company
		err = rows.Scan(&record.Id, &record.Name, &record.Ceo_admin_name, &record.Year_founded, &record.Country_origin)
		if err != nil {
			log.Fatal(err)
		}
		records = append(records, record)
	}
	return records
}

func (env *db_env) list_launches() []launch{
	rows, err := env.db.Query("SELECT * FROM launches;")
	if err != nil{
		log.Fatal(err)
	}
	
	records := []launch{}
	defer rows.Close()
	for rows.Next(){
		var record launch
		err = rows.Scan(&record.Id, &record.Mission_name, &record.Launch_date_timestamp, &record.Vehicle, &record.Tons_launched, &record.Outcome)
		if err != nil {
			log.Fatal(err)
		}
		record.Launch_date = env.parse_time_from_int(record.Launch_date_timestamp)
		records = append(records, record)
	}
	return records
}

func (env *db_env) rockets_get(w http.ResponseWriter, r *http.Request){
	var rockets []rocket
	
	rockets = list_rockets(env.db)	
	log.Println(len(rockets))
	json_rockets, _ := json.Marshal(rockets)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"rockets": `+ string(json_rockets) +`}`))
}

func (env *db_env) companies_get(w http.ResponseWriter, r *http.Request){
	var companies []company
	
	companies = env.list_companies()
	json_records, _ := json.Marshal(companies)
	log.Println("Getting companies list...")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(string(json_records)))
}

func (env *db_env) launches_get(w http.ResponseWriter, r *http.Request){
	var launches []launch
	
	launches = env.list_launches()
	json_records, _ := json.Marshal(launches)
	log.Println("Getting launches list...")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(string(json_records)))
	
}

func (env *db_env) get_rocket_record(w http.ResponseWriter, r *http.Request) {
	var vehicle rocket
	
	params := mux.Vars(r)
	record_id := params["id"]
	
	err := env.db.QueryRow("SELECT * FROM rockets WHERE id=?", record_id).Scan(&vehicle.Id, &vehicle.Name, &vehicle.No_of_stages, &vehicle.No_of_boosters, &vehicle.Producer, &vehicle.Tons_to_leo)
	if err !=nil{
		if err == sql.ErrNoRows{
			log.Println("Record with such id doesn't exist")
		}
	}
	
	json_record, _ := json.Marshal(vehicle)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(string(json_record)))
}

func (env *db_env) get_company_record(w http.ResponseWriter, r *http.Request){
	var record company
	params := mux.Vars(r)
	record_id := params["id"]
	err := env.db.QueryRow("SELECT * FROM companies WHERE id=?", record_id).Scan(&record.Id, &record.Name, &record.Ceo_admin_name, &record.Year_founded, &record.Country_origin)
	if err !=nil{
		if err == sql.ErrNoRows{
			log.Println("Record with such id doesn't exist")
		}
	}
	
	json_record, _ := json.Marshal(record)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(string(json_record)))
	
}

func (env *db_env) get_launch_record(w http.ResponseWriter, r *http.Request){
	var record launch
	params := mux.Vars(r)
	record_id := params["id"]
	err := env.db.QueryRow("SELECT * FROM launches WHERE id=?", record_id).Scan(&record.Id, &record.Mission_name, &record.Launch_date_timestamp, &record.Vehicle, &record.Tons_launched, &record.Outcome)
	if err !=nil{
		if err == sql.ErrNoRows{
			log.Println("Record with such id doesn't exist")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message": "Record with such id doesn't exist"}`))
		}
	}else{
	
		record.Launch_date = env.parse_time_from_int(record.Launch_date_timestamp)
	
		json_record, _ := json.Marshal(record)
	
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(string(json_record)))
	}
}

func (env *db_env) show_rocket_list(w http.ResponseWriter, r *http.Request) {
	fp := path.Join("./templates", "list_rockets.html")
	tmpl, err := template.ParseFiles(fp)
	
	companies_records := env.list_companies()
	companies := rocket_form_template{companies_records}
	if err != nil{
		log.Fatal(err)
	}
	err = tmpl.Execute(w, companies)
	if err != nil{
		log.Fatal(err)
	}
}


func (env *db_env) show_companies_list(w http.ResponseWriter, r *http.Request) {
	fp := path.Join("./templates", "list_companies.html")
	tmpl, err := template.ParseFiles(fp)
	
	companies_records := env.list_companies()
	if err != nil{
		log.Fatal(err)
	}
	err = tmpl.Execute(w, companies_records)
	if err != nil{
		log.Fatal(err)
	}
}

func (env *db_env) show_launch_list(w http.ResponseWriter, r *http.Request) {
	fp := path.Join("./templates", "list_launches.html")
	tmpl, err := template.ParseFiles(fp)
	
	rockets_records := list_rockets(env.db)
	rockets := launch_form_template{rockets_records}
	if err != nil{
		log.Fatal(err)
	}
	err = tmpl.Execute(w, rockets)
	if err != nil{
		log.Fatal(err)
	}
}

func (env *db_env) create_rocket_record(w http.ResponseWriter, r *http.Request){
	var vehicle rocket
	
	r.ParseForm()
	
	vehicle.Name = r.FormValue("name")
	vehicle.Producer, _ = strconv.Atoi(r.FormValue("producer"))
	tons_to_leo, _ := strconv.ParseFloat(r.FormValue("tons_to_leo"), 32)
	vehicle.Tons_to_leo = float32(tons_to_leo)
	vehicle.No_of_stages, _ = strconv.Atoi(r.FormValue("no_of_stages"))
	vehicle.No_of_boosters, _ = strconv.Atoi(r.FormValue("no_of_boosters"))
	db_query := fmt.Sprintf("INSERT INTO rockets(name, tons_to_leo, no_of_stages, no_of_boosters, producer) VALUES ('%s', %f, %d, %d, %d);", vehicle.Name, vehicle.Tons_to_leo, vehicle.No_of_stages, vehicle.No_of_boosters, vehicle.Producer)
	env.insert_row("rockets", db_query)
	log.Println(vehicle.Name)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Rocket record created succesfully"}`))
}

func (env *db_env) create_launch_record(w http.ResponseWriter, r *http.Request){
	var record launch
	
	r.ParseForm()
	
	record.Mission_name = r.FormValue("mission_name")
	record.Vehicle, _ = strconv.Atoi(r.FormValue("vehicle"))
	tons_launched, _ := strconv.ParseFloat(r.FormValue("tons_launched"), 32)
	record.Tons_launched = float32(tons_launched)
	record.Launch_date = r.FormValue("launch_date")
	record.Launch_date_timestamp = env.parse_time_from_str(r.FormValue("launch_date"))
	record.Outcome = r.FormValue("outcome")
	db_query := fmt.Sprintf("INSERT INTO launches(mission_name, tons_launched, launch_date, vehicle, outcome) VALUES ('%s', %f, %d, %d, '%s');", record.Mission_name, record.Tons_launched, record.Launch_date_timestamp, record.Vehicle, record.Outcome)
	log.Println(db_query)
	env.insert_row("launches", db_query)
	
	log.Println(record.Mission_name)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Launch record created succesfully"}`))
}

func (env *db_env) create_company_record(w http.ResponseWriter, r *http.Request){
	var record company
	
	r.ParseForm()
	
	record.Name = r.FormValue("name")
	record.Year_founded, _ = strconv.Atoi(r.FormValue("year_founded"))
	record.Ceo_admin_name = r.FormValue("ceo_name")
	record.Country_origin = r.FormValue("country_origin")
	
	db_query := fmt.Sprintf("INSERT INTO companies(name, year_founded, ceo_admin_name, country_origin) VALUES ('%s', %d, '%s', '%s');", record.Name, record.Year_founded, record.Ceo_admin_name, record.Country_origin)
	log.Println(db_query)
	env.insert_row("companies", db_query)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Company record created succesfully"}`))

}

func (env *db_env) parse_time_from_str(time_utc string) int{
	const longForm = "2006-01-02T15:04"
	t, err := time.Parse(longForm, time_utc)
	if err != nil{
		log.Fatal(err)
	}
	return int(t.Unix())
}


func (env *db_env) parse_time_from_int(timestamp int) string{
	time_time := time.Unix(int64(timestamp),0)
	time_s := fmt.Sprintf("%s", time_time)
	return time_s	
}

func get(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "get method called"}`))
}

func post(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "post method called"}`))
}

func notFound(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"message": "not found"}`))
}

func create_db(db_name string){
	log.Println("Creating database...")
	file, err := os.Create(db_name)
	if err != nil {
		log.Fatal(err.Error())
	}
	file.Close()
	log.Println("Created database")
}

func main(){
	db_name :=  os.Args[1]

	if _, err := os.Stat(db_name); errors.Is(err, os.ErrNotExist){
		log.Println("Database doesn't exist")
		create_db(db_name)
	}

	log.Println("Opening database")
	sqlitedb, _ := sql.Open("sqlite3", "./"+db_name)
	log.Println("Database opened")
	defer sqlitedb.Close()
	env := db_env{db: sqlitedb}
	env.init_tables()
	go env.populate_tables()
	
	env.parse_time_from_str("2022-01-10T13:00")

	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()
	
	api.HandleFunc("/", get).Methods(http.MethodGet)
	api.HandleFunc("/", post).Methods(http.MethodPost)
	api.HandleFunc("/", notFound)
	
	api.HandleFunc("/rockets", env.rockets_get).Methods(http.MethodGet)
	api.HandleFunc("/companies", env.companies_get).Methods(http.MethodGet)
	api.HandleFunc("/launches", env.launches_get).Methods(http.MethodGet)
	
	api.HandleFunc("/rockets/{id:[0-9]+}", env.get_rocket_record).Methods(http.MethodGet)
	api.HandleFunc("/companies/{id:[0-9]+}", env.get_company_record).Methods(http.MethodGet)
	api.HandleFunc("/launches/{id:[0-9]+}", env.get_launch_record).Methods(http.MethodGet)
	
	api.HandleFunc("/rockets", env.create_rocket_record).Methods(http.MethodPost)
	api.HandleFunc("/launches", env.create_launch_record).Methods(http.MethodPost)
	api.HandleFunc("/companies", env.create_company_record).Methods(http.MethodPost)
	
	
	r.HandleFunc("/rockets", env.show_rocket_list).Methods(http.MethodGet)
	r.HandleFunc("/launches", env.show_launch_list).Methods(http.MethodGet)
	r.HandleFunc("/companies", env.show_companies_list).Methods(http.MethodGet)
	
	log.Fatal(http.ListenAndServe(":8080", r))
}

