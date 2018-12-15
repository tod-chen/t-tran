var tag = {
    txtDepartureDate:'#departureDate',
    txtTranNum:'#tranNum',
    btnQuery: '#btn-query',
    btnAdd: '#btn-add',
    tableBody: '#schedulesBody',
}

var param = {
    departureDate:'',
    tranNum:'',
    page:1
}

$(function(){
    $(tag.txtDepartureDate).val(new Date().format());
    $(tag.btnQuery).click(function(){
        var date = new Date($(tag.txtDepartureDate).val());
        if (isNaN(date)){
            date = new Date();
        }
        param.departureDate = date.format();
        param.tranNum = $(tag.txtTranNum).val();
        param.page = 1;
        query();
    })
    $(tag.btnQuery).click();
})

function query(){
    $.ajax({
        url:'/admin/schedules/query',
        type:'GET',
        data:{depDate:param.departureDate, tranNum:param.tranNum, page: param.page},
        dataType:'json',
        success: function(result){
            $(tag.tableBody).empty();
            for(var i=0; result != null && i<result.schedules.length; i++){
                var s = result.schedules[i];
                var scheduleLink = '<a href="/admin/schedules/detail?scheduleID=' + s.id + '"><i class="fa fa-edit"></i></a>';
                var tr = '<tr><td>'+scheduleLink+'</td><td>'+s.departureDate+'</td><td>'+s.tranNum+'</td><td>'+s.saleTicketTime
                    +'</td><td>'+s.notSaleRemark+'</td></tr>';
                $(tag.tableBody).append(tr);
            }
            setPage(result.count, result.ps, param.page, function(page){
                param.page = page;
                query();
            })
        }
    })
}
