$(document).ready(function () {
    $('table').DataTable({
        /*"language": {
            "url": "/static/js/Russian.json"
        },*/

        "pageLength":100,
        "lengthMenu":[[25,50,100,250,500],[25,50,100,250,500]],
        "deferRender":true,
        "stateSave":true,

        "language":{
            "sProcessing":   "Подождите...",
            "sLengthMenu":   "Показать _MENU_ записей",
            "sZeroRecords":  "Записи отсутствуют.",
            "sInfo":         "Записи с _START_ до _END_ из _TOTAL_ записей",
            "sInfoEmpty":    "Записи с 0 до 0 из 0 записей",
            "sInfoFiltered": "(отфильтровано из _MAX_ записей)",
            "sInfoPostFix":  "",
            "sSearch":       "Поиск:",
            "sUrl":          "",
            "oPaginate": {
                "sFirst": "Первая",
                "sPrevious": "Предыдущая",
                "sNext": "Следующая",
                "sLast": "Последняя"
            },
            "oAria": {
                "sSortAscending":  ": активировать для сортировки столбца по возрастанию",
                "sSortDescending": ": активировать для сортировки столбцов по убыванию"
            }
        },
        dom: "<'row'<'col-sm-12 col-md-6' B><'col-sm-12 col-md-6'f>>" +
        "<'row'<'col-sm-12'tr>>" +
        "<'row'<'col-sm-12 col-md-5'i><'col-sm-12 col-md-7'p>>",
        buttons: [

            'excelHtml5',
            'csvHtml5',
            'pdfHtml5'
        ]

    });

    /* $('table').DataTable( {
         dom: 'B<"clear">lfrtip',
         buttons: [ 'copy', 'csv', 'excel' ]
    } );*/


});