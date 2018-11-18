var tag = {
    txtStationName:'#stationName',
    txtCityName:'#cityName',
    btnQuery: '#btn-query',
    btnAdd: '#btn-add',
    tableBody: '#stationsBody',
}

var param = {
    stationName:'',
    cityName:'',
    page:1
}

$(function(){
    query();
    $(tag.btnQuery).click(function(){
        param.stationName = $(tag.txtStationName).val();
        param.cityName = $(tag.txtCityName).val();
        param.page = 1;
        query();
    })
    
})

function query(){
    $.ajax({
        url:'/stations/query',
        type:'GET',
        data:{stationName:param.stationName, cityName:param.cityName, page: param.page},
        dataType:'json',
        success: function(result){
            $(tag.tableBody).empty();
            for(var i=0; result != null && i<result.stations.length; i++){
                var s = result.stations[i];
                var stationLink = '<a href="/admin/stations/detail?stationID=' + s.ID + '">' + s.StationName + '</a>';
                var passenger = s.IsPassenger == 1 ? '是' : '否';
                var tr = '<tr><td>'+stationLink+'</td><td>'+s.StationCode+'</td><td>'+passenger+'</td><td>'+s.CityCode
                    +'</td><td>'+s.CityName+'</td></tr>';
                $(tag.tableBody).append(tr);
            }
            setPage(result.count, result.ps, param.page, function(page){
                param.page = page;
                query();
            })
        }
    })
}
