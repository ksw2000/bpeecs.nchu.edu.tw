function stripHtml(html) {
    var tmp = document.createElement("div");
    tmp.innerHTML = html;
    return tmp.textContent || tmp.innerText || "";
}

function hash(str) {
    let h = 0;
    for (let i = 0; i < str.length; i++) {
        h = ((h << 5) - h) + str.charCodeAt(i);
        h |= 0;
    }
    return h;
}

// input: string(json)
// output: string(html code)
function loadAttchment(str) {
    // load attachment
    if (str === "") return ""

    try {
        let attachment = "";
        let parse = JSON.parse(str);
        for (let i = 0; parse.clientName.length < len; i++) {
            attachment += `<li><a href="${parse.path[i]}">${parse.clientName[i]}</a></li>`;
        }
        return attachment;
    } catch (e) {
        return "";
    }
}

function appendMoreInfo(obj) {
    $(obj).next().slideToggle();
}

function articleTypeDecoder(key) {
    let dict = {
        "normal": "一般消息",
        "activity": "演講 & 活動",
        "course": "課程 & 招生",
        "scholarships": "獎學金",
        "recruit": "徵才資訊"
    }
    return dict[key];
}

function loadNews(scope, type, from, to) {
    if (type === undefined) type = 'normal';
    if (from === undefined) from = '';
    if (to === undefined) to = '';

    return new Promise((resolve, reject) => {
        $.ajax({
            url: '/api/get_news',
            data: {
                'scope': scope,
                'type': type,
                'from': from,
                'to': to
            },
            type: 'GET',
            success: function (data) {
                resolve(data);
            },
            error: function (err) {
                reject(err);
            },
            dataType: 'json'
        });
    });
}

function LoadNewsForWhat(what, scope, type, from, to) {
    var self = this;

    this.from = from;
    this.to = to;
    this.len = to - from + 1;

    this.load = () => {
        return new Promise((resolve, reject) => {
            loadNews(scope, type, self.from, self.to).then((data) => {
                let ret = self.render(data.list);
                if (data.hasNext) {
                    ret += `<div>
                                <button style="margin:0px auto;" onclick="loadNext(this)">More</button>
                            </div>
                            `;
                }
                resolve(ret);
            }).catch((err) => {
                reject(err);
            });
        });
    }

    if (what == 'management') {
        this.render = (data) => {
            if (data == null) return "沒有文章...";
            let len = data.length;
            let ret = '';
            for (let i = 0; i < len; i++) {
                let isDraft = (data[i].publish === 0) ? true : false;
                data[i].create = $.format.date(new Date(data[i].create * 1000), "yyyy-MM-dd HH : mm");
                data[i].lastModified = (data[i].lastModified === 0) ? '-' : $.format.date(new Date(data[i].lastModified * 1000), "yyyy-MM-dd HH : mm");
                data[i].publish = (data[i].publish === 0) ? '-' : $.format.date(new Date(data[i].publish * 1000), "yyyy-MM-dd HH : mm");
                let newContent = stripHtml(data[i].content);
                if (newContent.length > 50) {
                    newContent = newContent.slice(0, 80);
                    newContent += `<a href="/news?id=${data[i].id}">...More</a><p></p>`;
                }
                let attachment = loadAttchment(data[i].attachment);
                let draftIcon = isDraft ? '<div class="draftIcon">draft</div>' : '';
                let draftColor = isDraft ? 'border-color:#fe6c6c;' : 'border-color:#14a1ff;';

                ret += `<div class="article" data-id="${data[i].id}" style="${draftColor}">
                    <h2 class="title">${draftIcon}${data[i].title}</h2>`;
                ret += `<div class="header" onclick="javascript:appendMoreInfo(this)">`;
                ret += `    <div class="candy-header"><span>分類</span><span>${articleTypeDecoder(data[i].type)}</span></div>`;
                ret += `    <div class="candy-header"><span>最後編輯</span><span class="orange">${data[i].lastModified}</span></div>`;
                ret += `</div>`;
                ret += `<div style="display: none;">`;
                ret += `  <div class="candy-header hide-less-500px"><span>建立於</span><span class="red">${data[i].create}</span></div>`;
                ret += `  <div class="candy-header hide-less-500px"><span>發佈於</span><span class="green">${data[i].publish}</span></div>`;
                ret += `</div>`;
                ret += `
                    <div class="content">
                        ${newContent}
                    </div>
                    <div id="attachmentArea">
                        <ul>${attachment}</ul>
                    </div>
                    <div class="buttonArea" style="text-align: right;">
                        <button id="read" onclick="window.location='/news?id=${data[i].id}'"" class="border">閱讀</button>
                        <button id="delete" onclick="javascript:delete_what(this, 'news', ${data[i].id})" class="red">刪除</button>
                        <button id="publish" onclick="javascript:edit_news(${data[i].id})" class="blue">編輯</button>
                    </div>
                </div>
                `;
            }
            return ret;
        }
    } else if (what == 'brief') {
        this.render = (data) => {
            if (data == null) return "沒有文章";
            let len = data.length;
            let ret = "";
            for (let i = 0; i < len; i++) {
                data[i].publish = $.format.date(new Date(data[i].publish * 1000), "yyyy-MM-dd");
                let newContent = stripHtml(data[i].content);
                if (newContent.length > 30) {
                    newContent = newContent.slice(0, 80);
                    newContent += `...<a href="/news?id=${data[i].id}">略</a>`;
                }
                let attachment = loadAttchment(data[i].attachment);
                ret += `
                <div class="article" data-id="${data[i].id}">
                    <h2 class="title">${data[i].title}</h2>
                    <div class="header" onclick="javascript:appendMoreInfo(this)">
                        <div class="candy-header"><span>發佈於</span><span>${data[i].publish}</span></div>
                    </div>
                    <div style="display:none;">
                `;
                ret += `<div class="candy-header"><span>分類</span><span class="green">${articleTypeDecoder(data[i].type)}</span></div>`;
                ret += `<div class="candy-header"><span>發文</span><span class="cyan">@${data[i].user}</span></div>`;
                ret += `
                    </div>
                    <div class="content">
                        ${newContent}
                    </div>
                    <div id="attachmentArea">
                        <ul>${attachment}</ul>
                    </div>
                    <p></p>
                    <div class="buttonArea" style="text-align: right;">
                        <button id="attachment" onclick="window.location='/news?id=${data[i].id}'"
                                style="display: inline-block;">閱讀全文</button>
                    </div>
                </div>
                `;
            }
            return ret;
        }
    }

    this.next = () => {
        self.from = self.to + 1;
        self.to += self.len;
        return self.load();
    }
}

function notice(msg) {
    $("#notice").html(msg);
    $("#notice").slideDown(100, function () {
        setTimeout(function () {
            $("#notice").slideUp(500);
        }, 10000);
    });
}

function slideToggole(id) {
    $('#' + id).slideToggle();
}
