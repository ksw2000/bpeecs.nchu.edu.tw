function notice(msg){
    $("#notice").html(msg);
    $("#notice").slideDown(100,function(){
        setTimeout(function(){
            $("#notice").slideUp(500);
        },10000);
    });
}

function slideToggole(id){
    $('#'+id).slideToggle();
}

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

function loadNews(type, from, to){
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

function loadNewsOnlyTitle(type){
    return new Promise((resolve, reject) => {
        loadNews(type).then((data) => {
            data = data.NewsList;
            let len = data.length;
            let ret = "";
            for(let i=0; i<len; i++){
                data[i].Publish_time = $.format.date(new Date(data[i].Publish_time * 1000), "yyyy-MM-dd");

                ret += `
                <div class="article" data-id="${data[i].Id}" style="cursor: pointer"
                    onclick="window.location='/news?id=${data[i].Id}'">
                    <div class="header">${data[i].Publish_time}</div>
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

function loadNewsForWhat(what, type, from, to){
    var self = this;

    this.from = from;
    this.to   = to;
    this.len  = to - from + 1;

    this.load = () => {
        return new Promise((resolve, reject) => {
            loadNews(type, self.from, self.to).then((data) => {
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
            if(data==null) return;
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

                let draftIcon = (isDraft)? '<div class="draftIcon">draft</div>' : '';
                let draftColor = (isDraft)? 'border-color:rgb(254, 108, 108);':'';

                ret += `
                <div class="article" data-id="${data[i].Id}"
                     style="${draftColor}">
                    <h2 class="title">${draftIcon}${data[i].Title}</h2>
                    <div class="header">
                        <p>Create at：${data[i].Create_time}</p>
                        <p>Last modified：${data[i].Last_modified}</p>
                `
                if(!isDraft){
                    ret+=`
                            <p>First publish：${data[i].Publish_time}</p>
                    `
                }

                ret+=`
                    </div>
                    <div class="content">
                        ${newContent}
                    </div>
                    <div class="buttonArea" style="text-align: right;">
                        <button id="attachment" onclick="javascript:delete_news(this, ${data[i].Id})" class="red">Delete</button>
                        <button id="publish" onclick="javascript:edit_news(${data[i].Id})" class="blue">Edit</button>
                        <button id="attachment" onclick="window.location='/news?id=${data[i].Id}'">Read</button>
                    </div>
                </div>
                `;
            }
            return ret;
        }
    }else if(what == 'brief'){
        this.render = (data) => {
            if(data == null) return
            let len = data.length;
            let ret = "";
            for(let i=0; i<len; i++){
                data[i].Publish_time = $.format.date(new Date(data[i].Publish_time * 1000), "yyyy-MM-dd HH : mm");
                let newContent = stripHtml(marked(data[i].Content));
                    if(newContent.length>30){
                        newContent = newContent.slice(0, 50);
                        newContent += `...<a href="/news?id=${data[i].Id}">略</a>`;
                    }

                ret += `
                <div class="article" data-id="${data[i].Id}">
                    <h2 class="title">${data[i].Title}</h2>
                    <div class="header">
                        <p>發佈時間：${data[i].Publish_time}</p>
                    </div>
                    <div class="content">
                        ${newContent}
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

            data.Last_modified = $.format.date(new Date(data.Last_modified * 1000), "yyyy-MM-dd HH : mm");
            data.Publish_time  = $.format.date(new Date(data.Publish_time * 1000), "yyyy-MM-dd HH : mm");
            data.Content       = marked(data.Content);

            ret.text += `
            <div class="article" data-id="${data.Id}" style="border:0px;">
                <div class="header">
                    <p>修改時間：${data.Last_modified}</p>
                    <p>發佈時間：${data.Publish_time}</p>
                </div>
                <div class="content">
                    ${data.Content}
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
