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
