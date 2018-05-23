$(function(){
    var uri = window.location.pathname;
    var path = uri.split('/')
    for (var i=0; i<path.length; i++) {        
        if (path[i] == ''){
            continue;
        }
        var name = path[i].substring(0,1).toUpperCase() + path[i].substring(1);
        if (i < path.length - 1) {
            $('ul.breadcrumb').append('<li class="breadcrumb-item"><a href="/' + path[i] + '">'+ name +'</a></li>')
        } else {
            $('ul.breadcrumb').append('<li class="breadcrumb-item active">'+ name +'</li>')
        }
    }
})