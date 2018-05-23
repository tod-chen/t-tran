var tag = {
    txtTranNum:'#tranNum',
    selTranType: '#tranType',
    btnQuery: '#btn-query',
    pages: '.pagination a',
    transBody: '#transBody',
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
        data:{tranNum:param.tranNum, tranType:param.tranType},
        dataType:'json',
        success: function(datas){
            $(tag.transBody).empty();
            for(var i=0; datas != null && i<datas.trans.length; i++){
                var t = datas.trans[i];
                var tranNumLink = '<a href="/trans/detail?tranNum=' + t.TranNum + '">' + t.TranNum + '</a>';
                var timetable = getTimeTableTd(t.RouteTimetable);                
                var cars = '';
                var level = t.EnableLevel;
                var start = new Date(t.EnableStartDate).toLocaleDateString();
                var end = new Date(t.EnableEndDate).toLocaleDateString();
                var html = '<tr><td>'+tranNumLink+'</td><td>'+timetable+'</td><td>'+cars+'</td><td>'+level
                    +'</td><td>'+start+'</td><td>'+end+'</td></tr>';
                $(tag.transBody).append(html);
            }
            $('[data-toggle="popover"]').popover({html : true});
        }
    })
}

function getTimeTableTd(RouteTimetable){
    var timetable = '';
    var routes = new Array();
    var rLen = RouteTimetable.length;
    var timetableTip = "<table class=\"table table-sm routeTable\"><thead><tr><th>序号</th><th>站名</th><th>到站时间</th><th>出发时间</th></tr></thead><tbody>";
    for(var j=0; j<rLen; j++){
        var arrTime = getDateHm(RouteTimetable[j].ArrTime);
        var depTime = getDateHm(RouteTimetable[j].DepTime);
        if (j == 0){
            arrTime = '--';
            timetable += RouteTimetable[j].StationName + ' --> ';
        }
        if (j == rLen - 1){
            depTime = '--';
            timetable += RouteTimetable[j].StationName;
        }
        timetableTip += '<tr><td>' + (j+1) + '</td><td>' + RouteTimetable[j].StationName + '</td><td>' + arrTime+ '</td><td>' +depTime+ '</td></tr>';
    }
    timetableTip += '</tbody></table>';
    return "<a href='#' title='' data-toggle='popover' data-trigger='focus' data-content='"+timetableTip+"'>"+timetable+"</a>";
}

function getDateHm(strDate){
    try{
        var d = new Date(strDate);
        var hour = d.getHours();
        var min = d.getMinutes();
        if (hour < 10) hour = '0' + hour;
        if (min < 10) min = '0' + min;
        return hour + ':' + min;
    }
    catch(err){
        return '--';
    }
}

function setCars(dtCar, carTip, cars){

}