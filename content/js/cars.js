var tag = {
    selSeatType:'#seatType',
    selTranType: '#tranType',
    btnQuery: '#btn-query',
    btnAdd: '#btn-add',
    tableBody: '#carsBody',
}

var param = {
    seatType: '',
    tranType: '',
    page:1
}

$(function(){
    commonObj.tranTypeMap.forEach(function(value, key){
        $(tag.selTranType).append('<option value="'+key+'">'+value+'</option>');
    })
    commonObj.seatTypeMap.forEach(function(value, key){
        $(tag.selSeatType).append('<option value="'+key+'">'+value+'</option>');
    })
    query();
    $(tag.btnQuery).click(function(){
        param.seatType = $(tag.selSeatType).val();
        param.tranType = $(tag.selTranType).val();
        param.page = 1;
        query();
    })
    
})

function query(){
    $.ajax({
        url:'/admin/cars/query',
        type:'GET',
        data:{seatType:param.seatType, tranType:param.tranType, page: param.page},
        dataType:'json',
        success: function(result){
            $(tag.tableBody).empty();
            for(var i=0; result != null && i<result.cars.length; i++){
                var c = result.cars[i];
                var carLink = '<a href="/admin/cars/detail?carId=' + c.id + '" target="_black"><i class="fa fa-edit"></i></a>';
                var tranType = commonObj.tranTypeMap.get(c.tranType);
                var seatType = commonObj.seatTypeMap.get(c.seatType);
                var tr = '<tr><td>'+carLink+'</td><td>'+tranType+'</td><td>'+seatType+'</td><td>'+c.SeatCount
                    +'</td><td>'+c.noSeatCount+'</td><td>'+c.remark+'</td></tr>';
                $(tag.tableBody).append(tr);
            }
            setPage(result.count, result.ps, param.page, function(page){
                param.page = page;
                query();
            })
        }
    })
}