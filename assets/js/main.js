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

/**
 * interface calender{
 *      id: number
 *      day: number
 *      month: number
 *      year: number     
 *      event: string
 *      link: string
 * } 
 */
function renderCalendar(infoList, editMode){
    if(infoList && infoList.length > 0){
        let ret = '<div class="calender">';
        infoList.forEach(e => {
            ret += '<div class="calender-list">'
            ret += `<div class="candy-header"><span class="single cyan big">${e.month} / ${e.day}</span></div>`;
            ret += (e.link) ? `<div class="calender-title"><a href="${e.link}">${e.event}</a></div>` : '';
            ret += `<div class="calender-title">${e.event}</div>`
            if(editMode){
                ret += ' <div class="calender-tail">';
                ret += `<i class="material-icons" onclick="btnEditCalendar(${e.id})" title="編輯">edit</i>`;
                ret += `<i class="material-icons delete" onclick="btnDelCalendar(${e.id})" title="刪除">close</i>`;
                ret += '</div>';
            }
            ret+= '</div>';
        });
        ret += '</div>';
        return ret;
    }

    return '無行程';
}

function loadCalendar(date, callback) {
    $('#load-calendar').fadeOut('fast', () => {
        $('#load-calendar').fadeIn('fast', () => {
            $.get('/api/get_calendar', {
                'year': date.getFullYear(),
                'month': Number(date.getMonth()) + 1 // javascript month [0, 12)
            }, (data) => {
                if (data.err) {
                    notice(data.err);
                    console.log(data);
                    return;
                }
                if(callback) callback(data);
            }, 'json');
        });
    });
}

