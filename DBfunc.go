package main

import (
	"log"
	"os"
	"database/sql"
	"io"
	"mime/multipart"
	"time"
	"strings"
	_ "github.com/max75025/go-sqlite3"
	"errors"
)


const dbFilePath  = "./database/db.db"

const dbDirPath = "./database/"
const editionDirPath  = "./editions/"
const publicationDirPath = "./publications/"
/*const contentListDirPath = "./contentLists/"*/

 func refreshCache()error{
 	var err error = nil
	 cacheData.authors, err = getAuthorsCache()
	 if err!=nil{return err}
	 cacheData.publications,err = getPublicationsCache()
	 cacheData.edition,err = getEditionsCache()
	 return err
 }


func upload(file multipart.File, filename string,  path string ) (string, error) {

	defer file.Close()
	filename = strings.Replace(filename, ".", "_"+string(time.Now().Unix()) +".",-1)

	f, err := os.OpenFile( "." + path+filename, os.O_WRONLY|os.O_CREATE, 0666)
	defer f.Close()
	if err != nil {
		log.Println(err)
		return "",err
	}
	io.Copy(f, file)
	return path+filename, nil
}

func checkFolders()error{
	err := os.MkdirAll(editionDirPath, os.ModePerm)
	if err != nil {return err}
	err = os.MkdirAll(publicationDirPath, os.ModePerm)
	if err != nil {return err}
	/*err = os.MkdirAll(contentListDirPath, os.ModePerm)
	if err != nil {return err}*/
	err = os.MkdirAll(dbDirPath, os.ModePerm)
	if err != nil {return err}
	return nil
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return true, err
}

func openDB()(*sql.DB, error){
	ex, err:= exists(dbFilePath)
	if err!=nil{
		log.Println(err)
		return nil,err
	}
	if !ex{
		_, err = os.OpenFile(dbFilePath, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Println(err)
			return nil,err
		}
	}

	db, err:= sql.Open("sqlite3", dbFilePath)
	if err!= nil{
		return nil, err
	}
	if !ex{
		/*statement, err := db.Prepare("CREATE TABLE IF NOT EXISTS author (ID PRIMARY KEY UNIQUE NOT NULL ,fio_ru TEXT  NOT NULL , fio_ua TEXT  NOT NULL, fio_en TEXT NOT NULL, email TEXT, phone TEXT, orsid_url TEXT, work TEXT  );")
		if err!=nil{
			return nil, err
		}
		statement.Exec()*/
		_, err := db.Exec("CREATE TABLE IF NOT EXISTS author (ID INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE  ,fio_ru TEXT  NOT NULL , fio_ua TEXT  NOT NULL, fio_en TEXT NOT NULL, email TEXT NOT NULL, phone TEXT NOT NULL, orcid_url TEXT NOT NULL, work TEXT NOT NULL  );")
		if err!=nil{return nil, err}

		_, err = db.Exec("CREATE TABLE IF NOT EXISTS publication (ID INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE ,name TEXT  NOT NULL ,  pages TEXT, doi_url TEXT NOT NULL, page_count INTEGER NOT NULL , page_print_count INTEGER  NOT NULL, pdf_url TEXT NOT NULL, id_edition INTEGER NOT NULL);")
		if err!=nil{return nil, err}

		_, err = db.Exec("CREATE TABLE IF NOT EXISTS edition (ID INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE  , number INTEGER  NOT NULL , tom INTEGER  NOT NULL, year INTEGER NOT NULL,  pdf_url TEXT NOT NULL );")
		if err!=nil{return nil, err}

		_,err = db.Exec("CREATE TABLE IF NOT EXISTS author_publication (ID INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE , id_author INTEGER NOT NULL, id_publication INTEGER NOT NULL);")
		if err!=nil{return nil, err}

	}
	return db, nil
}

func getPublicationsCache()(map[int]publicationCache, error){
	result:= make(map[int]publicationCache)
	var idPublication, idAuthor int
	db,err:= openDB()
	if err!=nil{return nil, err}
	rows, err := db.Query("SELECT * FROM publication")
	if err!=nil{return nil,err}
	for rows.Next(){
		p:= publicationCache{}
		rows.Scan(&idPublication,&p.name, &p.pages, &p.doi, &p.pageCount, &p.pagePrintCont, &p.pdf, &p.idEdition)
		p.idAuthors = p.idAuthors[:0]
		row, err := db.Query("SELECT id_author FROM author_publication WHERE id_publication = ?", idPublication)
		if err!=nil{return nil, err}
		for row.Next(){
			row.Scan(&idAuthor)
			p.idAuthors = append(p.idAuthors,idAuthor )
		}

		result[idPublication] = p
	}
	return result, nil
}


func getEditionsCache()(map[int]editionCache, error){
	result:= make(map[int]editionCache)
	var id int
	e:= editionCache{}
	db,err:= openDB()
	if err!=nil{return nil, err}
	rows, err := db.Query("SELECT * FROM edition")
	if err!=nil{return nil,err}
	for rows.Next(){
		rows.Scan(&id,&e.number, &e.tom, &e.year, &e.pdf)
		result[id] = e
	}
	return result, nil
}


func getAuthorsCache()(map[int]authorCache, error){
	result:= make(map[int]authorCache)
	var id int
	a:= authorCache{}
	db,err:= openDB()
	if err!=nil{return nil, err}
	rows, err := db.Query("SELECT * FROM author")
	if err!=nil{return nil,err}
	for rows.Next(){
		rows.Scan(&id,&a.fioRu, &a.fioUa, &a.fioEn, &a.email, &a.phone, &a.orcid, &a.work)
		result[id] = a
	}
	return result, nil
}


func getSortAuthors()([]AuthorModel, error){
	var result []AuthorModel
	a:= AuthorModel{}
	db,err:= openDB()
	if err!=nil{return nil, err}
	rows, err := db.Query("SELECT * FROM author ORDER BY fio_ru")
	if err!=nil{return nil,err}
	for rows.Next(){
		rows.Scan(&a.ID,&a.FioRu, &a.FioUa, &a.FioEn, &a.Email, &a.Phone, &a.ORCID, &a.Work)
		result = append(result, a)
	}
	return result, nil
}



func getSortEditions()([]EditionModel, error){
	var result []EditionModel
	e:= EditionModel{}
	db,err:= openDB()
	if err!=nil{return nil, err}
	rows, err := db.Query("SELECT * FROM edition ORDER BY tom, number")
	if err!=nil{return nil,err}
	for rows.Next(){
		rows.Scan(&e.ID,&e.Number, &e.Tom, &e.Year, &e.PDF)
		result = append(result, e)
	}
	return result, nil
}


func getPublicationByIdEdition(id int)([]PublicationModel, error){
	var result []PublicationModel
	var idPublication, idAuthor, idEdition int
	p:= PublicationModel{}
	db,err:= openDB()
	if err!=nil{return nil, err}
	rows, err := db.Query("SELECT * FROM publication WHERE id_edition =?", id)
	if err!=nil{return nil,err}
	for rows.Next(){
		rows.Scan(&idPublication,&p.Name, &p.Pages, &p.DOI, &p.PageCount, &p.PagePrintCount, &p.PDF, &idEdition)
		row, err := db.Query("SELECT id_author FROM author_publication WHERE id_publication = ?", idPublication)
		if err!=nil{return nil, err}
		for row.Next(){
			row.Scan(&idAuthor)
			p.Authors = append(p.Authors, AuthorModel{
				ID:    idAuthor,
				FioRu: cacheData.authors[idAuthor].fioRu,
				FioUa: cacheData.authors[idAuthor].fioUa,
				FioEn: cacheData.authors[idAuthor].fioEn,
				Email: cacheData.authors[idAuthor].email,
				Phone: cacheData.authors[idAuthor].phone,
				ORCID: cacheData.authors[idAuthor].orcid,
				Work:  cacheData.authors[idAuthor].work,
			})
		}

		result = append(result, p)
	}
	return result, nil
}

func getPublicationByIdAuthor(id int)([]PublicationModel, error){
	var result []PublicationModel
	var idPublication int
	p:= PublicationModel{}
	db,err:= openDB()
	if err!=nil{return nil, err}
	row, err := db.Query("SELECT id_pulication FROM author_publication WHERE id_author = ?", id)
	if err!=nil{return nil, err}
		for row.Next(){
			row.Scan(&idPublication)
			p = PublicationModel{
				ID:            idPublication,
				Name:          cacheData.publications[idPublication].name,
				Pages:         cacheData.publications[idPublication].pages,
				DOI:           cacheData.publications[idPublication].doi,
				PageCount:     cacheData.publications[idPublication].pageCount,
				PagePrintCount: cacheData.publications[idPublication].pagePrintCont,
				PDF:           cacheData.publications[idPublication].pdf,
				Edition:       EditionModel{
					ID:           cacheData.publications[idPublication].idEdition,
					Tom:          cacheData.edition[ cacheData.publications[idPublication].idEdition].tom,
					Number:       cacheData.edition[ cacheData.publications[idPublication].idEdition].number,
					Year:         cacheData.edition[ cacheData.publications[idPublication].idEdition].year,
					PDF:          cacheData.edition[ cacheData.publications[idPublication].idEdition].pdf,
					Publications: nil,
				},
			}
		result = append(result, p)
	}
	return result, nil
}

func deleteAuthor(id int)(error){
	db,err:=openDB()
	if err!=nil{return err}
	_,err = db.Exec("DELETE FROM author WHERE ID=?", id)
	return err
}

func deletePublication(id int)(error){
	db,err:=openDB()
	if err!=nil{return err}
	_,err = db.Exec("DELETE FROM publication WHERE ID=?", id)
	return err
}

func deleteEditon(id int)(error){
	db,err:=openDB()
	if err!=nil{return err}
	_,err = db.Exec("DELETE FROM edition WHERE ID=?", id)
	return err
}

func deleteAuthorPublication(idAuthor int, idPublication int)(error){
	db,err:=openDB()
	if err!=nil{return err}
	if idAuthor<0 && idPublication<0{return errors.New("both argument <0")}
	if(idAuthor>0) {
		_, err = db.Exec("DELETE FROM author_publication WHERE id_author=?", idAuthor)
	}else{
		_, err = db.Exec("DELETE FROM author_publication WHERE id_publication=?", idPublication)
	}
	return err
}

func addAuthor(fioRu string, fioUa string, fioEn string, email string, phone string, orcid string, work string)error{
	db,err:=openDB()
	if err!=nil{return err}
	stmt,err := db.Prepare("INSERT INTO author(fio_ru, fio_ua, fio_en, email, phone, orcid_url, work) VALUES(?,?,?,?,?,?,?)")
	if err!=nil{return err}
	_,err = stmt.Exec(fioRu,fioUa, fioEn, email, phone, orcid, work)
	return err

}

func addPublication(name string, pages string, doi string, pageCount int, pagePrintCount int, pdfUrl string, idEdition int, )(int, error){
	db,err:=openDB()
	if err!=nil{return 0,err}
	stmt,err := db.Prepare("INSERT INTO publication(name, pages, doi_url, page_count, page_print_count, pdf_url, id_edition) VALUES(?,?,?,?,?,?,?)")
	if err!=nil{return 0,err}
	r,err := stmt.Exec(name, pages, doi, pageCount, pagePrintCount, pdfUrl, idEdition)
	if err!=nil{return  0,err}
	id,err:= r.LastInsertId()//int64

	return int(id), err

}

func addAuthorPublication(idAuthors []int, idPublication int)error{
	db,err:=openDB()
	if err!=nil{return err}
	for _,i:= range idAuthors {
		stmt, err := db.Prepare("INSERT INTO author_publication(id_author, id_publication) VALUES(?,?)")
		if err != nil {return err}
		_, err = stmt.Exec(i, idPublication)
		if err!=nil{return err}
	}
	return nil
}

func addEdition(tom int, number int, year int, pdf string)error{
	db,err:=openDB()
	if err!=nil{return err}
	stmt,err := db.Prepare("INSERT INTO edition(tom, number, year, pdf_url) VALUES(?,?,?,?)")
	if err!=nil{return err}
	_,err = stmt.Exec(tom, number, year, pdf)
	return err
}

func updatePublication(id int, name string, pages string, doi string, pageCount int, pagePrintCount int, pdf string, idEdition int ) error {
	var err error
	db,err:= openDB()
	if pdf == ""{
		q, err := db.Prepare("UPDATE  publication SET name=?, pages = ?,doi_url = ?, page_count = ?, page_print_count = ?, id_edition = ?   WHERE ID=?  " )
		if err!=nil{ return  err}
		_,err = q.Exec(name, pages, doi,pageCount, pagePrintCount, idEdition, id)
	}else{
		q, err := db.Prepare("UPDATE  publication SET name=?, pages = ?,doi_url = ?, pageCount = ?, pagePrintCount = ?, idEdition = ?, pdf_url = ?   WHERE ID=?  " )
		if err!=nil{ return  err}
		_,err = q.Exec(name, pages, doi,pageCount, pagePrintCount, idEdition, pdf, id)

	}
	return err
}

func updateEdition(id int,tom int, number int, year int, pdf string ) error {
	var err error
	db,err:= openDB()
	if pdf == ""{
		q, err := db.Prepare("UPDATE  edition SET tom = ?, number = ?, year = ?   WHERE ID=?  " )
		if err!=nil{ return  err}
		_,err = q.Exec(tom, number, year, id)
	}else{
		q, err := db.Prepare("UPDATE  edition SET tom = ?, number = ?, year = ?, pdf_url = ?   WHERE ID=?  " )
		if err!=nil{ return  err}
		_,err = q.Exec(tom, number, year,pdf, id)

	}
	return err
}


func updateAuthor(id int,fioRu string, fioUa string, fioEn string, email string, phone string, orcid string, work string) error {
	var err error
	db,err:= openDB()

		q, err := db.Prepare("UPDATE  author SET fio_ru = ?, fio_ua = ?, fio_en = ?, email = ?, phone = ?, orcid_url = ?, work = ?   WHERE ID=?  " )
		if err!=nil{ return  err}
		_,err = q.Exec(fioRu,fioUa,fioEn, email, phone, orcid, work, id)

	return err
}


