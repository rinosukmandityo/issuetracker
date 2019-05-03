package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"
)

const configFile = "issues.json"

var db Db

type Change struct {
	ModifiedOn time.Time
	ModifiedBy string
	Status     int
	Comment    string
}

type Issue struct {
	Id      int
	Subject string
	Changes []Change
	Last    *Change
}

type Category struct {
	Name   string
	Id     int
	Issues []Issue
}

type User struct {
	Name    string
	Address string
}

type Db struct {
	NextId     int
	Users      []User
	Categories []Category
}

func (b *Issue) setLast() {
	size := len(b.Changes)
	if size > 0 {
		b.Last = &(b.Changes[size-1])
	}
}

func (p *Category) setLast() {
	for i := 0; i < len(p.Issues); i++ {
		b := &(p.Issues[i])
		b.setLast()
	}
}

func (db *Db) setLast() {
	for _, p := range db.Categories {
		p.setLast()
	}
}

func (db *Db) load() {
	f, err := os.OpenFile(configFile, os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
	defer f.Close()

	b, _ := ioutil.ReadAll(f)
	json.Unmarshal(b, db)
	db.setLast()
}

func (db *Db) save() {
	const tempFile = "issues.tmp"
	f, err := os.Create(tempFile)
	if err != nil {
		panic(err)
	}

	// fmt.Fprintf(f, xml.Header)
	b, err := json.MarshalIndent(db, "", "\t")
	if err != nil {
		panic(err)
	}
	f.Write(b)
	f.Close()
	oldFile := configFile + ".old"
	os.Remove(oldFile)
	os.Rename(configFile, oldFile)
	os.Rename(tempFile, configFile)
}

func findIssue(id int) *Issue {
	for _, p := range db.Categories {
		for i, b := range p.Issues {
			if b.Id == id {
				return &p.Issues[i]
			}
		}
	}
	return nil
}

func deleteIssue(id int) {
	for iC, p := range db.Categories {
		for i, b := range p.Issues {
			if b.Id == id {
				db.Categories[iC].Issues = append(db.Categories[iC].Issues[:i], db.Categories[iC].Issues[i+1:]...)
			}
		}
	}
}

func findCategory(id int) *Category {
	for i, p := range db.Categories {
		if p.Id == id {
			return &db.Categories[i]
		}
	}
	return nil
}

func loadTemplate(filename string) *template.Template {
	f, err := os.Open("tmpl/" + filename + ".template")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	b, _ := ioutil.ReadAll(f)
	tmpl, err := template.New(filename).Parse(string(b))
	if err != nil {
		panic(err)
	}
	return tmpl
}

func handlerRoot(w http.ResponseWriter, r *http.Request) {
	tmpl := loadTemplate("root")
	err := tmpl.Execute(w, db)
	if err != nil {
		panic(err)
	}
}

func handlerIssue(w http.ResponseWriter, r *http.Request) {
	s := r.FormValue("id")
	if s == "" {
		return
	}
	id, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}

	b := findIssue(id)
	if b == nil {
		return
	}
	type T struct {
		Users []User
		Bug   *Issue
	}
	var t T
	t.Bug = b
	t.Users = db.Users

	tmpl := loadTemplate("bug")
	err = tmpl.Execute(w, t)
	if err != nil {
		panic(err)
	}
}

func handlerChange1(w http.ResponseWriter, r *http.Request, id int) {
	b := findIssue(id)
	if b == nil {
		return
	}
	var c Change
	c.ModifiedOn = time.Now().UTC()
	subject := r.FormValue("subject")
	if subject == "" {
		return
	}
	b.Subject = subject
	var err error
	c.Status, err = strconv.Atoi(r.FormValue("status"))
	if err != nil {
		panic(err)
	}

	c.ModifiedBy = r.FormValue("who")
	c.Comment = r.FormValue("comment")
	b.Changes = append(b.Changes, c)
	b.setLast()

	db.save()
}

func handlerChange(w http.ResponseWriter, r *http.Request) {
	sId := r.FormValue("id")
	if sId == "" {
		return
	}
	id, err := strconv.Atoi(sId)
	if err != nil {
		panic(err)
	}
	handlerChange1(w, r, id)
	http.Redirect(w, r, "/issue?id="+sId, http.StatusSeeOther)
}

func handlerDelete(w http.ResponseWriter, r *http.Request) {
	sId := r.FormValue("id")
	if sId == "" {
		return
	}
	id, err := strconv.Atoi(sId)
	if err != nil {
		panic(err)
	}
	deleteIssue(id)
	db.save()
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func handlerNewIssue(w http.ResponseWriter, r *http.Request) {
	sId := r.FormValue("partId")
	if sId == "" {
		return
	}
	id, err := strconv.Atoi(sId)
	if err != nil {
		panic(err)
	}
	p := findCategory(id)

	var b Issue
	b.Subject = r.FormValue("subject")
	b.Id = db.NextId
	db.NextId++
	p.Issues = append(p.Issues, b)

	db.setLast()
	http.Redirect(w, r, "/issue?id="+strconv.Itoa(b.Id), http.StatusSeeOther)
}

func main() {
	db.load()
	http.HandleFunc("/", handlerRoot)
	http.HandleFunc("/issue", handlerIssue)
	http.HandleFunc("/delete", handlerDelete)
	http.HandleFunc("/change", handlerChange)
	http.HandleFunc("/newIssue", handlerNewIssue)
	http.Handle("/static/", http.StripPrefix("/", http.FileServer(http.Dir("."))))
	port := ":9000"
	fmt.Println(fmt.Sprintf("Server running on localhost%s", port))
	http.ListenAndServe(port, nil)
}
