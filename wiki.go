package main

import "os"
import "strings"
import "io/ioutil"
import "http"
import "template"
import "regexp"
import "gob"
import "fmt"
import "bytes"

type Page struct{
	Title string
	Body  []byte
}

type taxonomy map[string]string
// modify as per needs
const (
	extenstion = ".txt"
	permissions = 0600
	port = ":8119"
	templateEdit = "edit.html"
	templateView = "view.html"
	templateTag  = "tag.html"
	tagsFile     = "tags.bin"
	errorInvalidPageTitle = "Invalid page title"
	filePath = "./data/"
	templatePath = "./tmpl/"
	tagsPath    = "./tags"
	homePage = "homePage"

//function handlers
	view = "/view/"
	edit = "/edit/"
	save = "/save/"
	tag  = "/tag/"
	lenViewPath = len(view)
	lenEditPath = len(edit)
	lenSavePath = len(save)
	lenTagPath  = len(tag)

)

//modify as per needs
var titleValidator = regexp.MustCompile("^[a-zA-Z0-9]+$")	
	

 

func Store (t taxonomy, fname string) os.Error {
        b := new(bytes.Buffer)
        enc := gob.NewEncoder(b)
        err := enc.Encode(t)
        if err != nil {
                return err
        }

        fh, eopen := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY, 0666)
        defer fh.Close()
        if eopen != nil {
                return eopen
        }
        _,e := fh.Write(b.Bytes())
        if e != nil {
                return e
        }
        return nil
}

func Load (fname string) ( taxonomy, os.Error) {
        fh, err := os.Open(fname)
        if err != nil {
                return nil, err
        }
        t := make(map[string]string)
        dec := gob.NewDecoder(fh)
        err = dec.Decode(&t)
        if err != nil {
                return nil, err
        }
        return t, nil

}


func Add (m taxonomy,str1 string, str2 string) {
		
		for i := 0; i < len(m); i++ {
			 _, ok := m[str1]
			 switch {
			 case ok:
				m[str1] = m[str1] + "," + str2
				return
			  default:
				m[str1] = str2
				return
			}
		}

}
	
	
func (p *Page) save() os.Error {
	filename := filePath + p.Title + extenstion
	return ioutil.WriteFile(filename,p.Body,permissions)
}
	

func loadPage(Title string) (*Page, os.Error){
	Body,err := ioutil.ReadFile(filePath + Title+extenstion)
	if err != nil{
		return nil,err
	}
	return &Page{Title : Title,Body : Body}, nil
}

func getTitle(w http.ResponseWriter, r *http.Request) (title string, err os.Error) {


	switch {
	
	case strings.Contains(r.URL.Path,view):
		title = r.URL.Path[lenViewPath:]
	
	case strings.Contains(r.URL.Path,edit):
		title = r.URL.Path[lenEditPath:]
	
	
	case strings.Contains(r.URL.Path,save):
		title = r.URL.Path[lenSavePath:]
		
	case strings.Contains(r.URL.Path,tag):
		title = "dontcare"

	}
	

	if !titleValidator.MatchString(title) {
		http.NotFound(w, r)
		err = os.NewError(errorInvalidPageTitle)
	}
	return
}



func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	t, err := template.ParseFile(templatePath + tmpl, nil)
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, p)
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	if err != nil {
		return
	}
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, edit+title, http.StatusFound)
		return
	}
	renderTemplate(w, templateView, p)
}

func getMatches(m taxonomy,key string) string {
	for k, _ := range m {
		if k == key{
			return m[k]
		}
	}
	return ""
}

func tagHandler(w http.ResponseWriter, r *http.Request) {
	
	tags := r.FormValue("tags")
	fmt.Println(tags)
	
	_,err2 := Load(tagsPath+tagsFile)
	if err2!=nil{
		return

	}
	
	
	//fmt.Fprintf(w, "<h1>Matched</h1><div>%s</div>", getMatches(m,tags))
	
	
	t, err := template.ParseFile(templatePath + templateTag, nil)
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, nil)
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
	}

}

func editHandler(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	if err != nil {
		return
	}
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, templateEdit, p)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	if err != nil {
		return
	}
	body := r.FormValue("body")
	tags := r.FormValue("tags")

	m,err2 := Load(tagsPath+tagsFile)
	if err2!=nil{
		return

	}
	splitTokens := strings.Split(tags," ",-1)
	for i:=0;i<len(splitTokens);i++ {
		Add(m,splitTokens[i],title)
	}
	
	err2 = Store(m,tagsPath+tagsFile)
	if err2!=nil{
		return 
	}
	
	p := &Page{Title: title, Body: []byte(body)}
	err = p.save()
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, view+title, http.StatusFound)
}

func baseHandler(w http.ResponseWriter, r *http.Request) {
	t := homePage
	p, err := loadPage(t)
	if err != nil {
		http.Redirect(w, r, edit+t, http.StatusFound)
		return
	}
	renderTemplate(w, templateView, p)

}

func main() {
	http.HandleFunc("/", baseHandler)
	http.HandleFunc(view, viewHandler)
	http.HandleFunc(save, saveHandler)
	http.HandleFunc(edit, editHandler)
	http.HandleFunc(tag, tagHandler)
	http.ListenAndServe(port, nil)
	//get map loaded
	 Load(tagsPath+tagsFile)
}
