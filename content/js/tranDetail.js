var tag = {
    tranId: '#tranId',
    tranNum: '#tranNum',
    durationDay:'#durationDay',
    scheduleDay:'#scheduleDay',
    enableLevel:'#enableLevel',
    enableSD:'#enableSD',
    enableED:'#enableED',
    timetableTab:'#timetable-tab',
    btnAddTimetable:'#btn-add-timetable',
    timetableContent:'#timetable-content',
    timetableItemTpl:'#timetable-item-tpl',
    carsContent:'#cars-content',
    btnAddCar:'#btn-add-car',
    carItemTpl:'#cars-item-tpl',
    selectableCars:'#selectable-cars',
    routePriceConent:'#route-price-conent',
    btnAddRoutePrice:'#btn-add-routePrice',
    routePriceItemTpl:'#route-price-item-tpl',
    btnSave:'#btn-save',
    btnCancel:'#btn-cancel'
}

var tranDetail = {

}

$(function(){
    initData();
    initEvent();
})

// 初始化数据
function initData(){
    var id = getQueryString('tranId');
    if (id != null) {
        $.ajax({
            url:'/trans/getDetail',
            type:'GET',
            dataType:'json',
            data:{tranId:id},
            success: function(result){
                if (result == null || result.tranInfo == null ){
                    return;
                }
                tranDetail = result.tranInfo;
                var t = result.tranInfo;
                // 设置基础信息
                $(tag.tranId).val(t.id);
                $(tag.tranNum).val(t.tranNum);
                $(tag.durationDay).val(t.durationDays);
                $(tag.scheduleDay).val(t.scheduleDays);
                $(tag.enableLevel).val(t.enableLevel);
                $(tag.enableSD).val(t.enableStartDate.split('T')[0]);
                $(tag.enableED).val(t.enableEndDate.split('T')[0]);
                // 设置时刻表
                for(var i=0;i<t.routeTimetable.length;i++){
                    var r = t.routeTimetable[i];
                    $(tag.btnAddTimetable).parent().before('<li><a>'+(i+1) + '.' +r.stationName+'<i></i></a></li>');
                    $(tag.timetableContent).append($(tag.timetableItemTpl).html());
                    var rEle = $(tag.timetableContent + ' div.tab-pane:eq('+i+')');
                    rEle.find('input[name=stationName]').val(r.stationName);
                    if (i == 0) r.arrTime = '';
                    rEle.find('input[name=arrTime]').val(getDateHm(r.arrTime));
                    if (i == t.routeTimetable.length - 1) r.depTime = '';
                    rEle.find('input[name=depTime]').val(getDateHm(r.depTime));
                    rEle.find('input[name=checkTicketGate]').val(r.checkTicketGate);
                    rEle.find('input[name=platform]').val(r.platform);
                    rEle.find('input[name=maleageNext]').val(r.maleageNext);
                }
                resetTimetable();
                // 设置车厢信息
                var carIdCountArr = t.carIds.split(';');
                var carIdCountMap = new Map();
                for(var i=0;i<carIdCountArr.length;i++){
                    var id_count = carIdCountArr[i].split('-');
                    carIdCountMap.set(parseInt(id_count[0]), parseInt(id_count[1]));
                }
                var tranType = t.tranNum.substr(0 ,1);
                var g = new RegExp('^[1-9]');
                // 以数字开头，车次类型设置为“其他”
                if (g.test(tranType)) {
                    tranType = 'O';
                }
                $.ajax({
                    url:'/cars/query',
                    type:'GET',
                    dataType:'json',
                    data:{tranType:tranType},
                    success: function(data){
                        if (data == null || data.cars == null){
                            return;
                        }
                        var seatTypes = new Set();
                        for(var j=0;j<data.cars.length;j++){
                            var c = data.cars[j];
                            seatTypes.add(c.SeatType);
                            var cName = '座位数/床位数:' + c.SeatCount + ', 站票数:' + c.NoSeatCount + ', ' + c.Remark;
                            var op = '<option value="'+c.ID+'" data-seatType="'+c.SeatType+'">'+cName+'</option>';
                            $(tag.selectableCars).append(op);
                        }
                        
                        var stEle = $(tag.carItemTpl).find('select[name=seat]');
                        seatTypes.forEach(function(key){
                            stEle.append('<option value="'+key+'">'+commonObj.seatTypeMap.get(key)+'</option>');
                        });
                        for(var i=0;i<t.Cars.length;i++){
                            var c = t.Cars[i];
                            $(tag.carsContent).append($(tag.carItemTpl).html());
                            var cEle = $(tag.carsContent + ' div.form-inline:eq('+i+')');
                            $(tag.selectableCars).children().each(function(){
                                if ($(this).data('seatType') == c.SeatType) {
                                    cEle.find('select[name=cardId]').append($(this)[0]);
                                }
                            });
                            cEle.find('select[name=carId]').val(c.ID);
                            cEle.find('input[name=carCount]').val(carIdCountMap.get(c.ID));
                        }
                        resetCars();
                    }
                })
                // 设置各路段价格
                var rt = t.routeTimetable;
                for(var i=0;i<rt.length-1;i++){
                    var itemHtml = '<div class="row pt15 pr15 pl15">';
                    itemHtml += '<div class="col-md-3 pr0"><b>路段：'+ rt[i].stationName + '&nbsp;->&nbsp;' + rt[i+1].stationName +'</b></div>';
                    for(var j=0;j<t.seatPriceMap.length;j++){

                    }
                    itemHtml += '</div>';
                    $(tag.routePriceConent).append(itemHtml);
                }
                for(var i=0;i<t.seatPriceMap.length;i++){
                    
                }
            }
        })
    } else {
        $(tag.tranNum).removeAttr('disabled');
    }
}

function resetTimetable(){
    $(tag.timetableTab + ' a').each(function(index, ele){
        var idx = index+1;
        if ($(ele).attr('id') == 'btn-add-timetable') return;        
        $(ele).attr('href', '#timetable-item-' + idx).attr('data-toggle', 'tab').addClass('nav-link');
        $(ele).find('i').remove();
        var station = $(ele).html().split('.');
        $(ele).html(idx+'.');
        if (station.length == 2) {
            $(ele).append(station[1])
        }
        $(ele).parent().addClass('nav-item');
        $(ele).append('<i class="fa fa-remove" title="删除"></i>');
    })
    $(tag.timetableContent + ' div.tab-pane').each(function(idx, ele){
        $(this).attr('id', 'timetable-item-' + (idx+1));
    })
    $(tag.timetableTab + ' a:first').click();
}

function resetCars(){

}

// 初始化事件
function initEvent(){
    $(tag.btnAddTimetable).click(function(){
        $(this).parent().before('<li><a><i></i></a></li>');
        $(tag.timetableContent).append($(tag.timetableItemTpl).html());
        resetTimetable();
        $(this).parent().prev().find('a').click();
    })
    $(tag.timetableContent).on('keyup', 'input[name=stationName]', function(){
        var id =  $(this).parents('.tab-pane').attr('id');
        var tabSelector = $(tag.timetableTab + ' a[href="#' + id + '"]');
        var val = $(this).val();
        tabSelector.find('i').remove();
        var station = tabSelector.html().split('.');
        tabSelector.html(station[0] + '.' + val + '<i class="fa fa-remove" title="删除"></i>');
    })
    $(tag.timetableTab).on('click', 'i.fa-remove',function(){
        var sel = $(this).parent().attr('href');
        $(sel).remove();
        $(this).parent().remove();
        resetTimetable();
    })
    $(tag.btnAddCar).click(function(){

    })
    $(tag.carsContent).on('change', 'input[name=carCount]', function(){

    })
    $(tag.carsContent).on('click', 'button.btn', function(){

    })
    $(tag.btnSave).click(function(){

    })
    $(tag.btnCancel).click(function(){

    })
}

// 验证数据有效性
function validData(){

}

// 保存
function save(){

}