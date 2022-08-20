package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"html"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Service struct {
	Id          int
	Name        string
	Description string
	Status      string
	Created     string
	Updated     string
}

type Item struct {
	Id          int
	Sid         int
	Sname       string
	Name        string
	Description string
	Protocol    string
	Cidr        string
	Port        string
	Status      string
	Created     string
	Updated     string
}

type Rule struct {
	Id          int
	Sid         int
	Did         int
	Source      string
	Destination string
	Name        string
	Description string
	Status      string
	Created     string
	Updated     string
}

type ServiceEdit struct {
	Service Service
	Items   []Item
	SAcls   []AclSrch
	DAcls   []AclSrch
}

type ItemEdit struct {
	Item     Item
	Services []Service
	SAcls    []AclSrch
	DAcls    []AclSrch
}

type ItemAdd struct {
	ServiceID int
	Services  []Service
}

type RuleEdit struct {
	Rule     Rule
	Services []Service
	Acls     []AclSrch
}

type AclSrch struct {
	Protocol    string
	Source      string
	Destination string
	Port        string
	Rid         int
	Rname       string
	Rstatus     string
	Siid        int
	Sitem       string
	SIstatus    string
	Diid        int
	Ditem       string
	DIstatus    string
	Ssid        int
	Sservice    string
	SSstatus    string
	Dsid        int
	Dservice    string
	DSstatus    string
}

type Event struct {
	Id      int
	Type    string
	Event   string
	Created string
}

var database *sql.DB

// ====================================================================== Index ==============================================================================================
// функция стартовйо страницы
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/index.html")
}

// ====================================================================== Services ===========================================================================================
// ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// функция удаления данных и запись в лог - Service
func ServicesDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	_, err := database.Exec("insert into acl.events (type, event) values ('Delete Service:', (select concat('{ \"Name\": \"',name,'\"}' ) from services where id= ? ))", id)
	database.Exec("delete from acl.services  where id = ?", id)
	if err != nil {
		log.Println(err)
		tmpl, _ := template.ParseFiles("templates/dberr.html")
		tmpl.Execute(w, err)
		return
	}

	http.Redirect(w, r, "/services", http.StatusMovedPermanently)
}

// ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// функция предварительно возвращает пользователю страницу для редактирования объекта - Service
func ServicesEditPage(w http.ResponseWriter, r *http.Request) {
	var sdata ServiceEdit
	vars := mux.Vars(r)
	id := vars["id"]
	//  выборка всех Items которые добавлены в этот Service ----------------------------------------------
	rows, err := database.Query("SELECT i.id, s.id AS sid, s.name AS sname, i.name, i.description, i.protocol, i.cidr, i.port, i.status FROM items i JOIN services s ON i.service_id=s.id WHERE s.id = ? ORDER BY id DESC", id)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	for rows.Next() {
		itm := Item{}
		err := rows.Scan(&itm.Id, &itm.Sid, &itm.Sname, &itm.Name, &itm.Description, &itm.Protocol, &itm.Cidr, &itm.Port, &itm.Status)
		if err != nil {
			fmt.Println(err)
			continue
		}
		sdata.Items = append(sdata.Items, itm)
	}

	// выборка  Source Acls которое входят в этот Service ------------------------------------------------
	srows, err := database.Query("SELECT d.protocol, s.cidr AS source, d.cidr AS destination, d.port, r.id AS rid, r.name AS rname, r.status AS rstatus, s.id AS siid, s.name AS sitem, s.status AS sistatus, d.id AS diid, d.name AS ditem, d.status AS distatus, sr.id AS ssid, sr.name AS sservice, sr.status AS ssstatus, dr.id AS dsid, dr.name AS dservice, dr.status AS dsstatus FROM rules r JOIN items s ON r.source=s.service_id JOIN items d ON r.destination=d.service_id JOIN services sr ON r.source=sr.id JOIN services dr ON r.destination=dr.id  WHERE sr.id = ? ;", id)
	if err != nil {
		log.Println(err)
	}
	defer srows.Close()
	for srows.Next() {
		acl := AclSrch{}
		err := srows.Scan(&acl.Protocol, &acl.Source, &acl.Destination, &acl.Port, &acl.Rid, &acl.Rname, &acl.Rstatus, &acl.Siid, &acl.Sitem, &acl.SIstatus, &acl.Diid, &acl.Ditem, &acl.DIstatus, &acl.Ssid, &acl.Sservice, &acl.SSstatus, &acl.Dsid, &acl.Dservice, &acl.DSstatus)
		if err != nil {
			fmt.Println(err)
			continue
		}
		sdata.SAcls = append(sdata.SAcls, acl)
	}

	// выборка  Destination Acls которое входят в этот Service -------------------------------------------
	drows, err := database.Query("SELECT d.protocol, s.cidr AS source, d.cidr AS destination, d.port, r.id AS rid, r.name AS rname, r.status AS rstatus, s.id AS siid, s.name AS sitem, s.status AS sistatus, d.id AS diid, d.name AS ditem, d.status AS distatus, sr.id AS ssid, sr.name AS sservice, sr.status AS ssstatus, dr.id AS dsid, dr.name AS dservice, dr.status AS dsstatus FROM rules r JOIN items s ON r.source=s.service_id JOIN items d ON r.destination=d.service_id JOIN services sr ON r.source=sr.id JOIN services dr ON r.destination=dr.id  WHERE dr.id = ? ;", id)
	if err != nil {
		log.Println(err)
	}
	defer drows.Close()
	for drows.Next() {
		acl := AclSrch{}
		err := drows.Scan(&acl.Protocol, &acl.Source, &acl.Destination, &acl.Port, &acl.Rid, &acl.Rname, &acl.Rstatus, &acl.Siid, &acl.Sitem, &acl.SIstatus, &acl.Diid, &acl.Ditem, &acl.DIstatus, &acl.Ssid, &acl.Sservice, &acl.SSstatus, &acl.Dsid, &acl.Dservice, &acl.DSstatus)
		if err != nil {
			fmt.Println(err)
			continue
		}
		sdata.DAcls = append(sdata.DAcls, acl)
	}

	// выборка значений Service для редактирования -------------------------------------------------------
	row := database.QueryRow("select id, name, description, status from acl.services where id = ? ORDER BY id", id)
	//itm := Item{}
	err = row.Scan(&sdata.Service.Id, &sdata.Service.Name, &sdata.Service.Description, &sdata.Service.Status)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(404), http.StatusNotFound)
	} else {
		tmpl, _ := template.ParseFiles("templates/servicesedit.html")
		// tmpl.Execute(w, itm)
		e := tmpl.Execute(w, sdata)
		if e != nil {
			log.Println(e)
		}
	}
}

// --------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// функция изменения данных и запись в лог - Service
func ServicesEditHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}
	id := r.FormValue("id")
	name := r.FormValue("name")
	description := r.FormValue("description")
	status := r.FormValue("status")

	// транзакция
	ok := true
	database.Exec("begin")
	_, err = database.Exec("update acl.services set name=?, description=?, status = ? where id = ?", name, description, status, id)
	database.Exec("insert into acl.events (type, event) values ('Update Service:', ?)", "{ \"Name\": \""+name+"\", \"Description\": \""+description+"\", \"Status\": \""+status+"\" }")
	if err != nil {
		ok = false
	}
	if ok {
		database.Exec("commit")
	} else {
		database.Exec("rollback")
	}
	if err != nil {
		log.Println(err)
		tmpl, _ := template.ParseFiles("templates/dberr.html") // вывод ошибки бд
		tmpl.Execute(w, err)
		return
	}

	http.Redirect(w, r, "/services", http.StatusMovedPermanently)
}

// --------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// функция добавления данных и запись в лог - Service
func ServicesCreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}
		name := r.FormValue("name")
		description := r.FormValue("description")
		status := r.FormValue("status")

		// транзакция
		ok := true
		database.Exec("begin")
		_, err = database.Exec("insert into acl.services (name, description, status) values (?, ?, ?)", name, description, status)
		database.Exec("insert into acl.events (type, event) values ('Add Service:', ?)", "{ \"Name\": \""+name+"\", \"Description\": \""+description+"\", \"Status\": \""+status+"\" }")
		if err != nil {
			ok = false
		}
		if ok {
			database.Exec("commit")
		} else {
			database.Exec("rollback")
		}
		if err != nil {
			log.Println(err)
			tmpl, _ := template.ParseFiles("templates/dberr.html") // вывод ошибки бд
			tmpl.Execute(w, err)
			return
		}

		http.Redirect(w, r, "/services", http.StatusMovedPermanently)
	} else {
		http.ServeFile(w, r, "templates/servicesadd.html")
	}
}

//---------------------------------------------------------------------------------------------------------------------------------------------------------------------------
//  функция  вывода данных - Service
func ServicesIndexHandler(w http.ResponseWriter, r *http.Request) {

	rows, err := database.Query("select * from acl.services ORDER BY id DESC")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	services := []Service{}

	for rows.Next() {
		p := Service{}
		err := rows.Scan(&p.Id, &p.Name, &p.Description, &p.Status, &p.Created, &p.Updated)
		if err != nil {
			fmt.Println(err)
			continue
		}
		services = append(services, p)
	}

	tmpl, _ := template.ParseFiles("templates/services.html")
	tmpl.Execute(w, services)
}

// ====================================================================== Items ==============================================================================================
// ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// функция удаления данных и запись в лог - Item
func ItemsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	_, err := database.Exec("insert into acl.events (type, event) values ('Delete Item:', (select concat('{ \"Name\": \"',name,'\"}' ) from items where id= ? ))", id)
	database.Exec("delete from acl.items where id = ?", id)
	if err != nil {
		log.Println(err)
		tmpl, _ := template.ParseFiles("templates/dberr.html")
		tmpl.Execute(w, err)
		return
	}

	http.Redirect(w, r, "/items", http.StatusMovedPermanently)
}

//---------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// функция предварительно возвращает пользователю страницу для редактирования объекта - Item
func ItemsEditPage(w http.ResponseWriter, r *http.Request) {
	var sdata ItemEdit
	vars := mux.Vars(r)
	id := vars["id"]

	//  выборка всех Services для выбора -----------------------------------------------------------------
	rows, err := database.Query("SELECT * FROM acl.services ORDER BY id")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	//services := []Service{}
	for rows.Next() {
		p := Service{}
		err := rows.Scan(&p.Id, &p.Name, &p.Description, &p.Status, &p.Created, &p.Updated)
		if err != nil {
			fmt.Println(err)
			continue
		}
		sdata.Services = append(sdata.Services, p)
	}

	// выборка  Source Acls которое входят в  этот Items -------------------------------------------------
	srows, err := database.Query("SELECT d.protocol, s.cidr AS source, d.cidr AS destination, d.port, r.id AS rid, r.name AS rname, r.status AS rstatus, s.id AS siid, s.name AS sitem, s.status AS sistatus, d.id AS diid, d.name AS ditem, d.status AS distatus, sr.id AS ssid, sr.name AS sservice, sr.status AS ssstatus, dr.id AS dsid, dr.name AS dservice, dr.status AS dsstatus FROM rules r JOIN items s ON r.source=s.service_id JOIN items d ON r.destination=d.service_id JOIN services sr ON r.source=sr.id JOIN services dr ON r.destination=dr.id  WHERE s.id = ? ;", id)
	if err != nil {
		log.Println(err)
	}
	defer srows.Close()
	for srows.Next() {
		acl := AclSrch{}
		err := srows.Scan(&acl.Protocol, &acl.Source, &acl.Destination, &acl.Port, &acl.Rid, &acl.Rname, &acl.Rstatus, &acl.Siid, &acl.Sitem, &acl.SIstatus, &acl.Diid, &acl.Ditem, &acl.DIstatus, &acl.Ssid, &acl.Sservice, &acl.SSstatus, &acl.Dsid, &acl.Dservice, &acl.DSstatus)
		if err != nil {
			fmt.Println(err)
			continue
		}
		sdata.SAcls = append(sdata.SAcls, acl)
	}

	// выборка  Destination Acls которое входят в этот Items ---------------------------------------------
	drows, err := database.Query("SELECT d.protocol, s.cidr AS source, d.cidr AS destination, d.port, r.id AS rid, r.name AS rname, r.status AS rstatus, s.id AS siid, s.name AS sitem, s.status AS sistatus, d.id AS diid, d.name AS ditem, d.status AS distatus, sr.id AS ssid, sr.name AS sservice, sr.status AS ssstatus, dr.id AS dsid, dr.name AS dservice, dr.status AS dsstatus FROM rules r JOIN items s ON r.source=s.service_id JOIN items d ON r.destination=d.service_id JOIN services sr ON r.source=sr.id JOIN services dr ON r.destination=dr.id  WHERE d.id = ? ;", id)
	if err != nil {
		log.Println(err)
	}
	defer drows.Close()
	for drows.Next() {
		acl := AclSrch{}
		err := drows.Scan(&acl.Protocol, &acl.Source, &acl.Destination, &acl.Port, &acl.Rid, &acl.Rname, &acl.Rstatus, &acl.Siid, &acl.Sitem, &acl.SIstatus, &acl.Diid, &acl.Ditem, &acl.DIstatus, &acl.Ssid, &acl.Sservice, &acl.SSstatus, &acl.Dsid, &acl.Dservice, &acl.DSstatus)
		if err != nil {
			fmt.Println(err)
			continue
		}
		sdata.DAcls = append(sdata.DAcls, acl)
	}

	// выборка значений Item для редактирования ----------------------------------------------------------
	row := database.QueryRow("SELECT i.id, s.id AS sid, s.name AS sname, i.name, i.description, i.protocol, i.cidr, i.port, i.status FROM items i JOIN services s ON i.service_id=s.id WHERE i.id = ?", id)
	//itm := Item{}
	err = row.Scan(&sdata.Item.Id, &sdata.Item.Sid, &sdata.Item.Sname, &sdata.Item.Name, &sdata.Item.Description, &sdata.Item.Protocol, &sdata.Item.Cidr, &sdata.Item.Port, &sdata.Item.Status)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(404), http.StatusNotFound)
	} else {
		tmpl, _ := template.ParseFiles("templates/itemsedit.html")
		// tmpl.Execute(w, itm)
		e := tmpl.Execute(w, sdata)
		if e != nil {
			log.Println(e)
		}
	}
}

//---------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// функция изменения данных и запись в лог - Item
func ItemsEditHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}
	id := r.FormValue("id")
	service_id := r.FormValue("sid")
	name := r.FormValue("name")
	description := r.FormValue("description")
	protocol := r.FormValue("protocol")
	cidr := r.FormValue("cidr")
	port := r.FormValue("port")
	status := r.FormValue("status")

	//выборка  имени Services ----------------------------------------------------------------------------
	var sname string
	srow := database.QueryRow("select name from services where id = ?", service_id)
	err = srow.Scan(&sname)
	if err != nil {
		log.Println(err)
	}

	// транзакция
	ok := true
	database.Exec("begin")
	_, err = database.Exec("update acl.items set service_id=?, name=?, description=?, protocol=?, cidr=?, port=?, status = ? where id = ?", service_id, name, description, protocol, cidr, port, status, id)
	database.Exec("insert into acl.events (type, event) values ('Update Item:', ?)", "{ \"Service\": \""+sname+"\", \"Name\": \""+name+"\", \"Description\": \""+description+"\", \"Protocol\": \""+protocol+"\", \"Cidr\": \""+cidr+"\", \"Port\": \""+port+"\", \"Status\": \""+status+"\" }")
	if err != nil {
		ok = false
	}
	if ok {
		database.Exec("commit")
	} else {
		database.Exec("rollback")
	}
	if err != nil {
		log.Println(err)
		tmpl, _ := template.ParseFiles("templates/dberr.html")
		tmpl.Execute(w, err)
		return
	}

	http.Redirect(w, r, "/items", http.StatusMovedPermanently)
}

// --------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// функция возвращает пользователю страницу для добавления объекта items предварительно получив список services -  Items
func ItemsCreatePage(w http.ResponseWriter, r *http.Request) {

	//предварительно выбрать id сервиса
	var sdata ItemAdd
	ServiceID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err == nil {
		sdata.ServiceID = ServiceID
	}

	rows, err := database.Query("select id as sid, name as sname from acl.services ORDER BY id")
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(404), http.StatusNotFound)
	} else {
		defer rows.Close()
		sdata.Services = []Service{}

		for rows.Next() {
			p := Service{}
			err := rows.Scan(&p.Id, &p.Name)
			if err != nil {
				fmt.Println(err)
				continue
			}
			sdata.Services = append(sdata.Services, p)
		}

		tmpl, _ := template.ParseFiles("templates/itemsadd.html")
		tmpl.Execute(w, sdata)
	}
}

//--------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// функция добавления данных и запись в лог - Item
func ItemsCreateHandler(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)
	//service_id := vars["Sid"]
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}
	service_id := r.FormValue("sid")
	name := r.FormValue("name")
	description := r.FormValue("description")
	protocol := r.FormValue("protocol")
	cidr := r.FormValue("cidr")
	port := r.FormValue("port")
	status := r.FormValue("status")

	//выборка  имени Services ----------------------------------------------------------------------------
	var sname string
	srow := database.QueryRow("select name from services where id = ?", service_id)
	err = srow.Scan(&sname)
	if err != nil {
		log.Println(err)
	}

	// транзакция
	ok := true
	database.Exec("begin")
	_, err = database.Exec("insert into acl.items (service_id, name, description, protocol, cidr, port, status) values (?, ?, ?, ?, ?, ?, ?)", service_id, name, description, protocol, cidr, port, status)
	database.Exec("insert into acl.events (type, event) values ('Add Item:', ?)", "{ \"Service\": \""+sname+"\", \"Name\": \""+name+"\", \"Description\": \""+description+"\", \"Protocol\": \""+protocol+"\", \"Cidr\": \""+cidr+"\", \"Port\": \""+port+"\", \"Status\": \""+status+"\" }")
	if err != nil {
		ok = false
	}
	if ok {
		database.Exec("commit")
	} else {
		database.Exec("rollback")
	}
	if err != nil {
		log.Println(err)
		tmpl, _ := template.ParseFiles("templates/dberr.html")
		tmpl.Execute(w, err)
		return
	}

	http.Redirect(w, r, "/items", http.StatusMovedPermanently)
}

//---------------------------------------------------------------------------------------------------------------------------------------------------------------------------
//  функция  вывода данных - Item
func ItemsIndexHandler(w http.ResponseWriter, r *http.Request) {

	rows, err := database.Query("select items.id, services.id as sid, services.name as sname, items.name, items.description, items.protocol, items.cidr, items.port, items.status, items.created, items.updated from items join services on services.id = items.service_id ORDER BY id DESC;")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	items := []Item{}

	for rows.Next() {
		p := Item{}
		err := rows.Scan(&p.Id, &p.Sid, &p.Sname, &p.Name, &p.Description, &p.Protocol, &p.Cidr, &p.Port, &p.Status, &p.Created, &p.Updated)
		if err != nil {
			fmt.Println(err)
			continue
		}
		items = append(items, p)
	}

	tmpl, _ := template.ParseFiles("templates/items.html")
	tmpl.Execute(w, items)
}

// ====================================================================== Rules =============================================================================================

//---------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// функция удаления данных и запись в лог - Rule
func RulesDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	_, err := database.Exec("insert into acl.events (type, event) values ('Delete Rule:', (select concat('{ \"Name\": \"',name,'\"}' ) from rules where id= ? ))", id)
	database.Exec("delete from acl.rules where id = ?", id)
	if err != nil {
		log.Println(err)
		tmpl, _ := template.ParseFiles("templates/dberr.html")
		tmpl.Execute(w, err)
		return
	}

	http.Redirect(w, r, "/rules", http.StatusMovedPermanently)
}

//--------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// функция предварительно возвращает пользователю страницу для редактирования объекта - Rule
func RulesEditPage(w http.ResponseWriter, r *http.Request) {
	var sdata RuleEdit
	vars := mux.Vars(r)
	id := vars["id"]

	//  выборка всех Services  ---------------------------------------------------------------------------
	rows, err := database.Query("select * from acl.services ORDER BY id")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	//services := []Service{}
	for rows.Next() {
		p := Service{}
		err := rows.Scan(&p.Id, &p.Name, &p.Description, &p.Status, &p.Created, &p.Updated)
		if err != nil {
			fmt.Println(err)
			continue
		}
		sdata.Services = append(sdata.Services, p)
	}

	// выборка  Acls  которое входят в эту Rule ----------------------------------------------------------
	arows, err := database.Query("SELECT d.protocol, s.cidr AS source, d.cidr AS destination, d.port, r.id AS rid, r.name AS rname, r.status AS rstatus, s.id AS siid, s.name AS sitem, s.status AS sistatus, d.id AS diid, d.name AS ditem, d.status AS distatus, sr.id AS ssid, sr.name AS sservice, sr.status AS ssstatus, dr.id AS dsid, dr.name AS dservice, dr.status AS dsstatus FROM rules r JOIN items s ON r.source=s.service_id JOIN items d ON r.destination=d.service_id JOIN services sr ON r.source=sr.id JOIN services dr ON r.destination=dr.id  WHERE r.id = ?;", id)
	if err != nil {
		log.Println(err)
	}
	defer arows.Close()
	for arows.Next() {
		acl := AclSrch{}
		err := arows.Scan(&acl.Protocol, &acl.Source, &acl.Destination, &acl.Port, &acl.Rid, &acl.Rname, &acl.Rstatus, &acl.Siid, &acl.Sitem, &acl.SIstatus, &acl.Diid, &acl.Ditem, &acl.DIstatus, &acl.Ssid, &acl.Sservice, &acl.SSstatus, &acl.Dsid, &acl.Dservice, &acl.DSstatus)
		if err != nil {
			fmt.Println(err)
			continue
		}
		sdata.Acls = append(sdata.Acls, acl)
	}

	// выборка значений rule для редактирования ----------------------------------------------------------
	row := database.QueryRow("SELECT r.id, s.id AS sid, s.name AS sname, d.id AS did, d.name AS dname, r.name, r.description, r.status FROM rules r JOIN services s ON r.source=s.id JOIN services d ON r.destination=d.id WHERE r.id = ?;", id)
	//rul := Rule{}
	err = row.Scan(&sdata.Rule.Id, &sdata.Rule.Sid, &sdata.Rule.Source, &sdata.Rule.Did, &sdata.Rule.Destination, &sdata.Rule.Name, &sdata.Rule.Description, &sdata.Rule.Status)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(404), http.StatusNotFound)
	} else {
		tmpl, _ := template.ParseFiles("templates/rulesedit.html")
		// tmpl.Execute(w, rul)
		e := tmpl.Execute(w, sdata)
		if e != nil {
			log.Println(e)
		}
	}
}

// --------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// функция изменения данных и запись в лог - Rules
func RulesEditHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}
	id := r.FormValue("id")
	sid := r.FormValue("sid")
	did := r.FormValue("did")
	name := r.FormValue("name")
	description := r.FormValue("description")
	status := r.FormValue("status")

	//выборка  имени Services ----------------------------------------------------------------------------
	var sname string
	var dname string
	srow := database.QueryRow("select name from services where id = ?", sid)
	err = srow.Scan(&sname)
	if err != nil {
		log.Println(err)
	}
	drow := database.QueryRow("select name from services where id = ?", did)
	err = drow.Scan(&dname)
	if err != nil {
		log.Println(err)
	}

	// транзакция
	ok := true
	database.Exec("begin")
	_, err = database.Exec("update acl.rules set source=?, destination=?, name=?, description=?, status = ? where id = ?", sid, did, name, description, status, id)
	database.Exec("insert into acl.events (type, event) values ('Update Rule:', ?)", "{ \"Source\": \""+sname+"\", \"Destination\": \""+dname+"\", \"Name\": \""+name+"\", \"Description\": \""+description+"\", \"Status\": \""+status+"\" }")
	if err != nil {
		ok = false
	}
	if ok {
		database.Exec("commit")
	} else {
		database.Exec("rollback")
	}
	if err != nil {
		log.Println(err)
		tmpl, _ := template.ParseFiles("templates/dberr.html")
		tmpl.Execute(w, err)
		return
	}

	http.Redirect(w, r, "/rules", http.StatusMovedPermanently)
}

//--------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// функция возвращает пользователю страницу для добавления объекта Rule предварительно получив список services -  Rule
func RulesCreatePage(w http.ResponseWriter, r *http.Request) {
	rows, err := database.Query("SELECT id as sid, name as sname, id as did, name as dname FROM services ORDER BY id;")
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(404), http.StatusNotFound)
	} else {
		defer rows.Close()
		rules := []Rule{}

		for rows.Next() {
			p := Rule{}
			err := rows.Scan(&p.Sid, &p.Source, &p.Did, &p.Destination)
			if err != nil {
				fmt.Println(err)
				continue
			}
			rules = append(rules, p)
		}

		tmpl, _ := template.ParseFiles("templates/rulesadd.html")
		tmpl.Execute(w, rules)
	}
}

// функция добавления данных и запись в лог - Rule -------------------------------------------------------
func RulesCreateHandler(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)
	//service_id := vars["Sid"]
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}
	sid := r.FormValue("sid")
	did := r.FormValue("did")
	name := r.FormValue("name")
	description := r.FormValue("description")
	status := r.FormValue("status")

	//выборка  имени Services ----------------------------------------------------------------------------
	var sname string
	var dname string
	srow := database.QueryRow("select name from services where id = ?", sid)
	err = srow.Scan(&sname)
	if err != nil {
		log.Println(err)
	}
	drow := database.QueryRow("select name from services where id = ?", did)
	err = drow.Scan(&dname)
	if err != nil {
		log.Println(err)
	}

	// транзакция
	ok := true
	database.Exec("begin")
	_, err = database.Exec("insert into acl.rules (source, destination, name, description, status) values (?, ?, ?, ?, ?)", sid, did, name, description, status)
	database.Exec("insert into acl.events (type, event) values ('Add Rule:', ?)", "{ \"Source\": \""+sname+"\", \"Destination\": \""+dname+"\", \"Name\": \""+name+"\", \"Description\": \""+description+"\", \"Status\": \""+status+"\" }")
	if err != nil {
		ok = false
	}
	if ok {
		database.Exec("commit")
	} else {
		database.Exec("rollback")
	}
	if err != nil {
		log.Println(err)
		tmpl, _ := template.ParseFiles("templates/dberr.html")
		tmpl.Execute(w, err)
		return
	}

	http.Redirect(w, r, "/rules", http.StatusMovedPermanently)
}

//---------------------------------------------------------------------------------------------------------------------------------------------------------------------------
//  функция вывода данных - Rule
func RulesIndexHandler(w http.ResponseWriter, r *http.Request) {

	rows, err := database.Query("SELECT r.id, s.id AS sid, s.name AS sname, d.id AS did, d.name AS dname, r.name, r.description, r.status, r.created, r.updated FROM rules r JOIN services s ON r.source=s.id JOIN services d ON r.destination=d.id ORDER BY id DESC;")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	rules := []Rule{}

	for rows.Next() {
		p := Rule{}
		err := rows.Scan(&p.Id, &p.Sid, &p.Source, &p.Did, &p.Destination, &p.Name, &p.Description, &p.Status, &p.Created, &p.Updated)
		if err != nil {
			fmt.Println(err)
			continue
		}
		rules = append(rules, p)
	}

	tmpl, _ := template.ParseFiles("templates/rules.html")
	tmpl.Execute(w, rules)
}

// ====================================================================== Add ==============================================================================================
//--------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// функция добавления данных и запись в лог - ACLAdd
func AclAddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}
		sname := r.FormValue("sname")
		sdescription := r.FormValue("sdescription")
		sstatus := r.FormValue("sstatus")
		dname := r.FormValue("dname")
		ddescription := r.FormValue("ddescription")
		dstatus := r.FormValue("dstatus")
		siname := r.FormValue("siname")
		sidescription := r.FormValue("sidescription")
		siprotocol := r.FormValue("siprotocol")
		sicidr := r.FormValue("sicidr")
		siport := r.FormValue("siport")
		sistatus := r.FormValue("sistatus")
		diname := r.FormValue("diname")
		didescription := r.FormValue("didescription")
		diprotocol := r.FormValue("diprotocol")
		dicidr := r.FormValue("dicidr")
		diport := r.FormValue("diport")
		distatus := r.FormValue("distatus")
		rname := r.FormValue("rname")
		rdescription := r.FormValue("rdescription")
		rstatus := r.FormValue("rstatus")

		// транзакция
		ok := true
		database.Exec("begin")
		_, err =
			//  add source and destination service
			database.Exec("insert into acl.services (name, description, status) values (?, ?, ?)", sname, sdescription, sstatus)
		database.Exec("insert into acl.events (type, event) values ('Add Service:', ?)", "{ \"Name\": \""+sname+"\", \"Description\": \""+sdescription+"\", \"Status\": \""+sstatus+"\" }")
		database.Exec("insert into acl.services (name, description, status) values (?, ?, ?)", dname, ddescription, dstatus)
		database.Exec("insert into acl.events (type, event) values ('Add Service:', ?)", "{ \"Name\": \""+dname+"\", \"Description\": \""+ddescription+"\", \"Status\": \""+dstatus+"\" }")
		// add source and destination item
		database.Exec("insert into acl.items set service_id = (select id from services where name = '" + sname + "'), name = '" + siname + "', description = '" + sidescription + "', protocol = '" + siprotocol + "', cidr = '" + sicidr + "', port = '" + siport + "', status = '" + sistatus + "'")
		database.Exec("insert into acl.events (type, event) values ('Add Item:', ?)", "{ \"Service\": \""+sname+"\", \"Name\": \""+siname+"\", \"Description\": \""+sidescription+"\", \"Protocol\": \""+siprotocol+"\", \"Cidr\": \""+sicidr+"\", \"Port\": \""+siport+"\", \"Status\": \""+sistatus+"\" }")
		database.Exec("insert into acl.items set service_id = (select id from services where name = '" + dname + "'), name = '" + diname + "', description = '" + didescription + "', protocol = '" + diprotocol + "', cidr = '" + dicidr + "', port = '" + diport + "', status = '" + distatus + "'")
		database.Exec("insert into acl.events (type, event) values ('Add Item:', ?)", "{ \"Service\": \""+dname+"\", \"Name\": \""+diname+"\", \"Description\": \""+didescription+"\", \"Protocol\": \""+diprotocol+"\", \"Cidr\": \""+dicidr+"\", \"Port\": \""+diport+"\", \"Status\": \""+distatus+"\" }")
		// add rule
		database.Exec("insert into acl.rules set source = (select id from services where name = '" + sname + "'), destination =  (select id from services where name = '" + dname + "'), name = '" + rname + "', description = '" + rdescription + "', status = '" + rstatus + "'")
		database.Exec("insert into acl.events (type, event) values ('Add Rule:', ?)", "{ \"Source\": \""+sname+"\", \"Destination\": \""+dname+"\", \"Name\": \""+rname+"\", \"Description\": \""+rdescription+"\", \"Status\": \""+rstatus+"\" }")

		if err != nil {
			ok = false
		}
		if ok {
			database.Exec("commit")
		} else {
			database.Exec("rollback")
		}
		if err != nil {
			log.Println(err)
			tmpl, _ := template.ParseFiles("templates/dberr.html") // вывод ошибки бд
			tmpl.Execute(w, err)
			return
		}

		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	} else {
		http.ServeFile(w, r, "templates/acladd.html")
	}
}

// ====================================================================== Acl ==============================================================================================
//--------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// функция вывода и поиска данных - AclSearch
func AclSearchPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, ok := vars["id"]
	if !ok {
		id = r.URL.Query().Get("id")
	}
	like := "%" + id + "%"
	//like := "%" + vars["id"] + "%"

	protocol := like
	source := like
	destination := like
	port := like
	rname := like
	sitem := like
	ditem := like
	sservice := like
	dservice := like

	rows, err := database.Query("SELECT d.protocol, s.cidr AS source, d.cidr AS destination, d.port, r.id AS rid, r.name AS rname, s.id AS siid, s.name AS sitem, d.id AS diid, d.name AS ditem, sr.id AS ssid, sr.name AS sservice, dr.id AS dsid, dr.name AS dservice FROM rules r JOIN items s ON r.source=s.service_id JOIN items d ON r.destination=d.service_id JOIN services sr ON r.source=sr.id JOIN services dr ON r.destination=dr.id WHERE d.protocol LIKE ? OR s.cidr LIKE ? OR d.cidr LIKE ? OR d.port LIKE ? OR r.name LIKE ? OR s.name LIKE ? OR d.name LIKE ? OR sr.name LIKE ? OR dr.name LIKE ? ORDER BY r.id DESC LIMIT 1000;", protocol, source, destination, port, rname, sitem, ditem, sservice, dservice)
	if err != nil {
		log.Println(err)
	}

	defer rows.Close()
	acl := []AclSrch{}

	for rows.Next() {
		p := AclSrch{}
		err := rows.Scan(&p.Protocol, &p.Source, &p.Destination, &p.Port, &p.Rid, &p.Rname, &p.Siid, &p.Sitem, &p.Diid, &p.Ditem, &p.Ssid, &p.Sservice, &p.Dsid, &p.Dservice)
		if err != nil {
			fmt.Println(err)
			continue
		}
		acl = append(acl, p)
	}

	tmpl, _ := template.ParseFiles("templates/aclsearch.html")
	tmpl.Execute(w, acl)
}

// ====================================================================== Events ============================================================================================
//---------------------------------------------------------------------------------------------------------------------------------------------------------------------------
//  функция  вывода и поиска данных - Event
func EventsPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, ok := vars["id"]
	if !ok {
		id = r.URL.Query().Get("id")
	}
	like := "%" + id + "%"
	//like := "%" + vars["id"] + "%"

	etype := like
	event := like

	rows, err := database.Query("SELECT * FROM events WHERE type LIKE ? OR event LIKE ? ORDER BY id DESC LIMIT 1000;", etype, event)
	if err != nil {
		log.Println(err)
	}

	defer rows.Close()
	evn := []Event{}

	for rows.Next() {
		p := Event{}
		err := rows.Scan(&p.Id, &p.Type, &p.Event, &p.Created)
		if err != nil {
			fmt.Println(err)
			continue
		}
		evn = append(evn, p)
	}

	tmpl, _ := template.ParseFiles("templates/events.html")
	tmpl.Execute(w, evn)
}

// ====================================================================== Apply =============================================================================================
//---------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// функция обновления rule и запись в лог - Apply
func ApplyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}
		text := r.FormValue("text")

		_, err = database.Exec("insert into acl.events (type, event) values ('Apply acl:', ?)", "{ \"Apply\": \""+text+"\" }")
		if err != nil {
			log.Println(err)
			tmpl, _ := template.ParseFiles("templates/dberr.html")
			tmpl.Execute(w, err)
			return
		}

		// создание файла
		file, err := os.Create("/opt/last_acl_update")
		if err != nil {
			fmt.Println("Unable to create file:", err)
			//os.Exit(1)
		}
		defer file.Close()
		file.WriteString(text)

		// выполнение команды cmd
		// cmd := exec.Command(`/bin/echo`, `>`, `"1"`, `/opt/last_acl_update`)
		// stdoutStderr, err := cmd.CombinedOutput()
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// fmt.Printf("%s\n", stdoutStderr)

		http.Redirect(w, r, "/aclsearch", http.StatusMovedPermanently)
	} else {
		http.ServeFile(w, r, "templates/apply.html")
	}

}

// ====================================================================== Txt ===============================================================================================
//---------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// функция выводит текстовый файл
func TxtIndexHandler(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open("/opt/rului.conf")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer f.Close()

	_, err = io.WriteString(w, `<!DOCTYPE html><html><body>`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s := bufio.NewScanner(f)

	for s.Scan() {
		_, err = io.WriteString(w, html.EscapeString(s.Text())+`<br/>`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err = s.Err(); err != nil {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	_, err = io.WriteString(w, `</body></html>`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ====================================================================== Main ==============================================================================================
//---------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// функция  main
func main() {

	db, err := sql.Open("mysql", "acluser:aclpass@/acl")
	if err != nil {
		log.Println(err)
	}
	database = db
	defer db.Close()

	router := mux.NewRouter()
	router.HandleFunc("/", IndexHandler)

	router.HandleFunc("/services", ServicesIndexHandler)
	router.HandleFunc("/servicesadd", ServicesCreateHandler)
	router.HandleFunc("/servicesedit/{id:[0-9]+}", ServicesEditPage).Methods("GET")
	router.HandleFunc("/servicesedit/{id:[0-9]+}", ServicesEditHandler).Methods("POST")
	router.HandleFunc("/servicesdelete/{id:[0-9]+}", ServicesDeleteHandler)

	router.HandleFunc("/items", ItemsIndexHandler)
	router.HandleFunc("/itemsadd", ItemsCreatePage).Methods("GET")
	router.HandleFunc("/itemsadd", ItemsCreateHandler).Methods("POST")
	router.HandleFunc("/itemsedit/{id:[0-9]+}", ItemsEditPage).Methods("GET")
	router.HandleFunc("/itemsedit/{id:[0-9]+}", ItemsEditHandler).Methods("POST")
	router.HandleFunc("/itemsdelete/{id:[0-9]+}", ItemsDeleteHandler)

	router.HandleFunc("/rules", RulesIndexHandler)
	router.HandleFunc("/rulesadd", RulesCreatePage).Methods("GET")
	router.HandleFunc("/rulesadd", RulesCreateHandler).Methods("POST")
	router.HandleFunc("/rulesedit/{id:[0-9]+}", RulesEditPage).Methods("GET")
	router.HandleFunc("/rulesedit/{id:[0-9]+}", RulesEditHandler).Methods("POST")
	router.HandleFunc("/rulesdelete/{id:[0-9]+}", RulesDeleteHandler)

	router.HandleFunc("/acladd", AclAddHandler)

	router.HandleFunc("/aclsearch/{id}", AclSearchPage).Methods("GET")
	router.HandleFunc("/aclsearch/", AclSearchPage).Methods("GET")
	router.HandleFunc("/aclsearch", AclSearchPage).Methods("GET")

	router.HandleFunc("/events/{id}", EventsPage).Methods("GET")
	router.HandleFunc("/events/", EventsPage).Methods("GET")
	router.HandleFunc("/events", EventsPage).Methods("GET")

	router.HandleFunc("/apply", ApplyHandler)

	router.HandleFunc("/txt", TxtIndexHandler)

	router.PathPrefix("/templates/").Handler(http.StripPrefix("/templates/", http.FileServer(http.Dir("./templates/"))))
	//http.Handle("/templates/", http.StripPrefix("/templates/", http.FileServer(http.Dir("./templates"))))

	http.Handle("/", router)

	//fmt.Println("Server is listening...")
	//http.ListenAndServe(":8181", nil)
	cmd := flag.String("b", "localhost:8181", "for ex: ./acl -b 127.0.0.1:80") // for ex: ./acl -b 127.0.0.1:80
	flag.Parse()
	fmt.Printf("Server is listening on \"%v\"\n", string(*cmd))
	http.ListenAndServe(*cmd, nil)
}
