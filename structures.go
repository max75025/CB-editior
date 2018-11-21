package main



type AuthorModel struct {
	ID			int
	FioRu 		string
	FioUa 		string
	FioEn		string
	Email		string
	Phone		string
	ORCID		string
	Work		string
	Checked		bool
}



type EditPublicationModel struct {
	Publication PublicationModel
	Authors		[]AuthorModel
	Editions	[]EditionModel
}

type AddPublicationModel struct{
	Authors		[]AuthorModel
	Editions	[]EditionModel
}

type PublicationModel struct {
	ID 				int
	Name			string
	Pages			string
	DOI 			string
	PageCount		int
	PagePrintCount	int
	PDF				string
	//idEdition		int
	Authors			[]AuthorModel
	Edition			EditionModel
}




type EditionModel struct{
	ID				int
	Tom 			int
	Number  		int
	Year			int
	PDF				string
	Publications	[]PublicationModel
	Checked			bool
}



type cache struct {
	authors 		map[int]authorCache
	publications	map[int]publicationCache
	edition			map[int]editionCache
}

type publicationCache struct {
	name			string
	pages			string
	doi 			string
	pageCount		int
	pagePrintCont	int
	pdf				string
	idEdition		int
	idAuthors		[]int
}


type editionCache struct {
	tom 	int
	number  int
	year	int
	pdf		string
}

type authorCache struct{
	fioRu 		string
	fioUa 		string
	fioEn		string
	email		string
	phone		string
	orcid		string
	work		string

}

