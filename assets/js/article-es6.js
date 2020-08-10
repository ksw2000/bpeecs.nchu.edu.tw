function stripHtml(html){
   var tmp = document.createElement("div");
   tmp.innerHTML = html;
   return tmp.textContent || tmp.innerText || "";
}

function hash(str){
    let h = 0;
    for(let i = 0; i < str.length; i++) {
        h  = ((h << 5) - h) + str.charCodeAt(i);
        h |= 0;
    }
    return h;
}

// input: string(json)
// output: string(html code)
function loadAttchment(str){
    // load attachment
    if(str === ""){
        return ""
    }

    try{
        let attachment = "";
        let parse = JSON.parse(str);
        let len = parse.client_name.length;
        for(let i=0; i<len; i++){
            attachment += '<li><a href="'+ parse.path[i] + '">' +
                           parse.client_name[i] + '</a></li>';
        }
        return attachment;
    }catch(e){
        return "";
    }
}

function appendMoreInfo(obj){
    $(obj).next().slideToggle();
}

function articleTypeDecoder(key){
    let dict = {
        "normal" : "一般消息",
        "activity" : "演講 & 活動",
        "course" : "課程 & 招生",
        "scholarships" : "獎學金",
        "recruit" : "徵才資訊"
    }
    return dict[key];
}

function loadNews(scope, type, from, to){
    if(type === undefined){
        type = 'normal';
    }
    if(from === undefined){
        from = '';
    }
    if(to === undefined){
        to = '';
    }

    return new Promise((resolve, reject) => {
        $.ajax({
            url: '/function/get_news',
            data: {
                'scope' : scope,
                'type' : type,
                'from' : from,
                'to' : to
            },
            type: 'GET',
            success: function(data){
                resolve(data);
            },
            error: function(err) {
                reject(err);
            },
            dataType: 'json'
        });
    });
}

function loadNewsOnlyTitle(scope){
    return new Promise((resolve, reject) => {
        loadNews(scope, 'normal').then((data) => {
            data = data.NewsList;
            let len = data.length;
            let ret = "";
            for(let i=0; i<len; i++){
                data[i].Publish_time = $.format.date(new Date(data[i].Publish_time * 1000), "yyyy-MM-dd");

                ret += `
                <div class="article" data-id="${data[i].Id}" style="cursor: pointer"
                    onclick="window.location='/news?id=${data[i].Id}'">

                    <div class="candy-header"><span class="single">${data[i].Publish_time}</span></div>
                    <h2 class="title">${data[i].Title}</h2>
                </div>
                `;
            }
            resolve(ret);
        }).catch((err)=>{
            reject(err);
        });
    });
}

function loadNewsForWhat(what, scope, type, from, to){
    var self = this;

    this.from = from;
    this.to   = to;
    this.len  = to - from + 1;

    this.load = () => {
        return new Promise((resolve, reject) => {
            loadNews(scope, type, self.from, self.to).then((data) => {
                let ret = self.render(data.NewsList);
                if(data.HasNext){
                    ret += `<div>
                                <button style="margin:0px auto;" onclick="loadNext(this)">More</button>
                            </div>
                            `;
                }
                resolve(ret);
            }).catch((err)=>{
                reject(err);
            });
        });
    }

    if(what == 'management'){
        this.render = (data) => {
            if(data==null) return "No articles";
            let len = data.length;
            let ret = '';
            for(let i=0; i<len; i++){
                let isDraft = (data[i].Publish_time === 0)? true : false;
                data[i].Create_time   = $.format.date(new Date(data[i].Create_time * 1000), "yyyy-MM-dd HH : mm");
                data[i].Last_modified = (data[i].Last_modified === 0)? '-' : $.format.date(new Date(data[i].Last_modified * 1000), "yyyy-MM-dd HH : mm");
                data[i].Publish_time  = (data[i].Publish_time === 0)?  '-' : $.format.date(new Date(data[i].Publish_time * 1000), "yyyy-MM-dd HH : mm");
                let newContent = stripHtml(marked(data[i].Content));
                if(newContent.length > 50){
                    newContent = newContent.slice(0, 80);
                    newContent += `<a href="/news?id=${data[i].Id}">...More</a><p></p>`;
                }
                let attachment = loadAttchment(data[i].Attachment);
                let draftIcon = (isDraft)? '<div class="draftIcon">draft</div>' : '';
                let draftColor = (isDraft)? 'border-color:#fe6c6c;' : 'border-color:#14a1ff;';

                ret += `<div class="article" data-id="${data[i].Id}" style="${draftColor}">
                    <h2 class="title">${draftIcon}${data[i].Title}</h2>`;
                ret+=`<div class="header" onclick="javascript:appendMoreInfo(this)">`;
                ret+=`    <div class="candy-header"><span>分類</span><span>${articleTypeDecoder(data[i].Type)}</span></div>`;
                ret+=`    <div class="candy-header"><span>最後編輯</span><span class="orange">${data[i].Last_modified}</span></div>`;
                ret+=`</div>`;
                ret+=`<div style="display: none;">`;
                ret+=`  <div class="candy-header hide-less-500px"><span>建立於</span><span class="red">${data[i].Create_time}</span></div>`;
                ret+=`  <div class="candy-header hide-less-500px"><span>發佈於</span><span class="green">${data[i].Publish_time}</span></div>`;
                ret+=`</div>`;
                ret+=`
                    <div class="content">
                        ${newContent}
                    </div>
                    <div id="attachmentArea">
                        <ul>${attachment}</ul>
                    </div>
                    <div class="buttonArea" style="text-align: right;">
                        <button id="read" onclick="window.location='/news?id=${data[i].Id}'"" class="border">閱讀</button>
                        <button id="delete" onclick="javascript:delete_what(this, 'news', ${data[i].Id})" class="red">刪除</button>
                        <button id="publish" onclick="javascript:edit_news(${data[i].Id})" class="blue">編輯</button>
                    </div>
                </div>
                `;
            }
            return ret;
        }
    }else if(what == 'brief'){
        this.render = (data) => {
            if(data == null) return "No articles";
            let len = data.length;
            let ret = "";
            for(let i=0; i<len; i++){
                data[i].Publish_time  = $.format.date(new Date(data[i].Publish_time * 1000), "yyyy-MM-dd");
                let newContent = stripHtml(marked(data[i].Content));
                if(newContent.length>30){
                    newContent = newContent.slice(0, 80);
                    newContent += `...<a href="/news?id=${data[i].Id}">略</a>`;
                }
                let attachment = loadAttchment(data[i].Attachment);
                ret += `
                <div class="article" data-id="${data[i].Id}">
                    <h2 class="title">${data[i].Title}</h2>
                    <div class="header" onclick="javascript:appendMoreInfo(this)">
                        <div class="candy-header"><span>發佈於</span><span>${data[i].Publish_time}</span></div>
                    </div>
                    <div style="display:none;">
                `;
                ret+=`<div class="candy-header"><span>分類</span><span class="green">${articleTypeDecoder(data[i].Type)}</span></div>`;
                ret+=`<div class="candy-header"><span>發文</span><span class="cyan">@${data[i].User}</span></div>`;
                ret+=`
                    </div>
                    <div class="content">
                        ${newContent}
                    </div>
                    <div id="attachmentArea">
                        <ul>${attachment}</ul>
                    </div>
                    <p></p>
                    <div class="buttonArea" style="text-align: right;">
                        <button id="attachment" onclick="window.location='/news?id=${data[i].Id}'"
                                style="display: inline-block;">閱讀全文</button>
                    </div>
                </div>
                `;
            }
            return ret;
        }
    }

    this.next = () => {
        self.from  = self.to + 1;
        self.to   += self.len;
        return self.load();
    }
}

function loadNewsById(newsID){
    return new Promise((resolve, reject) => {
        $.ajax({
            url: '/function/get_news?id=' + newsID,
            type: 'GET',
            success: function(data){
                resolve(data);
            },
            error: function(err) {
                reject(err);
            },
            dataType: 'json'
        });
    });
}

function loadPublicNewsById(id){
    return new Promise((resolve, reject) => {
        loadNewsById(id).then((data) => {
            let ret = {};
            ret.text = "";

            data.Publish_time  = $.format.date(new Date(data.Publish_time * 1000), "yyyy-MM-dd");
            data.Content       = marked(data.Content);

            let attachment = loadAttchment(data.Attachment);
            ret.text += `
            <div class="article" data-id="${data.Id}" style="border:0px;">
                <div class="header" onclick="javascript:appendMoreInfo(this)">
                    <div class="candy-header"><span>發佈於</span><span>${data.Publish_time}</span></div>
                </div>
                <div class="header" style="display: none;">`;
            ret.text += `<div class="candy-header"><span>分類</span><span class="green">${articleTypeDecoder(data.Type)}</span></div>`;
            ret.text += `<div class="candy-header"><span>發文</span><span class="cyan">@${data.User}</span></div>`;
            ret.text += `</div>
                <div class="content">
                    ${data.Content}
                </div>
                <div id="attachmentArea">
                    <ul>${attachment}</ul>
                </div>
            </div>
            `;

            ret.json = data;
            resolve(ret);
        }).catch((err)=>{
            reject(err);
        });
    });
}
