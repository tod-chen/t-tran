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
    query();
    $(tag.btnQuery).click(function(){
        param.departureDate = $(tag.txtDepartureDate).val();
        param.tranNum = $(tag.txtTranNum).val();
        param.page = 1;
        query();
    })
})

function query(){
    $.ajax({
        url:'/schedules/query',
        type:'GET',
        data:{depDate:param.departureDate, tranNum:param.tranNum, page: param.page},
        dataType:'json',
        success: function(result){
            $(tag.tableBody).empty();
            for(var i=0; result != null && i<result.schedules.length; i++){
                var s = result.schedules[i];
                var scheduleLink = '<a href="/schedules/detail?scheduleID=' + s.ID + '"><i class="fa fa-edit"></i></a>';
                var tr = '<tr><td>'+scheduleLink+'</td><td>'+s.DepartureDate+'</td><td>'+s.TranNum+'</td><td>'+s.SaleTicketTime
                    +'</td><td>'+s.NotSaleRemak+'</td></tr>';
                $(tag.tableBody).append(tr);
            }
            setPage(result.count, result.ps, param.page, function(page){
                param.page = page;
                query();
            })
        }
    })
}
