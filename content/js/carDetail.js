var tag = {
    carId : '#carId',
    tranType:'#tranType',
    seatType:'#seatType',
    noSeatCount:'#noSeatCount',
    remark:'#remark',
    seatCount:'#seatCount',
    studentCount:'#studentCount',
    seatContent:'#seat-content',
    seats:'#seat-content div.seat-item',
    txtRowCount:'#seatRowCount',
    txtSeatCountInRow:'#rowSeatCount',
    txtSeatNumSuffix:'#seatNumSuffix',
    btnAddSeat:'#btn-add-seat',
    seatItemTpl:'#seat-item-tpl',
    btnSave:'#btn-save',
    btnCancel:'#btn-cancel'
}

var carDetail = {

}

$(function(){
    initData();
    initEvent();
})

function initData(){    
    commonObj.tranTypeMap.forEach(function(value, key){
        $(tag.tranType).append('<option value="'+key+'">'+value+'</option>');
    })
    commonObj.seatTypeMap.forEach(function(value, key){
        $(tag.seatType).append('<option value="'+key+'">'+value+'</option>');
    })
    setSeatSumInfo();
    
    var id = getQueryString('carId');    
    if (id == null) return;
    $.ajax({
        url:'/cars/getDetail',
        type:'GET',
        dataType:'json',
        data:{carId:id},
        success:function(result){
            if (result == null || result.car == null) return;
            carDetail = result.car;
            var c = result.car;
            $(tag.carId).val(c.id);
            $(tag.tranType).val(c.tranType);
            $(tag.seatType).val(c.seatType);
            $(tag.seatCount).html(c.SeatCount);
            $(tag.noSeatCount).val(c.noSeatCount);
            $(tag.remark).val(c.remark);
            for(var i=0; i<c.seats.length; i++){
                $(tag.seatContent).append($(tag.seatItemTpl).html());
                var item = $(tag.seats + ':eq(' + i + ')');
                var s = c.seats[i];
                item.find('input[name=seatNum]').val(s.seatNum);
                item.find('input[name=isStudent]').prop('checked', s.isStudent);                
            }
            setSeatSumInfo();
        }
    })
}

function initEvent(){
    $(tag.btnAddSeat).click(function(){
        $(tag.seats).remove();
        var rowCount = getParseInt($(tag.txtRowCount).val());
        var seatCountInRow = getParseInt($(tag.txtSeatCountInRow).val());
        var suffixArr = $(tag.txtSeatNumSuffix).val().split(',');
        for (var i=1, idx=0; i < rowCount+1;i++ ) {
            for(var si=0;si<seatCountInRow ;si++){
                $(tag.seatContent).append($(tag.seatItemTpl).html());
                var seatNum = '';
                if (i < 10){
                    seatNum = '0';
                } 
                seatNum += i + suffixArr[si];
                $(tag.seatContent + ' div.seat-item:eq('+idx+')').find('input[name="seatNum"]').val(seatNum);
                idx++;
            }
        }        
        var rowCount = getParseInt($(tag.txtRowSeatCount).val());
        for(var j=0;j<rowCount;j++){
            i++;
        }
        setSeatSumInfo();
    });
    $(tag.seatContent).on('click', 'i.fa-remove', function(){
        $(this).parent().remove();
        setSeatSumInfo();
    });
    $(tag.seatContent).on('change', 'input[name="isStudent"]', function(){
        setSeatSumInfo();
    })
    $(tag.btnCancel).click(function(){
        window.location.href = '/cars';
    });
    $(tag.btnSave).click(function(){
        save();    
    });
}

function setSeatSumInfo(){
    $(tag.seatCount).html($(tag.seats).length);
    var studentCount = 0;
    $(tag.seats).each(function(){
        if ($(this).find('input[name="isStudent"]').prop('checked')){
            studentCount++;
        }
    });
    $(tag.studentCount).html(studentCount);
}

function save(){
    carDetail = {
        id : getParseInt($(tag.carId).val()),
        tranType:$(tag.tranType).val(),
        seatType:$(tag.seatType).val(),
        noSeatCount: getParseInt($(tag.noSeatCount).val()),
        remark:$(tag.remark).val(),
        seats: new Array()
    };
    $(tag.seats).each(function(){
        var num = $(this).find('input[name="seatNum"]').val();
        if (num == '') return;
        var ck = $(this).find('input[name="isStudent"]').prop('checked');
        carDetail.seats.push({seatNum:num, isStudent:ck});
    });
    $.ajax({
        url:'/car/save',
        type:'POST',
        dataType:'json',
        data:JSON.stringify(carDetail),
        success:function(result){
            if (result.success){
                toastr.success('保存成功');
            } else {
                toastr.error(result.msg);
            }
        }
    })
}