var tag = {
    scheduleId : '',
    depDate : '#dep-date',
    tranNum:'#tran-num',
    saleTicketTime:'#sale-ticket-time',
    notSaleRemark:'#not-sale-remark',
    btnSave:'#btn-save'
}

$(function(){
    initDOM();
    initEvent();
});

function initDOM(){
    tag.scheduleId = parseInt(getQueryString('scheduleID'));
    $.ajax({
        url:'/schedules/getDetail',
        type:'GET',
        dataType:'json',
        data:{scheduleID: tag.scheduleId},
        success:function(result){
            var s = result.schedule;
            $(tag.depDate).val(s.departureDate);
            $(tag.tranNum).val(s.tranNum);
            var date = new Date(s.saleTicketTime).format('yyyy-MM-ddThh:mm:ss');
            console.log(date);
            $(tag.saleTicketTime).val(date);
            $(tag.notSaleRemark).val(s.notSaleRemark);
        }
    })
}

function initEvent(){
    $(tag.btnSave).click(function(){
        var param = {
            id: tag.scheduleId,
            departureDate: new Date($(tag.depDate).val()).format(),
            tranNum: $(tag.tranNum).val(),
            saleTicketTime: new Date($(tag.saleTicketTime).val()).toISOString(),
            notSaleRemark: $(tag.notSaleRemark).val()
        };
        $.ajax({
            url:'/schedules/save',
            type:'POST',
            dataType:'json',
            data: JSON.stringify(param),
            success:function(result){
                if (result.success){
                    toastr.success('保存成功');
                } else {
                    toastr.error(result.msg);
                }
            }
        })
    })
}