package renderer
import (
    "bytes"
    "encoding/json"
    "fmt"
    "html/template"
    "io/ioutil"
    "log"
    "time"
    "bpeecs.nchu.edu.tw/article"
)

const syllabusDataDir = "./assets/json/syllabus/"
const syllabusTemplate = "./include/syllabus/template.gohtml"

type fileInfo struct{
    ClientName string
    ServerName string
    Mime string
    Path string
}

type articleRenderInfo struct{
    ID uint32
    User string
    Type string
    CreateTime string
    PublishTime string
    LastModified string
    Title string
    Content template.HTML
    Attachment []fileInfo
    PhotoAttachment []fileInfo  // render photos by HTML <img>
}

func renderDate(timestamp uint64) string{
    //t := time.Unix(fmt.Sprintf("%u", timestamp), 0)
    t := time.Unix(int64(timestamp), 0)
    return t.Format("2006-01-02")
    //(t.Format("2006-01-02 15:04:05"))
}

func renderArticleType(key string) string{
    dict := map[string]string{
        "normal"       : "一般消息",
        "activity"     : "演講 & 活動",
        "course"       : "課程 & 招生",
        "scholarships" : "獎學金",
        "recruit"      : "徵才資訊",
    }

    val, ok := dict[key]
    if ok{
        return val
    }
    return ""
}

// RenderPublicArticle renders an article at url: /news/[articleID]
func RenderPublicArticle(artInfo *article.Format) template.HTML{
    data := new(articleRenderInfo)
    data.ID = artInfo.ID
    data.User = artInfo.User
    data.Type = renderArticleType(artInfo.Type)
    data.CreateTime = renderDate(artInfo.CreateTime)
    data.PublishTime = renderDate(artInfo.PublishTime)
    data.LastModified = renderDate(artInfo.LastModified)
    data.Title = artInfo.Title
    data.Content = template.HTML(artInfo.Content)
    data.PhotoAttachment = []fileInfo{}
    data.Attachment = []fileInfo{}

    for _, v := range artInfo.Attachment{
        var extName string
        matchNum, _ := fmt.Sscanf(v.Mime, "image/%s", &extName)
        if(matchNum > 0){
            data.PhotoAttachment = append(data.PhotoAttachment, fileInfo{
                Path: v.Path,
            })
        }else{
            data.Attachment = append(data.Attachment, fileInfo{
                ClientName : v.ClientName,
                ServerName : v.ServerName,
                Mime: v.Mime,
                Path : v.Path,
            })
        }
    }

    var buf bytes.Buffer
    t, _ := template.ParseFiles("./include/article_layout.gohtml")
    t.Execute(&buf, data)
    return template.HTML(buf.String())
}

// RenderPublicArticleBriefList dynamically renders article list an home page
func RenderPublicArticleBriefList(artInfoList []article.Format) template.HTML{
    data := new(articleRenderInfo)
    ret := ""
    for _, artInfo := range artInfoList{
        data.ID = artInfo.ID
        data.PublishTime = renderDate(artInfo.PublishTime)
        data.Title = artInfo.Title
        var buf bytes.Buffer
        t, _ := template.New("article_list_brief").Parse(`
        <div class="article brief">
            <div class="candy-header"><span class="single">{{.PublishTime}}</span></div>
            <a href="/news?id={{.ID}}">{{.Title}}</a>
        </div>`)
        t.Execute(&buf, data)
        ret += buf.String()
    }

    return template.HTML(ret)
}

// RenderSyllabus dynamically renders syllabus page
func RenderSyllabus(semester int, courseNumber int) (template.HTML, string, error){
    path := fmt.Sprintf("%s%d/%d.json", syllabusDataDir, semester, courseNumber)

    jsonData, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalln("render/dynamic.go RenderSyllabus() can not found " + path)
        return template.HTML(""), "", err
    }

    maps := map[string]interface{}{}
    json.Unmarshal(jsonData, &maps)

    courseName, _ := maps["Course_name_zh"].(string)

    for k, v := range maps{
        switch u := v.(type){
        case string:
            maps[k] = template.HTML(u)
        }
    }


    t, err := template.ParseFiles(syllabusTemplate)
    if err != nil{
        log.Fatalln("render/dynamic.go RenderSyllabus() can not found template " +
        syllabusTemplate)
    }

    var buf bytes.Buffer
    t.Execute(&buf, maps)
    return template.HTML(buf.String()), courseName, nil
}
