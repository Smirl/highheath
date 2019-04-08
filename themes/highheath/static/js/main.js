function cloneAndIncremetCat() {

    var newElement = $('div.tab-pane:last').clone(true);

    // Increment in index of the divs and inputs 
    newElement = changeCatIndex(newElement, getTotal())

    // Remove the content of the inputs
    newElement.find(':input').each(function() {
        $(this).val('').removeAttr('checked');
    });
    
    // Add to the total forms
    $('#id_cats-TOTAL_FORMS').val(getTotal()+1);
    return newElement;
}

function getTotal() {
    return parseInt($('#id_cats-TOTAL_FORMS').val());
}

function changeCatIndex(element, index) {

    element.attr('id', 'cat-'+ (parseInt(index)+1));
    element.find('.form-group').each(function() {
        var id = $(this).attr('id').replace(/-\d-/g,'-' + index + '-');
        $(this).attr('id', id);
    });
    element.find(':input').each(function() {
        var name = $(this).attr('name').replace(/-\d-/g,'-' + index + '-');
        var id = 'id_' + name;
        $(this).attr('name', name).attr('id', id);
    });
    element.find('label').each(function() {
        var newFor = $(this).attr('for').replace(/-\d-/g,'-' + index + '-');
        $(this).attr('for', newFor);
    });
    element.find('script').remove();
    return element;
}

var pageNum =  1;

function last() {
    $('li.tab-pane:last a').tab('show');
}

$(document).ready(function(){
    /**
     * Click Tab to show its contents
     */
    $(".nav.nav-tabs").on("click", "a", function(e) {
        e.preventDefault();
        $(this).tab('show');
    });

    /**
    * Add a Tab
    */
    $('a.add-cat').on('click', function(e) {
        e.preventDefault();
        pageNum++;
        $('.nav.nav-tabs li:last').after($('<li class="tab-pane"><a href="#cat-' + pageNum + '">' +'Cat ' + pageNum +'<button class="close" title="Remove this page" type="button">×</button>' +'</a></li>'));
     
        $('.tab-content').append(cloneAndIncremetCat());

        $('li.tab-pane:last a').tab('show');

        reset_pickers();
    });
     
    /**
    * Remove a Tab
    */
    $('.nav.nav-tabs').on('click', ' li a .close', function() {
        var tabId = $(this).parents('li').children('a').attr('href');
        $(this).parents('li').remove('li');
        $(tabId).remove();
        reNumberPages();
        $('#id_cats-TOTAL_FORMS').val(getTotal()-1);
        $('li.tab-pane:last a').tab('show');
        
        reset_pickers();
    });
     
 });
/**
* Reset numbering on tab buttons
*/
function reNumberPages() {
    pageNum = 1;
    var tabCount = $('.nav.nav-tabs > li').length;
    $('.nav.nav-tabs > li.tab-pane').each(function() {
        var pageId = $(this).children('a').attr('href');
        if (pageId == "#cat-1") {
            return true;
        }
        pageNum++;
        changeCatIndex($('div.tab-pane:nth-child('+pageNum+')'), pageNum-1);
        $(this).children('a').attr('href', '#cat-'+pageNum);
        $(this).children('a').html('Cat '+ pageNum +'<button class="close" title="Remove this page" type="button">×</button>');
    });
}

function reset_pickers() {
    //reset the vaccination datepickers
    $('[id*="vaccination_date_picker"]').each(function(){
        console.log('Remove ' + $(this).attr('id'));
        $(this).data().DateTimePicker.destroy();
    });
    $('[id*="vaccination_date_picker"] input').each(function(){
        console.log('Remove ' + $(this).attr('id'));
        $(this).data().DateTimePicker.destroy();
    });
    $('[id*="vaccination_date_picker"]').each(function(){
        console.log('Add '+ $(this).attr('id'));
        $(this).datetimepicker({'icons': {'date': 'fa fa-calendar'},'format': 'DD/MM/YYYY hh:mma','pickTime': false,});
    });
    $('[id*="vaccination_date_picker"] input').each(function(){
        console.log('Add ' + $(this).attr('id'));
        $(this).datetimepicker({'icons': {'date': 'fa fa-calendar'},'format': 'DD/MM/YYYY hh:mma','pickTime': false,});
    });
}

function clear_comment_form() {
    $('#id_name').val('');
    $('#id_email').val('');
    $('#id_comment').val('');
}