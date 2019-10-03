// Add field when Issue is Billing Terminated

$(document).on('change','#issue-input',function(){
    var selection = $(this).val();
    switch(selection){
    case "1":
        $("#IIA").show()
        break;
    default:
        $("#IIA").hide()
        break;
    }
});