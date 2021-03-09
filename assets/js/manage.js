window.directlyLeave = false;
window.onbeforeunload = function(){
    if(!leave_editing()){
        return '變更尚未儲存 (若確定已儲存請忽略此訊息)';
    }
    return;
};

var Editor = new (function(){
    let self = this;
    this.serial = -1;
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
        if(self.editor){
            self.editor.destroy();
            self.editor = null;
        }
    }

    this.getEditorValue = () => {
        return (self.editor)? self.editor.getData() : '';
    }
    this.setEditorValue = (val) => {
        if(self.editor){
            self.editor.setData(val);
        }
    }

    this.attachment = this.defaultAttahmentVal();
    this.getType = () => {
        return $("#new-article-area #type").val();
    }
    this.getTitle = () =>{
        return $("#new-article-area #title").val();
    }
    // this.hash is not zero because of attachment json string.
    this.preHash = hash(this.getType() + JSON.stringify(this.attachment));

    this.calcCurrentHash = () =>{
        return hash(self.getTitle() + self.getType() + self.getEditorValue() + JSON.stringify(self.attachment));
    }
})();

function leave_editing(){
    if(window.directlyLeave === true) return true;

    if(Editor.calcCurrentHash() != Editor.preHash){
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

function newArticle(){
    if($("#new-article-area").css("display") === "none"){
        if(Editor.serial === -1){
            $.post('/function/add_news',{}
            ,function(data){
                if(data['err']){
                    if(data['code'] === 1){
                        window.location = "/?notlogin"
                    }
                    notice(data['msg'])
                }else{
                    $("#new-article-area .buttonArea").html(btnAreaForDraft);
                    Editor.serial = Number(data['msg']);
                    openNewArticleArea();
                }
            },'json');
        }else{
            openNewArticleArea();
        }
    }else{
        closeNewArticleArea();
    }
}

function openNewArticleArea(callback){
    Editor.destroyCKEeditor();
    Editor.newCKEeditor();
    $("#new-article-area").slideDown('fast', () => {
        if(typeof callback === 'function'){
            callback();
        }
    });
    $("#btn-add").text('收起');
}

function closeNewArticleArea(callback){
    $("#new-article-area").slideUp('fast', ()=>{
        if(typeof callback === 'function'){
            callback();
        }
    });
    $("#btn-add").text('展開');
}

function attach(e){
    let form = new FormData();
    console.log(e.target.files);
    for(let i=0; i<e.target.files.length; i++){
        form.append("files", e.target.files[i]);
    }

    $("#new-article-area #attachmentArea div.loader").css('display', 'block');
    setTimeout(function(){
        $.ajax({
            url: '/function/upload',
            processData: false,
            contentType: false,
            mimeType: 'multipart/form-data',
            data: form,
            type: 'POST',
            success: function(data){
                if(data['Err']){
                    if(data['Code'] === 1){
                        window.location = '/?notlogin';
                    }
                    $("#new-article-area #attachmentArea").html('<span class="error">' + data['Msg'] + '</span>');
                    notice(data['Msg']);
                }else{
                    for(let i=0; i<e.target.files.length; i++){
                        Editor.attachment.push({
                            "client_name": e.target.files[i].name,
                            "server_name": data[i].Filename,
                            "path": data[i].Filepath
                        });

                        $("#new-article-area #attachmentArea ul").append(
                            `<li data-file-name="${data[i].Filename}">
                                <a href="${data[i].Filepath}">${e.target.files[i].name}</a>
                                <button class="red" onclick="delete_what(this, 'attachment', '${data[i].Filename}')">刪除</button>
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

/*
function uploadPic(e){
    let form = new FormData();

    for(let i=0; i<e.target.files.length; i++){
        form.append("files", e.target.files[i]);
    }

    $.ajax({
        url: '/function/upload',
        processData: false,
        contentType: false,
        mimeType: 'multipart/form-data',
        data: form,
        type: 'POST',
        success: function(data){
            if(data['Err']){
                if(data['Code'] === 1){
                    window.location = '/?notlogin';
                }
                $("#new-article-area #attachmentArea").html('<span class="error">' + data['Msg'] + '</span>');
                notice(data['Msg']);
                console.log(data);
            }else{
                let code = `![](${data.Filepath[0]})`;
                $("#new-article-area #return-url").slideDown();
                $("#new-article-area #img-upload-area").slideUp();
                $("#new-article-area #return-url input").val(code);
            }
        },
        dataType: 'json'
    });
}
*/

function save(isPrivate){
    $.post('/function/save_news',{
        serial: Editor.serial,
        title: Editor.getTitle(),
        type: Editor.getType(),
        content: Editor.getEditorValue(),
        attachment: JSON.stringify(Editor.attachment)
    },function(data){
        if(data['err']){
            if(data['code'] === 1){
                window.location = "/?notlogin";
            }
            notice(data['msg']);
        }else{
            if(isPrivate){
                Editor.preHash = Editor.calcCurrentHash();
                notice("Saved!");
            }else{
                // goto news page (for news which have been published already)
                window.directlyLeave = true;
                window.location = `/news?id=${Editor.serial}`;
            }
        }
    },'json');
}

function publish(){
    $.post('/function/publish_news',{
        serial: Editor.serial,
        title: Editor.getTitle(),
        type: Editor.getType(),
        content: Editor.getEditorValue(),
        attachment: JSON.stringify(Editor.attachment)
    }
    ,function(data){
        if(data['err']){
            if(data['code'] === 1){
                window.location = "/?notlogin"
            }
            notice(data['msg']);
            console.log(data);
        }else{
            // directly go to news page
            window.directlyLeave = true;
            window.location = `/news?id=${Editor.serial}`;
        }
    },'json');
}

function delete_what(obj, what, num){
    if(num === undefined){
        num = Editor.serial;
    }
    obj = $(obj);
    obj.fadeOut('fast', ()=>{
        obj.text("確定？");
        if(typeof num === 'string'){
            obj.attr('onclick', `real_delete_${what}('${num}')`);
        }else{
            obj.attr('onclick', `real_delete_${what}(${num})`);
        }
        obj.fadeIn('slow', ()=>{
            setTimeout(()=>{
                obj.fadeOut('fast',()=>{
                    obj.text("刪除");
                    if(typeof num === 'string'){
                        obj.attr('onclick', `delete_what(this, '${what}', '${num}')`);
                    }else{
                        obj.attr('onclick', `delete_what(this, '${what}', ${num})`);
                    }

                    obj.fadeIn('slow');
                });
            }, 3000);
        }, )
    });
}

function real_delete_attachment(filename){
    let target = -1;
    // Generate new attachment JSON
    // Find index of this file in Editor.attachment.client_name(server_name, path)
    for(let i=0; i<Editor.attachment.length; i++){
        if(Editor.attachment[i].server_name === filename){
            target = i;
            break;
        }
    }
    if(target < 0){
        notice('錯誤，欲刪除之檔案不在附加檔案列表中');
        return;
    }

    Editor.attachment.splice(target, 1);

    $.ajax({
        url: '/function/del_attachment',
        data: {
            'server_name' : filename,
            'serial_num': Editor.serial,
            'new_attachment': JSON.stringify(Editor.attachment)
        },
        type: 'POST',
        success: function(data){
            if(data['err']){
                if(data['code'] === 1){
                    window.location = '/?notlogin';
                }
                notice(data['msg']);
            }else{
                $('#new-article-area #attachmentArea ul li[data-file-name="' + filename + '"]').slideUp('slow');

            }
        },
        error: function(err) {
            console.log(err);
            notice('Error');
        },
        dataType: 'json'
    });
}

function real_delete_news(data_id){
    /*
        If user delete the news which is editing
        Refresh website instead of .slideUp()
    */

    $.ajax({
        url: '/function/del_news',
        data: {
            'serial' : data_id
        },
        type: 'POST',
        success: function(data){
            if(data['err']){
                if(data['code'] === 1){
                    window.location = '/?notlogin';
                }
                notice(data['msg']);
            }else{
                if(data_id === Editor.serial){
                    window.directlyLeave = true;
                    location.reload();
                }else{
                    $('.article[data-id="' + data_id + '"]').slideUp('slow');
                    // close news area
                    notice('刪除成功！');
                }
            }
        },
        error: function(err) {
            console.log(err);
            notice('Error');
        },
        dataType: 'json'
    });
}

function edit_news(data_id){
    //if(leave_editing()){
        // 將原來收起來的區域展開
        openNewArticleArea();
        if($('.article[data-id="' + Editor.serial + '"]').length){
            $('.article[data-id="' + Editor.serial + '"]').slideDown();
        }

        $.ajax({
            url: '/function/get_news',
            data: {
                'id' : data_id
            },
            type: 'GET',
            success: function(data){
                console.log(data);
                if(data['err']){
                    if(data['code'] === 1){
                        window.location = '/?notlogin';
                    }
                    notice(data['msg']);
                    console.log(data);
                }else{
                    // Step1: update value title, type, content, attachment and hash, serial
                    $("#new-article-area #title").val(data.Title);
                    for(let i=0; i<$("#new-article-area #type option").length; i++){
                        $("#new-article-area #type option").eq(i).removeAttr('selected');
                    }
                    $("#new-article-area #type option[value = '" + data.Type + "']").attr('selected', 'selected');

                    Editor.attachment = Editor.defaultAttahmentVal();
                    try{
                        Editor.attachment = data.Attachment;
                        let attachment = '';
                        for(let i=0; i < Editor.attachment.length; i++){
                            attachment += `
                                <li data-file-name="${Editor.attachment[i].server_name}">
                                    <a href="${Editor.attachment[i].path}">${Editor.attachment[i].client_name}</a>
                                    <button class="red" onclick="delete_what(this, 'attachment', '${Editor.attachment[i].server_name}')">刪除</button>
                                </li>
                            `;
                        }
                        $("#new-article-area #attachmentArea ul").html(attachment);
                    }catch(e){
                        console.log(e);
                    }
                    Editor.serial = data_id;

                    // Step2: slideUp news area
                    $('.article[data-id="' + data_id + '"]').slideUp('fast',()=>{
                        // Step3: Move editor after news area
                        $('#new-article-area').insertAfter($('.article[data-id="' + data_id + '"]'));

                        // Step4: Rebuild buttonArea
                        if(data.PublishTime === 0){
                            $('#new-article-area .buttonArea').html(btnAreaForDraft);
                        }else{
                            $('#new-article-area .buttonArea').html(btnAreaForPublished);
                        }

                        // Step5: SlideDown editor and jump to there
                        openNewArticleArea(function(){
                            window.scrollTo({
                                top : $("#new-article-area")[0].offsetTop - $("nav")[0].clientHeight,
                                behavior: 'smooth'
                            });
                            Editor.setEditorValue(data.Content);
                            Editor.preHash = Editor.calcCurrentHash();
                        });
                    });
                }
            },
            error: function(err) {
                console.log(err);
                notice('Error');
            },
            dataType: 'json'
        });
    //}
    // DO NOTHING
}

function loadNext(obj){
    window.lnfw.next().then((data) => {
        $("#article-parent").append(data);
        obj.remove();
    }).catch((reason) => {
        console.log(reason);
        notice("Error " + reason.status);
    })
}

function reload_news(scope){
    // Reload news
    let lnfw = new loadNewsForWhat('management', scope, 'normal', 0, 19);
    lnfw.load().then((data) => {
        $("#article-parent").html(data);
    }).catch((reason) => {
        console.log(reason);
        $("#article-parent").html("Error " + reason.status);
    });

    window.lnfw = lnfw;
}

function resetBtnColor(id){
    $('#btn-all').removeClass('blue-green');
    $('#btn-draft').removeClass('blue-green');
    $('#btn-published').removeClass('blue-green');
    $('#btn-'+id).addClass('blue-green');
}

function restoreTopEditorArea(){
    $('#top-editor-area').append($("#new-article-area"));
    $("#new-article-area").css('display', 'none');
}

function goToAll(){
    restoreTopEditorArea();
    resetBtnColor('all');
    reload_news('all');
}

function goToDraft(){
    restoreTopEditorArea();
    resetBtnColor('draft');
    reload_news('draft');
}

function goToPublished(){
    restoreTopEditorArea();
    resetBtnColor('published');
    reload_news('published');
}

goToAll();
