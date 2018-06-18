// $(function(){
//     // 设置面包屑
//     var uri = window.location.pathname;
//     var path = uri.split('/')
//     for (var i=0; i<path.length; i++) {        
//         if (path[i] == ''){
//             continue;
//         }
//         var name = path[i].substring(0,1).toUpperCase() + path[i].substring(1);
//         if (i < path.length - 1) {
//             $('ul.breadcrumb').append('<li class="breadcrumb-item"><a href="/' + path[i] + '">'+ name +'</a></li>')
//         } else {
//             $('ul.breadcrumb').append('<li class="breadcrumb-item active">'+ name +'</li>')
//         }
//     }
// })

var commonObj = {
    tranTypeMap : new Map([
        ['G', '高铁'],
        ['D', '动车'],
        ['C', '城际'],
        ['Z', '直达'],
        ['T', '特快'],
        ['K', '普快'],
        ['O', '其他'],
    ]),
    seatTypeMap : new Map([
        ['S', '商务座'],
        ['FC', '一等座'],
        ['SC', '二等座'],
        ['ASS', '高级软卧'],
        ['SS', '软卧'],
        ['DS', '动车卧铺'],
        ['MS', '动卧(床改座)'],
        ['HS', '硬卧'],
        ['SST', '软座'],
        ['HST', '硬座']
    ])
}

function getStrDate(strDate) {    
    try{
        return new Date(strDate).toLocaleString();
    }
    catch(err){
        return '';
    }
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

function getQueryString(name){
     var reg = new RegExp("(^|&)"+ name +"=([^&]*)(&|$)");
     var r = window.location.search.substr(1).match(reg);
     if(r!=null) return unescape(r[2]);
     return null;
}

function getParseInt(str, def){
    var num = parseInt(str);
    if (isNaN(num)){
        if (def == undefined){
            return 0;
        }
        return def;
    }
    return num;
}

function getParseFloat(str, def){
    var num = parseFloat(str);
    if (isNaN(num)){
        if (def == undefined){
            return 0;
        }
        return def;
    }
    return num;
}