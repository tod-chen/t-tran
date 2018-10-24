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
    btnSetRoutePrice:'#btn-reset-routeprice',
    routePriceTableBody:'#routePriceTableBody',
    btnSave:'#btn-save',
    btnCancel:'#btn-cancel'
}

$(function(){
    initData();
    initEvent();
})

// 初始化数据
function initData(){
    var id = getQueryString('tranId');
    if (id == null){
        $(tag.tranNum).removeAttr('disabled');
        return;
    }
    $.ajax({
        url:'/trans/getDetail',
        type:'GET',
        dataType:'json',
        data:{tranId:id},
        success: function(result){
            if (result == null || result.tranInfo == null ){
                return;
            }
            var t = result.tranInfo;
            // 设置基础信息
            $(tag.tranId).val(t.id);
            $(tag.tranNum).val(t.tranNum);
            $(tag.durationDay).val(t.durationDays);
            $(tag.scheduleDay).val(t.scheduleDays);
            $(tag.enableLevel).val(t.enableLevel);
            $(tag.enableSD).val(t.enableStartDate.split('T')[0]);
            $(tag.enableED).val(t.enableEndDate.split('T')[0]);
            // 设置车厢
            setCars(t.tranNum, t.carIds);
            // 设置时刻表
            for(var i=0;i<t.timetable.length;i++){
                var r = t.timetable[i];
                $(tag.btnAddTimetable).parent().before('<li><a><span class="timetable-idx"></span><span class="timetable-stationName">' +r.stationName+'</span><i></i></a></li>');
                $(tag.timetableContent).append($(tag.timetableItemTpl).html());
                var rEle = $(tag.timetableContent + ' div.tab-pane:eq('+i+')');
                rEle.find('input[name=stationName]').val(r.stationName);
                if (i == 0) r.arrTime = '';
                rEle.find('input[name=arrTime]').val(getDateHm(r.arrTime));
                if (i == t.timetable.length - 1) r.depTime = '';
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
            // 设置各路段价格
            var rt = t.timetable;
            for(var i=0;i<rt.length-1;i++){
                var tr = '<tr data-routeIdx="'+i+'"><td class="route">'+ rt[i].stationName + '&nbsp;->&nbsp;' + rt[i+1].stationName +'</td></tr>';
                $(tag.routePriceTableBody).append(tr);
            }
            for(var i=0;i<t.seatPriceMap.length;i++){
                var seatName = commonObj.seatTypeMap.get(t.seatPriceMap[i].key);
                $('th.route').parent().append('<th data-seatType="'+t.seatPriceMap[i].key+'">' + seatName + '</th>');
                for(var j=0;j<t.seatPriceMap[i].value.length;j++){
                    $('td[data-routeIdx="'+j+'"]').append('<td data-seatType="'+t.seatPriceMap[i].key
                    +'"><input type="number" value="'+t.seatPriceMap[i].value[j]+'" /></td>');
                }
            }
        }
    })
}

function resetTimetable(){
    $(tag.timetableTab + ' a').each(function(index, ele){
        var idx = index+1;
        if ($(ele).attr('id') == 'btn-add-timetable') return;        
        $(ele).attr('href', '#timetable-item-' + idx).attr('data-toggle', 'tab').addClass('nav-link');
        $(ele).find('i').attr({'class':'fa fa-remove', 'title':'删除'});
        $(ele).parent().addClass('nav-item');
    });
    $(tag.timetableTab + ' .timetable-idx').each(function(index, ele){
        $(ele).html((index+1)+'-');
    })
    $(tag.timetableContent + ' div.tab-pane').each(function(idx, ele){
        $(this).attr('id', 'timetable-item-' + (idx+1));
    })
    $(tag.timetableTab + ' a:first').click();
}

function setCars(tranNum, carIds){
    if (tranNum.length < 1) return;
    
    var tranType = tranNum.substr(0 ,1);
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
            var stEle = $(tag.carItemTpl).find('select[name=carId]');
            stEle.empty();
            stEle.data('tranType', tranType);
            for(var j=0;j<data.cars.length;j++){
                var c = data.cars[j];
                var cName = c.remark + ', 座位数/床位数:' + c.SeatCount + ', 站票数:' + c.noSeatCount;
                var op = '<option value="'+c.id+'" data-seatType="'+c.seatType+'">'+cName+'</option>';
                stEle.append(op);
            }
            var ids = carIds.split(';');
            for(var i=0;i<ids.length;i++){
                $(tag.carsContent + ' .cb').before($(tag.carItemTpl).html());
                var ic = ids[i].split('-');
                var carEle = $(tag.carsContent + ' .car-item:eq('+i+')');
                carEle.find('[name="carID"]').val(ic[0]);
                carEle.find('[name="carCount"]').val(ic[1]);
            }
            setCarSumInfo();
        }
    })
}

function setCarSumInfo(){
    var count = 0;
    $(tag.carsContent + ' input[name="carCount"]').each(function(){
        var c = parseInt($(this).val());
        if (!isNaN(c)) count += c;
    })
    $(tag.btnAddCar).prev().html('车厢数：' + count);
}

// 初始化事件
function initEvent(){
    $(tag.btnAddTimetable).click(function(){
        $(this).parent().before('<li><a><span class="timetable-idx"></span><span class="timetable-stationName"></span><i></i></a></li>');
        $(tag.timetableContent).append($(tag.timetableItemTpl).html());
        resetTimetable();
        $(this).parent().prev().find('a').click();
    })
    $(tag.timetableContent).on('keyup', 'input[name=stationName]', function(){
        var id =  $(this).parents('.tab-pane').attr('id');
        $(tag.timetableTab + ' a[href="#' + id + '"] .timetable-stationName').html($(this).val());
    })
    $(tag.timetableTab).on('click', 'i.fa-remove',function(){
        var sel = $(this).parent().attr('href');
        $(sel).remove();
        $(this).parent().remove();
        resetTimetable();
    })
    $(tag.btnAddCar).click(function(){
        $(tag.carsContent + ' .cb').before($(tag.carItemTpl).html());
        setCarSumInfo();
    })
    $(tag.carsContent).on('change', 'input[name=carCount]', setCarSumInfo)
    $(tag.carsContent).on('click', 'i', function(){
        $(this).parents('.car-item').remove();
        setCarSumInfo();
    })
    $(tag.btnSetRoutePrice).click(function(){
        var timetables = new Array();
        $(tag.timetableTab + ' .timetable-stationName').each(function(idx, ele){
            var stationName = $(ele).html();
            if (stationName != ''){
                timetables.push(stationName);
            }
        })
        var seatSet = new Set();
        $(tag.carsContent + ' select').each(function(){
            var val = $(this).val();
            $(this).children().each(function(){
                if ($(this).attr('value') == val){
                    seatSet.add($(this).data('seattype'));
                }
            });
        });
        $('th.route').nextAll().remove();
        $(tag.routePriceTableBody).empty();
        seatSet.forEach(function(val){
            $('th.route').parent().append('<th data-seatType="'+val+'">' + commonObj.seatTypeMap.get(val) + '</th>');
        })
        for(var i=1;i<timetables.length;i++){
            var from = timetables[i-1];
            var to = timetables[i];
            var tr = '<tr data-routeIdx="'+i+'"><td class="route">' + from + '&nbsp;->&nbsp;' + to + '</td>';
            seatSet.forEach(function(val){
                tr += '<td data-seatType="'+val+'"><input type="number" />';
            })
            tr += '</tr>';
            $(tag.routePriceTableBody).append(tr);
        }
    })
    $(tag.btnSave).click(save)
    $(tag.btnCancel).click(function(){

    })
}

// 验证数据有效性
function validData(){
    var data = {
        id : parseInt($(tag.tranId).val()),
        tranNum:$(tag.tranNum).val(),
        durationDay:parseInt($(tag.durationDay).val()),
        scheduleDay:parseInt($(tag.scheduleDay).val()),
        enableLevel:parseInt($(tag.enableLevel).val()),
        enableStartDate: new Date($(tag.enableSD).val()).toISOString(),
        enableEndDate: new Date($(tag.enableED).val()).toISOString(),
        timetable: new Array(),
        carIds:'',
        seatPriceMap:new Map()
    };
    $(tag.timetableContent + '>div').each(function(){
        var route = {
            stationName:$(this).find('input[name="stationName"]').val(),
            arrTime:$(this).find('input[name="arrTime"]').val(),
            depTime:$(this).find('input[name="depTime"]').val(),
            checkTicketGate:$(this).find('input[name="checkTicketGate"]').val(),
            platform: parseInt($(this).find('input[name="platform"]').val()),
            maleageNext: parseFloat($(this).find('input[name="maleageNext"]').val()),
        };
        if (isNaN(route.platform)) route.platform = 0;
        if (isNaN(route.maleageNext)) route.maleageNext = 0;
        if (route.arrTime == '') route.arrTime = '00:00';
        if (route.depTime == '') route.depTime = '00:00';
        route.arrTime = new Date('1971-01-01 ' + route.arrTime + ':00').toISOString();
        route.depTime = new Date('1971-01-01 ' + route.depTime + ':00').toISOString();
        data.timetable.push(route);
    });
    var cars = new Array();
    $(tag.carsContent + ' .car-item').each(function(){
        var carId = $(this).find('select[name="carId"]').val();
        var carCount = parseInt($.trim($(this).find('input[name="carCount"]').val()));
        if (isNaN(carCount)){
            carCount = 1;
        }
        cars.push(carId + ':' + carCount);
    })
    data.carIds = cars.join(';');
    var seatTypes = new Array();
    $('th.route').nextAll().each(function(){
        var st = $(this).data('seattype');
        seatTypes.push(st);
        data.seatPriceMap.set(st, new Array());
    });
    $(tag.routePriceTableBody + '>tr').each(function(){
        for(var i=0; i<seatTypes.length; i++){
            var st = seatTypes[i];
            var price = $(this).find('td[data-seattype="' + st + '"] input[type="number"]').val();
            price = parseFloat($.trim(price));
            if (isNaN(price)){
                price = 1;
            }
            var arr = data.seatPriceMap.get(st);
            arr.push(price);
            data.seatPriceMap.set(st, arr);
        }
    });
    data.seatPriceMap = _strMapToObj(data.seatPriceMap);
    return data;
}
function  _strMapToObj(strMap){
    let obj= Object.create(null);
    for (let[k,v] of strMap) {
      obj[k] = v;
    }
    return obj;
}
// 保存
function save(){
    var data = validData();
    if (data == null){
        return;
    }
    $.ajax({
        url:'/tran/save',
        type:'POST',
        dataType:'json',
        data:JSON.stringify(data),
        success:function(result){
            if (result.success){
                toastr.success('保存成功');
            } else {
                toastr.error(result.msg);
            }
        }
    })
}