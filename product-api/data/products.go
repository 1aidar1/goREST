package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"time"

	"github.com/go-playground/validator"
	"github.com/jmoiron/sqlx"
)

type Books []*Book
type Book struct {
	ID         int            `db:"id" json:"id"`
	UpdatedAt  sql.NullString `db:"updated_at" json:"-"`
	DeletedAt  sql.NullString `db:"deleted_at" json:"-"`
	CreatedAt  sql.NullString `db:"created_at" json:"-"`
	Title      string         `db:"title" json:"title" validate:"required"`
	Author     string         `db:"author" json:"author" validate:"required"`
	CallNumber string         `db:"call_number" json:"call_number"`
	PersonId   string         `db:"person_id" json:"person_id"`
}

var db *sqlx.DB

var ErrBookNotFound = fmt.Errorf("Book not found")

// *******************************

func SetDB(newDB *sqlx.DB) {
	db = newDB
}

func (p *Book) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(p)
}

func (p *Book) Validate() error {
	validate := validator.New()
	validate.RegisterValidation("sku", validateSKU)
	return validate.Struct(p)
}

func validateSKU(fl validator.FieldLevel) bool {
	re := regexp.MustCompile(`[a-z]+-[a-z]+-[a-z]+`)
	return re.MatchString(fl.Field().String())
}

func (p Books) ToJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	return encoder.Encode(p)
}

// GetBooks returns a list of Books
func GetBooks() (Books, error) {
	bookers := Books{}
	err := db.Select(&bookers, "SELECT * FROM books where deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	return bookers, nil
}

func AddBook(p *Book) error {

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec("INSERT INTO books (created_at, title, author, call_number,person_id) VALUES ($1,$2,$3,$4,$5)", time.Now(), p.Title, p.Author, p.CallNumber, p.PersonId)
	if err != nil {
		return err
	}
	tx.Commit()
	return err
}

func UpdateBook(id int, p *Book) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec("UPDATE books SET updated_at=$1, title=$2,author=$3,call_number=$4,person_id=$5 WHERE id=$6", time.Now(), p.Title, p.Author, p.CallNumber, p.PersonId, id)
	if err != nil {
		return err
	}
	tx.Commit()

	return err
}

func DeleteBook(id int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	result, err := tx.Exec("UPDATE books SET deleted_at=$1 WHERE id=$2", time.Now().Local().UTC(), id)
	if err != nil {
		return err
	}
	tx.Commit()
	s, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if s == 0 {
		return ErrBookNotFound
	}

	return nil
}

// func findBook(id int) (*Book, int, error) {

// 	for i, p := range bookList {
// 		if p.ID == id {
// 			return p, i, nil
// 		}
// 	}

// 	return nil, -1, ErrBookNotFound
// }

// func getNextID() int {
// 	lp := bookList[len(bookList)-1]
// 	return lp.ID + 1
// }

// bookList is a hard coded list of Books for this
// example data source
// var bookList = []*Book{
// 	&Book{
// 		ID:         1,
// 		Title:      "name1",
// 		Author:     "Frothy milky coffee",
// 		CallNumber: "2.45",
// 		PersonId:   "abc323",
// 		CreatedAt:  time.Now().UTC().String(),
// 		UpdatedAt:  time.Now().UTC().String(),
// 		DeletedAt:  time.Now().UTC().String(),
// 	},
// }
