window.directlyLeave = false;
window.onbeforeunload = function(){
    if(!leave_editing()){
        return 'Changes you made may not be saved.';
    }
    return;
};

var Editor = new (function(){
    this.editor = new SimpleMDE({
        placeholder: "Type here...",
        element: $("#new-article-area textarea")[0],
        spellChecker: false
    });
    this.serial = -1;

    this.defaultAttahmentVal = () =>{
        return {
            client_name: [],
            server_name: [],
            path :[]
        };
    }
    this.attachment = this.defaultAttahmentVal();

    // this.hash is not zero because of attachment json string.
    this.hash = hash(JSON.stringify(this.attachment));
})();

function leave_editing(){
    if(window.directlyLeave === true) return true;
    let now_content = $("#new-article-area #title").val() + Editor.editor.value() + JSON.stringify(Editor.attachment);
    if(hash(now_content) != Editor.hash){
        return confirm('Changes you made may not be saved.');
    }
    // default is 'leave'
    return true;
}

const btnAreaForDraft = `
<button id="attachment" onchange="javascript:attach(event)">Attachment
    <form enctype="multipart/form-data">
        <input type="file" multiple/>
    </form>
</button>
<button id="save" onclick="javascript:save(true)">Save (Draft)</button>
<button id="publish" onclick="javascript:publish()" class="blue">Publish</button>
`;

const btnAreaForPublished = `
<button id="attachment" onchange="javascript:attach(event)">Attachment<input type="file" multiple/></button>
<button id="save" onclick="javascript:save(false)" class="blue">Republish</button>
<button id="publish" onclick="javascript:delete_what(this, 'news')" class="red">Delete</button>
`

function newArticle(){
    if($("#new-article-area").css("display") === "none"){
        if(Editor.serial === -1){
            $.post('/function/add_news',{}
            ,function(data){
                console.log(data)
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
    $("#new-article-area").slideDown('fast', () => {
        if(typeof callback === 'function'){
            callback();
        }
    });
    $("#btn-add").text('Close');
}

function closeNewArticleArea(callback){
    $("#new-article-area").slideUp('fast', ()=>{
        if(typeof callback === 'function'){
            callback();
        }
    });
    $("#btn-add").text('Continue');
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
                    console.log(data);
                    console.log("成功");

                    for(let i=0; i<e.target.files.length; i++){
                        Editor.attachment.client_name.push(e.target.files[i].name);
                        Editor.attachment.server_name.push(data.Filename[i]);
                        Editor.attachment.path.push(data.Filepath[i]);

                        $("#new-article-area #attachmentArea ul").append(
                            `<li data-file-name="${data.Filename[i]}">
                                <a href="${data.Filepath[i]}">${e.target.files[i].name}</a>
                                <button class="red" onclick="delete_what(this, 'attachment', '${data.Filename[i]}')">delete</button>
                            </li>`
                        );
                    }
                    console.log(Editor);
                    $("#new-article-area #attachmentArea div.loader").css('display', 'none');
                }
            },
            dataType: 'json'
        });
    }, 500);
}

function save(isGotoDraft){
    $.post('/function/save_news',{
        serial: Editor.serial,
        title: $("#new-article-area #title").val(),
        content: Editor.editor.value(),
        attachment: JSON.stringify(Editor.attachment)
    },function(data){
        console.log(data)
        if(data['err']){
            if(data['code'] === 1){
                window.location = "/?notlogin"
            }
            notice(data['msg'])
        }else{
            // update editor hash
            Editor.hash = hash($("#new-article-area #title").val() + Editor.editor.value() + JSON.stringify(Editor.attachment));
            // refresh
            notice("Saved!");
            if(isGotoDraft){
                // goto draft (for news which is private, i.e. not published)
                goToDraft();
            }else{
                // goto news page (for news which have published already)
                window.directlyLeave = true;
                window.location = `/news?id=${Editor.serial}`;
            }
        }
    },'json');
}

function publish(){
    $.post('/function/publish_news',{
        serial: Editor.serial,
        title: $("#new-article-area #title").val(),
        content: Editor.editor.value(),
        attachment: JSON.stringify(Editor.attachment)
    }
    ,function(data){
        console.log(data)
        if(data['err']){
            if(data['code'] === 1){
                window.location = "/?notlogin"
            }
            notice(data['msg']);
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
        obj.text("SURE ?");
        if(typeof num === 'string'){
            obj.attr('onclick', `real_delete_${what}('${num}')`);
        }else{
            obj.attr('onclick', `real_delete_${what}(${num})`);
        }
        obj.fadeIn('slow', ()=>{
            setTimeout(()=>{
                obj.fadeOut('fast',()=>{
                    obj.text("Delete");
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
    // Find index of this file in Edirot.attachment.client_name(server_name, path)
    for(let i=0; i<Editor.attachment.server_name.length; i++){
        if(Editor.attachment.server_name[i] === filename){
            target = i;
            break;
        }
    }
    if(target < 0){
        notice('Error, the file you want to delete is not in the attachment list.');
        return;
    }

    Editor.attachment.server_name.splice(target, 1);
    Editor.attachment.client_name.splice(target, 1);
    Editor.attachment.path.splice(target, 1);

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
                    notice('Deleted!');
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
    if(leave_editing()){
        // 將原來收起來的區域展開
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
                console.log(data)
                if(data['err']){
                    if(data['code'] === 1){
                        window.location = '/?notlogin';
                    }
                    notice(data['msg']);
                }else{
                    // closeNewArticleArea();
                    // Step1: update value title, content, attachment and hash, serial
                    $("#new-article-area #title").val(data.Title);
                    Editor.editor.value(data.Content);
                    // Use default value (hash problem)
                    Editor.attachment = Editor.defaultAttahmentVal();
                    try{
                        Editor.attachment = JSON.parse(data.Attachment);
                        let parse = Editor.attachment;
                        let len = parse.client_name.length;
                        let attachment = "";
                        for(let i=0; i<len; i++){
                            attachment += `
                                <li data-file-name="${parse.server_name[i]}">
                                    <a href="${parse.path[i]}">${parse.client_name[i]}</a>
                                    <button class="red" onclick="delete_what(this, 'attachment', '${parse.server_name[i]}')">delete</button>
                                </li>
                            `;
                        }
                        $("#new-article-area #attachmentArea ul").html(attachment);
                    }catch(e){console.log(e)}
                    Editor.hash = hash(data.Title + data.Content + Editor.attachment);
                    Editor.serial = data_id;

                    // Step2: slideUp news area
                    $('.article[data-id="' + data_id + '"]').slideUp('fast',()=>{
                        // Step3: Move editor after news area
                        $('#new-article-area').insertAfter($('.article[data-id="' + data_id + '"]'));

                        // Step4: Rebuild buttonArea
                        if(data.Publish_time === 0){
                            $('#new-article-area .buttonArea').html(btnAreaForDraft);
                        }else{
                            $('#new-article-area .buttonArea').html(btnAreaForPublished);
                        }

                        // Step5: SlideDown editor and jump to there
                        openNewArticleArea(function(){
                            window.scrollTo({
                                top : $("#new-article-area")[0].offsetTop - $("nav")[0].clientHeight,
                                behavior: 'smooth'
                            })
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
    }
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

function reload_news(type){
    // Reload news
    let lnfw = new loadNewsForWhat('management', type, 0, 19);
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

function goToAll(){
    resetBtnColor('all');
    reload_news('all');
}

function goToDraft(){
    resetBtnColor('draft');
    reload_news('draft');
}

function goToPublished(){
    resetBtnColor('published');
    reload_news('public');
}

goToAll();
