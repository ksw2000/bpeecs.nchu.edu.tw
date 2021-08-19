// util function
function stripHtml(html) {
    var tmp = document.createElement("div");
    tmp.innerHTML = html;
    return tmp.textContent || tmp.innerText || "";
}

function keyEnter(e, callback) {
    let keycode = (window.event) ? window.event.keyCode : e.which;
    if (keycode === 13 && callback) callback();
}

function renderAttachment(list) {
    ret = '';
    list.forEach(e => {
        ret += `<li><a href="${e.path}">${e.client_name}</a></li>`;
    });
    console.log(ret);
    return ret;
}

/**
 * interface calendar{
 *      id: number
 *      day: number
 *      month: number
 *      year: number     
 *      event: string
 *      link: string
 * } 
 */
function renderCalendar(infoList, editMode) {
    if (infoList && infoList.length > 0) {
        let ret = '<div class="calendar">';
        infoList.forEach(e => {
            ret += '<div class="calendar-list">'
            ret += `<div class="calendar-head candy-header"><span class="single cyan big">${e.month} / ${e.day}</span></div>`;
            if (e.link) {
                ret += `<div class="calendar-title"><a href="${e.link}">${e.event}</a></div>`;
            } else {
                ret += `<div class="calendar-title">${e.event}</div>`;
            }
            if (editMode) {
                ret += ' <div class="calendar-tail">';
                ret += `<i class="material-icons" onclick="btnEditCalendar(${e.id})" title="編輯">edit</i>`;
                ret += `<i class="material-icons delete" onclick="btnDelCalendar(${e.id})" title="刪除">close</i>`;
                ret += '</div>';
            }
            ret += '</div>';
        });
        ret += '</div>';
        return ret;
    }

    return '無行程';
}

async function post(url, data) {
    return fetch(url, {
        body: data ? Object.keys(data).map(key => encodeURIComponent(key) + '=' + encodeURIComponent(data[key])).join('&') : "",
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        method: 'POST'
    }).then(res => {
        try {
            return res.json();
        } catch (e) {
            console.log(res);
        }
    });
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

function notice(msg, t) {
    const dom = document.getElementById('notice');
    const timeout = t || 10000; // unit: ms
    dom.innerHTML = msg;
    dom.classList.add('show');
    setTimeout(function () {
        dom.classList.remove('show');
    }, timeout);
}

// -------------------------------- layout page --------------------------------
function openSideBar(){
    const menuBtn = document.querySelector('#button-less-900px a i');
    const sideBar = document.getElementById('button-over-900px');
    if (sideBar.classList.contains('show')){
        sideBar.classList.remove('show');
        menuBtn.innerHTML = 'menu';
    }else{
        sideBar.classList.add('show');
        menuBtn.innerHTML = 'close';
    }
}

// --------------------------------- index page --------------------------------
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

// --------------------  http://localhost:9000/news?type=*----------------------
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

// --------------------------------- login page --------------------------------
function login() {
    const id = $("#login #id").val();
    const pwd = $("#login #pwd").val();
    if (id === "" || pwd === "") {
        notice("帳號或密碼不可為空！");
        return;
    }
    $.post('/api/login', {
        id: id,
        pwd: pwd
    }, (data) => {
        if (data.err) {
            notice(data.err);
        } else {
            window.location.href = "/manage";
        }
    }, 'json');
}

// ---------------------------------- reg page ---------------------------------
function register() {
    $.post('/api/reg', {
        id: $("#reg #id").val(),
        pwd: $("#reg #pwd").val(),
        re_pwd: $("#reg #re-pwd").val(),
        name: $("#reg #name").val()
    }, (data) => {
        if (data.err) {
            notice(data.err);
        } else {
            window.location.href = "/manage/reg-done";
        }
    }, 'json');
}
