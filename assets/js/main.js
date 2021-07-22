function stripHtml(html) {
    var tmp = document.createElement("div");
    tmp.innerHTML = html;
    return tmp.textContent || tmp.innerText || "";
}

function loadAttchment(list) {
    ret = '';
    list.forEach(e => {
        ret += `<li><a href="${e.path}">${e.client_name}</a></li>`;
    });
    console.log(ret);
    return ret;
}

function appendMoreInfo(obj) {
    $(obj).next().slideToggle();
}

const articleTypeMap = {
    'normal': '一般消息',
    'activity': '演講 & 活動',
    'course': '課程 & 招生',
    'scholarships': '獎學金',
    'recruit': '徵才資訊'
}

/**
 * @param {string} scope enum{'public', 'public-with-type'}
 * @param {string} type enum{'normal', 'activity', 'course', 'scholarships', 'recruit'}
 * @param {number|null} from? 
 * @param {number|null} to?
 * @return void 
 */

function loadNews(scope, type, from, to) {
    return new Promise((resolve, reject) => {
        $.ajax({
            url: '/api/get_news',
            data: {
                'scope': scope,
                'type': type || 'normal',
                'from': from || '',
                'to': to || ''
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
            if(e.link){
                ret += `<div class="calender-title"><a href="${e.link}">${e.event}</a></div>`;
            }else{
                ret += `<div class="calender-title">${e.event}</div>`;
            }
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
}

