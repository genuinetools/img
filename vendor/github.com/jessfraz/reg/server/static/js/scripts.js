// pretty date function
function prettyDate(time){
    var date = new Date((time || "").replace(/-/g,"/").replace(/[TZ]/g," ")),
    diff = (((new Date()).getTime() - date.getTime()) / 1000),
    day_diff = Math.floor(diff / 86400);

    if (isNaN(day_diff) || day_diff < 0)
        return;

    return day_diff == 0 && (
            diff < 60 && "just now" ||
            diff < 120 && "1 minute ago" ||
            diff < 3600 && Math.floor( diff / 60 ) + " minutes ago" ||
            diff < 7200 && "1 hour ago" ||
            diff < 86400 && Math.floor( diff / 3600 ) + " hours ago") ||
        day_diff == 1 && "Yesterday" ||
        day_diff < 7 && day_diff + " days ago" ||
        day_diff < 31 && Math.ceil( day_diff / 7 ) + " weeks ago" ||
        day_diff > 31 && Math.round(day_diff / 31) + " months ago";
}

// search function
function search(search_val){
    var suche = search_val.toLowerCase();
    var table = document.getElementById("directory");
    var cellNr = 1;
    var ele;
    for (var r = 1; r < table.rows.length; r++){
        ele = table.rows[r].cells[cellNr].innerHTML.replace(/<[^>]+>/g,"");
        if (ele.toLowerCase().indexOf(suche)>=0 ) {
            table.rows[r].style.display = '';
        } else {
            table.rows[r].style.display = 'none';
        }
    }
}

function loadVulnerabilityCount(url){
  var xhr = new XMLHttpRequest();
  xhr.open('GET', url);
  xhr.onload = function() {
      if (xhr.status === 200) {
          var report = JSON.parse(xhr.responseText);
          var id = report.Repo + ':' + report.Tag;
          var element = document.getElementById(id);

          if (element) {
            element.innerHTML = report.BadVulns;
          } else {
            console.log("element not found for given id ", id);
          }
      }
  };
  xhr.send();
}

var el = document.querySelectorAll('tr:nth-child(2)')[0].querySelectorAll('td:nth-child(2)')[0];
if (el.textContent == 'Parent Directory'){
    var parent_row = document.querySelectorAll('tr:nth-child(2)')[0];
    if (parent_row.classList){
        parent_row.classList.add('parent');
    } else {
        parent_row.className += ' ' + 'parent';
    }
}

var cells = document.querySelectorAll('td a');
Array.prototype.forEach.call(cells, function(item, index){
    var link = item.getAttribute('href');
    link = link.replace('.html', '');
    item.setAttribute('href', link);
});

var our_table = document.querySelectorAll('table')[0];
our_table.setAttribute('id', 'directory');

// search script
var search_input = document.querySelectorAll('input[name="filter"]')[0];
var clear_button = document.querySelectorAll('a.clear')[0];

if (search_input) {
  if (search_input.value !== ''){
      search(search_input.value);
  }

  search_input.addEventListener('keyup', function(e){
      e.preventDefault();
      search(search_input.value);
  });

  search_input.addEventListener('keypress', function(e){
      if ( e.which == 13 ) {
          e.preventDefault();
      }
  });
}

if (clear_button) {
  clear_button.addEventListener('click', function(e){
      search_input.value = '';
      search('');
  });
}
