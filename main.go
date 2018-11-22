package main

import (
	"github.com/max75025/httprouter"
	"github.com/max75025/open-golang/open"
	"net/http"
	"fmt"
	"html/template"
	"time"
	"log"
	"strconv"
	"os"

)

var cacheData cache

func init() {
	checkFolders()
	openDB()
	var err error
	cacheData.authors,err = getAuthorsCache()
	if err!=nil{
		panic(err)
	}
	cacheData.publications,err = getPublicationsCache()
	if err!=nil{
		panic(err)
	}
	cacheData.edition,err = getEditionsCache()
	if err!=nil{
		panic(err)
	}

}


func main(){
	open.Start("http://localhost:8080/login")

	router := httprouter.New()

	//Serve static
	router.ServeFiles("/static/*filepath", http.Dir("static"))
	router.ServeFiles("/newsImgs/*filepath", http.Dir("newsImgs"))
	router.ServeFiles("/photo/*filepath", http.Dir("photo"))
	router.ServeFiles("/editions/*filepath", http.Dir("editions"))
	router.ServeFiles("/publications/*filepath", http.Dir("publications"))


	router.GET("/login", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		t, _ := template.ParseFiles("tmpls/login.html")
		t.Execute(w, nil)
	})
	router.POST("/login", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		r.ParseForm()
		// logic part of log in
		/*fmt.Println("username:", r.Form["username"])
		fmt.Println("password:", r.Form["password"])*/
		fmt.Println("login success")
		//name := "admin"
		//password :="admin"
		/*if r.Form["username"][0] == name && r.Form["password"][0] == password {
			fmt.Fprintf(w, "success")
			//
		}*/
		name := r.FormValue("username")
		pass := r.FormValue("password")
		redirectTarget := "/errorLogin"
		if name != "" && pass != "" && checkLoginData(name, pass) {
			// .. check credentials ..
			setSession(name, w)
			redirectTarget = "/publications"
		}
		http.Redirect(w, r, redirectTarget, 302)
	} )

	router.GET("/logout", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		logoutHandler(w,r)
	})


	router.GET("/publications", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		checkLogin(w, r)
		t, _ := template.ParseFiles("tmpls/publications.html")
		var publications []PublicationModel
		for i,d:= range cacheData.publications{
			p:= PublicationModel{
				ID:             i,
				Name:           d.name,
				Pages:          d.pages,
				DOI:            d.doi,
				PageCount:      d.pageCount,
				PagePrintCount: d.pagePrintCont,
				PDF:            d.pdf,
				Authors:        nil,
				Edition:        EditionModel{
					ID:           d.idEdition,
					Tom:          cacheData.edition[d.idEdition].tom,
					Number:       cacheData.edition[d.idEdition].number,
					Year:         cacheData.edition[d.idEdition].year,
					PDF:          "",
					Publications: nil,
				},
			}
			for _,idAuthor:= range d.idAuthors{
				a:=AuthorModel{
					ID:    idAuthor,
					FioRu: cacheData.authors[idAuthor].fioRu,
					FioUa: "",
					FioEn: "",
					Email: "",
					Phone: "",
					ORCID: "",
					Work:  "",
				}
				p.Authors = append(p.Authors, a)
			}
			publications = append(publications,p)
		}


		//fmt.Println(publications)

		t.Execute(w, publications)
	})

	router.GET("/publication/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		checkLogin(w, r)

		id,err := strconv.Atoi(ps.ByName("id"))
		if err!=nil{
			fmt.Fprintf(w, "ошибка получения идентификатора")
			return
		}


		t, _ := template.ParseFiles("tmpls/publication.html")

		p:=PublicationModel{
			ID:             id,
			Name:           cacheData.publications[id].name,
			Pages:           cacheData.publications[id].pages,
			DOI:             cacheData.publications[id].doi,
			PageCount:       cacheData.publications[id].pageCount,
			PagePrintCount:  cacheData.publications[id].pagePrintCont,
			PDF:             cacheData.publications[id].pdf,
			Authors:        nil,
			Edition:        EditionModel{
				ID:            cacheData.publications[id].idEdition,
				Tom:          cacheData.edition[ cacheData.publications[id].idEdition].tom,
				Number:       cacheData.edition[ cacheData.publications[id].idEdition].number,
				Year:         cacheData.edition[ cacheData.publications[id].idEdition].year,
				PDF:          "",
				Publications: nil,
			},
		}
		for _,idAuthor:= range  cacheData.publications[id].idAuthors {
			a := AuthorModel{
				ID:    idAuthor,
				FioRu: cacheData.authors[idAuthor].fioRu,
				FioUa: "",
				FioEn: "",
				Email: "",
				Phone: "",
				ORCID: "",
				Work:  "",
			}
			p.Authors = append(p.Authors, a)
		}

		//fmt.Println(p)

		t.Execute(w,p)
	})

	router.GET("/addPublication", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		checkLogin(w, r)
		t, _ := template.ParseFiles("tmpls/addPublication.html")
		a,err:=getSortAuthors()
		if err!=nil{
			fmt.Fprintf(w, "ошибка получения данных")
			return
		}
		e,err:=getSortEditions()
		if err!=nil{
			fmt.Fprintf(w, "ошибка получения данных")
			return
		}
		t.Execute(w, AddPublicationModel{
			Authors:  a,
			Editions: e,
		})
	})

	router.POST("/addPublication", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		checkLogin(w, r)

		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("pdf")
		if err != nil {
			log.Println(err)
		}
		path,err := upload(file, handler.Filename, "/publications/")
		if err!=nil{
			fmt.Fprintf(w,"ошибка добавления файла")
			return
		}

		name:=r.FormValue("name")
		pages:=r.FormValue("pages")
		doi:=r.FormValue("doi")
		pc,err:=strconv.Atoi(r.FormValue("page_count"))
		if err!=nil{
			fmt.Fprintf(w, "ошибка преобразования данных")
			return
		}
		ppc,err:=strconv.Atoi(r.FormValue("page_print_count"))
		if err!=nil{
			fmt.Fprintf(w, "ошибка преобразования данных")
			return
		}

		edition,err:=strconv.Atoi(r.FormValue("edition"))
		if err!=nil{
			fmt.Fprintf(w, "ошибка преобразования данных")
			return
		}

		idPublication,err := addPublication(name,pages,doi,pc,ppc,path,edition)
		if err!=nil{
			fmt.Fprintf(w, "ошибка добавления")
			log.Println(err)
			return
		}
		r.ParseForm()
		authors:=r.Form["authors"]

		var is []int
		for _,d:=range authors{
			i,err:= strconv.Atoi(d)
			if err!=nil{
				log.Println(err)
				return
			}
			is = append(is, i)
		}

		err= addAuthorPublication(is,idPublication)
		if err!=nil{
			log.Println(err)
			fmt.Fprintf(w, "ошибка добавления связей")
			err = deletePublication(idPublication)
			if err!=nil{fmt.Fprintf(w, "ошибка удаления записи")}
			return
		}
		refreshCache()
		http.Redirect(w,r,"/publications",302)
		//fmt.Println(authors)

	})

	router.GET("/editPublication/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		checkLogin(w, r)

		id,err := strconv.Atoi(ps.ByName("id"))
		if err!=nil{
			fmt.Fprintf(w, "ошибка получения идентификатора")
			return
		}


		t, _ := template.ParseFiles("tmpls/editPublication.html")

		p:=EditPublicationModel{
			Publication: PublicationModel{
				ID:				id,
				Name:           cacheData.publications[id].name,
				Pages:          cacheData.publications[id].pages,
				DOI:            cacheData.publications[id].doi,
				PageCount:      cacheData.publications[id].pageCount,
				PagePrintCount: cacheData.publications[id].pagePrintCont,
				PDF:            cacheData.publications[id].pdf,
			},
			Authors:  nil,
			Editions: nil,
		}
		authors,err:= getSortAuthors()
		if err!=nil{
			fmt.Fprintf(w, "ошибка получения авторов")
			log.Println(err)
			return
		}
		editions,err:= getSortEditions()
		if err!=nil{
			fmt.Fprintf(w, "ошибка получения выпусков")
			log.Println(err)
			return
		}

		for _,idAuthor:= range  cacheData.publications[id].idAuthors {
			for i, author := range authors{
				if author.ID == idAuthor{
					authors[i].Checked = true
				}
			}
		}
		p.Authors = authors
		idEdition := cacheData.publications[id].idEdition
		for i, edition := range editions{
			if edition.ID == idEdition{
				editions[i].Checked = true
			}
		}
		p.Editions = editions
		//fmt.Println(p)

		t.Execute(w,p)
	})

	router.POST("/editPublication/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		checkLogin(w, r)

		id,err := strconv.Atoi(ps.ByName("id"))
		if err!=nil{
			fmt.Fprintf(w, "ошибка получения идентификатора")
			return
		}

		file, handler, err := r.FormFile("pdf")
		if err != nil {
			log.Println(err)
			return
		}
		path:= ""
		if handler.Filename !=""{
			path,err = upload(file, handler.Filename, "/publications/")
			if err!=nil{
				fmt.Fprintf(w,"ошибка добавления файла")
				return
			}
			err=os.Remove("." +cacheData.publications[id].pdf)
			if err!=nil{
				fmt.Fprintf(w,"ошибка удаления  файла")
				return
			}
		}

		name:=r.FormValue("name")
		pages:=r.FormValue("pages")
		doi:=r.FormValue("doi")
		pc,err:=strconv.Atoi(r.FormValue("page_count"))
		if err!=nil{
			fmt.Fprintf(w, "ошибка преобразования данных")
			return
		}
		ppc,err:=strconv.Atoi(r.FormValue("page_print_count"))
		if err!=nil{
			fmt.Fprintf(w, "ошибка преобразования данных")
			return
		}

		edition,err:=strconv.Atoi(r.FormValue("edition"))
		if err!=nil{
			fmt.Fprintf(w, "ошибка преобразования данных")
			return
		}

		err = updatePublication(id,name,pages,doi,pc,ppc,path,edition)
		if err!=nil{
			fmt.Fprintf(w, "ошибка добавления")
			log.Println(err)
			return
		}
		r.ParseForm()
		authors:=r.Form["authors"]

		var is []int
		for _,d:=range authors{
			i,err:= strconv.Atoi(d)
			if err!=nil{
				log.Println(err)
				return
			}
			is = append(is, i)
		}

		err = deleteAuthorPublication(-1, id)
		if err!=nil{
			fmt.Fprintf(w, "ошибка уделения связей")
			fmt.Println(err)
			err = deletePublication(id)
			if err!=nil{
				fmt.Fprintf(w, "ошибка удаления записи")
				fmt.Println(err)}
			return
		}

		err= addAuthorPublication(is,id)
		if err!=nil{
			log.Println(err)
			fmt.Fprintf(w, "ошибка добавления связей")
			err = deletePublication(id)
			if err!=nil{fmt.Fprintf(w, "ошибка удаления записи")}
			return
		}
		refreshCache()
		http.Redirect(w,r,"/publications",302)


	})

	router.POST("/deletePublication/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		checkLogin(w, r)


		id,err:= strconv.Atoi(ps.ByName("id"))
		if err!=nil{
			fmt.Fprintf(w, "ошибка получения идентификатора")
			return
		}
		err=deletePublication(id)
		if err!=nil{
			fmt.Fprintf(w, "ошибка удаления статьи")
		}
		err = deleteAuthorPublication(-1, id)
		if err!=nil{
			fmt.Fprintf(w, "ошибка удаления зависимостей")
		}
		refreshCache()
		http.Redirect(w,r,"/publications", 302)

	})

	router.GET("/authors", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		checkLogin(w, r)
		t, _ := template.ParseFiles("tmpls/authors.html")
		var authors []AuthorModel
		for i,d:= range cacheData.authors{
			authors = append(authors , AuthorModel{
				ID:    i,
				FioRu: d.fioRu,
				FioUa: d.fioUa,
				FioEn: d.fioEn,
				Email: d.email,
				Phone: d.phone,
				ORCID: d.orcid,
				Work:  d.work,
			})
		}
		//fmt.Println(authors)
		t.Execute(w, authors)
	})

	router.GET("/author/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		checkLogin(w, r)
		t, _ := template.ParseFiles("tmpls/author.html")
		id,err:= strconv.Atoi(ps.ByName("id"))
		if err!=nil{
			log.Println(err)
			fmt.Fprintf(w,"ошибка получения данных")
			return
		}
		//fmt.Println(authors)
		t.Execute(w, AuthorModel{
			ID:   id ,
			FioRu: cacheData.authors[id].fioRu,
			FioUa: cacheData.authors[id].fioUa,
			FioEn: cacheData.authors[id].fioEn,
			Email: cacheData.authors[id].email,
			Phone: cacheData.authors[id].phone,
			ORCID: cacheData.authors[id].orcid,
			Work:  cacheData.authors[id].work,
		})

	})

	router.GET("/editAuthor/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		checkLogin(w, r)
		t, _ := template.ParseFiles("tmpls/editAuthor.html")
		id,err:= strconv.Atoi(ps.ByName("id"))
		if err!=nil{
			log.Println(err)
			fmt.Fprintf(w,"ошибка получения данных")
			return
		}
		//fmt.Println(authors)
		t.Execute(w, AuthorModel{
			ID:   id ,
			FioRu: cacheData.authors[id].fioRu,
			FioUa: cacheData.authors[id].fioUa,
			FioEn: cacheData.authors[id].fioEn,
			Email: cacheData.authors[id].email,
			Phone: cacheData.authors[id].phone,
			ORCID: cacheData.authors[id].orcid,
			Work:  cacheData.authors[id].work,
		})

	})

	router.POST("/editAuthor/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		checkLogin(w, r)

		id,err:= strconv.Atoi(ps.ByName("id"))
		if err!=nil{
			log.Println(err)
			fmt.Fprintf(w,"ошибка получения данных")
			return
		}
		ru:= r.FormValue("fio-ru")
		ua:= r.FormValue("fio-ua")
		en:= r.FormValue("fio-en")
		email:= r.FormValue("email")
		phone:= r.FormValue("phone")
		orcid:= r.FormValue("orcid")
		work:= r.FormValue("work")
		err = updateAuthor(id,ru,ua,en, email, phone, orcid, work)
		if err!=nil{
			log.Println(err)
			fmt.Fprintf(w,"ошибка редактирования данных")
			return
		}
		refreshCache()
		http.Redirect(w,r,"/authors", 302)

	})

	router.POST("/deleteAuthor/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		checkLogin(w, r)
		id,err:= strconv.Atoi(ps.ByName("id"))
		if err!=nil{
			log.Println(err)
			fmt.Fprintf(w,"ошибка получения данных")
			return
		}
		err = deleteAuthor(id)
		if err!= nil{
			fmt.Fprintf(w,"ошибка удаления")
			return
		}
		refreshCache()
		http.Redirect(w,r, "/authors",302)

	})

	router.GET("/addAuthor", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		checkLogin(w, r)
		t, _ := template.ParseFiles("tmpls/addAuthor.html")
		t.Execute(w, nil)
	})

	router.POST("/addAuthor", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		checkLogin(w, r)
		r.ParseForm()
		err:= addAuthor(r.FormValue("fio-ru"), r.FormValue("fio-ua"), r.FormValue("fio-en"),r.FormValue("email"), r.FormValue("phone"), r.FormValue("orcid"), r.FormValue("work"))
		if err!=nil {
			log.Println(err)
			fmt.Fprintf(w, "ошибка добавления, попробуйте еще раз или перезапустите программу")
			return
		}
		//log.Println("author add")
		refreshCache()


		http.Redirect(w,r,"/authors", 302)
	})


	router.GET("/editions", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		checkLogin(w, r)
		t, _ := template.ParseFiles("tmpls/editions.html")
		var editions []EditionModel
		for i,d:= range cacheData.edition{
			editions = append(editions, EditionModel{
				ID:           	i,
				Tom:         	d.tom,
				Number:       	d.number,
				Year:         	d.year,
				PDF:          	d.pdf,
				Publications: nil,
			})
		}
		t.Execute(w, editions)
	})

	router.GET("/edition/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		checkLogin(w, r)
		id,err:=strconv.Atoi( ps.ByName("id"))
		if err!=nil{
			fmt.Fprintf(w, "ошибка получения данных")
			return
		}
		t, _ := template.ParseFiles("tmpls/edition.html")
		t.Execute(w, EditionModel{
			ID:           id,
			Tom:          cacheData.edition[id].tom,
			Number:       cacheData.edition[id].number,
			Year:         cacheData.edition[id].year,
			PDF:          cacheData.edition[id].pdf,
			Publications: nil,
		})
	})


	router.GET("/addEdition", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		checkLogin(w, r)
		t, _ := template.ParseFiles("tmpls/addEdition.html")
		t.Execute(w, nil)
	})

	router.POST("/addEdition", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		checkLogin(w, r)
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("pdf")
		if err != nil {
			log.Println(err)
		}
		path,err := upload(file, handler.Filename, "/editions/")
		if err!=nil{
			fmt.Fprintf(w,"ошибка добавления файла")
			return
		}
		tom,err:= strconv.Atoi(r.FormValue("tom"))
		if err!=nil{
			fmt.Fprintf(w,"ошибка преобразования, возможно, вы ввели не правильные значения")
			return
		}
		number,err:=strconv.Atoi(r.FormValue("number"))
		if err!=nil{
			fmt.Fprintf(w,"ошибка преобразования, возможно, вы ввели не правильные значения")
			return
		}
		year,err:=strconv.Atoi(r.FormValue("year"))
		if err!=nil{
			fmt.Fprintf(w,"ошибка преобразования, возможно, вы ввели не правильные значения")
			return
		}

		err = addEdition(tom,number,year,path)
		if err!=nil {
			log.Println(err)
			fmt.Fprintf(w, "ошибка добавления, попробуйте еще раз или перезапустите программу")
			return
		}
		refreshCache()
		http.Redirect(w,r,"/editions", 302)
	})

	router.POST("/deleteEdition/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		checkLogin(w, r)
		id,err:= strconv.Atoi(ps.ByName("id"))
		if err!=nil{
			log.Println(err)
			fmt.Fprintf(w,"ошибка получения данных")
			return
		}
		err = deleteEditon(id)
		if err!= nil{
			fmt.Fprintf(w,"ошибка удаления")
			return
		}
		refreshCache()
		http.Redirect(w,r, "/editions",302)

	})

	router.GET("/editEdition/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		checkLogin(w, r)
		t, _ := template.ParseFiles("tmpls/editEdition.html")
		id,err:= strconv.Atoi(ps.ByName("id"))
		if err!=nil{
			log.Println(err)
			fmt.Fprintf(w,"ошибка получения данных")
			return
		}
		//fmt.Println(authors)
		t.Execute(w, EditionModel{
			ID:           id,
			Tom:          cacheData.edition[id].tom,
			Number:       cacheData.edition[id].number,
			Year:         cacheData.edition[id].year,
			PDF:          cacheData.edition[id].pdf,
			Publications: nil,
		})

	})

	router.POST("/editEdition/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		checkLogin(w, r)

		id,err:= strconv.Atoi(ps.ByName("id"))
		if err!=nil{
			log.Println(err)
			fmt.Fprintf(w,"ошибка конвертации данных, проверьте правильность введенных данных")
			return
		}
		tom,err := strconv.Atoi(r.FormValue("tom"))
		if err!=nil{
			log.Println(err)
			fmt.Fprintf(w,"ошибка конвертации данных, проверьте правильность введенных данных")
			return
		}
		number,err := strconv.Atoi(r.FormValue("number"))
		if err!=nil{
			log.Println(err)
			fmt.Fprintf(w,"ошибка конвертации данных, проверьте правильность введенных данных")
			return
		}
		year,err := strconv.Atoi(r.FormValue("year"))
		if err!=nil{
			log.Println(err)
			fmt.Fprintf(w,"ошибка конвертации данных, проверьте правильность введенных данных")
			return
		}
		file, handler, err := r.FormFile("pdf")
		if err != nil {
			log.Println(err)
			return
		}
		path:= ""
		if handler.Filename !=""{
			path,err = upload(file, handler.Filename, "/editions/")
			if err!=nil{
				fmt.Fprintf(w,"ошибка добавления файла")
				return
			}
			err=os.Remove("." +cacheData.edition[id].pdf)
			if err!=nil{
				fmt.Fprintf(w,"ошибка удаления  файла")
				return
			}
		}




		err = updateEdition(id,tom,number,year,path)
		if err!=nil{
			log.Println(err)
			fmt.Fprintf(w,"ошибка редактирования данных")
			return
		}
		refreshCache()
		http.Redirect(w,r,"/editions", 302)

	})


	router.GET("/diagram", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		checkLogin(w, r)
		t, _ := template.ParseFiles("tmpls/diagram.html")
		t.Execute(w, nil)
	})

	router.GET("/errorLogin",func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fmt.Fprintf(w,"не верные данные авторизации")
	})

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r* http.Request) {
		fmt.Fprintf(w,"42")

	})



	//Iit Server
	server := http.Server{
		Addr:         ":8080",
		ReadTimeout:  time.Duration(30) * time.Second,
		WriteTimeout: time.Duration(30) * time.Second,
		Handler:      router,
	}
	//refreshCache(db)
	//go autoRefreshCache(db)
	fmt.Println("server listen and serve...")
	/*start:= time.Now()
	for i:=1; i!=3000;i++{
		err:=addAuthor("author"+strconv.Itoa(i),"author"+strconv.Itoa(i),"author"+strconv.Itoa(i),"author"+strconv.Itoa(i)+"@test.com",strconv.Itoa(i),"test","work")
		if err!=nil{
			log.Println(err)
		}
		err=addEdition(i,i,i,"test")
		if err!=nil{
			log.Println(err)
		}
	}

	t := time.Now()
	 fmt.Println(t.Sub(start))*/

	panic(server.ListenAndServe())


}
