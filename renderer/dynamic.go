package renderer
import (
    "bytes"
    "encoding/json"
    "html/template"
    "time"
    "bpeecs.nchu.edu.tw/article"
)

type Files_render struct{
    Client_name []string `json:"client_name"`
    Server_name []string `json:"server_name"`
    Path []string        `json:"path"`
}

type File_render struct{
    Client_name string
    Server_name string
    Path string
}

type Article_render struct{
    Id uint32
    User string
    Type string
    Create_time string
    Publish_time string
    Last_modified string
    Title string
    Content template.HTML
    Attachment []File_render
}

func RenderSimpleTime(timestamp uint64) string{
    //t := time.Unix(fmt.Sprintf("%u", timestamp), 0)
    t := time.Unix(int64(timestamp), 0)
    return t.Format("2006-01-02")
    //(t.Format("2006-01-02 15:04:05"))
}

func RenderArticleType(key string) string{
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

func RenderPublicArticle(artInfo *article.Article_Format) template.HTML{
    data := new(Article_render)
    data.Id = artInfo.Id
    data.User = artInfo.User
    data.Type = RenderArticleType(artInfo.Type)
    data.Create_time = RenderSimpleTime(artInfo.Create_time)
    data.Publish_time = RenderSimpleTime(artInfo.Publish_time)
    data.Last_modified = RenderSimpleTime(artInfo.Last_modified)
    data.Title = artInfo.Title
    data.Content = template.HTML(artInfo.Content)

    res := Files_render{}
    json.Unmarshal([]byte(artInfo.Attachment), &res)
    data.Attachment = make([]File_render, len(res.Path))
    for i:=0; i < len(res.Path); i++{
        data.Attachment[i] = File_render{
            Client_name : res.Client_name[i],
            Server_name : res.Server_name[i],
            Path : res.Path[i],
        }
    }

    var buf bytes.Buffer
    t, _ := template.ParseFiles("./include/article_layout.html")
    t.Execute(&buf, data)
    return template.HTML(buf.String())
}

func RenderPublicArticleBriefList(artInfoList []article.Article_Format) template.HTML{
    data := new(Article_render)
    ret := ""
    for _, artInfo := range artInfoList{
        data.Id = artInfo.Id
        data.Publish_time = RenderSimpleTime(artInfo.Publish_time)
        data.Title = artInfo.Title
        var buf bytes.Buffer
        t, _ := template.New("article_list_brief").Parse(`
        <div class="article brief">
            <div class="candy-header"><span class="single">{{.Publish_time}}</span></div>
            <a href="/news?id={{.Id}}">{{.Title}}</a>
        </div>`)
        t.Execute(&buf, data)
        ret += buf.String()
    }

    return template.HTML(ret)
}
