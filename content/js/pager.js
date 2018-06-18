var pageTag = {    
    pageUl:'.pagination',
    pageItem: '.pagination li.page-other',
    pageTurningFunc: ''
}

$(function(){
    // 设置翻页事件
    $(pageTag.pageUl).on("click", 'li.page-item', function(){
        if (typeof(pageTag.pageTurningFunc) === 'function'){
            pageTag.pageTurningFunc(parseInt($(this).data('page')));
        }
    })
})

// 设置分页
function setPage(count, pageSize, page, pageTurningFunc){
    pageTag.pageTurningFunc = pageTurningFunc;
    var pageCount = Math.ceil(count / pageSize);
    $(pageTag.pageUl).empty();    
    var prev = page - 1;
    if (prev < 1) prev = 1;
    var next = page + 1;
    if (next > pageCount) next = pageCount;
    $(pageTag.pageUl).append('<li class="page-item page-other" data-page="1"><a class="page-link" href="#">首页</a></li>');
    $(pageTag.pageUl).append('<li class="page-item page-other" data-page="'+prev+'"><a class="page-link" href="#">上一页</a></li>');
    var nearCount = 4;    
    for (var i=page - nearCount; i<=page + nearCount && i<= pageCount; i++) {
        if (i < 1){
            i = 1;
        }
        var active = '';
        if (i == page){
            active = ' active';
        }
        $(pageTag.pageUl).append('<li class="page-item' + active + '" data-page="' + i
            +'"><a class="page-link" href="#">' + i +'</a></li>');
    }
    $(pageTag.pageUl).append('<li class="page-item page-other" data-page="'+ next +'"><a class="page-link" href="#">下一页</a></li>');
    $(pageTag.pageUl).append('<li class="page-item page-other" data-page="'+ pageCount +'"><a class="page-link" href="#">尾页</a></li>');
    $(pageTag.pageItem).each(function(){
        var dp = parseInt($(this).data('page'));
        if(dp == page){
            $(this).addClass('disabled');
        }
    })
}