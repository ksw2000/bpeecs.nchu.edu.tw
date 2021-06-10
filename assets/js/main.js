function stripHtml(html) {
    var tmp = document.createElement("div");
    tmp.innerHTML = html;
    return tmp.textContent || tmp.innerText || "";
}

// @param string(json)
// @return string(html code)
function loadAttchment(str) {
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

const articleTypeMap = {
    "normal": "一般消息",
    "activity": "演講 & 活動",
    "course": "課程 & 招生",
    "scholarships": "獎學金",
    "recruit": "徵才資訊"
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
