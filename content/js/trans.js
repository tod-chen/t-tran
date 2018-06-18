var tag = {
    txtTranNum:'#tranNum',
    selTranType: '#tranType',
    btnQuery: '#btn-query',
    btnAdd: '#btn-add',
    tableBody: '#transBody',
}

var param = {
    tranNum:'',
    tranType: '',
    page:1
}

$(function(){
    query();
    $(tag.btnQuery).click(function(){
        param.tranNum = $(tag.txtTranNum).val();
        param.tranType = $(tag.selTranType).val();
        param.page = 1;
        query();
    })
    
})

function query(){
    $.ajax({
        url:'/trans/query',
        type:'GET',
        data:{tranNum:param.tranNum, tranType:param.tranType, page: param.page},
        dataType:'json',
        success: function(result){
            $(tag.tableBody).empty();
            for(var i=0; result != null && i<result.trans.length; i++){
                var t = result.trans[i];
                var tranNumLink = '<a href="/trans/detail?tranId=' + t.id + '">' + t.tranNum + '</a>';
                var timetable = getTimeTableTd(t.routeTimetable);                
                var cars = '';
                var level = t.enableLevel;
                var start = new Date(t.enableStartDate).toLocaleDateString();
                var end = new Date(t.enableEndDate).toLocaleDateString();
                var tr = '<tr><td>'+tranNumLink+'</td><td>'+timetable+'</td><td>'+cars+'</td><td>'+level
                    +'</td><td>'+start+'</td><td>'+end+'</td></tr>';
                $(tag.tableBody).append(tr);
            }
            $('[data-toggle="popover"]').popover({html : true});
            setPage(result.count, result.ps, param.page, function(page){
                param.page = page;
                query();
            })
        }
    })
}

function getTimeTableTd(RouteTimetable){
    var timetable = '';
    var routes = new Array();
    var rLen = RouteTimetable.length;
    var timetableTip = "<table class=\"table table-sm routeTable\"><thead><tr><th>序号</th><th>站名</th><th>到站时间</th><th>出发时间</th></tr></thead><tbody>";
    for(var j=0; j<rLen; j++){
        var arrTime = getDateHm(RouteTimetable[j].arrTime);
        var depTime = getDateHm(RouteTimetable[j].depTime);
        if (j == 0){
            arrTime = '--';
            timetable += RouteTimetable[j].stationName + ' --> ';
        }
        if (j == rLen - 1){
            depTime = '--';
            timetable += RouteTimetable[j].stationName;
        }
        timetableTip += '<tr><td>' + (j+1) + '</td><td>' + RouteTimetable[j].stationName + '</td><td>' + arrTime+ '</td><td>' +depTime+ '</td></tr>';
    }
    timetableTip += '</tbody></table>';
    return "<a href='#' title='' data-toggle='popover' data-trigger='focus' data-content='"+timetableTip+"'>"+timetable+"</a>";
}

function setCars(dtCar, carTip, cars){

}