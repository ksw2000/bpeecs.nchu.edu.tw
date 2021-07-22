function hash(str) {
    let h = 0;
    for (let i = 0; i < str.length; i++) {
        h = ((h << 5) - h) + str.charCodeAt(i);
        h |= 0;
    }
    return h;
}

window.directlyLeave = false;
window.onbeforeunload = function () {
    if (!leave_editing()) {
        return '變更尚未儲存 (若確定已儲存請忽略此訊息)';
    }
    return;
};

var Editor = new (function () {
    let self = this;
    this.aid = -1;
    this.defaultAttahmentVal = () => {
        return [{
            client_name: "",
            server_name: "",
            path: ""
        }];
    }
    this.editor = null;

    this.newCKEeditor = () => {
        CKEDITOR.replace('mainEditor');
        self.editor = CKEDITOR.instances.mainEditor;
    }

    this.destroyCKEeditor = () => {
        if (self.editor) {
            self.editor.destroy();
            self.editor = null;
        }
    }

    this.getEditorValue = () => {
        return (self.editor) ? self.editor.getData() : '';
    }
    this.setEditorValue = (val) => {
        if (self.editor) {
            self.editor.setData(val);
        }
    }

    this.attachment = this.defaultAttahmentVal();
    this.getType = () => {
        return $("#new-article-area #type").val();
    }
    this.getTitle = () => {
        return $("#new-article-area #title").val();
    }
    // this.hash is not zero because of attachment json string.
    this.preHash = hash(this.getType() + JSON.stringify(this.attachment));

    this.calcCurrentHash = () => {
        return hash(self.getTitle() + self.getType() + self.getEditorValue() + JSON.stringify(self.attachment));
    }
})();

function leave_editing() {
    if (window.directlyLeave === true) return true;

    if (Editor.calcCurrentHash() != Editor.preHash) {
        return confirm('Changes you made may not be saved.');
    }
    // default is 'leave'
    return true;
}

const btnAreaForDraft = `
<button id="attachment" onchange="javascript:attach(event)">附件
    <form enctype="multipart/form-data">
        <input type="file" multiple/>
    </form>
</button>
<button id="save" onclick="javascript:save(true)">儲存至草稿</button>
<button id="publish" onclick="javascript:publish()" class="blue">發佈</button>
`;

const btnAreaForPublished = `
<button id="attachment" onchange="javascript:attach(event)">附件<input type="file" multiple/></button>
<button id="save" onclick="javascript:save(false)" class="blue">重新發佈</button>
<button id="publish" onclick="javascript:delete_what(this, 'news')" class="red">刪除</button>
`

function newArticle() {
    if ($("#new-article-area").css("display") === "none") {
        openNewArticleArea();
        $("#new-article-area .buttonArea").html(btnAreaForDraft);
    } else {
        closeNewArticleArea();
    }
}

function openNewArticleArea(callback) {
    Editor.destroyCKEeditor();
    Editor.newCKEeditor();
    $("#new-article-area").slideDown('fast', () => {
        if (typeof callback === 'function') {
            callback();
        }
    });
    $("#btn-add").text('收起');
}

function closeNewArticleArea(callback) {
    $("#new-article-area").slideUp('fast', () => {
        if (typeof callback === 'function') {
            callback();
        }
    });
    $("#btn-add").text('展開');
}

function attach(e) {
    let form = new FormData();
    console.log(e.target.files);
    for (let i = 0; i < e.target.files.length; i++) {
        form.append("files", e.target.files[i]);
    }

    $("#new-article-area #attachmentArea div.loader").css('display', 'block');
    setTimeout(function () {
        $.ajax({
            url: '/api/upload',
            processData: false,
            contentType: false,
            mimeType: 'multipart/form-data',
            data: form,
            type: 'POST',
            success: function (data) {
                if (data.err) {
                    if (data.errNotLogin) {
                        window.location = '/?notlogin';
                    }
                    notice(data.err);
                } else {
                    let fl = data.fileList;
                    for (let i = 0; i < e.target.files.length; i++) {
                        Editor.attachment.push({
                            "client_name": e.target.files[i].name,
                            "server_name": fl[i].fileName,
                            "path": fl[i].filePath
                        });

                        $("#new-article-area #attachmentArea ul").append(
                            `<li data-file-name="${fl[i].fileName}">
                                <a href="${fl[i].filePath}">${e.target.files[i].name}</a>
                                <button class="red" onclick="delete_what(this, 'attachment', '${fl[i].fileName}')">刪除</button>
                            </li>`
                        );
                    }
                    $("#new-article-area #attachmentArea div.loader").css('display', 'none');
                }
            },
            dataType: 'json'
        });
    }, 500);
}

function save(isPrivate) {
    $.post('/api/news?save', {
        aid: Editor.aid,
        title: Editor.getTitle(),
        type: Editor.getType(),
        content: Editor.getEditorValue(),
        attachment: JSON.stringify(Editor.attachment)
    }, (data) => {
        if (data.err) {
            if(data.errNotLogin){
                window.location = '/?notlogin';
            }
            notice(data.err);
            console.log(data);
        } else if (isPrivate) {
            Editor.preHash = Editor.calcCurrentHash();
            notice("儲存成功！");
        } else {
            Editor.aid = data.aid;
            // goto news page (for news which have been published already)
            window.directlyLeave = true;
            window.location = `/news?id=${Editor.aid}`;
        }
    }, 'json');
}

function publish() {
    $.post('/api/news?publish', {
        aid: Editor.aid,
        title: Editor.getTitle(),
        type: Editor.getType(),
        content: Editor.getEditorValue(),
        attachment: JSON.stringify(Editor.attachment)
    }, (data) => {
        if (data.err) {
            if (data.errNotLogin) {
                window.location = '/?notlogin';
            }
            notice(data.err);
            console.log(data);
        } else {
            Editor.aid = data.aid;
            // directly go to news page
            window.directlyLeave = true;
            window.location = `/news?id=${Editor.aid}`;
        }
    }, 'json');
}

function delete_what(obj, what, num) {
    if (num === undefined) {
        num = Editor.aid;
    }
    obj = $(obj);
    obj.fadeOut('fast', () => {
        obj.text("確定？");
        if (typeof num === 'string') {
            obj.attr('onclick', `real_delete_${what}('${num}')`);
        } else {
            obj.attr('onclick', `real_delete_${what}(${num})`);
        }
        obj.fadeIn('slow', () => {
            setTimeout(() => {
                obj.fadeOut('fast', () => {
                    obj.text("刪除");
                    if (typeof num === 'string') {
                        obj.attr('onclick', `delete_what(this, '${what}', '${num}')`);
                    } else {
                        obj.attr('onclick', `delete_what(this, '${what}', ${num})`);
                    }

                    obj.fadeIn('slow');
                });
            }, 3000);
        })
    });
}

function real_delete_attachment(fileName) {
    let target = -1;
    // Generate new attachment JSON
    // Find index of this file in Editor.attachment.client_name(server_name, path)
    for (let i = 0; i < Editor.attachment.length; i++) {
        if (Editor.attachment[i].server_name === fileName) {
            target = i;
            break;
        }
    }
    if (target < 0) {
        notice('錯誤，欲刪除之檔案不在附加檔案列表中');
        return;
    }

    Editor.attachment.splice(target, 1);

    $.ajax({
        url: '/api/del_attachment',
        data: {
            'server_name': fileName,
            'aid_num': Editor.aid,
            'new_attachment': JSON.stringify(Editor.attachment)
        },
        type: 'POST',
        success: function (data) {
            if (data.err) {
                if (data.errNotLogin) {
                    window.location = '/?notlogin';
                }
                notice(data.err);
            } else {
                $('#new-article-area #attachmentArea ul li[data-file-name="' + fileName + '"]').slideUp('slow');
            }
        },
        error: function (err) {
            console.log(err);
            notice('Error' + err);
        },
        dataType: 'json'
    });
}

function real_delete_news(data_id) {
    // If user delete the news which is editing
    // Refresh website instead of .slideUp()
    $.ajax({
        url: '/api/news?del',
        data: {
            'aid': data_id
        },
        type: 'POST',
        success: (data) =>{
            if (data.err) {
                notice(data.err);
                console.log(data);
            } else if (data_id === Editor.aid) {
                window.directlyLeave = true;
                location.reload();
            } else {
                $('.article[data-id="' + data_id + '"]').slideUp('slow');
                // close news area
                notice('刪除成功！');
            }
        },
        error: function (err) {
            console.log(err);
            notice('Error');
        },
        dataType: 'json'
    });
}

function edit_news(data_id) {
    //if(leave_editing()){
    // 將原來收起來的區域展開
    openNewArticleArea();
    if ($('.article[data-id="' + Editor.aid + '"]').length) {
        $('.article[data-id="' + Editor.aid + '"]').slideDown();
    }

    $.ajax({
        url: '/api/get_news',
        data: {
            'id': data_id
        },
        type: 'GET',
        success: function (data) {
            console.log(data);
            if (data['err']) {
                if (data['code'] === 1) {
                    window.location = '/?notlogin';
                }
                notice(data['msg']);
                console.log(data);
            } else {
                // Step1: update value title, type, content, attachment and hash, aid
                $("#new-article-area #title").val(data.title);
                for (let i = 0; i < $("#new-article-area #type option").length; i++) {
                    $("#new-article-area #type option").eq(i).removeAttr('selected');
                }
                $("#new-article-area #type option[value = '" + data.type + "']").attr('selected', 'selected');

                Editor.attachment = Editor.defaultAttahmentVal();
                try {
                    Editor.attachment = data.attachment;
                    let attachment = '';
                    for (let i = 0; i < Editor.attachment.length; i++) {
                        attachment += `
                                <li data-file-name="${Editor.attachment[i].server_name}">
                                    <a href="${Editor.attachment[i].path}">${Editor.attachment[i].client_name}</a>
                                    <button class="red" onclick="delete_what(this, 'attachment', '${Editor.attachment[i].server_name}')">刪除</button>
                                </li>
                            `;
                    }
                    $("#new-article-area #attachmentArea ul").html(attachment);
                } catch (e) {
                    console.log(e);
                }
                Editor.aid = data_id;

                // Step2: slideUp news area
                $('.article[data-id="' + data_id + '"]').slideUp('fast', () => {
                    // Step3: Move editor after news area
                    $('#new-article-area').insertAfter($('.article[data-id="' + data_id + '"]'));

                    // Step4: Rebuild buttonArea
                    if (data.publish === 0) {
                        $('#new-article-area .buttonArea').html(btnAreaForDraft);
                    } else {
                        $('#new-article-area .buttonArea').html(btnAreaForPublished);
                    }

                    // Step5: SlideDown editor and jump to there
                    openNewArticleArea(function () {
                        window.scrollTo({
                            top: $("#new-article-area")[0].offsetTop - $("nav")[0].clientHeight,
                            behavior: 'smooth'
                        });
                        Editor.setEditorValue(data.content);
                        Editor.preHash = Editor.calcCurrentHash();
                    });
                });
            }
        },
        error: function (err) {
            console.log(err);
            notice('Error');
        },
        dataType: 'json'
    });
    //}
    // DO NOTHING
}

function artListRenderer(dataList){
    if (dataList == null) return "沒有文章";
    let ret = '';
    dataList.forEach((data) => {
        let isDraft = (data.publish === 0) ? true : false;
        data.create = $.format.date(new Date(data.create * 1000), "yyyy-MM-dd HH : mm");
        data.lastModified = (data.lastModified === 0) ? '-' : $.format.date(new Date(data.lastModified * 1000), "yyyy-MM-dd HH : mm");
        data.publish = (data.publish === 0) ? '-' : $.format.date(new Date(data.publish * 1000), "yyyy-MM-dd HH : mm");
        let newContent = stripHtml(data.content);
        if (newContent.length > 50) {
            newContent = newContent.slice(0, 80);
            newContent += `<a href="/news?id=${data.id}">...More</a><p></p>`;
        }
        let attachment = renderAttachment(data.attachment);
        let draftIcon = isDraft ? '<div class="draftIcon">draft</div>' : '';
        let draftColor = isDraft ? 'border-color:#fe6c6c;' : 'border-color:#14a1ff;';

        ret += `<div class="article" data-id="${data.id}" style="${draftColor}">
                    <h2 class="title">${draftIcon}${data.title}</h2>`;
        ret += `<div class="header" onclick="javascript:appendMoreInfo(this)">`;
        ret += `    <div class="candy-header"><span>分類</span><span>${articleTypeMap[data.type]}</span></div>`;
        ret += `    <div class="candy-header"><span>最後編輯</span><span class="orange">${data.lastModified}</span></div>`;
        ret += `</div>`;
        ret += `<div style="display: none;">`;
        ret += `  <div class="candy-header hide-less-500px"><span>建立於</span><span class="red">${data.create}</span></div>`;
        ret += `  <div class="candy-header hide-less-500px"><span>發佈於</span><span class="green">${data.publish}</span></div>`;
        ret += `</div>`;
        ret += `
                    <div class="content">
                        ${newContent}
                    </div>
                    <div id="attachmentArea">
                        <ul>${attachment}</ul>
                    </div>
                    <div class="buttonArea" style="text-align: right;">
                        <button id="read" onclick="window.location='/news?id=${data.id}'"" class="border">閱讀</button>
                        <button id="delete" onclick="javascript:delete_what(this, 'news', ${data.id})" class="red">刪除</button>
                        <button id="publish" onclick="javascript:edit_news(${data.id})" class="blue">編輯</button>
                    </div>
                </div>
                `;
    });
    return ret;
}

function load(from, to, scope){
    return new Promise((resolve, reject) => {
        loadNews(scope, 'normal', from, to).then((data)=>{
            let content = artListRenderer(data.list);
            if (data.hasNext) {
                content += `<button style="margin:0px auto;" onclick="loadNext(${to + 1}, ${to + to - from + 1}, "${scope}", this)">More</button></div>`;
            }
            resolve(content);
        }).catch((e)=>{
            reject(e);
        });
    });
}

function loadNext(from, to, scope, obj) {
    load(from, to, scope).then((text) => {
        if (from === 0) {
            $("#article-parent").html(text);
        } else {
            $("#article-parent").append(text);
        }
    }).catch((e) => {
        console.log(e);
        $("#article-parent").html("Error " + e.status);
    });
    if (obj) {
        obj.remove();
    }
}

function resetBtnColor(id) {
    $('#btn-all').removeClass('blue-green');
    $('#btn-draft').removeClass('blue-green');
    $('#btn-published').removeClass('blue-green');
    $('#btn-' + id).addClass('blue-green');
}

function restoreTopEditorArea() {
    $('#top-editor-area').append($("#new-article-area"));
    $("#new-article-area").css('display', 'none');
}

function goToAll() {
    restoreTopEditorArea();
    resetBtnColor('all');
    loadNext(0, 19, 'all');
}

function goToDraft() {
    restoreTopEditorArea();
    resetBtnColor('draft');
    loadNext(0, 19, 'draft');
}

function goToPublished() {
    restoreTopEditorArea();
    resetBtnColor('published');
    loadNext(0, 19, 'published');
}

goToAll();
